package waiter

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Starter must be implemented for those types, that are going to be run concurrently and
// will be used by Runner.
type Starter interface {
	Start(ctx context.Context) chan error
}

// Runner is the wrapper for sync.WG module. It accumulates Doners and waits until they all
// finish their work.
type Runner struct {
	ctx context.Context
	d   []Starter
	sync.RWMutex
	eg *errgroup.Group
}

// Add adds new delta to underlying WG counter. It starts new routine and
// that waits doner to be done, and then decrements WG counter for that doner.
func (w *Runner) Add(d ...Starter) {
	w.Lock()
	w.d = append(w.d, d...)
	w.Unlock()
}

// Wait calls underlying sync.Wg wait method to wait all doners to finish their work.
func (w *Runner) Run() error {
	defer func() {
		w.d = make([]Starter, 0)
	}()
	for _, d := range w.d {
		d := d
		w.eg.Go(func() error {
			return <-d.Start(w.ctx)
		})
	}

	return w.eg.Wait()
}

// New creates new instance of Runner.
func New(ctx context.Context) *Runner {
	eg, ctx := errgroup.WithContext(ctx)
	return &Runner{
		ctx: ctx,
		d:   make([]Starter, 0),
		eg:  eg,
	}
}
