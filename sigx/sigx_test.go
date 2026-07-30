package sigx

import (
	"syscall"
	"testing"
	"time"
)

func TestNotifyContext(t *testing.T) {
	ctx, stop := NotifyContext(syscall.SIGUSR1)
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
