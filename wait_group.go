package syncsafe

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// WaitGroup provides a way to wait for all async routines to be done exactly as sync.WaitGroup does it but providing
// more controllable ways of waiting, avoiding infinite blocking in case of any unexpected circumstances
// WaitGroup instance is a single use only to prevent potential unnecessary mess in case of re-using
type WaitGroup struct {
	cnt  int64
	done chan struct{}
	lck  *sync.Mutex
}

// NewWaitGroup returns a new instance of WaitGroup
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		done: make(chan struct{}),
		lck:  &sync.Mutex{},
	}
}

// adds delta to the wait group
func (g *WaitGroup) add(delta int64) {
	g.lck.Lock()
	defer g.lck.Unlock()
	v := atomic.AddInt64(&g.cnt, delta)
	switch {
	case v == 0:
		close(g.done)
	case v < 0:
		panic("wait group counter has gone negative")
	}
}

// Add adds delta to the wait group counter
// delta could be negative to implement done behavior with value >1, but it will panic if wait group counter goes
// to negative
func (g *WaitGroup) Add(delta int64) {
	g.add(delta)
}

// Done decreases wait group counter by 1
func (g *WaitGroup) Done() {
	g.add(-1)
}

// Wait blocks the routine until counter is zero in exactly the same behavior as native sync.WaitGroup does
func (g *WaitGroup) Wait() {
	if atomic.LoadInt64(&g.cnt) == 0 {
		return
	}
	<-g.done
	return
}

// WaitContext blocks the routine until counter is zero or ctx is done, whatever comes first
// An appropriate error will be returned if ctx is done before the counter is zero
func (g *WaitGroup) WaitContext(ctx context.Context) StackError {
	if atomic.LoadInt64(&g.cnt) == 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return newWaitError(errors.Wrap(ctx.Err(), "context done unexpectedly"))
	case <-g.done:
		return nil
	}
}

// WaitChan returns a channel which can be used to implement custom wait handling behavior
// Channel will be closed once wait group counter is zero
func (g *WaitGroup) WaitChan() <-chan struct{} {
	select {
	case <-g.done:
		return g.done
	default:
		if atomic.LoadInt64(&g.cnt) == 0 {
			close(g.done)
		}
	}
	return g.done
}
