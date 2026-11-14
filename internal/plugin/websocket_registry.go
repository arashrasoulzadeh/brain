package plugin

import (
	"errors"
	"sync"
)

var ErrWebSocketConnectionStillTracked = errors.New("WebSocket connection still tracked")
var ErrWebSocketProjectRequired = errors.New("WebSocket project required")

type WebSocketConnection interface {
	Close() error
}

type WebSocketRegistry struct {
	mu          sync.Mutex
	connections map[string]webSocketConnectionEntry
}

type webSocketConnectionEntry struct {
	projectID  string
	connection WebSocketConnection
}

func NewWebSocketRegistry() *WebSocketRegistry {
	return &WebSocketRegistry{connections: make(map[string]webSocketConnectionEntry)}
}

func (r *WebSocketRegistry) Register(id string, connection WebSocketConnection) {
	r.register(id, "", connection)
}

func (r *WebSocketRegistry) RegisterProject(id, projectID string, connection WebSocketConnection) error {
	if !isValidProjectID(projectID) {
		return ErrWebSocketProjectRequired
	}
	r.register(id, projectID, connection)
	return nil
}

func (r *WebSocketRegistry) register(id, projectID string, connection WebSocketConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[id] = webSocketConnectionEntry{projectID: projectID, connection: connection}
}

func (r *WebSocketRegistry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.connections, id)
}

func (r *WebSocketRegistry) CloseAll() error {
	r.mu.Lock()
	connections := make([]WebSocketConnection, 0, len(r.connections))
	for _, entry := range r.connections {
		connections = append(connections, entry.connection)
	}
	r.connections = make(map[string]webSocketConnectionEntry)
	r.mu.Unlock()

	var closeError error
	for _, connection := range connections {
		if err := connection.Close(); err != nil && closeError == nil {
			closeError = err
		}
	}
	return closeError
}

func (r *WebSocketRegistry) CloseProject(projectID string) error {
	if !isValidProjectID(projectID) {
		return ErrWebSocketProjectRequired
	}
	r.mu.Lock()
	connections := make([]WebSocketConnection, 0)
	for id, entry := range r.connections {
		if entry.projectID == projectID {
			connections = append(connections, entry.connection)
			delete(r.connections, id)
		}
	}
	r.mu.Unlock()
	var closeError error
	for _, connection := range connections {
		if err := connection.Close(); err != nil && closeError == nil {
			closeError = err
		}
	}
	return closeError
}

func (r *WebSocketRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.connections)
}
