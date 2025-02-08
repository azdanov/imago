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

	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTmpl *template.Template
}

func Parse(fs fs.FS, pattern ...string) (*Template, error) {
	t, err := template.New(pattern[0]).Funcs(funcMap).ParseFS(fs, pattern...)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
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
		return template.HTML(""), fmt.Errorf("csrfField called in template without a request")
	},
}
