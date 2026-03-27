package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"example.com/test/internal/agent/runtime"
	"example.com/test/internal/domain"
	"github.com/gorilla/websocket"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	jobQueueSize = 100
	numWorkers   = 4
)

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "clientSide.log",
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	url := "ws://localhost:8000/ws/agent1"
	log.Printf("Connecting to the URL : %s", url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer conn.Close()

	log.Println("Connected to the server! Waiting for the jobs")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	jobQueue := make(chan *domain.Job, jobQueueSize)
	scheduler := runtime.NewJobScheduler(conn, runtime.NewExecutor())

	go func() {
		<-interrupt
		cancel()
	}()

	go scheduler.WritePump()

	var wg sync.WaitGroup
	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.Worker(ctx, id, jobQueue, scheduler)
		}(workerID)
	}

	runtime.ReadSocket(ctx, jobQueue, conn)
	cancel()
	wg.Wait()
	log.Println("Shutting down")
}
