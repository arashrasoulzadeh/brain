package operator

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func (s *Store) EnableMFA(secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.account == nil || s.authenticatedOperator == "" || strings.TrimSpace(secret) == "" || totp(secret, 0) == "" {
		return ErrInvalidCredentials
	}
	s.account.MFASecret = secret
	return nil
}

func (s *Store) VerifyMFA(code string, now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.account == nil || s.account.MFASecret == "" || len(code) != 6 {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		if subtle.ConstantTimeCompare([]byte(code), []byte(totp(s.account.MFASecret, now.Unix()/30+offset))) == 1 {
			return true
		}
	}
	return false
}

func totp(secret string, counter int64) string {
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.ReplaceAll(secret, " ", "")))
	if err != nil || len(decoded) == 0 {
		return ""
	}
	var input [8]byte
	binary.BigEndian.PutUint64(input[:], uint64(counter))
	h := hmac.New(sha1.New, decoded)
	_, _ = h.Write(input[:])
	sum := h.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (uint32(sum[offset])&0x7f)<<24 | uint32(sum[offset+1])<<16 | uint32(sum[offset+2])<<8 | uint32(sum[offset+3])
	return fmt.Sprintf("%06d", value%1000000)
}
