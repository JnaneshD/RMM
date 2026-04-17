package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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

func RegisterClient(serverURL string, agentUUID string) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("get hostname: %v", err)
	}

	hardwareFingerPrint, err := runtime.HardwareFingerprint()
	if err != nil {
		log.Fatalf("[agent] Hardware fingerprint is not working %v", err)
	}

	// --- Step 1: Register with server over HTTPS (cert pinned) ---
	httpClient := runtime.BuildHTTPClient()
	token, err := runtime.Register(serverURL, httpClient, agentUUID, hardwareFingerPrint, hostname)
	if err != nil {
		return "", fmt.Errorf("[agent] registration failed: %v", err)
	}
	log.Printf("[agent] registered successfully, token=%s", token)
	return token, nil
}

func runAgent(serverURL string, token string, agentUUID string) error {
	// generate once on first run and persist it.

	// --- Step 2: Connect WebSocket over WSS (cert pinned) ---
	wsDialer := runtime.BuildWSDialer()
	conn, err := runtime.ConnectWS(serverURL, wsDialer, token, agentUUID)
	if err != nil {
		return fmt.Errorf("[agent] websocket connection failed: %w", err)
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
	return fmt.Errorf("[agent] shutdown complete")
}

func main() {
	serverFlag := flag.String("server", "", "Server URL (e.g. https://localhost:8081)")
	flag.Parse()

	if *serverFlag != "" {
		if err := runtime.SaveServerURL(*serverFlag); err != nil {
			log.Fatalf("[agent] failed to save server URL: %v", err)
		}
		log.Printf("[agent] saved server URL: %s", *serverFlag)
	}

	serverURL, err := runtime.LoadServerURL()
	if err != nil {
		log.Fatalf("[agent] MISSING SERVER URL! Please start the agent with: -server <URL>")
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   "clientSide.log",
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	backoff := time.Second * 2
	maxBackoff := time.Minute * 5

	agentUUID, err := runtime.AgentUUID()
	if err != nil {
		log.Fatalf("[agent] uuid is not correct")
	}

	for {
		// 1. Try loading existing token, otherwise register
		token, err := runtime.LoadAgentToken()
		if err != nil {
			log.Printf("[agent] no token found, attempting registration...")
			token, err = RegisterClient(serverURL, agentUUID)
			if err != nil {
				log.Printf("[agent] registration failed: %v", err)
				log.Printf("[agent] retrying in %v...", backoff)
				time.Sleep(backoff)
				
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			runtime.SaveAgentToken(token)
			// Reset backoff on successful registration
			backoff = time.Second * 2
		}

		// 2. We have a token, start the agent
		err = runAgent(serverURL, token, agentUUID)
		log.Printf("[agent] run ended: %v", err)

		// 3. If unauthorized, clear token and retry without sleeping
		if errors.Is(err, runtime.ErrUnauthorized) {
			log.Printf("[agent] token unauthorized or expired. Clearing token...")
			runtime.ClearAgentToken()
			backoff = time.Second * 2
			continue
		}

		// 4. Sleep and retry connecting
		log.Printf("[agent] reconnecting in %v...", backoff)
		time.Sleep(backoff)

		// exponential backoff
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
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
