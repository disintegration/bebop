package jwt

import (
	"strings"
	"testing"
	"time"
)

func TestCreateAndVerify(t *testing.T) {
	userID := int64(10)
	secret := strings.Repeat("0", 64)
	badSecret := strings.Repeat("1", 64)

	s, err := NewService(secret)
	if err != nil {
		t.Fatalf("failed to create a new service: %s", err)
	}

	tokenString, err := s.Create(userID)
	if err != nil {
		t.Fatalf("failed to create a token: %s", err)
	}

	gotUserID, gotIssuedAt, err := s.Verify(tokenString)
	if err != nil {
		t.Fatalf("failed to verify a token: %s", err)
	}

	if gotUserID != userID {
		t.Fatalf("bad verified userID: got %v; want %v", gotUserID, userID)
	}

	now := time.Now()
	since := now.Sub(gotIssuedAt)
	if since < 0 || since > 3*time.Second {
		t.Fatalf("bad issuedAt of the verified token: %v; now: %v", gotIssuedAt, now)
	}

	s, err = NewService(badSecret)
	if err != nil {
		t.Fatalf("failed to create a new service: %s", err)
	}
	_, _, err = s.Verify(tokenString)
	if err == nil {
		t.Fatalf("no error on verifying with bad secret")
	}
}
