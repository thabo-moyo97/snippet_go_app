package services

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/templatemanager"
)

type TemplateService struct {
	manager       *templatemanager.Manager
	sessions      *scs.SessionManager
	logger        *slog.Logger
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

func NewTemplateService(manager *templatemanager.Manager, sessions *scs.SessionManager, logger *slog.Logger, isDevelopment *bool) *TemplateService {

	return &TemplateService{
		manager:       manager,
		sessions:      sessions,
		logger:        logger,
		isDevelopment: *isDevelopment,
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
	ts, err := s.manager.GetTemplateFile("error")
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
	ts, err := s.manager.GetTemplateFile(page)
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
