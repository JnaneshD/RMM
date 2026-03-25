package main

import (
	"log"

	"example.com/test/handlers"
	"example.com/test/ws"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Cleanup(hub *ws.Hub) {
	hub.Stop()
}

func main() {

	// logger

	log.SetOutput(&lumberjack.Logger{
		Filename:   "clientSide.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})
	// Now i have to initialize a new hub
	our_hub := ws.NewHub()

	go our_hub.Run()
	// Creating the Gin Router with default middleware
	r := gin.Default()

	// write some api
	r.GET("/ws/:id", func(ctx *gin.Context) {
		handlers.HandleServerSideSocket(ctx, our_hub, log.Default())
	})

	r.POST("/push/:id", func(ctx *gin.Context) {
		handlers.HandlePushMessage(ctx, our_hub, log.Default())
	})

	r.GET("/clients", func(ctx *gin.Context) {
		handlers.ReturnClients(ctx, our_hub)
	})

	r.GET("/jobs", func(ctx *gin.Context) {
		handlers.ReturnJobs(ctx, our_hub)
	})

	r.Run("0.0.0.0:8000")

	// Add cleanup
	defer Cleanup(our_hub)
}
