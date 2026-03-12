package logic

import (
	"fmt"
	"log"
	"net/http"

	"example.com/test/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleServerSideSocket(ctx *gin.Context, hub *Hub, log *log.Logger) {
	// Now we have the gin's context and websocket created,
	// Handle the connection to websocket and make a persistent messaging
	// channel
	conn, _ := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	client_id := ctx.Param("id")

	client := models.NewClient(client_id, conn)

	hub.Register(client)

	// Now run and listen to the client
	done := make(chan bool)

	defer func() {
		close(done)
		hub.Unregister(client)
		conn.Close()
	}()

	go SendCommandsToClient(client, conn, done)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading from the agent sockett %s: %v", "test", err)
			break
		}
		fmt.Printf("Received this message : %s\n", message)
	}

}

func SendCommandsToClient(client *models.Client, conn *websocket.Conn, done chan bool) {
	// This function is responsible for sending a message
	// to the client over the websocket after the POST request

	// Listen to the client send channel
	defer fmt.Println("Stopping writes")
	for {
		select {
		case message := <-client.Send:
			conn.WriteMessage(websocket.TextMessage, []byte(message))
		case <-done:
			return
		}
	}
}

// HandlePushMessage handles POST requests to send messages to connected clients
func HandlePushMessage(ctx *gin.Context, hub *Hub, log *log.Logger) {
	id := ctx.Param("id")

	var job models.Job
	if err := ctx.BindJSON(&job); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hub.Mu.RLock()
	client, exists := hub.Clients()[id]
	hub.Mu.RUnlock()

	if exists {
		select {
		case client.Send <- job.Command:
			ctx.JSON(200, gin.H{"status": "Sent to agent"})
		default:
			ctx.JSON(504, gin.H{"error": "meh"})
		}
	} else {
		ctx.JSON(404, gin.H{"error": "Agent is down"})
	}
}
