package app

import (
	"database/sql"
	"relay/internal/config"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
	Router *chi.Mux
}

func New(cfg *config.Config, db *sql.DB) *App {
	return &App{
		Config: cfg,
		DB:     db,
		Router: chi.NewRouter(),
	}
}
