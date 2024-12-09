package routes

import (
	"net/http"

	"github.com/justinas/alice"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (rr *RouteResource) SnippetRoutes(protected, dynamic alice.Chain) http.Handler {
	mux := http.NewServeMux()

	snippetHandler := &handlers.SnippetHandler{App: rr.app}

	// Public routes with dynamic middleware
	mux.Handle("GET /", dynamic.ThenFunc(snippetHandler.Home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(snippetHandler.SnippetView))

	// Protected routes
	mux.Handle("GET /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreatePost))

	return mux
}
