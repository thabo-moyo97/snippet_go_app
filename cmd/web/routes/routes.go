package routes

import (
	"github.com/justinas/alice"
	"net/http"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/ui"
)

type RouteResource struct {
	app *config.Application
}

func cacheControlFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set cache control headers
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
		http.FileServer(fs).ServeHTTP(w, r)
	})
}

func Routes(app *config.Application) http.Handler {
	mux := http.NewServeMux()

	//TODO - Add a redirect route for http requests to https

	mux.Handle("/static/", cacheControlFileServer(http.FS(ui.Files)))

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	routeResources := &RouteResource{
		app: app,
	}

	mux.Handle("/", routeResources.SnippetRoutes(mux))
	mux.Handle("/user/", http.StripPrefix("/user", routeResources.UserRoutes(mux)))

	standard := alice.New(routeResources.recoverPanic, routeResources.logRequest, commonHeaders)
	return standard.Then(mux)
}
