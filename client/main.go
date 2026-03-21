package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"example.com/test/models"
	"github.com/gorilla/websocket"
)

const (
	JobQueueSize = 100
	NumWorkers   = 4
)

func main() {
	// We should configure the log file also
	logfile, err := os.OpenFile("clientside.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error Opening log file")
	}

	log.SetOutput(logfile)

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

	scheduler := JobScheduler{conn: conn}

	// Start those workers here
	var wg sync.WaitGroup
	for i := range NumWorkers {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			Worker(ctx, workerId, jobQueue, &scheduler)
		}(i)
	}
	read_socket(ctx, jobQueue, conn)
	cancel()
	wg.Wait()
	log.Println("Shutting down")
}

func read_socket(ctx context.Context, jobQueue chan *models.Job, conn *websocket.Conn) {
	defer close(jobQueue)
	for {
		select {
		case <-ctx.Done():
			log.Println("read_socket: context cancelled, stopping the reader")
			return

		default:
			var newjob models.Job
			err := conn.ReadJSON(&newjob)
			if err != nil {
				log.Println("Read error from the server", err)
				return
			}
			log.Printf("Received this job %s", newjob.Command)

			select {
			case jobQueue <- &newjob:
			case <-ctx.Done():
				log.Println("read_socket: context cancelled while enqueuing, stopping")
				return
			}
		}

	}
}
