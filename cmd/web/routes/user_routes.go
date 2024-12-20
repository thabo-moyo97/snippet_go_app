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
	mux.Handle("GET /signup", dynamic.ThenFunc(userHandler.UserSignupView))
	mux.Handle("POST /signup", dynamic.ThenFunc(userHandler.UserSignupPostAction))
	mux.Handle("GET /login", dynamic.ThenFunc(userHandler.UserLoginView))
	mux.Handle("POST /login", dynamic.ThenFunc(userHandler.UserLoginPostAction))

	// Protected routes
	mux.Handle("POST /logout", protected.ThenFunc(userHandler.UserLogoutPostAction))
	mux.Handle("GET /account/view", protected.ThenFunc(userHandler.UserAccountView))

	return mux
}
