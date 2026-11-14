package plugin

import (
	"errors"
	"sync"
)

var ErrWebSocketConnectionStillTracked = errors.New("WebSocket connection still tracked")

type WebSocketConnection interface {
	Close() error
}

type WebSocketRegistry struct {
	mu          sync.Mutex
	connections map[string]WebSocketConnection
}

func NewWebSocketRegistry() *WebSocketRegistry {
	return &WebSocketRegistry{connections: make(map[string]WebSocketConnection)}
}

func (r *WebSocketRegistry) Register(id string, connection WebSocketConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[id] = connection
}

func (r *WebSocketRegistry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.connections, id)
}

func (r *WebSocketRegistry) CloseAll() error {
	r.mu.Lock()
	connections := make([]WebSocketConnection, 0, len(r.connections))
	for _, connection := range r.connections {
		connections = append(connections, connection)
	}
	r.connections = make(map[string]WebSocketConnection)
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
