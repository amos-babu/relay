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
	"relay/internal/websocket"
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

	// User Repository Injections
	userRepo := postgres.NewUserRepository(db)
	conversationRepo := postgres.NewConversationRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	//RefreshToken Repository
	refreshToken := postgres.NewRefreshTokenRepository(db)

	// JWT token Service Injections
	tokenService := token.NewService(cfg.JWT.Secret)

	//Hub Injections
	hub := websocket.NewHub()

	// Service Injections
	userService := services.NewUserService(userRepo, tokenService, refreshToken)
	conversationService := services.NewConversationService(conversationRepo, userRepo)
	messageService := services.NewMessageService(messageRepo, conversationRepo, hub)
	websocketService := services.NewWebsocketService(conversationRepo, hub)

	// Handler Injections
	userHandler := handlers.NewUserHandler(userService)
	conversationHandler := handlers.NewConversationHandle(conversationService)
	messageHandler := handlers.NewMessageHandle(messageService)
	websocketHandler := websocket.NewHandler(hub, websocketService.HandleEvent)

	//app
	application := app.New(cfg, db, userHandler, tokenService)

	//router
	router.Register(
		application,
		userHandler,
		conversationHandler,
		messageHandler,
		websocketHandler,
	)

	log.Printf("Starting server on :%s", cfg.App.Port)

	err = http.ListenAndServe(":"+cfg.App.Port, application.Router)
	if err != nil {
		log.Fatal(err)
	}
}
