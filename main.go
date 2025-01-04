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
	r.Use(middleware.Logger)

	tmpl := views.Must(views.Parse(templates.FS, "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "contact.tmpl.html"))
	r.Get("/contact", controllers.StaticHandler(tmpl))

	tmpl = views.Must(views.Parse(templates.FS, "faq.tmpl.html"))
	r.Get("/faq", controllers.FAQ(tmpl))

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}
