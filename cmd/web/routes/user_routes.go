package routes

import (
	"github.com/justinas/alice"
	"net/http"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (route *RouteResource) UserRoutes(mux *http.ServeMux) http.Handler {
	dynamic := alice.New(route.app.SessionManager.LoadAndSave, noSurf, route.authenticate)

	userResource := &handlers.UserHandler{
		App: route.app,
	}

	mux.Handle("GET /user/signup", dynamic.ThenFunc(userResource.UserSignup))
	mux.Handle("GET /user/login", dynamic.ThenFunc(userResource.UserLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(userResource.UserLoginPost))

	mux.Handle("POST /user/signup", dynamic.ThenFunc(userResource.UserSignupPost))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(userResource.UserLogoutPost))

	return mux
}
