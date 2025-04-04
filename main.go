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

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(config.CSRF.Key), csrf.Secure(config.CSRF.Secure)))

	um := controllers.NewUserMiddleware(ss, sc)
	r.Use(um.SetUser)

	// Setup routes
	tmpl := views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	usersC := controllers.NewUsers(us, ss, sc)

	usersC.Templates.SignUp = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "signup.tmpl.html"))
	r.Get("/signup", usersC.NewSignup)
	r.Post("/signup", usersC.HandleSignup)

	usersC.Templates.SignIn = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "signin.tmpl.html"))
	r.Get("/signin", usersC.NewSignin)
	r.Post("/signin", usersC.HandleSignin)
	r.Post("/signout", usersC.HandleSignout)

	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "me.tmpl.html"))
	r.Route("/users/me", func(r chi.Router) {
		r.Use(um.RequireUser)
		r.Get("/", controllers.StaticHandler(tmpl))
	})

	// Start server
	addr := config.Server.GetAddr()
	log.Printf("Starting server on %s\n", addr)
	http.ListenAndServe(addr, r)
}
