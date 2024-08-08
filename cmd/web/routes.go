package main

import (
	"net/http"
)

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	// Create a file server which serves files out of the ./ui/static directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// Use the mux.Handle() function to register the file server as the handler for
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, http.StatusOK, "error.tmpl", app.newTemplateData(r))
	})

	return app.recoverPanic(app.logRequest(commonHeaders(mux)))
}
