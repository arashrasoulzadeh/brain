package plugin

import "errors"

var ErrWebSocketSendQueueFull = errors.New("WebSocket send queue full")

type WebSocketSendQueue struct {
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
	if len(q.frames) >= q.capacity {
		return ErrWebSocketSendQueueFull
	}
	copyOfFrame := append([]byte(nil), frame...)
	q.frames = append(q.frames, copyOfFrame)
	return nil
}

func (q *WebSocketSendQueue) Dequeue() []byte {
	if len(q.frames) == 0 {
		return nil
	}
	frame := q.frames[0]
	q.frames = q.frames[1:]
	return frame
}
