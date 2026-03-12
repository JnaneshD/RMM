package main

import (
	"fmt"
	"log"
	"os"

	"example.com/test/handlers"
	serverside "example.com/test/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Cleanup(hub *serverside.Hub) {
	close(hub.Stop)
}

func main() {

	// logger

	fmt.Println(uuid.New())
	logfile, err := os.OpenFile("backend.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	log.SetOutput(logfile)
	// Now i have to initialize a new hub
	our_hub := serverside.NewHub()

	go our_hub.Run()
	// Creating the Gin Router with default middleware
	r := gin.Default()

	// write some api
	r.GET("/ws/:id", func(ctx *gin.Context) {
		serverside.HandleServerSideSocket(ctx, our_hub, log.Default())
	})

	r.POST("/push/:id", func(ctx *gin.Context) {
		serverside.HandlePushMessage(ctx, our_hub, log.Default())
	})

	r.GET("/clients", func(ctx *gin.Context) {
		handlers.ReturnClients(ctx, our_hub)
	})

	r.Run("0.0.0.0:8000")

	// Add cleanup
	defer Cleanup(our_hub)
}
