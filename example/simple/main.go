package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kkkbird/concurrent"
)

func workerFunc(ctx context.Context, d interface{}) error {
	fmt.Println("work start:", d)
	time.Sleep(time.Second)
	fmt.Println("  work done:", d)
	return nil
}

func main() {
	m, _ := concurrent.NewBufferModule(concurrent.ModuleOptions{
		WorkerCount: 10,
		F:           workerFunc,
	})

	go func() {
		for i := 0; i < 100; i++ {
			m.Feed(i)
		}
		fmt.Println("Feed done")
	}()

	m.Start(nil)
}
