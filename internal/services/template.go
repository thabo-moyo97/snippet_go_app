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
	ts, ok := s.cache["error"]
	if !ok {
		s.logger.Error("the error template does not exist")
		return
	}
	buffer := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buffer, "base", data)
	if err != nil {
		s.logger.Error("template execution failed", "error", err)
		return
	}

	w.WriteHeader(status)
	buffer.WriteTo(w)
}

func (s *TemplateService) RenderView(w http.ResponseWriter, r *http.Request, status int, page string, data *TemplateData) {
	fmt.Println("partial", r.URL.Query().Get("partial"))

	page = s.manager.NormaliseTemplateName(page)
	ts, ok := s.cache[page]
	if !ok {
		s.logger.Error("template not found", "name", page)
		s.RenderViewError(w, r, http.StatusInternalServerError, data)
		return
	}

	buffer := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buffer, "base", data)
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

func (s *TemplateService) NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	var pages []string
	err := fs.WalkDir(ui.Files, "html", func(path string, d fs.DirEntry, err error) error {
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
	if err != nil {
		return nil, fmt.Errorf("error walking templates directory: %w", err)
	}

	for _, page := range pages {
		name := s.manager.NormaliseTemplateName(page)
		s.logger.Info("processing template", "name", name)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(s.functions).ParseFS(ui.Files, patterns...)
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
