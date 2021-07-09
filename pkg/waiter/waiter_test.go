package waiter

import (
	"context"
	"testing"
	"time"
)

func TestWaiter(t *testing.T) {
	waiter := New(context.Background())
	deadline := time.Millisecond * 50
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(deadline))
	doners := []testStarter{
		{
			ctx: ctx,
		},
		{
			ctx: ctx,
		},
	}

	for _, d := range doners {
		waiter.Add(d)
	}

	if len(waiter.d) != len(doners) {
		t.Fatalf("waiter exptected to add doners to its list. Got: %d", len(waiter.d))
	}

	go func(t *testing.T) {
		timer := time.NewTimer(deadline + deadline)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			t.Errorf("All doners are expected to done by this moment")
			cancel()
		}
	}(t)

	err := waiter.Run()

	if len(waiter.d) != 0 {
		t.Fatalf("waiter expected to clear all its doners")
	}

	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
}

type testStarter struct {
	ctx context.Context
}

func (d testStarter) Start(ctx context.Context) chan error {
	done := make(chan error)
	go func() {
		<-d.ctx.Done()
		done <- nil
	}()
	return done
}
