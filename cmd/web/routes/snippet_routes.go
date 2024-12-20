package routes

import (
	"net/http"

	"github.com/justinas/alice"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (rh *RouteHandler) SnippetRoutes(protected, dynamic alice.Chain) http.Handler {
	mux := http.NewServeMux()

	snippetHandler := handlers.NewSnippetHandler(rh.services)

	// Public routes with dynamic middleware
	mux.Handle("GET /", dynamic.ThenFunc(snippetHandler.Home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(snippetHandler.SnippetView))
	mux.Handle("GET /snippet/edit/{id}", dynamic.ThenFunc(snippetHandler.SnippetEditView))

	// Protected routes
	mux.Handle("GET /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreateView))
	mux.Handle("POST /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreatePostAction))

	return mux
}
