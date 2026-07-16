package main

import (
	"fmt"
	"log"
	"net/http"

	"relay/internal/app"
	"relay/internal/config"
	"relay/internal/database"
	"relay/internal/router"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("✅ Database connected")

	application := app.New(cfg, db)

	router.Register(application)

	log.Printf("Starting server on :%s", cfg.App.Port)

	err = http.ListenAndServe(":"+cfg.App.Port, application.Router)
	if err != nil {
		log.Fatal(err)
	}
}
