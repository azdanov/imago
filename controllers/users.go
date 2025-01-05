package controllers

import (
	"fmt"
	"net/http"
)

type Users struct {
	Templates struct {
		New Template
	}
}

func (u Users) NewSignup(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	data := struct {
		Email string
	}{
		Email: email,
	}
	u.Templates.New.Execute(w, data)
}

func (u Users) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		email := r.PostForm.Get("email")
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s", email), http.StatusSeeOther)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	if email == "" {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}
	if password == "" {
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s", email), http.StatusSeeOther)
		return
	}
	if len(password) < 8 {
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s", email), http.StatusSeeOther)
		return
	}

	fmt.Printf("Email: %s, Password: %s\n", email, password)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
