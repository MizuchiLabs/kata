// Package sigx provides graceful-shutdown signal handling.
package sigx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// NotifyContext returns a context that is cancelled when the process
// receives one of sigs (default: SIGINT, SIGTERM). After the first
// signal the default disposition is restored, so a second signal kills
// the process immediately instead of waiting for graceful shutdown.
func NotifyContext(sigs ...os.Signal) context.Context {
	if len(sigs) == 0 {
		sigs = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	go func() {
		<-ctx.Done()
		stop() // restore default behavior: next signal kills instantly
	}()
	return ctx
}
