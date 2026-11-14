package plugin

import "errors"

var ErrInvalidWebSocketMessage = errors.New("invalid WebSocket message")

type WebSocketMessage struct {
	Protocol  int    `json:"protocol"`
	RequestID string `json:"request_id"`
	ProjectID string `json:"project_id"`
	RoomID    string `json:"room_id"`
	Type      string `json:"type"`
	Sequence  uint64 `json:"sequence"`
	Payload   []byte `json:"payload"`
}

func (m WebSocketMessage) Validate(protocol int) error {
	if m.Protocol != protocol || m.RequestID == "" || m.ProjectID == "" || m.RoomID == "" || m.Type == "" || len(m.Payload) > 1<<20 {
		return ErrInvalidWebSocketMessage
	}
	return nil
}

func (m WebSocketMessage) ValidateForProject(protocol int, projectID string) error {
	if err := m.Validate(protocol); err != nil {
		return err
	}
	if projectID == "" || m.ProjectID != projectID {
		return ErrInvalidWebSocketMessage
	}
	return nil
}
