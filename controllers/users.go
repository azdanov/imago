package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/azdanov/imago/models"
)

type Users struct {
	Templates struct {
		SignUp Template
		SignIn Template
	}
	UserService *models.UserService
}

func (u Users) NewSignup(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	errorMessage := r.URL.Query().Get("error")
	data := struct {
		Email string
		Error string
	}{
		Email: email,
		Error: errorMessage,
	}
	u.Templates.SignUp.Execute(w, r, data)
}

func (u Users) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signup?error=%s", url.QueryEscape("Something went wrong")), http.StatusSeeOther)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	if email == "" {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}
	if password == "" {
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s&error=%s", email, "Password is required"), http.StatusSeeOther)
		return
	}
	if len(password) < 8 {
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s&error=%s", email, "Password must be at least 8 characters long"), http.StatusSeeOther)
		return
	}

	_, err := u.UserService.Create(email, password)
	if err != nil {
		log.Printf("create user: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s&error=%s", email, "Error creating user"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (u Users) NewSignin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	errorMessage := r.URL.Query().Get("error")
	data := struct {
		Email string
		Error string
	}{
		Email: email,
		Error: errorMessage,
	}
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) HandleSignin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signin?error=%s", "Something went wrong"), http.StatusSeeOther)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	if email == "" {
		http.Redirect(w, r, "/signin?error=Email is required", http.StatusSeeOther)
		return
	}
	if password == "" {
		http.Redirect(w, r, fmt.Sprintf("/signin?email=%s&error=%s", email, "Password is required"), http.StatusSeeOther)
		return
	}

	_, err := u.UserService.Authenticate(email, password)
	if err != nil {
		log.Printf("authenticate user: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signin?email=%s&error=%s", email, "Invalid email or password"), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
