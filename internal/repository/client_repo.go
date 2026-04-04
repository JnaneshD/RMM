package repository

import (
	"context"
	"errors"
	"time"

	"example.com/test/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository struct {
	db *pgxpool.Pool
}

func NewClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{
		db: db,
	}
}

func (r *ClientRepository) UpsertRegistration(ctx context.Context, c *domain.ClientModel) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO clients (id, fingerprint, hostname, session_token, token_expires_at, last_seen_at, os)
		 VALUES ($1, $2, $3, $4, $5, NOW(), $6)
		 ON CONFLICT (id)
		 DO UPDATE SET
		     fingerprint = EXCLUDED.fingerprint,
		     hostname = EXCLUDED.hostname,
		     session_token = EXCLUDED.session_token,
		     token_expires_at = EXCLUDED.token_expires_at,
		     last_seen_at = NOW(),
			 os = EXCLUDED.os`,
		c.ID, c.Fingerprint, c.HostName, c.SessionToken, c.TokenExpiresAt, c.OS,
	)
	return err
}

func (r *ClientRepository) AuthenticateSession(ctx context.Context, clientID, token string) (*domain.ClientModel, error) {
	var client domain.ClientModel
	err := r.db.QueryRow(ctx,
		`SELECT id, fingerprint, hostname, session_token, token_expires_at, created_at, last_seen_at
		 FROM clients
		 WHERE id = $1
		   AND session_token = $2
		   AND token_expires_at > NOW()`,
		clientID, token,
	).Scan(
		&client.ID,
		&client.Fingerprint,
		&client.HostName,
		&client.SessionToken,
		&client.TokenExpiresAt,
		&client.CreatedAt,
		&client.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &client, nil
}

func (r *ClientRepository) TouchLastSeen(ctx context.Context, clientID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE clients
		 SET last_seen_at = NOW()
		 WHERE id = $1`,
		clientID,
	)
	return err
}

func (r *ClientRepository) RevokeSession(ctx context.Context, clientID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE clients
		 SET session_token = NULL,
		     token_expires_at = NULL
		 WHERE id = $1`,
		clientID,
	)
	return err
}

func (r *ClientRepository) ListClients(ctx context.Context) ([]domain.ClientSummary, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, hostname, created_at, last_seen_at, os
		 FROM clients
		 ORDER BY created_at DESC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []domain.ClientSummary
	for rows.Next() {
		var client domain.ClientSummary
		client.Online = false

		if err := rows.Scan(
			&client.ID,
			&client.HostName,
			&client.CreatedAt,
			&client.LastSeenAt,
			&client.OS,
		); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

func SessionExpiry(hours int) *time.Time {
	expiresAt := time.Now().UTC().Add(time.Duration(hours) * time.Hour)
	return &expiresAt
}
