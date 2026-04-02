package domain

import "time"

type ClientModel struct {
	ID             string     `db:"id"`
	Fingerprint    string     `db:"fingerprint"`
	HostName       string     `db:"hostname"`
	SessionToken   string     `db:"session_token"`
	TokenExpiresAt *time.Time `db:"token_expires_at"`
	CreatedAt      time.Time  `db:"created_at"`
	LastSeenAt     *time.Time `db:"last_seen_at"`
}

type ClientSummary struct {
	ID          string     `json:"id"`
	HostName    string     `json:"hostname"`
	Fingerprint string     `json:"fingerprint"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
	Online      bool       `json:"online"`
}

type ClientSession struct {
	ID             string     `db:"id"`
	ClientID       string     `db:"client_id"`
	ConnectedAt    time.Time  `db:"connected_at"`
	DisconnectedAt *time.Time `db:"disconnected_at"`
}
