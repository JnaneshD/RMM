package repository

import (
	"context"
	"time"

	"example.com/test/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, clientID string) (*domain.ClientSession, error) {
	session := &domain.ClientSession{
		ID:          uuid.NewString(),
		ClientID:    clientID,
		ConnectedAt: time.Now().UTC(),
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO agent_sessions (id, client_id, connected_at)
		 VALUES ($1, $2, $3)`,
		session.ID, session.ClientID, session.ConnectedAt,
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) MarkDisconnected(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE agent_sessions
		 SET disconnected_at = NOW()
		 WHERE id = $1`,
		sessionID,
	)
	return err
}
