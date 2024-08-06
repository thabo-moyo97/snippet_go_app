package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Initialize a slice containing the paths to the two files. It's important
	// to note that the file containing our base template must be the *first*
	// file in the slice.
	files := []string{
		"./ui/html/pages/home.gohtml", "./ui/html/template.gohtml", "./ui/html/partials/nav.gohtml",
	}

	html, err := template.ParseFiles(files...)

	if err != nil {
		app.serverError(w, r, err) // Use the serverError() helper.
	}

	responseData := struct{ Items []string }{
		Items: []string{"Item 1", "Item 2", "Item 3"},
	}

	err = html.ExecuteTemplate(w, "template", responseData)

	if err != nil {
		app.serverError(w, r, err) // Use the serverError() helper.
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	_, err = fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Display a form for creating a new snippet..."))
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("Save a new snippet..."))
	if err != nil {
		app.logger.Error("Failed to write response: ", slog.Any("error", err))
		return
	}
}
