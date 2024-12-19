package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type ErrorService struct {
	logger    *slog.Logger
	templates *TemplateService
	debugMode bool
}

func NewErrorService(logger *slog.Logger, templates *TemplateService, debugMode bool) *ErrorService {
	return &ErrorService{
		logger:    logger,
		templates: templates,
		debugMode: debugMode,
	}
}

func (s *ErrorService) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	if s.debugMode {
		body := fmt.Sprintf("%s\n%s", err, trace)
		http.Error(w, body, http.StatusInternalServerError)
		return
	}

	s.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	data := s.templates.NewTemplateData(r)
	s.templates.RenderViewError(w, r, http.StatusInternalServerError, data)
}

func (s *ErrorService) ClientError(w http.ResponseWriter, r *http.Request, status int) {

	http.Error(w, http.StatusText(status), status)
}
