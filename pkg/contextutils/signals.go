package contextutils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WithSignals(parent context.Context, signals ...os.Signal) context.Context {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)
	ctx, cancel := context.WithCancel(parent)

	go func() {
		<-stop
		cancel()
	}()

	return ctx
}
