package main

import (
	"html/template"
	"io/fs"
	"log"
	"path/filepath"
	"time"

	"thabomoyo.co.uk/ui"

	"thabomoyo.co.uk/internal/models"
)

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

type templateData struct {
	Snippet         models.Snippet
	Snippets        []models.Snippet
	CurrentYear     int
	ErrorMessage    string
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string // Add a CSRFToken field.
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		log.Printf("Error finding template pages: %v", err)
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		log.Printf("Processing template: %s", name)

		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			log.Printf("Error parsing template %s: %v", name, err)
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
