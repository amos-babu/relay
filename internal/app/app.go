package app

import (
	"database/sql"
	"relay/internal/config"
	"relay/internal/handlers"
	"relay/internal/token"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Router *chi.Mux

	UserHandler *handlers.UserHandler
	Token       *token.Service
}

func New(cfg *config.Config, db *sql.DB, userHandler *handlers.UserHandler, token *token.Service) *App {
	return &App{
		Config:      cfg,
		DB:          db,
		Router:      chi.NewRouter(),
		UserHandler: userHandler,
		Token:       token,
	}
}
