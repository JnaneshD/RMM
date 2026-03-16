package handlers

import (
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
