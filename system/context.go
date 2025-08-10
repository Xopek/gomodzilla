package system

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Callback func(ctx context.Context, signal os.Signal)

// NewSystemContext returns new Context, which will be cancelled on receiving SIGTERM and
// SIGINT signals after supplied delay. Additionally multiple Callback functions can be passed,
// they will be called immediately (each one in a separate goroutine) after receiving signals,
// before delay.
//
// Parameter delay specifies time to wait before cancelling the context. The primary need for this
// is cloud environment where the application may receive SIGTERM signal, but load balancers take time
// to update their routing tables. During this gap, traffic might still be sent to the terminating
// application, causing connection errors, if the app shuts down too quickly.
func NewSystemContext(ctx context.Context, delay time.Duration, callbacks ...Callback) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		for _, cb := range callbacks {
			go cb(ctx, sig)
		}

		time.Sleep(delay)

		cancel()
	}()

	return ctx
}

func NewLogSystemSignalCallback(logger *slog.Logger) Callback {
	return func(ctx context.Context, signal os.Signal) {
		logger.InfoContext(ctx, fmt.Sprintf("system signal %d (%s) received, context will be canceled shortly", signal, signal.String()))
	}
}
