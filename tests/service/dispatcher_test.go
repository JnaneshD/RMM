package service_test

import (
	"testing"

	"example.com/test/internal/domain"
	"example.com/test/internal/server/realtime"
	"example.com/test/internal/server/service"
)

func TestDispatcher_IsClientExists(t *testing.T) {
	hub := realtime.NewHub()
	
	// Create the dispatcher using the exported constructor
	dispatcher := service.NewDispatcher(hub, nil)
	
	clientID := "test-client-123"
	
	// 1. Test client doesn't exist initially
	if dispatcher.IsClientExists(clientID) {
		t.Errorf("expected client to not exist initially")
	}
	
	// 2. Register a mock client
	client := realtime.NewClient(clientID, nil)
	dispatcher.RegisterClient(client)
	
	// 3. Test client exists after registration
	if !dispatcher.IsClientExists(clientID) {
		t.Errorf("expected client to exist after registration")
	}
	
	// 4. Unregister the client
	dispatcher.UnregisterClient(client)
	
	// 5. Test client doesn't exist after unregistering
	if dispatcher.IsClientExists(clientID) {
		t.Errorf("expected client to not exist after being unregistered")
	}
}

func TestDispatcher_Dispatch(t *testing.T) {
	hub := realtime.NewHub()
	dispatcher := service.NewDispatcher(hub, nil)
	
	clientID := "client-abc"
	
	// 1. Test dispatch to non-existent client
	_, err := dispatcher.Dispatch(clientID, "echo test", "cmd", "C:\\")
	if err != service.ErrClientNotFound {
		t.Errorf("expected ErrClientNotFound, got %v", err)
	}
	
	// 2. Register the client
	client := realtime.NewClient(clientID, nil)
	// Make the channel buffered so the test doesn't block
	client.Send = make(chan domain.Job, 1)
	dispatcher.RegisterClient(client)
	
	// 3. Test successful dispatch
	job, err := dispatcher.Dispatch(clientID, "echo test", "cmd", "C:\\")
	if err != nil {
		t.Errorf("expected successful dispatch, got %v", err)
	}
	
	if job.Command != "echo test" {
		t.Errorf("expected command 'echo test', got %s", job.Command)
	}
	if job.Status != domain.WAIT {
		t.Errorf("expected status WAIT, got %v", job.Status)
	}
	
	// 4. Test client busy (the buffered channel is now full)
	_, err = dispatcher.Dispatch(clientID, "echo test", "cmd", "C:\\")
	if err != service.ErrClientBusy {
		t.Errorf("expected ErrClientBusy, got %v", err)
	}
}
