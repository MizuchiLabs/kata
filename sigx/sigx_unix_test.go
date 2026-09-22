//go:build unix

package sigx

import (
	"syscall"
	"testing"
	"time"
)

func TestNotifyContext(t *testing.T) {
	ctx, unregistered := notifyContext(syscall.SIGUSR1)

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("send signal: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled after SIGUSR1")
	}

	select {
	case <-unregistered:
	case <-time.After(2 * time.Second):
		t.Fatal("signal handler was not unregistered after the first signal")
	}
}
