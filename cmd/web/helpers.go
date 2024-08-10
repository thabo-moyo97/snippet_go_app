package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4" // New import
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
		Flash:       app.sessionManager.PopString(r.Context(), "flash"),
	}
}

func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// snippetCreatePost handler.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call Decode() on our decoder instance, passing the target destination as
	// the first parameter.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError.We use
		// errors.As() to check for this and raise a panic rather than returning
		// the error.
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// For all other errors, we return them as normal.
		return err
	}

	return nil
}
