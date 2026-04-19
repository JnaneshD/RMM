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
	jobrepo    *repository.JobRepository
}

func NewHTTPHandler(dispatcher *service.Dispatcher, clientRepo *repository.ClientRepository,
	jobrepo *repository.JobRepository) *HTTPHandler {
	return &HTTPHandler{
		dispatcher: dispatcher,
		clientRepo: clientRepo,
		jobrepo:    jobrepo,
	}
}

// ReturnClients godoc
// @Summary List registered clients
// @Description Returns all known clients and marks whether each one is currently online.
// @Tags clients
// @Produce json
// @Success 200 {object} ClientsResponse
// @Failure 500 {object} ErrorResponse
// @Router /clients [get]
func (h *HTTPHandler) ReturnClients(ctx *gin.Context) {
	clients, err := h.clientRepo.ListClients(ctx.Request.Context())
	if err != nil {
		log.Printf("API Response error for list clients: %v", err)
		writeError(ctx, http.StatusInternalServerError, "failed to fetch clients")
		return
	}

	for i, cl := range clients {
		if h.dispatcher.IsClientExists(cl.ID) {
			clients[i].Online = true
		}
	}
	if clients == nil {
		clients = []domain.ClientSummary{}
	}
	ctx.JSON(http.StatusOK, ClientsResponse{Clients: clients})
}

func (h *HTTPHandler) ReturnClientDetails(ctx *gin.Context) {
	client_id := ctx.Param("id")
	client, err := h.clientRepo.GetClientDetails(ctx, client_id)
	if h.dispatcher.IsClientExists(client_id) {
		client.Online = true
	}
	if err != nil {
		log.Printf("API Error from DB : %v", err)
		writeError(ctx, http.StatusInternalServerError, "Failed to fetch client")
		return
	}

	ctx.JSON(http.StatusOK, client)
}

// ReturnJobs godoc
// @Summary List jobs grouped by client
// @Description Returns the current persisted job snapshot grouped by client ID.
// @Tags jobs
// @Produce json
// @Success 200 {object} JobsResponse
// @Router /jobs [get]
func (h *HTTPHandler) ReturnJobs(ctx *gin.Context) {
	client_id := ctx.Query("clientId")

	if client_id != "" {
		jobs, err := h.jobrepo.ListByClient(ctx, client_id)
		if err != nil {
			log.Printf("API Error from DB : %v", err)
			writeError(ctx, http.StatusInternalServerError, "Failed to fetch Jobs")
			return
		}
		ctx.JSON(http.StatusOK, jobs)
		return
	}
	jobs := h.dispatcher.JobsSnapshot()
	if jobs == nil {
		jobs = map[string][]domain.Job{}
	}
	ctx.JSON(http.StatusOK, JobsResponse{Jobs: jobs})
}

// HandlePushMessage godoc
// @Summary Dispatch a command to a client
// @Description Queues a command for the target online client and returns the created job.
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param payload body PushMessageRequest true "Command payload"
// @Success 200 {object} PushMessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 504 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /push/{id} [post]
func (h *HTTPHandler) HandlePushMessage(ctx *gin.Context) {
	id := ctx.Param("id")

	var payload PushMessageRequest
	if err := ctx.BindJSON(&payload); err != nil {
		writeError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.dispatcher.Dispatch(id, payload.Command, payload.ShellType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrClientNotFound):
			writeError(ctx, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrClientBusy):
			writeError(ctx, http.StatusGatewayTimeout, err.Error())
		default:
			writeError(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, PushMessageResponse{
		Status: "Sent to agent",
		Job:    job,
	})
}

// HandleRegistration godoc
// @Summary Register or refresh a client session
// @Description Validates a client registration request, persists the client, and returns a session token used for websocket authentication.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "Client registration payload"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (h *HTTPHandler) HandleRegistration(ctx *gin.Context) {
	var body RegisterRequest
	if err := ctx.ShouldBindJSON(&body); err != nil || body.UUID == "" {
		writeError(ctx, http.StatusBadRequest, "Invalid Payload")
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
		writeError(ctx, http.StatusUnauthorized, "invalid request")
		return
	}

	sessionToken, err := service.NewSessionToken()
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, "failed to create session token")
		return
	}

	client := &domain.Client{
		ID:             body.UUID,
		Fingerprint:    body.FingerPrint,
		HostName:       body.Hostname,
		SessionToken:   sessionToken,
		TokenExpiresAt: repository.SessionExpiry(24),
		OS:             body.OS,
	}
	if err := h.clientRepo.UpsertRegistration(ctx.Request.Context(), client); err != nil {
		writeError(ctx, http.StatusInternalServerError, "failed to persist client")
		return
	}

	ctx.JSON(http.StatusOK, RegisterResponse{
		SessionToken: sessionToken,
		WSURL:        "",
	})
}

// HandleDeleteJobsOfClient godoc
// @Summary Delete all jobs for a client
// @Description Removes every persisted job belonging to the supplied client ID.
// @Tags jobs
// @Produce json
// @Param clientId query string true "Client ID"
// @Success 200 {object} DeleteJobsResponse
// @Failure 400 {object} DeleteJobsResponse
// @Failure 500 {object} DeleteJobsResponse
// @Router /delete/jobs [delete]
func (h *HTTPHandler) HandleDeleteJobsOfClient(ctx *gin.Context) {
	clientId := ctx.Query("clientId")
	var resp DeleteJobsResponse
	if clientId == "" {
		resp.ErrorCode = "-1"
		resp.ErrorMsg = "Invalid Client Id"
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	err := h.jobrepo.DeleteAllJobsOfClient(ctx, clientId)
	if err != nil {
		resp.ErrorCode = "-2"
		resp.ErrorMsg = err.Error()
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp.ErrorCode = "0"
	resp.ErrorMsg = ""
	ctx.JSON(http.StatusOK, resp)
}
