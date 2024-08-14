package routes

import (
	"github.com/justinas/alice"
	"net/http"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (route *RouteResource) SnippetRoutes(mux *http.ServeMux) http.Handler {
	protected := alice.New(route.app.SessionManager.LoadAndSave, noSurf, route.authenticate, route.requireAuthentication)

	snippetResource := &handlers.SnippetHandler{
		App: route.app,
	}

	mux.Handle("GET /{$}", protected.ThenFunc(snippetResource.Home))
	mux.Handle("GET /snippet/view/{id}", protected.ThenFunc(snippetResource.SnippetView))
	mux.Handle("GET /snippet/create", protected.ThenFunc(snippetResource.SnippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(snippetResource.SnippetCreatePost))

	return mux
}
