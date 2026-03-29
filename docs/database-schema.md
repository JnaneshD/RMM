# Database Schema Design

## Context

This project is a remote job execution system with:

- a server that accepts agent connections over WebSocket
- agents/clients that register with an `id`
- jobs pushed to a specific agent
- job execution status and output returned back to the server

Right now, most state is in memory inside the WebSocket hub. That is fine for early prototyping, but it means:

- all jobs disappear on process restart
- online/offline state is not historically traceable
- job retries and execution history are not durable
- multiple server instances will be hard to coordinate

## Recommended Database Stack

### Primary database

Use **PostgreSQL** as the system of record.

Why PostgreSQL fits this project well:

- strong support for relational data and constraints
- excellent indexing and transactional guarantees
- good support for enums, JSONB, partial indexes, and auditing patterns
- easy to evolve with migrations
- good fit for jobs, agents, attempts, schedules, and event history

### Optional supporting infrastructure

- **Redis**: optional for ephemeral queueing, distributed locks, and presence caching
- **Object storage**: optional if command output/logs become too large for a table row

For this codebase, PostgreSQL alone is enough to start.

## What "Schema" Means

Schema does **not** only mean `CREATE DATABASE` and `CREATE TABLE`.

In a real app team, schema usually includes:

- tables
- columns and data types
- primary keys and foreign keys
- unique constraints
- check constraints
- indexes
- join tables
- enums or lookup tables
- views and materialized views when needed
- partitioning strategy for large tables
- migration files
- seed/reference data
- naming conventions
- retention and archival rules

Think of schema as the **formal structure of persisted data** and the rules around it.

## Domain Model

The current project naturally maps to these core entities:

- `agents`: remote clients that connect to the server
- `agent_sessions`: each connection lifecycle
- `jobs`: requested units of work
- `job_attempts`: each execution attempt of a job
- `job_events`: append-only audit trail of state changes
- `job_schedules`: optional future recurring jobs
- `job_tags`: optional labels for filtering/grouping

## Long-Lived Design Principles

- Keep **identity tables** separate from **activity/history tables**
- Keep **current state** and **event history** both available
- Avoid storing many comma-separated values in one column
- Model retries as separate rows in an attempts table
- Prefer append-only event/history tables for debugging
- Use foreign keys for relationships
- Use timestamps everywhere for lifecycle tracking
- Use soft deletes only where business recovery matters
- Keep output payloads separate if they may grow large

## Suggested Tables

### 1. `agents`

Represents a logical worker/agent machine or process.

