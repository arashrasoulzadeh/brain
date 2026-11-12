package plugin

import "time"

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

func (p RetryPolicy) Valid() bool { return p.MaxAttempts > 0 && p.MaxAttempts <= 8 && p.BaseDelay >= 0 }

func (p RetryPolicy) Retryable(code ErrorCode, attempt int) bool {
	return p.Valid() && attempt >= 1 && attempt < p.MaxAttempts && (code == ErrorRetryable || code == ErrorUnavailable)
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}
	delay := p.BaseDelay
	for i := 1; i < attempt && delay < time.Hour/2; i++ {
		delay *= 2
	}
	return delay
}
