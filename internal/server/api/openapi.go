package api

import "example.com/test/internal/domain"

type ErrorResponse struct {
	Error string `json:"error"`
}

type PushMessageRequest struct {
	Command string `json:"command" binding:"required"`
}

type PushMessageResponse struct {
	Status string     `json:"status"`
	Job    domain.Job `json:"job"`
}

type ClientsResponse struct {
	Clients []domain.ClientSummary `json:"clients"`
}

type JobsResponse struct {
	Jobs map[string][]domain.Job `json:"jobs"`
}

type RegisterRequest struct {
	UUID        string `json:"uuid" binding:"required"`
	FingerPrint string `json:"fingerprint" binding:"required"`
	TimeStamp   string `json:"timestamp"`
	Signature   string `json:"signature" binding:"required"`
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
	WSURL        string `json:"ws_url"`
}
