package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	email     = "contact@example.com"
	templates = map[string]*template.Template{}
)

func main() {
	loadTemplates()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}

func loadTemplates() {
	templates["home"] = parseTemplate("home.tmpl.html")
	templates["contact"] = parseTemplate("contact.tmpl.html")
	templates["faq"] = parseTemplate("faq.tmpl.html")
}

func parseTemplate(filename string) *template.Template {
	tPath := filepath.Join("templates", filename)
	t, err := template.ParseFiles(tPath)
	if err != nil {
		log.Fatalf("parse template %s: %v", filename, err)
	}
	return t
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, ok := templates[tmpl]
	if !ok {
		http.Error(w, "The requested template does not exist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.Execute(w, data)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home", nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "contact", struct{ Email string }{email})
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "faq", struct{ Email string }{email})
}
