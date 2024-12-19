package routes

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/justinas/alice"
	"thabomoyo.co.uk/internal/services"
	"thabomoyo.co.uk/ui"
)

type RouteHandler struct {
	services *services.Services
}

func NewRouteHandler(services *services.Services) *RouteHandler {
	return &RouteHandler{services: services}
}

func (rh *RouteHandler) cacheControlFileServer(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := fs.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		if rh.services.IsDevelopment {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}

		http.FileServer(fs).ServeHTTP(w, r)
	})
}

func (rh *RouteHandler) Routes() http.Handler {
	fileServer := rh.cacheControlFileServer(http.FS(ui.Files))
	staticHandler := http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/" + r.URL.Path
		fileServer.ServeHTTP(w, r)
	}))

	mux := http.NewServeMux()

	mux.Handle("/static/", staticHandler)

	dynamic := alice.New(
		recoverPanic(rh.services),
		logRequest(rh.services),
		commonHeaders,
		rh.services.Sessions.LoadAndSave,
		noSurf,
		authenticate(rh.services),
	)

	protected := dynamic.Append(requireAuthentication(rh.services))
	snippetRoutes := rh.SnippetRoutes(protected, dynamic)
	userRoutes := rh.UserRoutes(protected, dynamic)

	mux.Handle("/", snippetRoutes)
	mux.Handle("/user/", http.StripPrefix("/user", userRoutes))

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, you should check the origin
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			rh.services.Logger.Error("websocket upgrade failed", "error", err)
			return
		}

		rh.services.Watcher.AddClient(conn)
		//	defer rh.services.Watcher.RemoveClient(conn)
	})

	return mux
}
