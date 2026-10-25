package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if in == nil {
		return nil, nil
	}
	var wg sync.WaitGroup
	ch := make(chan int)
	results := make([]Result[R], len(in))
	for range workers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					res, err := fn(ctx, in[v])
					if err != nil {
						results[v].Err = err
					} else {
						results[v] = Result[R]{Value: res, Err: err}
					}
				}
			}
		})
	}
index_loop:
	for i := range in {
		select {
		case <-ctx.Done():
			break index_loop
		case ch <- i:
		}
	}
	close(ch)
	wg.Wait()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return results, nil

}
