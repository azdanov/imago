package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var email = "contact@example.com"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>Home Page</h1>\n<p>Welcome to the home page!</p>")
	})

	r.Get("/contact", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>Contact Page</h1>\n<p>To get in touch, please send an email to <a href=\"mailto:%s\">%s</a>.</p>", email, email)
	})

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", r)
}
