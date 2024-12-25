package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
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

	// Extract relevant stack frames
	frames := s.parseStack(trace)

	fmt.Print(err.Error(),
		"method", method,
		"uri", uri,
		"stack", s.formatStack(frames),
	)

	if s.debugMode {
		// Format error message with relevant stack frames
		body := fmt.Sprintf("Error: %v\n\nStack Trace:\n%s", err, s.formatStack(frames))
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

func (s *ErrorService) ClientError(w http.ResponseWriter, r *http.Request, status int, err *error) {
	s.logger.Info("Client error", "status", status, "error", err)
	http.Error(w, http.StatusText(status), status)
}

type StackFrame struct {
	Function string
	File     string
	Line     int
}

func (s *ErrorService) parseStack(stack string) []StackFrame {
	var frames []StackFrame
	lines := strings.Split(stack, "\n")

	for i := 1; i < len(lines)-1; i += 2 {
		functionLine := lines[i]
		fileLine := lines[i+1]

		if !strings.Contains(fileLine, "/cmd/") &&
			!strings.Contains(fileLine, "/internal/") {
			continue
		}

		funcName := strings.TrimSpace(functionLine)
		if idx := strings.LastIndex(funcName, "("); idx != -1 {
			funcName = funcName[:idx]
		}
		if idx := strings.LastIndex(funcName, "."); idx != -1 {
			funcName = funcName[idx+1:]
		}

		fileInfo := strings.TrimSpace(fileLine)
		if idx := strings.LastIndex(fileInfo, ":"); idx != -1 {
			lineStr := fileInfo[idx+1:]
			fileInfo = fileInfo[:idx]

			if plusIdx := strings.Index(lineStr, " +"); plusIdx != -1 {
				lineStr = lineStr[:plusIdx]
			}

			line, _ := strconv.Atoi(lineStr)

			if idx := strings.Index(fileInfo, "/app/"); idx != -1 {
				fileInfo = fileInfo[idx+5:]
			}

			frames = append(frames, StackFrame{
				Function: funcName,
				File:     fileInfo,
				Line:     line,
			})
		}
	}
	return frames
}

func (s *ErrorService) formatStack(frames []StackFrame) string {
	var buf strings.Builder
	fmt.Fprint(&buf, "\n")
	for i, frame := range frames {
		var methodName string

		if frame.Function == "1" {
			methodName = "-"
		} else {
			methodName = frame.Function
		}

		methodName = strings.TrimSpace(methodName)
		// Using %s instead of \n for line breaks
		fmt.Fprintf(&buf, "%d. %s %s:%d%s",
			i+1,
			methodName,
			frame.File,
			frame.Line,
			"\n")
	}

	return buf.String()
}
