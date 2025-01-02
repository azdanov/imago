package views

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Template struct {
	htmlTpl *template.Template
}

func LoadTemplates() (map[string]Template, error) {
	templateNames := []string{"home", "contact", "faq"}

	templates := make(map[string]Template)
	for _, n := range templateNames {
		t, err := parse(n + ".tmpl.html")
		if err != nil {
			return nil, err
		}
		templates[n] = Template{htmlTpl: t}
	}
	return templates, nil
}

func parse(filename string) (*template.Template, error) {
	p := filepath.Join("templates", filename)
	t, err := template.ParseFiles(p)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", filename, err)
	}
	return t, nil
}

func (t Template) Execute(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.htmlTpl.Execute(w, data)
	if err != nil {
		log.Printf("execute template: %v", err)
		http.Error(w, "There was an error processing your request", http.StatusInternalServerError)
	}
}
