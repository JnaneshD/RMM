package service

import (
	"errors"
	"sort"

	"example.com/test/internal/domain"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/store/memory"
)

var (
	ErrClientNotFound = errors.New("agent is down")
	ErrClientBusy     = errors.New("agent channel full")
)

type Dispatcher struct {
	hub   *realtime.Hub
	store *memory.JobStore
}

func NewDispatcher(hub *realtime.Hub, store *memory.JobStore) *Dispatcher {
	return &Dispatcher{
		hub:   hub,
		store: store,
	}
}

func (d *Dispatcher) RegisterClient(client *realtime.Client) {
	d.hub.Register(client)
}

func (d *Dispatcher) UnregisterClient(client *realtime.Client) {
	d.hub.Unregister(client)
}

func (d *Dispatcher) ClientIDs() []string {
	ids := d.hub.ClientIDs()
	sort.Strings(ids)
	return ids
}

func (d *Dispatcher) JobsSnapshot() map[string][]domain.Job {
	return d.store.Snapshot()
}

func (d *Dispatcher) Dispatch(clientID, command string) (domain.Job, error) {
	client, exists := d.hub.GetClient(clientID)
	if !exists {
		return domain.Job{}, ErrClientNotFound
	}

	job := d.store.Create(clientID, command)
	select {
	case client.Send <- job:
		return job, nil
	default:
		return domain.Job{}, ErrClientBusy
	}
}

func (d *Dispatcher) RecordJobUpdate(job domain.Job) bool {
	return d.store.Update(job)
}
