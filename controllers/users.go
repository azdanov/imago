package controllers

import (
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
	data := struct {
		Email string
	}{
		Email: r.URL.Query().Get("email"),
	}

	u.Templates.SignUp.Execute(w, r, data)
}

func (u Users) HandleSignup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		vals := url.Values{
			"error": {"Something went wrong"},
		}
		http.Redirect(w, r, "/signup?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	vals := url.Values{
		"email": {email},
	}

	if email == "" {
		vals.Set("error", "Email is required")
		http.Redirect(w, r, "/signup?"+vals.Encode(), http.StatusSeeOther)
		return
	}
	if password == "" {
		vals.Set("error", "Password is required")
		http.Redirect(w, r, "/signup?"+vals.Encode(), http.StatusSeeOther)
		return
	}
	if len(password) < 8 {
		vals.Set("error", "Password must be at least 8 characters long")
		http.Redirect(w, r, "/signup?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	user, err := u.UserService.Create(email, password)
	if err != nil {
		log.Printf("create user: %v", err)
		if err == models.ErrEmailAlreadyExists {
			vals.Set("error", "Email already exists")
		} else {
			vals.Set("error", "Error creating user")
		}
		http.Redirect(w, r, "/signup?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("create session: %v", err)
		vals := url.Values{
			"email": {email},
			"error": {"Error creating session"},
		}
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	u.SessionCookie.Set(w, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusSeeOther)
}

func (u Users) NewSignin(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.URL.Query().Get("email"),
	}
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) HandleSignin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		vals := url.Values{
			"error": {"Something went wrong"},
		}
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	vals := url.Values{
		"email": {email},
	}

	if email == "" {
		vals.Set("error", "Email is required")
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}
	if password == "" {
		vals.Set("error", "Password is required")
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	user, err := u.UserService.Authenticate(email, password)
	if err != nil {
		log.Printf("authenticate user: %v", err)
		vals.Set("error", "Invalid email or password")
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		log.Printf("create session: %v", err)
		vals.Set("error", "Error creating session")
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
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
		vals := url.Values{
			"error": {"Error signing out"},
		}
		http.Redirect(w, r, "/signin?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	u.SessionCookie.Clear(w)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (u Users) NewForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string
	}{
		Email: r.FormValue("email"),
	}

	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")

	vals := url.Values{
		"email": {email},
	}

	passwordReset, err := u.PasswordResetService.Generate(email)
	if err != nil {
		log.Printf("generate password reset: %v", err)
		vals.Set("error", "Something went wrong")
		http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	resetVals := url.Values{
		"token": {passwordReset.Token},
	}
	resetURL := u.serverURL + "?" + resetVals.Encode()

	err = u.EmailService.SendResetPassword(email, resetURL)
	if err != nil {
		log.Printf("send email: %v", err)
		vals.Set("error", "Something went wrong")
		http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	vals.Set("success", "An email has been sent with instructions to reset your password")
	http.Redirect(w, r, "/forgot-password?"+vals.Encode(), http.StatusSeeOther)
}

func (u Users) NewResetPassword(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Token string
	}{
		Token: r.FormValue("token"),
	}

	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("parse form: %v", err)
		vals := url.Values{
			"error": {"Something went wrong"},
		}
		http.Redirect(w, r, "/reset-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}
	token := r.PostForm.Get("token")
	password := r.PostForm.Get("password")

	vals := url.Values{
		"token": {token},
	}

	user, err := u.PasswordResetService.GetUserByToken(token)
	if err != nil {
		log.Printf("get user by token: %v", err)
		vals.Set("error", "Invalid or expired token")
		http.Redirect(w, r, "/reset-password?"+vals.Encode(), http.StatusSeeOther)
		return
	}

	if err = u.UserService.UpdatePassword(user.ID, password); err != nil {
		log.Printf("update password: %v", err)
		vals.Set("error", "Internal server error")
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
