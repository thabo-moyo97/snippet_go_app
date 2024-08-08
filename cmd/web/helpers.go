package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	re := regexp.MustCompile(`(/[\w/.-]+):(\d+)`)
	traceWithLinks := re.ReplaceAllStringFunc(trace, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			filePath := parts[1]
			lineNumber := parts[2]
			return fmt.Sprintf(`"file://%s"%s:%s`, filePath, filePath, lineNumber)
		}
		return match
	})

	humanFriendlyTrace := formatStackTrace(traceWithLinks)

	app.logger.Error(err.Error(), slog.Any("method", method), slog.Any("uri", uri), slog.Any("trace", humanFriendlyTrace))

	template := app.newTemplateData(r)
	template.ErrorMessage = "Something went wrong. If the problem persists, please email"

	app.render(w, r, http.StatusInternalServerError, "error.tmpl", template)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func formatStackTrace(trace string) string {
	lines := strings.Split(trace, "\n")
	var formattedLines []string

	for _, line := range lines {
		if strings.Contains(line, "file://") {
			parts := strings.Split(line, "\"")
			if len(parts) > 1 {
				filePath := parts[1]
				formattedLines = append(formattedLines, fmt.Sprintf("File: %s", filePath))
			}
		} else {
			formattedLines = append(formattedLines, line)
		}
	}

	return strings.Join(formattedLines, "\n")
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buffer := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buffer, "base", data)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buffer.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}
