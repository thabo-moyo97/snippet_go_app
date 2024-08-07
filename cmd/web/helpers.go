package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"
	// New import
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

/**
 * The render() method is a wrapper around the http.ResponseWriter.Write() method
 * that provides some convenience features for rendering templates. It takes the
 * name of a template file as its first parameter, then a slice of strings
 * (representing the names of any partial templates that you want to include in
	 * the main template), followed by the usual data parameter. The method returns
	 * an error, which will be nil if the template renders correctly.
*/
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
