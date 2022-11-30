package syncsafe

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWaitGroup(t *testing.T) {
	t.Parallel()
	got := NewWaitGroup()
	require.NotNil(t, got)
	assert.NotNil(t, got.done)
	assert.Zero(t, got.cnt)
}

func TestWaitGroup_add(t *testing.T) {
	t.Parallel()
	type args struct {
		delta int64
	}
	tests := []struct {
		name    string
		wg      *WaitGroup
		args    args
		wantCnt int64
	}{
		{
			name:    "add positive value",
			wg:      NewWaitGroup(),
			args:    args{delta: 1},
			wantCnt: 1,
		},
		{
			name:    "add zero value",
			wg:      NewWaitGroup(),
			args:    args{delta: 0},
			wantCnt: 0,
		},
		{
			name:    "counter goes to negative",
			wg:      NewWaitGroup(),
			args:    args{delta: -1},
			wantCnt: -1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantCnt < 0 {
				require.Panics(t, func() {
					tt.wg.add(tt.args.delta)
				})
				return
			}
			tt.wg.add(tt.args.delta)
			assert.Equal(t, tt.wantCnt, tt.wg.cnt)
		})
	}
}

func TestWaitGroup_Done(t *testing.T) {
	t.Parallel()
	t.Run("should decrement by 1", func(t *testing.T) {
		t.Parallel()
		wg := NewWaitGroup()
		wg.Add(5)
		wg.Done()
		wg.Done()
		assert.Equal(t, int64(3), wg.cnt)
	})
	t.Run("should panic due to going negative", func(t *testing.T) {
		t.Parallel()
		wg := NewWaitGroup()
		wg.Add(1)
		wg.Done()
		assert.Panics(t, func() {
			wg.Done()
		})
	})
}

func TestWaitGroup_WaitContext(t *testing.T) {
	t.Parallel()
	t.Run("should be done with no error", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		wg := NewWaitGroup()
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		wg := NewWaitGroup()
		wg.Add(1)
		go func() {
			time.Sleep(time.Second * 2)
			defer wg.Done()
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
		wg := NewWaitGroup()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		assert.NoError(t, wg.WaitContext(context.Background()))
		guard.dismiss()
	})
}

func TestWaitGroup_WaitChan(t *testing.T) {
	t.Parallel()
	t.Run("should be closed when counter is 0", func(t *testing.T) {
		t.Parallel()
		guard := newAsyncGuard(t)
		wg := NewWaitGroup()
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()

		go guard.run(time.Second*1, "looks like wg was not done properly and still waiting")
		<-wg.WaitChan()
		guard.dismiss()
	})
	t.Run("should not be closed", func(t *testing.T) {
		t.Parallel()
		wg := NewWaitGroup()
		wg.Add(1)
		select {
		case <-wg.WaitChan():
			t.Error("wg channel should not be closed")
		default:
		}
	})
	t.Run("should be closed as untouched", func(t *testing.T) {
		t.Parallel()
		wg := NewWaitGroup()
		select {
		case <-wg.WaitChan():
		default:
			t.Error("wg channel should be closed")
		}
	})
}
