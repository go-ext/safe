package syncsafe

import (
	"context"
	"fmt"
	"log"
	"time"
)

func ExampleTaggedWaitGroup_Add() {
	wg := NewTaggedWaitGroup()
	doneCalcJob := wg.Add("calculate-job", 1)
	doneSendJob := wg.Add("send-job", 1)
	go func() {
		// After a while
		doneCalcJob()
		doneSendJob()
	}()
	wg.Wait()
}

func ExampleTaggedWaitGroup_Counters() {
	wg := NewTaggedWaitGroup()
	_ = wg.Add("calculate-job", 1)
	doneSendJob := wg.Add("send-job", 1)
	fmt.Println("Before done", wg.Counters()) // Will print map[calculate-job:1 send-job:1]

	// After a while
	doneSendJob()
	fmt.Println("After done", wg.Counters()) // Will print map[calculate-job:1]
}

func ExampleTaggedWaitGroup_Wait() {
	wg := NewTaggedWaitGroup()
	for i := 0; i < 3; i++ {
		done := wg.Add(fmt.Sprintf("some-job-%d", i), 1)
		go func() {
			defer done()
			// Some work
			time.Sleep(time.Millisecond * 100)
		}()
	}
	wg.Wait()
}

func ExampleTaggedWaitGroup_WaitContext() {
	// Set timeout to 1 second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	wg := NewTaggedWaitGroup()
	for i := 0; i < 3; i++ {
		done := wg.Add(fmt.Sprintf("some-job-%d", i), 1)
		go func(i int) {
			defer done()
			// Some work will take longer than timeout
			time.Sleep(time.Second * time.Duration(i))
		}(i)
	}
	if err := wg.WaitContext(ctx); err != nil {
		log.Fatal(err, ": ", err.StackTrace())
	}
}

func ExampleTaggedWaitGroup_WaitChan() {
	// Set timeout to 1 second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	wg := NewTaggedWaitGroup()
	for i := 0; i < 3; i++ {
		done := wg.Add(fmt.Sprintf("some-job-%d", i), 1)
		go func(i int) {
			defer done()
			// Some work will take longer than timeout
			time.Sleep(time.Second * time.Duration(i))
		}(i)
	}
	select {
	case <-wg.WaitChan():
	case <-ctx.Done():
		log.Fatal("context cancelled before wait group done")
	}
}
