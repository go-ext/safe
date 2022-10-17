package syncsafe

import (
	"context"
	"log"
	"time"
)

func ExampleWaitGroup_Wait() {
	wg := NewWaitGroup()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 100)
		}()
	}
	wg.Wait()
}

func ExampleWaitGroup_WaitContext() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	wg := NewWaitGroup()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * time.Duration(i))
		}()
	}
	if err := wg.WaitContext(ctx); err != nil {
		log.Fatal(err, err.StackTrace())
	}
}

func ExampleWaitGroup_WaitChan() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	wg := NewWaitGroup()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * 2)
		}()
	}
	select {
	case <-wg.WaitChan():
	case <-ctx.Done():
		log.Fatal("context cancelled before wait ends")
	}
}
