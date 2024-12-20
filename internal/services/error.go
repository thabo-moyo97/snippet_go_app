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

func (s *ErrorService) ServerError(w http.ResponseWriter, r *http.Request, err error, status ...int) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)
	s.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)

	if s.debugMode {
		body := fmt.Sprintf("%s\n%s", err, trace)
		statusCode := http.StatusInternalServerError
		if len(status) > 0 {
			statusCode = status[0]
		}
		http.Error(w, body, statusCode)
		return
	}

	data := s.templates.NewTemplateData(r)
	s.templates.RenderViewError(w, r, http.StatusInternalServerError, data)
}

func (s *ErrorService) ClientError(w http.ResponseWriter, r *http.Request, status int) {
	http.Error(w, http.StatusText(status), status)
}
