package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"
)

type Template struct {
	htmlTmpl *template.Template
}

func Parse(fs fs.FS, pattern ...string) (*Template, error) {
	pattern = append([]string{"layouts/base.tmpl.html"}, pattern...)
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

func (t Template) Execute(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Println("t.htmlTmpl: ", t.htmlTmpl.Name())
	err := t.htmlTmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
	}
}

var funcMap = template.FuncMap{
	"currentYear": func() int {
		return time.Now().Year()
	},
}
