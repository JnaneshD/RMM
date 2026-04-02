package api

import (
	"errors"
	"log"
	"net/http"

	"example.com/test/internal/domain"
	"example.com/test/internal/repository"
	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	dispatcher *service.Dispatcher
	clientRepo *repository.ClientRepository
}

func NewHTTPHandler(dispatcher *service.Dispatcher, clientRepo *repository.ClientRepository) *HTTPHandler {
	return &HTTPHandler{
		dispatcher: dispatcher,
		clientRepo: clientRepo,
	}
}

func (h *HTTPHandler) ReturnClients(ctx *gin.Context) {
	clients, err := h.clientRepo.ListClients(ctx.Request.Context())
	for i, cl := range clients {
		if h.dispatcher.IsClientExists(cl.ID) {
			clients[i].Online = true
		}
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch clients"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"clients": clients,
	})
}

func (h *HTTPHandler) ReturnJobs(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"jobs": h.dispatcher.JobsSnapshot(),
	})
}

func (h *HTTPHandler) HandlePushMessage(ctx *gin.Context) {
	id := ctx.Param("id")

	var payload struct {
		Command string `json:"command"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.dispatcher.Dispatch(id, payload.Command)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClientNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrClientBusy):
			ctx.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "Sent to agent",
		"job":    job,
	})
}

func (h *HTTPHandler) HandleRegistration(ctx *gin.Context) {
	var body struct {
		UUID        string `json:"uuid"`
		FingerPrint string `json:"fingerprint"`
		TimeStamp   string `json:"timestamp"`
		Signature   string `json:"signature"`
		Hostname    string `json:"hostname"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil || body.UUID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Payload",
		})
		return
	}

	// Check the timestamp
	// t, err := time.Parse(time.RFC3339, body.TimeStamp)
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
	// 	return
	// }
	// if time.Since(t) > 5*time.Minute {
	// 	log.Printf("[register] stale request from uuid=%s", body.UUID)
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "request expired"})
	// 	return
	// }

	// Now we will do the actual validation
	log.Printf("[register] this agent with uuid %s", body.UUID)
	if !service.ValidateAgentRegistration(body.UUID, body.FingerPrint, body.Signature) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid request",
		})
		return
	}

	sessionToken, err := service.NewSessionToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session token"})
		return
	}

	client := &domain.ClientModel{
		ID:             body.UUID,
		Fingerprint:    body.FingerPrint,
		HostName:       body.Hostname,
		SessionToken:   sessionToken,
		TokenExpiresAt: repository.SessionExpiry(24),
	}
	if err := h.clientRepo.UpsertRegistration(ctx.Request.Context(), client); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist client"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"session_token": sessionToken,
		"ws_url":        "",
	})
}
