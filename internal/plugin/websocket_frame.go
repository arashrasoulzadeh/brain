package plugin

import "errors"

var ErrWebSocketFrameTooLarge = errors.New("WebSocket frame too large")

func ValidateWebSocketFrame(frame []byte, maxBytes int) error {
	if maxBytes <= 0 || len(frame) > maxBytes {
		return ErrWebSocketFrameTooLarge
	}
	return nil
}
