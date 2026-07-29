package router

import (
	"relay/internal/app"
	"relay/internal/handlers"
	"relay/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func Register(app *app.App, userHandle *handlers.UserHandler, conversationHandler *handlers.ConversationHandle) {
	r := app.Router

	//Request Id Middleware
	r.Use(middleware.RequestID)

	//Logging Middleware
	r.Use(middleware.Logging)

	//Panic Recovery Middleware
	r.Use(middleware.Recovery)

	//Public routes
	r.Get("/health", handlers.Health)

	r.Post("/users/register", userHandle.Register)
	r.Post("/users/login", userHandle.Login)
	r.Post("/users/refresh", userHandle.Refresh)
	r.Post("/users/logout", userHandle.Logout)

	//Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(app.Token))

		r.Get("/users/profile", userHandle.Profile)
		r.Post("/conversations", conversationHandler.Create)
		r.Get("/conversations", conversationHandler.ListForUser)
		// r.Get("/conversations/{id}/messages", conversationHandler.ListForUser)
	})
}
