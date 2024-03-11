package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	fmt.Printf("Process ID: %v\n", os.Getpid())
}

func main() {
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	generator := func(dataItem string, stream chan interface{}) {
		for {
			select {
			case <-ctx.Done():
				return
			case stream <- dataItem:
			}
		}
	}

	infiniteApples := make(chan interface{})
	go generator("apple", infiniteApples)

	infiniteOranges := make(chan interface{})
	go generator("orange", infiniteOranges)

	infiniteBananas := make(chan interface{})
	go generator("banana", infiniteBananas)

	wg.Add(1)
	go func1(ctx, &wg, infiniteApples)

	func2 := genericFunc("func2")
	func3 := genericFunc("func3")

	wg.Add(1)
	go func2(ctx, &wg, infiniteOranges)

	wg.Add(1)
	go func3(ctx, &wg, infiniteBananas)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		<-sc
		fmt.Println("=== Shutdown received ===")
		cancel()
	}()

	wg.Wait()
	fmt.Println("All workers done, shutting down!")
}

func func1(parentCtx context.Context, parentWg *sync.WaitGroup, stream <-chan interface{}) {
	defer parentWg.Done()
	var wg sync.WaitGroup

	doWork := func(ctx context.Context, id int) {
		defer wg.Done()
		fmt.Printf("@@ doWork[%d]: start\n", id)

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("@@ doWork[%d]: done\n", id)
				return
			case d, ok := <-stream:
				if !ok {
					return
				}
				fmt.Println(d)
			default:
				time.Sleep(time.Second)
			}
		}
	}

	ctx, cancel := context.WithTimeout(parentCtx, time.Second*3)
	defer cancel()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go doWork(ctx, i)
	}

	wg.Wait()
}

type genericReturnFunc func(ctx context.Context, wg *sync.WaitGroup, stream <-chan interface{})

func genericFunc(name string) genericReturnFunc {
	return func(ctx context.Context, wg *sync.WaitGroup, stream <-chan interface{}) {
		defer wg.Done()
		fmt.Printf("-- genericFunc[%s]: start\n", name)

		ctxDeadline, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
		defer cancel()

		for {
			select {
			case <-ctxDeadline.Done():
				fmt.Printf("-- genericFunc[%s]: done\n", name)
				return
			case d, ok := <-stream:
				if !ok {
					return
				}
				fmt.Println(d)

			default:
				time.Sleep(time.Second)
			}
		}
	}
}
