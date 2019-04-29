package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
		FetchChanCount:    1,
		FetchBufferLen:    25,
		WorkerCount:       10,
		F:                 workerFunc,
		HandleResult:      true,
		OverflowBehaivour: concurrent.OverFlowDrop,
	})

	var (
		allJobCount    = 100
		finishJobCount = 0
	)

	go func() {
		for i := 0; i < allJobCount; i++ {
			//time.Sleep(time.Millisecond * 50)
			m.Feed(i)
		}
		fmt.Println("Feed done")
	}()

	go m.Start(nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	doneJobs := make([]int, 0)
	dropedJobs := make([]int, 0)
	replacedJobs := make([]int, 0)
EXIT:
	for {
		select {
		case <-sigs:
			m.Stop()
			break EXIT
		case r := <-m.ResultChan():
			if r.IsDropped() {
				dropedJobs = append(dropedJobs, r.Data.(int))
			} else if r.IsReplaced() {
				replacedJobs = append(replacedJobs, r.Data.(int))
			} else {
				doneJobs = append(doneJobs, r.Data.(int))
			}

			finishJobCount++
			if finishJobCount == allJobCount {
				m.Stop()

				fmt.Printf("Done(%d):%v\n", len(doneJobs), doneJobs)
				fmt.Printf("Dropped(%d):%v\n", len(dropedJobs), dropedJobs)
				fmt.Printf("Replaced(%d):%v\n", len(replacedJobs), replacedJobs)

				break EXIT
			}
		}
	}
}
