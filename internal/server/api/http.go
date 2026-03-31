package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
)

const AgentSecret = "replace-with-a-long-random-secret-string"

type HTTPHandler struct {
	dispatcher *service.Dispatcher
}

func NewHTTPHandler(dispatcher *service.Dispatcher) *HTTPHandler {
	return &HTTPHandler{dispatcher: dispatcher}
}

func (h *HTTPHandler) ReturnClients(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"clients": h.dispatcher.ClientIDs(),
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
	t, err := time.Parse(time.RFC3339, body.TimeStamp)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
		return
	}
	if time.Since(t) > 5*time.Minute {
		log.Printf("[register] stale request from uuid=%s", body.UUID)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "request expired"})
		return
	}

	// Now we will do the actual validation
	log.Printf("[register] this agent with uuid %s", body.UUID)
	valid_req := HandleAgentAuth(body.UUID, body.FingerPrint, body.TimeStamp, body.Signature)
	if !valid_req {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid request",
		})
		return
	}
	client := realtime.NewClient(body.UUID, body.FingerPrint, body.Hostname)
	h.dispatcher.RegisterClient(client)

	ctx.JSON(http.StatusOK, gin.H{
		"session_token": body.UUID,
		"ws_url":        "",
	})
}

func HandleAgentAuth(uuid string, fingerprint string, timestamp string, signature string) bool {
	// Lets create the same mac and validate that it came from our same binary
	mac := hmac.New(sha256.New, []byte(AgentSecret))
	mac.Write([]byte(uuid + "|" + fingerprint + "|" + timestamp))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return false
	}
	return true
}
