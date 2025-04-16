package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/azdanov/imago/config"
	"github.com/azdanov/imago/controllers"
	"github.com/azdanov/imago/database"
	"github.com/azdanov/imago/models"
	"github.com/azdanov/imago/templates"
	"github.com/azdanov/imago/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 60 * time.Second
)

func main() {
	// Load environment variables
	cnf := config.NewEnvConfig()

	// Setup database
	db, err := setupDatabase(cnf)
	if err != nil {
		log.Fatalf("Unable to setup database: %v", err)
	}

	// Setup services
	services := setupServices(db, cnf)

	// Setup router and routes
	r := setupRouter(cnf, services)

	// Start server
	log.Printf("Starting server on %s\n", cnf.Server.GetURL())
	srv := &http.Server{
		Handler:      r,
		Addr:         cnf.Server.GetAddr(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	if err = srv.ListenAndServe(); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}

func setupDatabase(cnf *config.Config) (*sql.DB, error) {
	db, err := database.NewDB(cnf)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	err = database.Migrate(db, database.FS, database.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("unable to migrate database: %w", err)
	}

	return db, nil
}

type services struct {
	sessionService       *models.SessionService
	userService          *models.UserService
	sessionCookie        *controllers.SessionCookie
	passwordResetService *models.PasswordResetService
	emailService         *models.EmailService
	galleryService       *models.GalleryService
}

func setupServices(db *sql.DB, cnf *config.Config) *services {
	ss := models.NewSessionService(db, models.MinSessionTokenBytes)
	us := models.NewUserService(db)
	sc := controllers.NewSessionCookie(cnf.Server.SSLMode)
	ps := models.NewPasswordResetService(db, models.MinSessionTokenBytes, models.DefaultTokenLifetime)
	es, err := models.NewEmailService(cnf)
	if err != nil {
		log.Fatalf("Unable to create email service: %v", err)
	}
	gs := models.NewGalleryService(db)

	return &services{
		sessionService:       ss,
		userService:          us,
		sessionCookie:        sc,
		passwordResetService: ps,
		emailService:         es,
		galleryService:       gs,
	}
}

func setupRouter(cnf *config.Config, s *services) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(cnf.CSRF.Key),
		csrf.Secure(cnf.CSRF.Secure),
		csrf.TrustedOrigins([]string{cnf.Server.GetAddr()}),
	))

	um := controllers.NewUserMiddleware(s.sessionService, s.sessionCookie)
	r.Use(um.SetUser)

	notificationMiddleware := controllers.NewNotificationMiddleware()
	r.Use(notificationMiddleware.ExtractNotifications)

	setupRoutes(r, s, um, cnf)

	return r
}

func setupRoutes(r *chi.Mux, s *services, um *controllers.UserMiddleware, cnf *config.Config) {
	// Static routes
	tmpl := views.Must(views.Parse(templates.FS, "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	// User routes
	usersC := controllers.NewUsers(
		s.userService, s.sessionService, s.sessionCookie, s.passwordResetService, s.emailService, cnf)

	usersC.Templates.SignUp = views.Must(views.Parse(templates.FS, "signup.tmpl.html"))
	r.Get("/signup", usersC.NewSignup)
	r.Post("/signup", usersC.HandleSignup)

	usersC.Templates.SignIn = views.Must(views.Parse(templates.FS, "signin.tmpl.html"))
	r.Get("/signin", usersC.NewSignin)
	r.Post("/signin", usersC.HandleSignin)
	r.Post("/signout", usersC.HandleSignout)

	usersC.Templates.ForgotPassword = views.Must(views.Parse(templates.FS, "forgot_password.tmpl.html"))
	r.Get("/forgot-password", usersC.NewForgotPassword)
	r.Post("/forgot-password", usersC.HandleForgotPassword)

	usersC.Templates.ResetPassword = views.Must(views.Parse(templates.FS, "reset_password.tmpl.html"))
	r.Get("/reset-password", usersC.NewResetPassword)
	r.Post("/reset-password", usersC.HandleResetPassword)

	tmpl = views.Must(views.Parse(templates.FS, "me.tmpl.html"))
	r.Route("/users/me", func(r chi.Router) {
		r.Use(um.RequireUser)
		r.Get("/", controllers.StaticHandler(tmpl))
	})

	// Gallery routes
	galleriesC := controllers.NewGalleries(s.galleryService)
	galleriesC.Templates.New = views.Must(views.Parse(templates.FS, "galleries/new.tmpl.html"))
	galleriesC.Templates.Edit = views.Must(views.Parse(templates.FS, "galleries/edit.tmpl.html"))
	galleriesC.Templates.Show = views.Must(views.Parse(templates.FS, "galleries/show.tmpl.html"))
	galleriesC.Templates.List = views.Must(views.Parse(templates.FS, "galleries/list.tmpl.html"))
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(um.RequireUser)
			r.Get("/", galleriesC.List)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images", galleriesC.UploadImage)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
	})

	// 404 handler
	tmpl = views.Must(views.Parse(templates.FS, "404.tmpl.html"))
	r.NotFound(controllers.StaticHandler(tmpl))
}
