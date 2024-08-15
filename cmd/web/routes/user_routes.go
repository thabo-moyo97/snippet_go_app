package routes

import (
	"github.com/justinas/alice"
	"net/http"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (route *RouteResource) UserRoutes(mux *http.ServeMux) http.Handler {
	dynamic := alice.New(route.app.SessionManager.LoadAndSave, noSurf, route.authenticate)
	protected := dynamic.Append(route.requireAuthentication)

	userResource := &handlers.UserHandler{
		App: route.app,
	}

	/**
	 * Prefix all routes with /user
	 */
	mux.Handle("GET /signup", dynamic.ThenFunc(userResource.UserSignup))
	mux.Handle("GET /login", dynamic.ThenFunc(userResource.UserLogin))
	mux.Handle("GET /account/view", protected.ThenFunc(userResource.UserAccountView))

	mux.Handle("POST /login", dynamic.ThenFunc(userResource.UserLoginPost))
	mux.Handle("POST /signup", dynamic.ThenFunc(userResource.UserSignupPost))
	mux.Handle("POST /logout", dynamic.ThenFunc(userResource.UserLogoutPost))

	return mux
}
