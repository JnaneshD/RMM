package handlers

import (
	"fmt"
	"log"
	"net/http"

	"example.com/test/models"
	"example.com/test/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleServerSideSocket(ctx *gin.Context, hub *ws.Hub, log *log.Logger) {
	conn, _ := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	client_id := ctx.Param("id")

	client := ws.NewClient(client_id, conn)

	hub.Register(client)

	done := make(chan bool)

	defer func() {
		close(done)
		hub.Unregister(client)
		conn.Close()
	}()

	go SendCommandsToClient(client, conn, done)

	// Lets declare a variable to hold incoming jobs
	for {
		var jobdata_from_client models.Job
		err := conn.ReadJSON(&jobdata_from_client)
		if err != nil {
			fmt.Printf("Error reading from the agent socket %s: %v", client_id, err)
			break
		}

		// Do we have the same job in our array
		if _, ok := hub.Client_Jobs[client][jobdata_from_client.ID]; ok == true {
			hub.Client_Jobs[client][jobdata_from_client.ID] = jobdata_from_client

		} else {
			fmt.Println("We messed up")
		}
		// Now add the job to the

	}

}

func SendCommandsToClient(client *ws.Client, conn *websocket.Conn, done chan bool) {
	defer fmt.Println("Stopping writes")
	for {
		select {
		case message := <-client.Send:
			conn.WriteJSON(message)
			//conn.WriteMessage(websocket.TextMessage, []byte(message))
		case <-done:
			return
		}
	}
}

func HandlePushMessage(ctx *gin.Context, hub *ws.Hub, log *log.Logger) {
	id := ctx.Param("id")

	var job models.Job
	if err := ctx.BindJSON(&job); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	job.ClientID = id

	// Now assign a id to the job
	job_id := hub.NextID()
	job.ID = job_id

	client, exists := hub.GetClient(id)
	hub.AddJobToClient(job, client)
	if exists {
		select {
		case client.Send <- job:
			ctx.JSON(200, gin.H{"status": "Sent to agent"})
		default:
			ctx.JSON(504, gin.H{"error": "Agent channel full"})
		}
	} else {
		ctx.JSON(404, gin.H{"error": "Agent is down"})
	}
}
