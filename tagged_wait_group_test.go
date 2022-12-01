package syncsafe

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewTaggedWaitGroup(t *testing.T) {
	t.Parallel()
	got := NewTaggedWaitGroup()
	require.NotNil(t, got)
	assert.NotNil(t, got.done)
	assert.NotNil(t, got.counters)
	assert.NotNil(t, got.lck)
}

func TestTaggedWaitGroup_add(t *testing.T) {
	t.Parallel()
	type args struct {
		tag   string
		delta int64
	}
	tests := []struct {
		name string
		wg   *TaggedWaitGroup
		args args
		want int64
	}{
		{
			name: "add positive value",
			wg:   NewTaggedWaitGroup(),
			args: args{tag: "tag1", delta: 1},
			want: 1,
		},
		{
			name: "add zero value",
			wg:   NewTaggedWaitGroup(),
			args: args{tag: "tag1", delta: 0},
			want: 0,
		},
		{
			name: "counter goes to negative",
			wg:   NewTaggedWaitGroup(),
			args: args{tag: "tag1", delta: -1},
			want: -1,
		},
		{
			name: "add value with empty tag",
			wg:   NewTaggedWaitGroup(),
			args: args{tag: "", delta: 1},
			want: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.want < 0 {
				require.Panics(t, func() {
					tt.wg.add(tt.args.tag, tt.args.delta)
				})
				return
			}
			tt.wg.add(tt.args.tag, tt.args.delta)
			assert.Equal(t, tt.want, tt.wg.counters[tt.args.tag])
		})
	}
}

func TestTaggedWaitGroup_Add(t *testing.T) {
	t.Parallel()
	wg := NewTaggedWaitGroup()
	tag1 := "some-tag-1"
	tag2 := "some-tag-2"
	wg.Add(tag1, 1)
	assert.Equal(t, int64(1), wg.counters[tag1])
	assert.Equal(t, 1, len(wg.counters))
	wg.Add(tag2, 2)
	assert.Equal(t, int64(2), wg.counters[tag2])
	assert.Equal(t, 2, len(wg.counters))
}

func TestTaggedWaitGroup_done(t *testing.T) {
	t.Parallel()
	t.Run("should decrement by 1", func(t *testing.T) {
		t.Parallel()
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		done := wg.Add(tag, 5)
		done()
		done()
		assert.Equal(t, int64(3), wg.counters[tag])
	})
	t.Run("should panic due to going negative", func(t *testing.T) {
		t.Parallel()
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		done := wg.Add(tag, 1)
		done()
		assert.Panics(t, func() {
			done()
		})
	})
}

func TestTaggedWaitGroup_Wait(t *testing.T) {
	t.Parallel()
	wg := NewTaggedWaitGroup()
	tag := "some-tag"
	done := wg.Add(tag, 1)
	waited := false
	go func() {
		time.Sleep(time.Second * 1)
		waited = true
		done()
	}()
	wg.Wait()
	assert.True(t, waited, "looks like wg is not waiting properly")
}

func TestTaggedWaitGroup_WaitContext(t *testing.T) {
	t.Parallel()
	t.Run("should be done with no error", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		done := wg.Add(tag, 1)
		go func() {
			defer done()
		}()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		assert.NoError(t, wg.WaitContext(ctx))
		guard.dismiss()
	})
	t.Run("should be done with error by timeout", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		done := wg.Add(tag, 1)
		go func() {
			time.Sleep(time.Second * 2)
			defer done()
		}()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		err := wg.WaitContext(ctx)
		guard.dismiss()
		assert.Error(t, err)
		assert.NotEmpty(t, err.StackTrace())
	})
	t.Run("should be done as untouched", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		wg := NewTaggedWaitGroup()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		assert.NoError(t, wg.WaitContext(context.Background()))
		guard.dismiss()
	})
}

func TestTaggedWaitGroup_WaitChan(t *testing.T) {
	t.Parallel()
	t.Run("should be closed when counter is 0", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		done := wg.Add(tag, 1)
		go func() {
			defer done()
		}()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		<-wg.WaitChan()
		guard.dismiss()
	})
	t.Run("should not be closed", func(t *testing.T) {
		t.Parallel()
		wg := NewTaggedWaitGroup()
		tag := "some-tag"
		wg.Add(tag, 1)
		select {
		case <-wg.WaitChan():
			t.Error("wg channel should not be closed")
		default:
		}
	})
	t.Run("should be closed as untouched", func(t *testing.T) {
		t.Parallel()
		wg := NewTaggedWaitGroup()
		select {
		case <-wg.WaitChan():
		default:
			t.Error("wg channel should be closed")
		}
	})
}

func TestTaggedWaitGroup_PendingCounters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prepFn func(wg *TaggedWaitGroup)
		want   map[string]int64
	}{
		{
			name: "should return all counters as they were added",
			prepFn: func(wg *TaggedWaitGroup) {
				wg.Add("one", 1)
				wg.Add("two", 5)
			},
			want: map[string]int64{
				"one": 1,
				"two": 5,
			},
		},
		{
			name: "should return all counters but some of them was done",
			prepFn: func(wg *TaggedWaitGroup) {
				wg.Add("one", 1)
				wg.Add("two", 5)()
			},
			want: map[string]int64{
				"one": 1,
				"two": 4,
			},
		},
		{
			name: "should return only those counters which is > 0",
			prepFn: func(wg *TaggedWaitGroup) {
				wg.Add("one", 1)()
				wg.Add("two", 5)
			},
			want: map[string]int64{
				"two": 5,
			},
		},
		{
			name: "should return counter with empty tag",
			prepFn: func(wg *TaggedWaitGroup) {
				wg.Add("", 1)
			},
			want: map[string]int64{
				"": 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wg := NewTaggedWaitGroup()
			tt.prepFn(wg)
			assert.Equal(t, tt.want, wg.Counters())
		})
	}
}

// asyncGuard is just a testing helper to prevent hanging of async test cases
type asyncGuard struct {
	t  *testing.T
	ch chan struct{}
}

func newAsyncGuard(t *testing.T) *asyncGuard {
	return &asyncGuard{
		t:  t,
		ch: make(chan struct{}),
	}
}

func (g *asyncGuard) run(timeout time.Duration, errMsg string) {
	time.Sleep(timeout)
	select {
	case <-g.ch:
	default:
		g.t.Fatal(errMsg)
	}
}

func (g *asyncGuard) dismiss() {
	close(g.ch)
}
