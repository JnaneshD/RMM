package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Fixed UUID for now — in production read from a file,
	// generate once on first run and persist it.
	agentUUID := "agent-dev-001"

	// --- Step 1: Register with server over HTTPS (cert pinned) ---
	httpClient := runtime.BuildHTTPClient()
	token, err := runtime.Register(httpClient, agentUUID)
	if err != nil {
		log.Fatalf("[agent] registration failed: %v", err)
	}
	log.Printf("[agent] registered successfully, token=%s", token)

	// --- Step 2: Connect WebSocket over WSS (cert pinned) ---
	wsDialer := runtime.BuildWSDialer()
	conn, err := runtime.ConnectWS(wsDialer, token)
	if err != nil {
		log.Fatalf("[agent] websocket connection failed: %v", err)
	}
	defer conn.Close()

	// --- Step 3: Your existing worker pool + scheduler (unchanged) ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	go func() {
		sig := <-interrupt
		log.Printf("[agent] received signal: %v — shutting down cleanly", sig)

		// 1. Cancel context — stops workers and ReadSocket
		cancel()

		// 2. Send proper WS close frame — server sees a clean disconnect
		closeWS(conn)

		// 3. Close the connection
		conn.Close()
	}()

	jobQueue := make(chan *domain.Job, jobQueueSize)
	scheduler := runtime.NewJobScheduler(conn, runtime.NewExecutor())

	// Graceful shutdown on Ctrl+C
	go func() {
		<-interrupt
		log.Println("[agent] interrupt received, shutting down...")
		cancel()
	}()

	// WritePump sends results back to server
	go scheduler.WritePump()

	// Worker pool — processes jobs from the queue
	var wg sync.WaitGroup
	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runtime.Worker(ctx, id, jobQueue, scheduler)
		}(workerID)
	}

	// ReadSocket blocks — reads jobs from server, pushes to queue
	// Returns when ctx is cancelled or connection drops
	runtime.ReadSocket(ctx, jobQueue, conn)

	cancel()
	wg.Wait()
	log.Println("[agent] shutdown complete")
}

func closeWS(conn *websocket.Conn) {
	log.Println("[ws] sending close frame to server...")

	// WriteMessage with CloseMessage sends the WS close frame
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "agent shutting down")
	err := conn.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		log.Printf("[ws] close frame error (ok if server already closed): %v", err)
		return
	}

	// Give the server up to 3 seconds to acknowledge the close
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// This is expected — either server sent its close frame back
			// or the deadline hit. Either way we're done.
			log.Println("[ws] connection closed cleanly")
			return
		}
	}
}
