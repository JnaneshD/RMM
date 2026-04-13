package api

import "example.com/test/internal/domain"

type ErrorResponse struct {
	Error string `json:"error"`
}

type ClientsResponse struct {
	Clients []domain.ClientSummary `json:"clients"`
}

type JobsResponse struct {
	Jobs map[string][]domain.Job `json:"jobs"`
}

type PushMessageResponse struct {
	Status string     `json:"status"`
	Job    domain.Job `json:"job"`
}

type RegisterResponse struct {
	SessionToken string `json:"session_token"`
	WSURL        string `json:"ws_url"`
}

type DeleteJobsResponse struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

type JobsByClientResponse struct {
	Jobs []domain.Job
}
