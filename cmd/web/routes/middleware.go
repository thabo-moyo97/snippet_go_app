package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"thabomoyo.co.uk/internal/services"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := map[string]string{
			"Content-Security-Policy": " style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; script-src 'self' 'unsafe-inline'; img-src 'self'",
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

func logRequest(services *services.Services) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ip     = r.RemoteAddr
				proto  = r.Proto
				method = r.Method
				uri    = r.URL.RequestURI()
			)

			services.Logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
			next.ServeHTTP(w, r)
		})
	}
}

/**
 * The recoverPanic middleware is used to recover from any panics that occur during the request/response cycle.
 */
func recoverPanic(services *services.Services) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					services.Logger.Error("panic recovered", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

/**
 * handler wrapper to check if the user is authenticated before allowing access to the handler.
 */
func requireAuthentication(services *services.Services) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !services.Templates.IsAuthenticated(r) {
				services.Sessions.Put(r.Context(), "flash", "You must be authenticated to access this page.")
				http.Redirect(w, r, "/user/login", http.StatusSeeOther)
				return
			}
			w.Header().Add("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
}

// CSRF protection middleware
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Name:     "csrf_token",
		Domain:   "",
	})

	csrfHandler.ExemptPath("/static/")
	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("CSRF Validation Failed for %s\n", r.URL.Path)
		http.Error(w, "CSRF token validation failed", http.StatusBadRequest)
	}))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfHandler.ServeHTTP(w, r)
	})
}

func authenticate(services *services.Services) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := services.Sessions.Get(r.Context(), "authenticatedUserID")
			if id == nil {
				next.ServeHTTP(w, r)
				return
			}

			exists, err := services.Users.Exists(id.(int))
			if err != nil {
				services.Logger.Error("user existence check failed", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if exists {
				ctx := context.WithValue(r.Context(), "isAuthenticated", true)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
