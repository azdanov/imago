package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/azdanov/imago/context"
	"github.com/azdanov/imago/models"
)

type Users struct {
	Templates struct {
		SignUp Template
		SignIn Template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
	SessionCookie  *SessionCookie
}

func NewUsers(us *models.UserService, ss *models.SessionService, sc *SessionCookie) *Users {
	return &Users{
		UserService:    us,
		SessionService: ss,
		SessionCookie:  sc,
	}
}

type UserMiddleware struct {
	SessionService *models.SessionService
	SessionCookie  *SessionCookie
}

func NewUserMiddleware(ss *models.SessionService, sc *SessionCookie) *UserMiddleware {
	return &UserMiddleware{
		SessionService: ss,
		SessionCookie:  sc,
	}
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

	user, err := u.UserService.Create(email, password)
	if err != nil {
		log.Printf("create user: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signup?email=%s&error=%s", email, "Error creating user"), http.StatusSeeOther)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("create session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signin?email=%s&error=%s", email, "Error creating session"), http.StatusSeeOther)
		return
	}

	u.SessionCookie.Set(w, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusSeeOther)
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

	user, err := u.UserService.Authenticate(email, password)
	if err != nil {
		log.Printf("authenticate user: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signin?email=%s&error=%s", email, "Invalid email or password"), http.StatusSeeOther)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("create session: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/signin?email=%s&error=%s", email, "Error creating session"), http.StatusSeeOther)
		return
	}

	u.SessionCookie.Set(w, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusSeeOther)
}

func (u Users) HandleSignout(w http.ResponseWriter, r *http.Request) {
	token, err := u.SessionCookie.Get(r)
	if err != nil {
		log.Printf("get token: %v", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if err := u.SessionService.Delete(token); err != nil {
		log.Printf("delete session: %v", err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	u.SessionCookie.Clear(w)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (m UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.SessionCookie.Get(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
