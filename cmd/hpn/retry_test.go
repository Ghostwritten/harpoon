package main

import (
	"context"
	"errors"
	"testing"
	"time"

	containerruntime "github.com/harpoon/hpn/internal/runtime"
)

var fastCfg = containerruntime.RetryConfig{
	MaxAttempts: 3,
	Delay:       time.Millisecond,
	MaxDelay:    5 * time.Millisecond,
}

func TestRetryWithBackoff_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return nil
	}, fastCfg)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryWithBackoff_SucceedsOnSecondAttempt(t *testing.T) {
	calls := 0
	sentinel := errors.New("transient")
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		if calls < 2 {
			return sentinel
		}
		return nil
	}, fastCfg)
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryWithBackoff_ExhaustsAllAttempts(t *testing.T) {
	calls := 0
	sentinel := errors.New("permanent failure")
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return sentinel
	}, fastCfg)
	if err == nil {
		t.Fatal("expected error after all attempts, got nil")
	}
	if calls != fastCfg.MaxAttempts {
		t.Errorf("expected %d calls, got %d", fastCfg.MaxAttempts, calls)
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected wrapped sentinel, got %v", err)
	}
}

func TestRetryWithBackoff_WrapsLastError(t *testing.T) {
	sentinel := errors.New("root cause")
	err := retryWithBackoff(context.Background(), func() error {
		return sentinel
	}, fastCfg)
	if !errors.Is(err, sentinel) {
		t.Errorf("errors.Is should find sentinel in chain, got %v", err)
	}
}

func TestRetryWithBackoff_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	calls := 0
	err := retryWithBackoff(ctx, func() error {
		calls++
		return errors.New("should not reach here often")
	}, fastCfg)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// With an already-cancelled context the first pre-check returns early.
	if calls > 1 {
		t.Errorf("expected at most 1 call with cancelled context, got %d", calls)
	}
}

func TestRetryWithBackoff_CancelDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	cfg := containerruntime.RetryConfig{
		MaxAttempts: 5,
		Delay:       50 * time.Millisecond,
		MaxDelay:    200 * time.Millisecond,
	}

	done := make(chan error, 1)
	go func() {
		done <- retryWithBackoff(ctx, func() error {
			calls++
			if calls == 1 {
				cancel() // cancel during the first backoff sleep
			}
			return errors.New("fail")
		}, cfg)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled after cancel, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("retryWithBackoff did not respect context cancellation during wait")
	}
}

func TestRetryWithBackoff_SingleAttempt(t *testing.T) {
	sentinel := errors.New("fail")
	cfg := containerruntime.RetryConfig{MaxAttempts: 1, Delay: time.Millisecond, MaxDelay: time.Millisecond}
	calls := 0
	err := retryWithBackoff(context.Background(), func() error {
		calls++
		return sentinel
	}, cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}
