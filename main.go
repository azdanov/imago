package main

import (
	"log"
	"net/http"

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

func main() {
	// Load environment variables
	config := config.NewEnvConfig()

	// Setup database
	db, err := database.NewDB(config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	err = database.Migrate(db, database.FS, database.MigrationsDir)
	if err != nil {
		log.Fatalf("Unable to migrate database: %v", err)
	}

	// Setup services
	ss := models.NewSessionService(db, models.MinSessionTokenBytes)
	us := models.NewUserService(db)
	sc := controllers.NewSessionCookie(config.Server.SSLMode)
	ps := models.NewPasswordResetService(db, models.MinSessionTokenBytes, models.DefaultTokenLifetime)
	es, err := models.NewEmailService(config)
	if err != nil {
		log.Fatalf("Unable to create email service: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(config.CSRF.Key), csrf.Secure(config.CSRF.Secure)))

	um := controllers.NewUserMiddleware(ss, sc)
	r.Use(um.SetUser)

	notificationMiddleware := controllers.NewNotificationMiddleware()
	r.Use(notificationMiddleware.ExtractNotifications)

	// Setup routes
	tmpl := views.Must(views.Parse(templates.FS, "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	usersC := controllers.NewUsers(us, ss, sc, ps, es, &config)

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

	// Start server
	log.Printf("Starting server on %s\n", config.Server.GetURL())
	http.ListenAndServe(config.Server.GetAddr(), r)
}
