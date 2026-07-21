package main

import (
	"fmt"
	"log"
	"net/http"

	"relay/internal/app"
	"relay/internal/config"
	"relay/internal/database"
	"relay/internal/handlers"
	"relay/internal/repositories/postgres"
	"relay/internal/router"
	"relay/internal/services"
	"relay/internal/token"
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

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close the database: %v", err)
		}
	}()

	fmt.Println("✅ Database connected")

	// Repository Injections
	userRepo := postgres.NewUserRepository(db)

	// JWT token Service Injections
	tokenService := token.NewService(cfg.JWT.Secret)

	// Service Injections
	userService := services.NewUserService(userRepo, tokenService)

	// Service Injections
	userHandler := handlers.NewUserHandler(userService)

	//app
	application := app.New(cfg, db, userHandler, tokenService)

	//router
	router.Register(application, userHandler)

	log.Printf("Starting server on :%s", cfg.App.Port)

	err = http.ListenAndServe(":"+cfg.App.Port, application.Router)
	if err != nil {
		log.Fatal(err)
	}
}
