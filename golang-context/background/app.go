package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(time.Second)*3))
	defer cancel()
	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	case <-time.After(time.Duration(time.Second) * 1):
		fmt.Println("finish task")
	}
}
