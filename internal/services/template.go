package services

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/templatemanager"
	"thabomoyo.co.uk/ui"
)

type TemplateService struct {
	cache         map[string]*template.Template
	manager       *templatemanager.Manager
	sessions      *scs.SessionManager
	logger        *slog.Logger
	functions     template.FuncMap
	isDevelopment bool
}

type TemplateData struct {
	Snippet         models.Snippet
	Snippets        []models.Snippet
	CurrentYear     int
	ErrorMessage    string
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            models.User
}

func NewTemplateService(cache map[string]*template.Template, manager *templatemanager.Manager, sessions *scs.SessionManager, logger *slog.Logger, isDevelopment bool) *TemplateService {
	functions := template.FuncMap{
		"humanDate": humanDate,
	}

	return &TemplateService{
		cache:         cache,
		manager:       manager,
		sessions:      sessions,
		logger:        logger,
		functions:     functions,
		isDevelopment: isDevelopment,
	}
}

func (s *TemplateService) NewTemplateData(r *http.Request) *TemplateData {
	return &TemplateData{
		CurrentYear:     time.Now().Year(),
		Flash:           s.GetFlash(r),
		IsAuthenticated: s.IsAuthenticated(r),
		CSRFToken:       s.GetCSRFToken(r),
	}
}

func (s *TemplateService) RenderViewError(w http.ResponseWriter, r *http.Request, status int, data *TemplateData) {
	ts, err := s.getTemplateFile("error")
	if err != nil {
		s.logger.Error("the error template does not exist", "error", err)
		return
	}
	buffer := new(bytes.Buffer)
	err = ts.ExecuteTemplate(buffer, "base", data)
	if err != nil {
		s.logger.Error("template execution failed", "error", err)
		return
	}

	w.WriteHeader(status)
	buffer.WriteTo(w)
}

func (s *TemplateService) RenderView(w http.ResponseWriter, r *http.Request, status int, page string, data *TemplateData) {
	page = s.manager.NormaliseTemplateName(page)
	ts, err := s.getTemplateFile(page)
	if err != nil {
		s.logger.Error("template not found", "name", page)
		s.RenderViewError(w, r, http.StatusInternalServerError, data)
		return
	}

	buffer := new(bytes.Buffer)
	err = ts.ExecuteTemplate(buffer, "base", data)

	if err != nil {
		s.logger.Error("template execution failed", "error", err)
		s.RenderViewError(w, r, http.StatusInternalServerError, data)
		return
	}

	w.WriteHeader(status)
	buffer.WriteTo(w)
}

func (s *TemplateService) GetFlash(r *http.Request) string {
	flash := s.sessions.Get(r.Context(), "flash")
	s.sessions.Remove(r.Context(), "flash")
	if flash != nil {
		return flash.(string)
	}
	return ""
}

func (s *TemplateService) IsAuthenticated(r *http.Request) bool {
	isAuthenticated := s.sessions.Get(r.Context(), "authenticatedUserID")
	return isAuthenticated != nil
}

func (s *TemplateService) GetCSRFToken(r *http.Request) string {
	return nosurf.Token(r)
}

func (s *TemplateService) loadTemplates(fsys fs.FS) ([]string, error) {
	var pages []string
	err := fs.WalkDir(fsys, "html", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			for _, ext := range s.manager.GetExtensions() {
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

func (s *TemplateService) parseTemplate(fsys fs.FS, name string, patterns []string) (*template.Template, error) {
	return template.New(name).Funcs(s.functions).ParseFS(fsys, patterns...)
}

func (s *TemplateService) NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	fsys := ui.Files

	pages, err := s.loadTemplates(fsys)
	if err != nil {
		s.logger.Error("error walking templates directory", "error", err)
		return nil, err
	}

	for _, page := range pages {
		name := s.manager.NormaliseTemplateName(page)
		s.logger.Info("processing template", "name", name)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := s.parseTemplate(fsys, name, patterns)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func (s *TemplateService) Get(name string) *template.Template {
	return s.cache[name]
}

func GetProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func (s *TemplateService) getTemplateFile(name string) (*template.Template, error) {
	if !s.isDevelopment {
		return s.cache[name], nil
	}

	fsys := ui.ViewFiles
	pages, err := s.loadTemplates(fsys)
	if err != nil {
		s.logger.Error("error walking templates directory", "error", err)
		return nil, err
	}

	for _, page := range pages {
		normalisedName := s.manager.NormaliseTemplateName(page)
		if normalisedName == name {
			patterns := []string{
				"html/base.html",
				"html/partials/*.html",
				page,
			}

			return s.parseTemplate(fsys, name, patterns)
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}
