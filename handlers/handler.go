package handlers

import (
	"example.com/test/models"
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
	exportable := make(map[string][]models.Job)
	for client, jobs := range hub.Client_Jobs {
		clientID := client.ID

		if exportable[clientID] == nil {
			exportable[clientID] = make([]models.Job, 0)

		}
		for _, job := range jobs {
			exportable[clientID] = append(exportable[clientID], *job)
		}
	}

	ctx.JSON(200, gin.H{
		"jobs": exportable,
	})
}
