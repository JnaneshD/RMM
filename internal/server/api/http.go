package api

import (
	"errors"
	"net/http"

	"example.com/test/internal/server/service"
	"github.com/gin-gonic/gin"
)

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
