package router

import (
	"relay/internal/app"
	"relay/internal/handlers"
)

func Register(app *app.App) {
	app.Router.Get("/health", handlers.Health)
}
