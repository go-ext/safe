package syncsafe

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

// DoneFn specifies a type of TaggedWaitGroup done function which decreases counter increased with Add call previously
type DoneFn func()

// TaggedWaitGroup provides a way to wait for all async routines to be done exactly as sync.WaitGroup does it but
// providing more controllable ways of waiting, avoiding infinite blocking in case of any unexpected circumstances.
// It also gives a way of tagging every Add operation and have insights on pending counters (by tags) at any time.
// Due to tagging specifics, this kind of wait group doesn't provide a Done method from top level explicitly but returns
// a done function for a specific tag from Add call. Done function behaves in exactly the same way as native
// sync.WaitGroup decreasing particular tag counter by 1.
// TaggedWaitGroup instance is a single use only to prevent potential unnecessary mess in case of re-using
type TaggedWaitGroup struct {
	counters map[string]int64
	lck      *sync.Mutex
	done     chan struct{}
}

// NewTaggedWaitGroup returns a new instance of WaitGroup
func NewTaggedWaitGroup() *TaggedWaitGroup {
	return &TaggedWaitGroup{
		done:     make(chan struct{}),
		counters: make(map[string]int64),
		lck:      &sync.Mutex{},
	}
}

// adds delta with specific tag in wait group
func (g *TaggedWaitGroup) add(tag string, delta int64) {
	g.lck.Lock()
	defer g.lck.Unlock()
	g.counters[tag] = g.counters[tag] + delta
	v := g.counters[tag]
	switch {
	case v == 0:
		delete(g.counters, tag)
		if g.countersAreZero() {
			close(g.done)
		}
	case v < 0:
		panic(fmt.Sprintf("wait group counter has gone negative for tag %s", tag))
	}
}

// Add increases by 1 wait group counter having provided tag
func (g *TaggedWaitGroup) Add(tag string, delta int64) DoneFn {
	g.add(tag, delta)
	return func() {
		g.add(tag, -1)
	}
}

// Wait blocks the routine until all tagged counters are zero
func (g *TaggedWaitGroup) Wait() {
	g.lck.Lock()
	if g.countersAreZero() {
		g.lck.Unlock()
		return
	}
	g.lck.Unlock()
	<-g.done
	return
}

// WaitContext blocks the routine until all counters are zero or ctx is done, whatever comes first
// An appropriate error will be returned if ctx is done before counters are zero
func (g *TaggedWaitGroup) WaitContext(ctx context.Context) StackError {
	g.lck.Lock()
	if g.countersAreZero() {
		g.lck.Unlock()
		return nil
	}
	g.lck.Unlock()
	select {
	case <-ctx.Done():
		return newWaitError(errors.Wrap(ctx.Err(), "context done"))
	case <-g.done:
		return nil
	}
}

// WaitChan returns a channel which can be used to implement custom wait handling behavior
// Channel will be closed once wait group all counters are zero
func (g *TaggedWaitGroup) WaitChan() <-chan struct{} {
	select {
	case <-g.done:
		return g.done
	default:
		g.lck.Lock()
		if g.countersAreZero() {
			close(g.done)
		}
		g.lck.Unlock()
	}
	return g.done
}

// countersAreZero reports whether all counters are zero
func (g *TaggedWaitGroup) countersAreZero() bool {
	return len(g.counters) == 0
}

// Counters returns counters' current state
func (g *TaggedWaitGroup) Counters() map[string]int64 {
	g.lck.Lock()
	defer g.lck.Unlock()
	// Clone the counters map
	counters := make(map[string]int64, len(g.counters))
	for tag, cnt := range g.counters {
		counters[tag] = cnt
	}
	return counters
}
