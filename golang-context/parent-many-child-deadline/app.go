package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func parent(c context.Context) ([]int, error) {
	ctx, cancel := context.WithTimeout(c, time.Millisecond*1000) // 1000ms, 下游并发rpc不可超过1000ms
	defer cancel()
	arr := make([]int, 10)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < len(arr); i++ {
		go child(ctx, &arr[i], func() {
			wg.Done()
		})
	}
	wg.Wait()
	return arr, ctx.Err()
}

func child(ctx context.Context, i *int, doneWgFunc func()) {
	defer doneWgFunc()
	rand.Seed(time.Now().UnixNano())
	timeout := rand.Intn(1500)
	select {
	case <-ctx.Done():
		fmt.Println(*i, ctx.Err())
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		*i = timeout
	}
}

func main() {
	// 开启parent goroutine
	// 开启10个child goroutine，每个响应时间在100ms到300ms内随机
	ctx := context.Background()
	arr, err := parent(ctx)
	fmt.Println(arr)
	fmt.Println(err)
}