```sql
CREATE TABLE agents (
    id              UUID PRIMARY KEY,
    external_key    TEXT NOT NULL UNIQUE,
    name            TEXT,
    platform        TEXT,
    hostname        TEXT,
    version         TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Notes:

- `external_key` can map to your current `agent1` style identifier
- `id` should be an internal stable UUID even if user-facing IDs change
- `last_seen_at` gives fast access to current freshness

### 2. `agent_sessions`

Represents each connect/disconnect window.

```sql
CREATE TABLE agent_sessions (
    id                  UUID PRIMARY KEY,
    agent_id            UUID NOT NULL REFERENCES agents(id),
    connected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at     TIMESTAMPTZ,
    disconnect_reason   TEXT,
    server_instance_id  TEXT,
    remote_addr         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_sessions_agent_connected
    ON agent_sessions(agent_id, connected_at DESC);
```

Why it matters:

- you can answer "when was this agent online?"
- helps debug flaky connections
- supports future multi-server deployments

### 3. `jobs`

Represents the logical job request.

```sql
CREATE TYPE job_status AS ENUM (
    'queued',
    'dispatched',
    'running',
    'succeeded',
    'failed',
    'cancelled',
    'timed_out'
);

CREATE TABLE jobs (
    id                  BIGSERIAL PRIMARY KEY,
    public_id           UUID NOT NULL UNIQUE,
    agent_id            UUID NOT NULL REFERENCES agents(id),
    command             TEXT NOT NULL,
    status              job_status NOT NULL DEFAULT 'queued',
    priority            SMALLINT NOT NULL DEFAULT 100,
    requested_by        TEXT,
    idempotency_key     TEXT,
    queued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at       TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    current_attempt_no  INTEGER NOT NULL DEFAULT 0,
    max_attempts        INTEGER NOT NULL DEFAULT 1,
    timeout_seconds     INTEGER,
    latest_output       TEXT,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_jobs_attempts CHECK (max_attempts >= 1),
    CONSTRAINT chk_jobs_current_attempt_no CHECK (current_attempt_no >= 0)
);

CREATE INDEX idx_jobs_agent_status ON jobs(agent_id, status);
CREATE INDEX idx_jobs_status_queued_at ON jobs(status, queued_at);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
CREATE UNIQUE INDEX idx_jobs_idempotency_key
    ON jobs(idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

Why `jobs` is separate from attempts:

- one job can retry multiple times
- "logical job" and "individual run" are not the same thing

### 4. `job_attempts`

Represents each execution attempt for a job.

```sql
CREATE TYPE attempt_status AS ENUM (
    'queued',
    'running',
    'succeeded',
    'failed',
    'cancelled',
    'timed_out'
);

CREATE TABLE job_attempts (
    id                  BIGSERIAL PRIMARY KEY,
    job_id              BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    attempt_no          INTEGER NOT NULL,
    agent_id            UUID NOT NULL REFERENCES agents(id),
    agent_session_id    UUID REFERENCES agent_sessions(id),
    status              attempt_status NOT NULL DEFAULT 'queued',
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    exit_code           INTEGER,
    stdout_text         TEXT,
    stderr_text         TEXT,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_job_attempt UNIQUE (job_id, attempt_no)
);

CREATE INDEX idx_job_attempts_job_id ON job_attempts(job_id);
CREATE INDEX idx_job_attempts_agent_id ON job_attempts(agent_id);
```

This is the most important normalization step for long-term durability.

### 5. `job_events`

Append-only audit log for every meaningful transition.

```sql
CREATE TABLE job_events (
    id              BIGSERIAL PRIMARY KEY,
    job_id          BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    job_attempt_id  BIGINT REFERENCES job_attempts(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL,
    event_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_type      TEXT,
    actor_id        TEXT,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_job_events_job_id_event_at
    ON job_events(job_id, event_at DESC);
```

Example events:

- `job_queued`
- `job_dispatched`
- `attempt_started`
- `attempt_finished`
- `agent_disconnected`
- `job_marked_failed`

This is invaluable for debugging.

### 6. `agent_heartbeats`

Use only if you want full heartbeat history. If you just need the latest value, `agents.last_seen_at` is enough.

```sql
CREATE TABLE agent_heartbeats (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    session_id      UUID REFERENCES agent_sessions(id) ON DELETE SET NULL,
    heartbeat_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status_payload  JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_agent_heartbeats_agent_time
    ON agent_heartbeats(agent_id, heartbeat_at DESC);
```

### 7. `job_schedules`

Useful if you later add recurring or delayed jobs.

```sql
CREATE TABLE job_schedules (
    id                  UUID PRIMARY KEY,
    agent_id            UUID NOT NULL REFERENCES agents(id),
    name                TEXT NOT NULL,
    command_template    TEXT NOT NULL,
    cron_expression     TEXT NOT NULL,
    timezone_name       TEXT NOT NULL DEFAULT 'UTC',
    is_enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at         TIMESTAMPTZ,
    last_run_at         TIMESTAMPTZ,
    max_attempts        INTEGER NOT NULL DEFAULT 1,
    timeout_seconds     INTEGER,
    metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 8. `tags` and `job_tags`

Optional, but useful for grouping jobs by environment, owner, or purpose.

```sql
CREATE TABLE tags (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE job_tags (
    job_id       BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    tag_id       BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (job_id, tag_id)
);
```

This is a classic many-to-many join table.

## Relationship Summary

- one `agent` has many `agent_sessions`
- one `agent` has many `jobs`
- one `job` has many `job_attempts`
- one `job` has many `job_events`
- one `job_attempt` may belong to one `agent_session`
- one `job` has many `job_tags`
- one `tag` has many `job_tags`

## Which Tables Are Core vs Optional

### Core now

- `agents`
- `jobs`
- `job_attempts`
- `job_events`
- `agent_sessions`

### Add later if needed

- `agent_heartbeats`
- `job_schedules`
- `tags`
- `job_tags`

## Why This Will Last Longer Than a Single `jobs` Table

A naive design is often:

```sql
CREATE TABLE jobs (
    id BIGSERIAL PRIMARY KEY,
    client_id TEXT,
    command TEXT,
    status TEXT,
    output TEXT
);
```

That works at first, but it breaks down when you need:

- retries
- multiple executions
- event history
- agent reconnect history
- metrics
- queue priority
- scheduling
- debugging of partial failures

The schema above avoids that trap.

## Example Query Patterns

### Find online agents

```sql
SELECT id, external_key, name, last_seen_at
FROM agents
WHERE is_active = TRUE
  AND last_seen_at > NOW() - INTERVAL '30 seconds';
```

### List latest jobs for an agent

```sql
SELECT j.id, j.public_id, j.command, j.status, j.queued_at, j.finished_at
FROM jobs j
JOIN agents a ON a.id = j.agent_id
WHERE a.external_key = 'agent1'
ORDER BY j.created_at DESC
LIMIT 50;
```

### Show execution history for one job

```sql
SELECT ja.attempt_no, ja.status, ja.started_at, ja.finished_at, ja.exit_code
FROM job_attempts ja
WHERE ja.job_id = $1
ORDER BY ja.attempt_no ASC;
```

### Show audit trail for one job

```sql
SELECT event_type, event_at, payload
FROM job_events
WHERE job_id = $1
ORDER BY event_at ASC;
```

## Team Practices That Make a Schema Last

### 1. Treat schema as code

Store schema changes in migrations inside the repo.

Example folder:

```text
db/
  migrations/
    0001_create_agents.sql
    0002_create_jobs.sql
    0003_create_job_attempts.sql
```

Do not manually edit production databases without a migration.

### 2. Use forward-only migrations

Each change should be:

- reviewed in PRs
- tested in staging
- safe to run exactly once

### 3. Separate application models from database truth

Go structs are not the full schema. The database is the source of truth for:

- constraints
- foreign keys
- indexes
- data integrity

ORM models help, but they are not enough by themselves.

### 4. Add constraints early

A durable schema relies on the database preventing bad states.

Examples:

- `NOT NULL`
- `CHECK (max_attempts >= 1)`
- `UNIQUE`
- `FOREIGN KEY`

### 5. Design for query patterns

Ask before creating a table:

- what are the most common reads?
- what needs to be filtered?
- what needs to be sorted?
- what must be unique?

Then add indexes intentionally.

### 6. Prefer append-only history for operations systems

Mutable current state is useful, but history is what helps you recover incidents.

For this app:

- `jobs` = current summary
- `job_attempts` = execution history
- `job_events` = audit trail

### 7. Avoid premature over-normalization

Normalize where relationships matter. Do not split every text field into its own table just because you can.

### 8. Plan for growth points

These usually appear later:

- large job outputs
- retry policies
- scheduled jobs
- multi-server coordination
- tenancy and auth

Leave room in the schema with:

- UUID public IDs
- JSONB metadata for non-critical extensibility
- dedicated history tables

## Suggested Naming Conventions

- table names: plural snake_case
- primary key: `id`
- foreign key: `<parent>_id`
- timestamps: `created_at`, `updated_at`, `started_at`, `finished_at`
- booleans: `is_*`, `has_*`
- status columns: enum or constrained text

Consistency matters more than the exact convention.

## Minimal First Version

If you want a smaller first production schema, start with:

- `agents`
- `agent_sessions`
- `jobs`
- `job_attempts`
- `job_events`

That gives you a strong base without overbuilding.

## Mapping From Current Code

Current code concepts:

- `models.Client` -> `agents`
- in-memory websocket connect/disconnect -> `agent_sessions`
- `models.Job` -> `jobs`
- current `Status` and `Output` fields -> summary fields on `jobs`
- each execution/result update -> `job_attempts` and `job_events`

## Final Answer

For this project, the schema should be much more than a few `CREATE TABLE` statements. A good long-lived schema is the combination of table design, joins, constraints, indexes, migrations, and operational rules for how the team evolves data safely.

If you implement persistence next, PostgreSQL is the right default choice, and the most important design move is to split:

- agent identity from agent connection sessions
- logical jobs from execution attempts
- current job state from append-only event history

That structure will hold up much longer than a single in-memory jobs map or a single flat `jobs` table.
