package plugin

import (
	"errors"
	"sync"
)

var ErrWebSocketSendQueueFull = errors.New("WebSocket send queue full")

type WebSocketSendQueue struct {
	mu       sync.Mutex
	capacity int
	frames   [][]byte
}

func NewWebSocketSendQueue(capacity int) *WebSocketSendQueue {
	if capacity <= 0 {
		panic("WebSocket send queue capacity must be positive")
	}
	return &WebSocketSendQueue{capacity: capacity}
}

func (q *WebSocketSendQueue) Enqueue(frame []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.frames) >= q.capacity {
		return ErrWebSocketSendQueueFull
	}
	copyOfFrame := append([]byte(nil), frame...)
	q.frames = append(q.frames, copyOfFrame)
	return nil
}

func (q *WebSocketSendQueue) Dequeue() []byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.frames) == 0 {
		return nil
	}
	frame := q.frames[0]
	q.frames = q.frames[1:]
	return frame
}
