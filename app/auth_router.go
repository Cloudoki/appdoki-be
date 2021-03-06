package app

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (a *Application) AuthRouter(router *mux.Router) {
	authHandler := NewAuthHandler(a.conf.AppConfig, a.usersRepository, a.notifier)

	// for local testing purposes
	router.
		Methods(http.MethodGet).
		Path("/auth/login").
		HandlerFunc(authHandler.Login)

	// for local testing purposes
	router.
		Methods(http.MethodGet).
		Path("/auth/google/callback").
		HandlerFunc(authHandler.Callback)

	//router.
	//	Methods(http.MethodGet).
	//	Path("/auth/token").
	//	HandlerFunc(authHandler.Token)

	router.
		Methods(http.MethodGet).
		Path("/auth/url").
		HandlerFunc(a.JwtVerify(authHandler.GetURL))

	router.
		Methods(http.MethodGet).
		Path("/auth/user").
		HandlerFunc(a.JwtVerify(authHandler.FindCreateUser))
}
