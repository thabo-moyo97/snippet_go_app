package routes

import (
	"net/http"

	"github.com/justinas/alice"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/ui"
)

type RouteResource struct {
	app *config.Application
}

func cacheControlFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := fs.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		w.Header().Set("Cache-Control", "public, max-age=31536000")

		http.FileServer(fs).ServeHTTP(w, r)
	})
}

func Routes(app *config.Application) http.Handler {
	routeResources := &RouteResource{app: app}

	fileServer := cacheControlFileServer(http.FS(ui.Files))
	staticHandler := http.StripPrefix("/static", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static" + r.URL.Path
		fileServer.ServeHTTP(w, r)
	}))

	mux := http.NewServeMux()

	mux.Handle("GET /static/", staticHandler)

	dynamic := alice.New(
		routeResources.recoverPanic,
		routeResources.logRequest,
		commonHeaders,
		app.SessionManager.LoadAndSave,
		noSurf,
		routeResources.authenticate,
	)

	protected := dynamic.Append(routeResources.requireAuthentication)

	snippetRoutes := routeResources.SnippetRoutes(protected, dynamic)
	userRoutes := routeResources.UserRoutes(protected, dynamic)

	mux.Handle("/", snippetRoutes)
	mux.Handle("/user/", http.StripPrefix("/user", userRoutes))

	return mux
}
