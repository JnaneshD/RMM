package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"example.com/test/internal/domain"
	"example.com/test/internal/heartbeat"
	"example.com/test/internal/repository"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SocketHandler struct {
	dispatcher  *service.Dispatcher
	clientRepo  *repository.ClientRepository
	sessionRepo *repository.SessionRepository
	logger      *log.Logger
}

func NewSocketHandler(
	dispatcher *service.Dispatcher,
	clientRepo *repository.ClientRepository,
	sessionRepo *repository.SessionRepository,
	logger *log.Logger,
) *SocketHandler {
	return &SocketHandler{
		dispatcher:  dispatcher,
		clientRepo:  clientRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (h *SocketHandler) HandleServerSideSocket(ctx *gin.Context) {

	clientID := ctx.Param("id")
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Token",
		})
		return
	}

	authenticatedClient, err := h.clientRepo.AuthenticateSession(ctx.Request.Context(), clientID, token)
	if err != nil {
		h.logger.Printf("websocket auth lookup failed for client %s: %v", clientID, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}
	if authenticatedClient == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		h.logger.Printf("websocket upgrade failed: %v", err)
		return
	}
	session, err := h.sessionRepo.Create(context.Background(), clientID)
	if err != nil {
		h.logger.Printf("create session for client %s: %v", clientID, err)
		conn.Close()
		return
	}
	client := realtime.NewClient(authenticatedClient.ID, authenticatedClient.Fingerprint, authenticatedClient.HostName)
	client.UpdateClient(conn)
	h.dispatcher.RegisterClient(client)
	if err := h.clientRepo.TouchLastSeen(context.Background(), clientID); err != nil {
		h.logger.Printf("touch last_seen for client %s: %v", clientID, err)
	}

	done := make(chan bool, 1)

	defer func() {
		h.dispatcher.UnregisterClient(client)
		if err := h.sessionRepo.MarkDisconnected(context.Background(), session.ID); err != nil {
			h.logger.Printf("mark disconnected for client %s session %s: %v", clientID, session.ID, err)
		}
		if err := h.clientRepo.TouchLastSeen(context.Background(), clientID); err != nil {
			h.logger.Printf("touch last_seen on disconnect for client %s: %v", clientID, err)
		}
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(heartbeat.PongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(heartbeat.PongWait))
		return nil
	})

	go h.sendCommandsToClient(client, conn, done)
	go h.receiveJobOutputFromClient(conn, done, clientID)
	<-done
}

func (h *SocketHandler) sendCommandsToClient(client *realtime.Client, conn *websocket.Conn, done chan bool) {
	ticker := time.NewTicker(heartbeat.PingInterval)
	defer ticker.Stop()
	defer h.logger.Println("Stopping writes")

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(
				websocket.PingMessage, nil, time.Now().Add(heartbeat.WriteWait),
			); err != nil {
				select {
				case done <- true:
				default:
				}
				return
			}
		case message, ok := <-client.Send:
			if !ok {
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				h.logger.Printf("failed to write job %d to client %s: %v", message.ID, client.ID, err)
				select {
				case done <- true:
				default:
				}
				return
			}
		case <-done:
			return
		}
	}
}

func (h *SocketHandler) receiveJobOutputFromClient(conn *websocket.Conn, done chan bool, clientID string) {
	for {
		var job domain.Job
		err := conn.ReadJSON(&job)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				h.logger.Printf("unexpected close from agent %s: %v", clientID, err)
			} else {
				h.logger.Printf("agent %s disconnected: %v", clientID, err)
			}
			select {
			case done <- true:
			default:
			}
			return
		}

		job.ClientID = clientID
		if !h.dispatcher.RecordJobUpdate(job) {
			h.logger.Printf("received update for unknown job %d from agent %s", job.ID, clientID)
		}
	}
}
