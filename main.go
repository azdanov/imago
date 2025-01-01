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

var email = "contact@example.com"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tPath := filepath.Join("templates", "home.html")
	t, err := template.ParseFiles(tPath)
	if err != nil {
		log.Printf("parse template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
		return
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Contact Page</h1>\n<p>To get in touch, please send an email to <a href=\"mailto:%s\">%s</a>.</p>", email, email)
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<h1>FAQ</h1>
	<dl>
		<dt>What is this service?</dt>
		<dd>This is a sample FAQ page.</dd>
		<dt>How can I contact support?</dt>
		<dd>You can contact support by sending an email to <a href="mailto:%s">%s</a>.</dd>
	</dl>`, email, email)
}
