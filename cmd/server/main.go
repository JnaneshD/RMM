package main

import (
	"context"
	"log"

	"example.com/test/internal/repository"
	"example.com/test/internal/server/api"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

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

	// Handle DB
	ctx := context.Background()
	dbPool, err := repository.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to Supabase : %v", err)
	}
	defer dbPool.Close()

	// Create Repos

	clientRepo := repository.NewClientRepository(dbPool)
	jobrepo := repository.NewJobRepository(dbPool)
	sessionRepo := repository.NewSessionRepository(dbPool)

	hub := realtime.NewHub()
	dispatcher := service.NewDispatcher(hub, jobrepo)
	httpHandler := api.NewHTTPHandler(dispatcher, clientRepo)
	socketHandler := api.NewSocketHandler(dispatcher, clientRepo, sessionRepo, log.Default())

	go hub.Run()
	defer cleanup(hub)

	router := gin.Default()
	router.GET("/ws/:id", socketHandler.HandleServerSideSocket)
	router.POST("/push/:id", httpHandler.HandlePushMessage)
	router.GET("/clients", httpHandler.ReturnClients)
	router.GET("/jobs", httpHandler.ReturnJobs)
	router.POST("/register", httpHandler.HandleRegistration)
	if err := router.RunTLS(":8080", "cert.pem", "key.pem"); err != nil {
		log.Fatalf("failed to start TLS server: %v", err)
	}
	//router.Run("0.0.0.0:8000")
}
