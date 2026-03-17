package handlers

import (
	"fmt"

	models "example.com/test/models"
	"example.com/test/ws"
	"github.com/gin-gonic/gin"
)

func ReturnClients(ctx *gin.Context, hub *ws.Hub) {
	// all clients are identified by the strings
	keys := make([]string, 0, len(hub.Clients()))
	for k := range hub.Clients() {
		keys = append(keys, k)
	}
	ctx.JSON(200, gin.H{
		"clients": keys,
	})
}

func ReturnJobs(ctx *gin.Context, hub *ws.Hub) {

	// Create a temporary map with string keys
	fmt.Println("what is happening")
	exportable := make(map[string][]models.Job)
	for client, jobs := range hub.Client_Jobs {
		exportable[client.ID] = jobs // Or use a name/address
	}

	ctx.JSON(200, gin.H{
		"mad": exportable,
	})
}
