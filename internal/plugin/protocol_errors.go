package plugin

type ErrorCode string

const (
	ErrorInvalidRequest ErrorCode = "invalid_request"
	ErrorUnauthorized   ErrorCode = "unauthorized"
	ErrorNotFound       ErrorCode = "not_found"
	ErrorConflict       ErrorCode = "conflict"
	ErrorRetryable      ErrorCode = "retryable"
	ErrorUnavailable    ErrorCode = "unavailable"
)

type ProtocolError struct {
	RequestID string    `json:"request_id"`
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
}

func (e ProtocolError) Valid() bool {
	if e.RequestID == "" || e.Message == "" {
		return false
	}
	switch e.Code {
	case ErrorInvalidRequest, ErrorUnauthorized, ErrorNotFound, ErrorConflict, ErrorRetryable, ErrorUnavailable:
		return true
	default:
		return false
	}
}
