package main

import (
	"context"
	"errors"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWorkerPool_Empty(t *testing.T) {
	results := runWorkerPool(context.Background(), []string{}, 5, func(s string) error {
		return nil
	})
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty jobs, got %d", len(results))
	}
}

func TestRunWorkerPool_AllSuccess(t *testing.T) {
	jobs := []string{"a", "b", "c", "d", "e"}
	results := runWorkerPool(context.Background(), jobs, 3, func(s string) error {
		return nil
	})
	if len(results) != len(jobs) {
		t.Fatalf("expected %d results, got %d", len(jobs), len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for job %q: %v", r.Job, r.Err)
		}
	}
}

func TestRunWorkerPool_AllFail(t *testing.T) {
	jobs := []int{1, 2, 3}
	sentinel := errors.New("always fails")
	results := runWorkerPool(context.Background(), jobs, 2, func(_ int) error {
		return sentinel
	})
	if len(results) != len(jobs) {
		t.Fatalf("expected %d results, got %d", len(jobs), len(results))
	}
	for _, r := range results {
		if r.Err == nil {
			t.Errorf("expected error for job %d, got nil", r.Job)
		}
	}
}

func TestRunWorkerPool_PartialFailure(t *testing.T) {
	jobs := []int{1, 2, 3, 4, 5}
	fail := errors.New("fail")
	results := runWorkerPool(context.Background(), jobs, 2, func(n int) error {
		if n%2 == 0 {
			return fail
		}
		return nil
	})

	if len(results) != len(jobs) {
		t.Fatalf("expected %d results, got %d", len(jobs), len(results))
	}

	var succeeded, failed int
	for _, r := range results {
		if r.Err != nil {
			failed++
		} else {
			succeeded++
		}
	}
	if succeeded != 3 || failed != 2 {
		t.Errorf("expected 3 success + 2 failure, got %d + %d", succeeded, failed)
	}
}

func TestRunWorkerPool_AllJobsProcessed(t *testing.T) {
	const n = 50
	jobs := make([]int, n)
	for i := range jobs {
		jobs[i] = i
	}

	var processed atomic.Int64
	results := runWorkerPool(context.Background(), jobs, 10, func(job int) error {
		processed.Add(1)
		return nil
	})

	if len(results) != n {
		t.Fatalf("expected %d results, got %d", n, len(results))
	}
	if int(processed.Load()) != n {
		t.Errorf("expected %d processed jobs, got %d", n, processed.Load())
	}
}

func TestRunWorkerPool_WorkerCountCappedAtJobs(t *testing.T) {
	// More workers than jobs should not panic or deadlock
	jobs := []string{"only-one"}
	results := runWorkerPool(context.Background(), jobs, 100, func(s string) error {
		return nil
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestRunWorkerPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create many slow jobs
	jobs := make([]int, 20)
	for i := range jobs {
		jobs[i] = i
	}

	var started atomic.Int64
	done := make(chan struct{})

	go func() {
		runWorkerPool(ctx, jobs, 2, func(job int) error {
			started.Add(1)
			// Cancel after first job starts
			if started.Load() == 1 {
				cancel()
			}
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		close(done)
	}()

	select {
	case <-done:
		// All results collected (cancelled jobs return ctx.Err(), not hang)
	case <-time.After(3 * time.Second):
		t.Fatal("worker pool did not finish after context cancellation")
	}
}

func TestRunWorkerPool_ResultsContainAllJobs(t *testing.T) {
	jobs := []string{"x", "y", "z"}
	results := runWorkerPool(context.Background(), jobs, 2, func(s string) error {
		return nil
	})

	got := make([]string, len(results))
	for i, r := range results {
		got[i] = r.Job
	}
	sort.Strings(got)
	sort.Strings(jobs)

	for i := range jobs {
		if jobs[i] != got[i] {
			t.Errorf("result set mismatch at %d: want %q got %q", i, jobs[i], got[i])
		}
	}
}
