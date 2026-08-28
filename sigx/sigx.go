// Package sigx provides graceful-shutdown signal handling.
package sigx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Context returns a copy of parent that is cancelled when the process
// receives one of sigs (default: SIGINT, SIGTERM). After the first
// signal the default disposition is restored, so a second signal kills
// the process immediately instead of waiting for graceful shutdown.
//
// Call the returned stop function to release the signal handler before
// the returned context is otherwise done, mirroring
// signal.NotifyContext.
func Context(
	parent context.Context,
	sigs ...os.Signal,
) (ctx context.Context, stop context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if len(sigs) == 0 {
		sigs = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	ctx, stop = signal.NotifyContext(parent, sigs...)
	go func() {
		<-ctx.Done()
		stop() // restore default behavior: next signal kills instantly
	}()
	return ctx, stop
}
