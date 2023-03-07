package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnzhanghua/workerpool/pool"
)

func main() {
	e := make(chan os.Signal)
	signal.Notify(e, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	jobs := []any{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p := pool.NewJobPool(ctx, jobs, 2, sqrTimeout)
	err := p.Process()
	fmt.Printf("values: %v, error:%v\n", p.Results(), err)

	<-e
	os.Exit(1)
}

func sqrTimeout(ctx context.Context, i any) (any, error) {
	d, ok := i.(int)
	if !ok {
		return nil, fmt.Errorf("invalid value: %v, %w", i, pool.ErrInvalidValue)
	}
	time.Sleep(6 * time.Second)
	return any(d * d), nil
}
