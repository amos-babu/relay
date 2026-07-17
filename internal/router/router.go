package router

import (
	"relay/internal/app"
	"relay/internal/handlers"
	"relay/internal/middleware"
)

func Register(app *app.App, userHandle *handlers.UserHandler) {
	app.Router.Use(middleware.RequestID)
	app.Router.Use(middleware.Logging)
	app.Router.Use(middleware.Rocovery)
	app.Router.Get("/health", handlers.Health)
	app.Router.Post("/users/register", userHandle.Register)
}
