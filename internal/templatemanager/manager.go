package templatemanager

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"time"

	"thabomoyo.co.uk/ui"
)

type Manager struct {
	extensions    []string
	cache         map[string]*template.Template
	logger        *slog.Logger
	isDevelopment bool
}

func NewManager(logger *slog.Logger, isDevelopment *bool) *Manager {
	return &Manager{
		extensions:    []string{".tmpl", ".html", ".phtml"},
		logger:        logger,
		isDevelopment: *isDevelopment,
	}
}

func (m *Manager) loadTemplates(fsys fs.FS) ([]string, error) {
	var pages []string
	err := fs.WalkDir(fsys, "html", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			for _, ext := range m.extensions {
				if strings.HasSuffix(path, ext) {
					pages = append(pages, path)
					break
				}
			}
		}
		return nil
	})
	return pages, err
}

func (m *Manager) CreateTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	var fsys fs.FS
	fsys = ui.Files
	/* if m.isDevelopment {
		fsys = ui.ViewFiles //Load straight from disk during dev
	} else {
		fsys = ui.Files
	} */

	pages, err := m.loadTemplates(fsys)
	if err != nil {
		m.logger.Error("error walking templates directory", "error", err)
		return nil, err
	}

	for _, page := range pages {
		name := m.NormaliseTemplateName(page)
		m.logger.Info("processing template", "name", name)

		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := m.parseTemplate(fsys, name, patterns)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func (m *Manager) parseTemplate(fsys fs.FS, name string, patterns []string) (*template.Template, error) {
	return template.New(name).Funcs(functions).ParseFS(fsys, patterns...)
}

func GetProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func (m *Manager) GetTemplateFile(name string) (*template.Template, error) {
	if !m.isDevelopment {
		//return m.cache[name], nil
	}

	fsys := ui.ViewFiles
	pages, err := m.loadTemplates(fsys)
	if err != nil {
		m.logger.Error("error walking templates directory", "error", err)
		return nil, err
	}

	for _, page := range pages {
		normalisedName := m.NormaliseTemplateName(page)
		if normalisedName == name {
			patterns := []string{
				"html/base.tmpl",
				"html/partials/*.tmpl",
				page,
			}

			return m.parseTemplate(fsys, name, patterns)
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

func (m *Manager) SetExtensions(exts []string) {
	m.extensions = exts
}

func (m *Manager) NormaliseTemplateName(name string) string {
	name = strings.TrimPrefix(name, "ui/html/pages/")
	name = strings.TrimPrefix(name, "html/pages/")
	name = strings.TrimPrefix(name, "html/partials/")
	name = strings.TrimPrefix(name, "html/")

	name = strings.ReplaceAll(name, "/", ".")

	for _, ext := range m.extensions {
		name = strings.TrimSuffix(name, ext)
	}

	return name
}

func (m *Manager) GetExtensions() []string {
	return m.extensions
}

var functions = template.FuncMap{
	"humanDate":      humanDate,
	"createFormDate": createFormDate,
}

func humanDate(t interface{}) (result string) {
	var timeVal time.Time

	switch v := t.(type) {
	case sql.NullTime:
		if v.Valid {
			timeVal = v.Time
		} else {
			return "never"
		}
	case time.Time:
		timeVal = v
	case *time.Time:
		if v != nil {
			timeVal = *v
		} else {
			return "Never"
		}
	default:
		return "Never"
	}

	if timeVal.IsZero() {
		return "Never"
	}

	return timeVal.UTC().Format("02 Jan 2006 at 15:04")
}

func createFormDate(t interface{}) (result string) {
	var timeVal string

	switch v := t.(type) {
	case sql.NullTime:
		if v.Valid {
			timeVal = v.Time.Format("02 Jan 2006 at 15:04")
		} else {
			return "01 Jan 0001 at 00:00"
		}
	case time.Time:
		timeVal = v.Format("02 Jan 2006 at 15:04")
	default:
		timeVal = "01 Jan 0001 at 00:00"
	}

	return timeVal
}
