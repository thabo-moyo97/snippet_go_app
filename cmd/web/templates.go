package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"strings"
	"time"

	"thabomoyo.co.uk/ui"

	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/templatemanager"
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

func newTemplateCache(templateManager *templatemanager.Manager) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	var pages []string
	err := fs.WalkDir(ui.Files, "html", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			for _, ext := range templateManager.GetExtensions() {
				if strings.HasSuffix(path, ext) {
					pages = append(pages, path)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking templates directory: %w", err)
	}

	for _, page := range pages {
		name := templateManager.NormaliseTemplateName(page)
		log.Printf("Processing template: %s", name)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
