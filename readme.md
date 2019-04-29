# sample codes implement advanced-go-concurrency-patterns

ref to [Advanced Go Concurrency Patterns](!https://blog.golang.org/advanced-go-concurrency-patterns)

## example

``` golang
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

```

more complex [example](/example/main.go)

## options

## FetchBufferLen

length to store job if all worker is busy

### OverflowBehaivour

* OverFlowNolimit: always do jobs, `FetchBufferLen` will be ignored in this case
* OverFlowDrop : new job will be dropped if jobs in queue exceed `FetchBufferLen`
* OverFlowReplace : new job will replace first job in waiting job queue if jobs in queue exceed `FetchBufferLen`

### HandleResult

Check if job is done, dropped or Replaced, see [example](/example/main.go) for detail