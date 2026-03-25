package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"example.com/test/models"
	"example.com/test/wsheartbeat"
	"github.com/gorilla/websocket"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	JobQueueSize = 100
	NumWorkers   = 4
)

func main() {
	// We should configure the log file also

	log.SetOutput(&lumberjack.Logger{
		Filename:   "clientSide.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})

	url := "ws://localhost:8000/ws/agent1"
	log.Printf("Connecting to the URL : %s", url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial error", err)
	}

	log.Println("Connected to the server! Waiting for the jobs")

	ctx, cancel := context.WithCancel(context.Background())

	// Create aa channel to catch Ctrl + C so we close gracefully
	interrupt := make(chan os.Signal, 1)

	// Lets create a Job Queue and Schduler
	jobQueue := make(chan *models.Job, JobQueueSize)

	go func() {
		<-interrupt
		cancel()
	}()

	scheduler := &JobScheduler{
		Jobs: make([]models.Job, 0),
		conn: conn,
		send: make(chan *models.Job, 10),
	}

	// Setup the write pump
	go scheduler.writePump()

	// Start those workers here
	var wg sync.WaitGroup
	for i := range NumWorkers {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			Worker(ctx, workerId, jobQueue, scheduler)
		}(i)
	}
	read_socket(ctx, jobQueue, conn)
	cancel()
	wg.Wait()
	log.Println("Shutting down")
}

func read_socket(ctx context.Context, jobQueue chan *models.Job, conn *websocket.Conn) {
	defer close(jobQueue)

	conn.SetReadDeadline(time.Now().Add(wsheartbeat.PongWait))

	conn.SetPingHandler(func(data string) error {
		log.Println("writing pong")
		conn.SetReadDeadline(time.Now().Add(wsheartbeat.PongWait))
		return conn.WriteControl(
			websocket.PongMessage, []byte(data),
			time.Now().Add(wsheartbeat.WriteWait),
		)
	})
	for {

		var newjob models.Job
		err := conn.ReadJSON(&newjob)
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
				// deadline timeout — server missed heartbeat
				log.Println("read_socket: connection dead (missed heartbeat):", err)
			}
			return
		}
		log.Printf("Received this job %s", newjob.Command)

		log.Printf("Received this job %s", newjob.Command)
		newjob.Status = models.WAIT

		// Make the status of job like wait
		select {
		case jobQueue <- &newjob:
		case <-ctx.Done():
			log.Println("read_socket: context cancelled while enqueuing, stopping")
			return
		}

	}
}
