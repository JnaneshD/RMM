package main

import (
	"log"

	"example.com/test/internal/server/api"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
	"example.com/test/internal/server/store/memory"
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

	hub := realtime.NewHub()
	store := memory.NewJobStore()
	dispatcher := service.NewDispatcher(hub, store)
	httpHandler := api.NewHTTPHandler(dispatcher)
	socketHandler := api.NewSocketHandler(dispatcher, log.Default())

	go hub.Run()
	defer cleanup(hub)

	router := gin.Default()
	router.GET("/ws/:id", socketHandler.HandleServerSideSocket)
	router.POST("/push/:id", httpHandler.HandlePushMessage)
	router.GET("/clients", httpHandler.ReturnClients)
	router.GET("/jobs", httpHandler.ReturnJobs)

	router.Run("0.0.0.0:8000")
}
