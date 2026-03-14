package main

import (
	"context"
	"sync"
)

// workerResult holds the outcome of a single worker job.
type workerResult[J any] struct {
	Job J
	Err error
}

// runWorkerPool executes fn for every job in jobs, using up to maxWorkers goroutines.
// When ctx is cancelled any unstarted job is skipped and marked with ctx.Err().
// Results are returned in completion order (non-deterministic relative to input order).
func runWorkerPool[J any](ctx context.Context, jobs []J, maxWorkers int, fn func(J) error) []workerResult[J] {
	n := len(jobs)
	if n == 0 {
		return nil
	}

	workers := maxWorkers
	if workers > n {
		workers = n
	}
	if workers < 1 {
		workers = 1
	}

	jobCh := make(chan J, n)
	resCh := make(chan workerResult[J], n)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				// Honour context cancellation: skip remaining jobs instead of
				// launching new sub-processes that will fail immediately anyway.
				if ctx.Err() != nil {
					resCh <- workerResult[J]{Job: job, Err: ctx.Err()}
					continue
				}
				resCh <- workerResult[J]{Job: job, Err: fn(job)}
			}
		}()
	}

	// Feed jobs
	go func() {
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)
	}()

	// Close result channel once all workers finish
	go func() {
		wg.Wait()
		close(resCh)
	}()

	results := make([]workerResult[J], 0, n)
	for r := range resCh {
		results = append(results, r)
	}
	return results
}
