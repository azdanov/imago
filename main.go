package main

import (
	"fmt"
	"net/http"

	"github.com/azdanov/imago/controllers"
	"github.com/azdanov/imago/templates"
	"github.com/azdanov/imago/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var email = "contact@example.com"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tmpl := views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	usersC := &controllers.Users{}
	usersC.Templates.New = views.Must(views.Parse(templates.FS, "layouts/base.tmpl.html", "signup.tmpl.html"))
	r.Get("/signup", usersC.NewSignup)
	r.Post("/signup", usersC.HandleSignup)

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}
