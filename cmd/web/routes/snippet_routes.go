package routes

import (
	"net/http"

	"github.com/justinas/alice"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (rh *RouteHandler) SnippetRoutes(protected, dynamic alice.Chain) http.Handler {
	mux := http.NewServeMux()

	snippetHandler := handlers.NewSnippetHandler(rh.services)

	// Public routes with dynamic middleware {s}
	mux.Handle("GET /", dynamic.ThenFunc(snippetHandler.Home))
	mux.Handle("GET /snippet/{id}/view", protected.ThenFunc(snippetHandler.SnippetShow))
	mux.Handle("GET /snippet/{id}/edit", protected.ThenFunc(snippetHandler.SnippetEditView))

	// Protected routes
	mux.Handle("GET /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreateView))
	mux.Handle("POST /snippet/create", protected.ThenFunc(snippetHandler.SnippetCreateAction))
	mux.Handle("POST /snippet/{id}/edit", protected.ThenFunc(snippetHandler.SnippetEditAction))

	// Add the new delete route - protected since only authenticated users should delete
	mux.Handle("POST /snippet/{id}/delete", protected.ThenFunc(snippetHandler.SnippetDeleteAction))

	return mux
}
