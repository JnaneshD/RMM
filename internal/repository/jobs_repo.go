package repository

import (
	"context"
	"sort"

	"example.com/test/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, clientID, command string) (domain.Job, error) {
	job := domain.Job{
		ClientID: clientID,
		Command:  command,
		Status:   domain.WAIT,
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO jobs (client_id, command, status, result)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		job.ClientID, job.Command, job.Status.String(), job.Output,
	).Scan(&job.ID)
	if err != nil {
		return domain.Job{}, err
	}

	return job, nil
}

func (r *JobRepository) UpdateStatus(ctx context.Context, jobID uint64, status, result string) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE jobs
		 SET status=$1, result=$2
		 WHERE id=$3`,
		status, result, jobID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *JobRepository) GetPendingByClient(ctx context.Context, clientID string) ([]domain.Job, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, command, result
		 FROM jobs
		 WHERE client_id=$1 AND status='pending'`,
		clientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(&j.ID, &j.ClientID, &j.Command, &j.Output); err != nil {
			return nil, err
		}
		j.Status = domain.WAIT
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *JobRepository) ListAll(ctx context.Context) (map[string][]domain.Job, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, command, status, COALESCE(result, '')
		 FROM jobs
		 ORDER BY client_id ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobsByClient := make(map[string][]domain.Job)
	for rows.Next() {
		var (
			job       domain.Job
			statusStr string
		)
		if err := rows.Scan(&job.ID, &job.ClientID, &job.Command, &statusStr, &job.Output); err != nil {
			return nil, err
		}
		job.Status = parseJobStatus(statusStr)
		jobsByClient[job.ClientID] = append(jobsByClient[job.ClientID], job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for clientID := range jobsByClient {
		sort.Slice(jobsByClient[clientID], func(i, j int) bool {
			return jobsByClient[clientID][i].ID < jobsByClient[clientID][j].ID
		})
	}

	return jobsByClient, nil
}

func parseJobStatus(status string) domain.JobStatus {
	switch status {
	case "pending":
		return domain.WAIT
	case "running":
		return domain.RUNNING
	case "finished", "succeeded":
		return domain.FINISHED
	case "failed":
		return domain.FAILED
	default:
		return domain.WAIT
	}
}
