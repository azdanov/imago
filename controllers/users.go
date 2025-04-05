package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/azdanov/imago/config"
	"github.com/azdanov/imago/context"
	"github.com/azdanov/imago/models"
)

type Users struct {
	Templates struct {
		SignUp         Template
		SignIn         Template
		ForgotPassword Template
		ResetPassword  Template
	}

	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService

	SessionCookie *SessionCookie
	serverURL     string
}

func NewUsers(
	us *models.UserService,
	ss *models.SessionService,
	sc *SessionCookie,
	ps *models.PasswordResetService,
	es *models.EmailService,
	cn *config.Config,
) *Users {
	return &Users{
		UserService:          us,
		SessionService:       ss,
		SessionCookie:        sc,
		PasswordResetService: ps,
		EmailService:         es,
		serverURL:            cn.Server.GetURL(),
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

func (u Users) NewForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email   string
		Error   string
		Success string
	}{
		Email:   r.FormValue("email"),
		Error:   r.URL.Query().Get("error"),
		Success: r.URL.Query().Get("success"),
	}

	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")

	var vals url.Values

	passwordReset, err := u.PasswordResetService.Generate(data.Email)
	if err != nil {
		log.Printf("generate password reset: %v", err)
		vals = url.Values{
			"email": {data.Email},
			"error": {"Something went wrong"},
		}
		http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	vals = url.Values{
		"token": {passwordReset.Token},
	}
	resetURL := u.serverURL + "?" + vals.Encode()

	err = u.EmailService.SendResetPassword(data.Email, resetURL)
	if err != nil {
		log.Printf("send email: %v", err)
		vals = url.Values{
			"email": {data.Email},
			"error": {"Something went wrong"},
		}
		http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	vals = url.Values{
		"email":   {data.Email},
		"success": {"true"},
	}

	http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
}

func (u Users) NewResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
		Error string
	}
	data.Token = r.FormValue("token")
	data.Error = r.URL.Query().Get("error")

	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.GetUserByToken(data.Token)
	if err != nil {
		log.Printf("get user by token: %v", err)
		vals := url.Values{
			"error": {"Invalid or expired token"},
			"token": {data.Token},
		}
		http.Redirect(w, r, "/reset-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	if err = u.UserService.UpdatePassword(user.ID, data.Password); err != nil {
		log.Printf("update password: %v", err)
		vals := url.Values{
			"error": {"Internal server error"},
			"token": {data.Token},
		}
		http.Redirect(w, r, "/reset-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("create session: %v", err)
		vals := url.Values{
			"error": {"Error creating session. Please try again"},
			"email": {user.Email},
		}
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	u.SessionCookie.Set(w, session.Token)

	http.Redirect(w, r, "/users/me", http.StatusSeeOther)
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
