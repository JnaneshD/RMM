package service

import (
	"context"
	"errors"
	"log"

	"example.com/test/internal/domain"
	"example.com/test/internal/repository"
	"example.com/test/internal/server/realtime"
)

var (
	ErrClientNotFound = errors.New("agent is down")
	ErrClientBusy     = errors.New("agent channel full")
)

type Dispatcher struct {
	hub     *realtime.Hub
	jobRepo *repository.JobRepository
}

func NewDispatcher(hub *realtime.Hub,
	jobrepo *repository.JobRepository) *Dispatcher {
	return &Dispatcher{
		hub:     hub,
		jobRepo: jobrepo,
	}
}

func (d *Dispatcher) RegisterClient(client *realtime.ActiveClient) {
	d.hub.Register(client)
}

func (d *Dispatcher) UnregisterClient(client *realtime.ActiveClient) {
	d.hub.Unregister(client)
}

func (d *Dispatcher) IsClientExists(clientId string) bool {
	_, exists := d.hub.GetClient(clientId)
	return exists

}

func (d *Dispatcher) JobsSnapshot() map[string][]domain.Job {
	if d.jobRepo == nil {
		return map[string][]domain.Job{}
	}

	jobs, err := d.jobRepo.ListAll(context.Background())
	if err != nil {
		log.Printf("list jobs: %v", err)
		return map[string][]domain.Job{}
	}
	return jobs
}

func (d *Dispatcher) Dispatch(clientID, command string) (domain.Job, error) {
	client, exists := d.hub.GetClient(clientID)
	if !exists {
		return domain.Job{}, ErrClientNotFound
	}

	var job domain.Job
	if d.jobRepo != nil {
		var err error
		job, err = d.jobRepo.Create(context.Background(), clientID, command)
		if err != nil {
			return domain.Job{}, err
		}
	} else {
		job = domain.Job{
			ClientID: clientID,
			Command:  command,
			Status:   domain.WAIT,
		}
	}
	select {
	case client.Send <- job:
		return job, nil
	default:
		return domain.Job{}, ErrClientBusy
	}
}

func (d *Dispatcher) RecordJobUpdate(job domain.Job) bool {
	if d.jobRepo != nil {
		ok, err := d.jobRepo.UpdateStatus(context.Background(), job.ID, job.Status.String(), job.Output)
		if err != nil {
			log.Printf("update job %d: %v", job.ID, err)
			return false
		}
		return ok
	}
	return false
}
