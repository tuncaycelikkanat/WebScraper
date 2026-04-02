package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_ImmediateSuccess(t *testing.T) {
	calls := 0
	err := Do(context.Background(), 3, 10*time.Millisecond, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_EventualSuccess(t *testing.T) {
	calls := 0
	err := Do(context.Background(), 5, 10*time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_AllAttemptsFail(t *testing.T) {
	calls := 0
	err := Do(context.Background(), 2, 10*time.Millisecond, func() error {
		calls++
		return errors.New("always fails")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	// Cancel after first failure
	err := Do(ctx, 5, 50*time.Millisecond, func() error {
		calls++
		if calls == 1 {
			cancel() // cancel context after first attempt
		}
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should have stopped after context cancellation detected in wait
	if calls > 2 {
		t.Fatalf("expected at most 2 calls (cancel after 1st), got %d", calls)
	}
}

func TestDo_SingleAttempt(t *testing.T) {
	err := Do(context.Background(), 1, 10*time.Millisecond, func() error {
		return errors.New("single fail")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
