package router

import (
	"relay/internal/app"
	"relay/internal/handlers"
	"relay/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func Register(app *app.App, userHandle *handlers.UserHandler) {
	r := app.Router

	//Request Id Middleware
	r.Use(middleware.RequestID)

	//Logging Middleware
	r.Use(middleware.Logging)

	//Panic Recovery Middleware
	r.Use(middleware.Rocovery)

	r.Get("/health", handlers.Health)
	r.Post("/users/register", userHandle.Register)

	r.With(middleware.Auth(app.Token)).Get("/profile", userHandle.Profile)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(app.Token))

		r.Get("/profile", app.UserHandler.Profile)
	})
}
