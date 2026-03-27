package runtime

import (
	"context"
	"log"
	"time"

	"example.com/test/internal/domain"
	"example.com/test/internal/heartbeat"
	"github.com/gorilla/websocket"
)

func ReadSocket(ctx context.Context, jobQueue chan *domain.Job, conn *websocket.Conn) {
	defer close(jobQueue)

	conn.SetReadDeadline(time.Now().Add(heartbeat.PongWait))
	conn.SetPingHandler(func(data string) error {
		conn.SetReadDeadline(time.Now().Add(heartbeat.PongWait))
		return conn.WriteControl(
			websocket.PongMessage, []byte(data),
			time.Now().Add(heartbeat.WriteWait),
		)
	})

	for {
		var newJob domain.Job
		err := conn.ReadJSON(&newJob)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("read_socket: context cancelled, stopping the reader")
				return
			}
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Println("read_socket: unexpected close error:", err)
			} else {
				log.Println("read_socket: connection dead (missed heartbeat):", err)
			}
			return
		}

		log.Printf("Received this job %s", newJob.Command)
		newJob.Status = domain.WAIT

		select {
		case jobQueue <- &newJob:
		case <-ctx.Done():
			log.Println("read_socket: context cancelled while enqueuing, stopping")
			return
		}
	}
}
