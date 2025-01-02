package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/azdanov/imago/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	email     = "contact@example.com"
	templates map[string]views.Template
)

func main() {
	var err error
	templates, err = views.LoadTemplates()
	if err != nil {
		log.Fatalf("load templates: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templates["home"].Execute(w, nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	templates["contact"].Execute(w, struct{ Email string }{email})
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	templates["faq"].Execute(w, struct{ Email string }{email})
}
