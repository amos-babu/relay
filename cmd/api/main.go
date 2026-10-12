package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"relay/internal/app"
	"relay/internal/config"
	"relay/internal/database"
	"relay/internal/events"
	"relay/internal/handlers"
	"relay/internal/repositories/postgres"
	"relay/internal/router"
	"relay/internal/services"
	"relay/internal/token"
	"relay/internal/websocket"

	"github.com/redis/go-redis/v9"
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

	//Creating a redisClient
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//Redis Publisher Injections
	publisher := events.NewRedisPublisher(redisClient)

	//Redis Subscriber Injections
	subscriber := events.NewRedisSubscriber(redisClient, hub)

	//Start redis subscriber
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		if err := subscriber.Subscribe(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("Redis subscriber stopped: %v", err)
		}
	}()

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("failed to close Redis client: %v", err)
		}
	}()

	// Service Injections
	userService := services.NewUserService(userRepo, tokenService, refreshToken)
	conversationService := services.NewConversationService(conversationRepo, userRepo)
	messageService := services.NewMessageService(messageRepo, conversationRepo, publisher)
	websocketService := services.NewWebsocketService(conversationRepo, messageService, hub)

	// Handler Injections
	userHandler := handlers.NewUserHandler(userService)
	conversationHandler := handlers.NewConversationHandle(conversationService)
	messageHandler := handlers.NewMessageHandler(messageService)
	websocketHandler := websocket.NewHandler(
		hub,
		websocketService.HandleConnect,
		websocketService.HandleDisconnect,
		websocketService.HandleEvent,
	)

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
