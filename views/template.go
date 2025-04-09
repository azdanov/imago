package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/azdanov/imago/context"
	"github.com/azdanov/imago/models"
	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTmpl *template.Template
}

const baseTemplate = "layouts/base.tmpl.html"

var baseTemplates = []string{
	baseTemplate,
	"layouts/notifications.tmpl.html",
}

func Parse(fs fs.FS, pattern ...string) (*Template, error) {
	if len(pattern) == 0 {
		return nil, fmt.Errorf("no template files provided")
	}
	pattern = append(pattern, baseTemplates...)

	t, err := template.New(baseTemplate).Funcs(funcMap).ParseFS(fs, pattern...)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return &Template{htmlTmpl: t}, nil
}

func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any) {
	tpl, err := t.htmlTmpl.Clone()
	if err != nil {
		log.Printf("clone template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
		return
	}

	tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"currentUser": func() *models.User {
			return context.User(r.Context())
		},
		"notifications": func() []models.Notification {
			notifications := context.Notifications(r.Context())
			if notifications == nil {
				return nil
			}
			return models.SortNotifications(notifications)
		},
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var buf bytes.Buffer
	err = tpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

var funcMap = template.FuncMap{
	"currentYear": func() int {
		return time.Now().Year()
	},
	"csrfField": func() (template.HTML, error) {
		return template.HTML(""), fmt.Errorf("csrfField: called in template without a request")
	},
	"currentUser": func() (*models.User, error) {
		return nil, fmt.Errorf("currentUser: called in template without a request")
	},
	"notifications": func() ([]models.Notification, error) {
		return nil, fmt.Errorf("notifications: called in template without a request")
	},
}
