package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryHandoffSingleUseAndExpiry(t *testing.T) {
	m := NewMemory()
	ctx := context.Background()
	now := time.Now()
	code := []byte("hashed-code")
	if err := m.CreateHandoff(ctx, code, "session-token", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	token, err := m.ConsumeHandoff(ctx, code, now)
	if err != nil || token != "session-token" {
		t.Fatalf("first exchange: %q, %v", token, err)
	}
	if _, err := m.ConsumeHandoff(ctx, code, now); !errors.Is(err, ErrNotFound) {
		t.Fatalf("replayed code: %v", err)
	}
	if err := m.CreateHandoff(ctx, code, "expired", now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.ConsumeHandoff(ctx, code, now); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired code: %v", err)
	}
}
