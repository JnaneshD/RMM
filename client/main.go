package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
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
	go func() {
		<-interrupt
		cancel()
	}()

	scheduler := JobScheduler{conn: conn}

	for {
		select {
		case <-ctx.Done():
			log.Println("Closing the client as per request")
		default:
			// We will read the pipe that was just opened
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error", err)
				return
			}
			fmt.Println("Hmm here it is")
			log.Printf("Currently we have received this job %s", message)
			scheduler.AddJobImmediate(string(message))

		}
	}
}
