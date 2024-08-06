package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
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

	app.logger.Error(err.Error(), slog.Any("method", method), slog.Any("uri", uri), slog.Any("trace", traceWithLinks))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
