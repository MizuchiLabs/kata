package sigx

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestContextCancelsOnSignal(t *testing.T) {
	ctx, stop := Context(context.Background(), syscall.SIGUSR1)
	defer stop()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send signal: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled after SIGUSR1")
	}
}

func TestContextUsesDefaultSignals(t *testing.T) {
	// SIGUSR1 is not in the default set, so this context must survive it.
	ctx, stop := Context(context.Background())
	defer stop()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send signal: %v", err)
	}

	select {
	case <-ctx.Done():
		t.Fatalf("context cancelled by %v, want it to ignore non-default signals", ctx.Err())
	case <-time.After(100 * time.Millisecond):
	}
}

func TestContextParentCancellation(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	ctx, stop := Context(parent, syscall.SIGUSR1)
	defer stop()

	parentCancel()

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("child context was not cancelled after parent cancellation")
	}
}

func TestContextNilParent(t *testing.T) {
	// A nil parent must not panic; it falls back to context.Background.
	ctx, stop := Context(context.TODO(), syscall.SIGUSR1)
	defer stop()

	if ctx == nil || stop == nil {
		t.Fatal("Context(nil, ...) returned nil values")
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send signal: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled after SIGUSR1")
	}
}
