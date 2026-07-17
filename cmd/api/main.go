package main

import (
	"fmt"
	"log"
	"net/http"

	"relay/internal/app"
	"relay/internal/config"
	"relay/internal/database"
	"relay/internal/repositories/postgres"
	"relay/internal/router"
	"relay/internal/services"
)

func main() {

	// Env Configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Database Configurations
	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("✅ Database connected")

	// Repository Injections
	userRepo := postgres.NewUserRepository(db)

	// Service Injections
	userService := services.NewUserService(userRepo)

	application := app.New(cfg, db)

	router.Register(application)

	log.Printf("Starting server on :%s", cfg.App.Port)

	err = http.ListenAndServe(":"+cfg.App.Port, application.Router)
	if err != nil {
		log.Fatal(err)
	}
}
