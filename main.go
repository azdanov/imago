package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/azdanov/imago/controllers"
	"github.com/azdanov/imago/models"
	"github.com/azdanov/imago/templates"
	"github.com/azdanov/imago/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var email = "contact@example.com"

func main() {
	csrfSecret := os.Getenv("CSRF_SECRET")
	isProduction := os.Getenv("ENV") == "production"
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(csrfSecret), csrf.Secure(isProduction)))

	tmpl := views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))
	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))
	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	usersC := &controllers.Users{
		UserService: &models.UserService{DB: db},
	}
	usersC.Templates.SignUp = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "signup.tmpl.html"))
	r.Get("/signup", usersC.NewSignup)
	r.Post("/signup", usersC.HandleSignup)
	usersC.Templates.SignIn = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "signin.tmpl.html"))
	r.Get("/signin", usersC.NewSignin)
	r.Post("/signin", usersC.HandleSignin)

	fmt.Println("Server is running on http://localhost:3000")
	http.ListenAndServe(":3000", r)
}
