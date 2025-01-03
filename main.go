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

	tpl := views.Must(views.Parse(templates.FS, "home.tmpl.html"))
	r.Get("/", controllers.StaticHandler(tpl))

	tpl = views.Must(views.Parse(templates.FS, "contact.tmpl.html"))
	tpl.Data = struct{ Email string }{Email: email}
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl = views.Must(views.Parse(templates.FS, "faq.tmpl.html"))
	tpl.Data = struct{ Email string }{Email: email}
	r.Get("/faq", controllers.StaticHandler(tpl))

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}
