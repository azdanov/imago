package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/azdanov/imago/controllers"
	"github.com/azdanov/imago/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var email = "contact@example.com"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	tpl, err := views.Parse("home")
	if err != nil {
		log.Fatalf("load template: %v", err)
	}
	r.Get("/", controllers.StaticHandler(tpl))

	tpl, err = views.Parse("contact")
	if err != nil {
		log.Fatalf("load template: %v", err)
	}
	tpl.Data = struct{ Email string }{Email: email}
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl, err = views.Parse("faq")
	if err != nil {
		log.Fatalf("load template: %v", err)
	}
	tpl.Data = struct{ Email string }{Email: email}
	r.Get("/faq", controllers.StaticHandler(tpl))

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}
