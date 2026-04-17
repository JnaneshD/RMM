package main

import (
	"context"
	"log"
	"time"

	_ "example.com/test/docs/openapi"
	"example.com/test/internal/repository"
	"example.com/test/internal/server/api"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/server/main.go -d ../.. -o ../../docs/openapi --parseInternal

// @title Gin Agent Control API
// @version 1.0
// @description HTTPS and websocket-backed API for agent registration, client presence, and job dispatch.
// @BasePath /
// @schemes https

func cleanup(hub *realtime.Hub) {
	hub.Stop()
}

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "backend.log",
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Handle DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	dbPool, err := repository.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Supabase : %v", err)
	}
	defer dbPool.Close()
	defer cancel()

	// Create Repos

	clientRepo := repository.NewClientRepository(dbPool)
	jobrepo := repository.NewJobRepository(dbPool)
	sessionRepo := repository.NewSessionRepository(dbPool)

	hub := realtime.NewHub()
	dispatcher := service.NewDispatcher(hub, jobrepo)
	httpHandler := api.NewHTTPHandler(dispatcher, clientRepo, jobrepo)
	socketHandler := api.NewSocketHandler(dispatcher, clientRepo, sessionRepo, log.Default())

	go hub.Run()
	defer cleanup(hub)

	router := gin.Default()
	api.RegisterRoutes(router, httpHandler, socketHandler)
	if err := router.RunTLS(":8081", "cert.pem", "key.pem"); err != nil {
		log.Fatalf("failed to start TLS server: %v", err)
	}
	//router.Run("0.0.0.0:8000")
}
