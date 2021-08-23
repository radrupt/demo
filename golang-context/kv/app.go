package main

import (
	"context"
	"fmt"
)

func child(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	child1(ctx, 1)
	go child1(ctx, 2)
}

func child1(ctx context.Context, index int) {
	fmt.Println(index, ctx.Value("name"))
}

func main() {
	ctx := context.WithValue(context.Background(), "name", "王迪")
	child(ctx)
}
