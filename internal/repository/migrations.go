package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type migration struct {
	version    int
	name       string
	statements []string
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_clients",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS clients (
				id TEXT PRIMARY KEY,
				fingerprint TEXT NOT NULL,
				hostname TEXT NOT NULL DEFAULT '',
				session_token TEXT,
				token_expires_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				last_seen_at TIMESTAMPTZ,
				os TEXT NOT NULL DEFAULT ''
			)`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS fingerprint TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS hostname TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS session_token TEXT`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMPTZ`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ`,
			`ALTER TABLE clients ADD COLUMN IF NOT EXISTS os TEXT NOT NULL DEFAULT ''`,
			`CREATE INDEX IF NOT EXISTS idx_clients_last_seen_at ON clients(last_seen_at DESC)`,
			`CREATE INDEX IF NOT EXISTS idx_clients_token_expires_at ON clients(token_expires_at)`,
		},
	},
	{
		version: 2,
		name:    "create_jobs",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS jobs (
				id BIGSERIAL PRIMARY KEY,
				client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
				command TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				shell_type TEXT NOT NULL DEFAULT '',
				job_dir TEXT NOT NULL DEFAULT '',
				result TEXT NOT NULL DEFAULT '',
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				CONSTRAINT chk_jobs_status CHECK (status IN ('pending', 'running', 'finished', 'failed', 'succeeded'))
			)`,
			`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS shell_type TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS job_dir TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS result TEXT NOT NULL DEFAULT ''`,
			`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
			`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
			`CREATE INDEX IF NOT EXISTS idx_jobs_client_id ON jobs(client_id)`,
			`CREATE INDEX IF NOT EXISTS idx_jobs_client_status ON jobs(client_id, status)`,
			`CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC)`,
		},
	},
	{
		version: 3,
		name:    "create_agent_sessions",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS agent_sessions (
				id TEXT PRIMARY KEY,
				client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
				connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				disconnected_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
			`ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
			`ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS disconnected_at TIMESTAMPTZ`,
			`ALTER TABLE agent_sessions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
			`CREATE INDEX IF NOT EXISTS idx_agent_sessions_client_connected ON agent_sessions(client_id, connected_at DESC)`,
		},
	},
}

func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	if db == nil {
		return fmt.Errorf("database pool is nil")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(912754392)`); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}

	applied, err := appliedMigrations(ctx, tx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.version] {
			continue
		}

		for _, statement := range migration.statements {
			if _, err := tx.Exec(ctx, statement); err != nil {
				return fmt.Errorf("apply migration %03d_%s: %w", migration.version, migration.name, err)
			}
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`,
			migration.version,
			migration.name,
		); err != nil {
			return fmt.Errorf("record migration %03d_%s: %w", migration.version, migration.name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration transaction: %w", err)
	}
	return nil
}

func appliedMigrations(ctx context.Context, tx pgx.Tx) (map[int]bool, error) {
	rows, err := tx.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}

	return applied, nil
}
