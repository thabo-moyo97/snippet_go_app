package routes

import (
	"net/http"

	"github.com/justinas/alice"
	"thabomoyo.co.uk/cmd/web/handlers"
)

func (rh *RouteHandler) UserRoutes(protected, dynamic alice.Chain) http.Handler {
	mux := http.NewServeMux()

	userHandler := handlers.NewUserHandler(rh.services)

	// Public routes with dynamic middleware
	mux.Handle("GET /signup", dynamic.ThenFunc(userHandler.UserSignup))
	mux.Handle("POST /signup", dynamic.ThenFunc(userHandler.UserSignupPost))
	mux.Handle("GET /login", dynamic.ThenFunc(userHandler.UserLogin))
	mux.Handle("POST /login", dynamic.ThenFunc(userHandler.UserLoginPost))

	// Protected routes
	mux.Handle("POST /logout", protected.ThenFunc(userHandler.UserLogoutPost))
	mux.Handle("GET /account/view", protected.ThenFunc(userHandler.UserAccountView))

	return mux
}
