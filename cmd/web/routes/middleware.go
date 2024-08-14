package routes

import (
	"context"
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
	"thabomoyo.co.uk/cmd/web/config"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := map[string]string{
			"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline'; font-src fonts.gstatic.com",
			"Referrer-Policy":         "origin-when-cross-origin",
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "deny",
			"X-XSS-Protection":        "0",
			"Server":                  "Go",
		}

		for key, value := range headers {
			w.Header().Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}

func (route *RouteResource) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		route.app.Logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

/**
 * The recoverPanic middleware is used to recover from any panics that occur during the request/response cycle.
 */
func (route *RouteResource) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				route.app.ServerError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

/**
 * handler wrapper to check if the user is authenticated before allowing access to the handler.
 */
func (route *RouteResource) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !route.app.IsAuthenticated(r) {
			route.app.SessionManager.Put(r.Context(), "flash", "You must be authenticated to access this page.")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// CSRF protection middleware
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (route *RouteResource) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := route.app.SessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		exists, err := route.app.Users.Exists(id)
		if err != nil {
			route.app.ServerError(w, r, err)
			return
		}

		if exists {
			ctx := context.WithValue(r.Context(), config.IsAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
