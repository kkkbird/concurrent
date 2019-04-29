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
	time.Sleep(time.Second * 5)
	fmt.Println("work done:", d)
	return nil
}

func main() {
	m, _ := concurrent.NewBufferModule(concurrent.ModuleOptions{
		FetchChanCount: 1,
		FetchBufferLen: 200,
		WorkerCount:    10,
		F:              workerFunc,
		HandleError:    false,
	})

	var (
		allJobs = 20
		doneJob = 0
	)

	go func() {
		for i := 0; i < allJobs; i++ {
			m.Feed(i)
		}
		fmt.Println("Feed done")
	}()

	go m.Start(nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

EXIT:
	for {
		select {
		case <-sigs:
			m.Stop()
			break EXIT
		case <-m.ErrChan():
			doneJob++
			if doneJob == allJobs {
				m.Stop()
				break EXIT
			}
		}
	}
}
