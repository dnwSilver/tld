package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/dnwSilver/tld/internal/ui"
)

func sendJobMsg[T any](ctx context.Context, ch chan<- T, msg T) {
	select {
	case ch <- msg:
	case <-ctx.Done():
	}
}

type projectJobEvent struct {
	message string
	err     error
	done    bool
	step    bool
	current int
	total   int
}

type batchResult struct {
	processed int
	failures  []string
}

func runBatch[T any](ctx context.Context, items []T, name func(T) string, kind, completeMessage string, run func(context.Context, T, func(string)) error, emit func(projectJobEvent)) batchResult {
	return runBatchConcurrent(ctx, items, 1, name, kind, completeMessage, run, emit)
}

func runBatchConcurrent[T any](ctx context.Context, items []T, concurrency int, name func(T) string, kind, completeMessage string, run func(context.Context, T, func(string)) error, emit func(projectJobEvent)) batchResult {
	if concurrency < 1 {
		concurrency = 1
	}
	result := batchResult{failures: make([]string, 0)}
	total := len(items)
	if total == 0 {
		if emit != nil && ctx.Err() == nil {
			emit(projectJobEvent{message: completeMessage, done: true})
		}
		return result
	}
	concurrency = min(concurrency, total)
	type batchItem struct{ value T }
	jobs := make(chan batchItem)
	var completed atomic.Int64
	var processed atomic.Int64
	var failureMu sync.Mutex
	var emitMu sync.Mutex
	emitEvent := func(event projectJobEvent) {
		if emit == nil {
			return
		}
		emitMu.Lock()
		emit(event)
		emitMu.Unlock()
	}
	var workers sync.WaitGroup
	workers.Add(concurrency)
	for range concurrency {
		go func() {
			defer workers.Done()
			for job := range jobs {
				if ctx.Err() != nil {
					return
				}
				itemName := name(job.value)
				progress := func(message string) {
					emitEvent(projectJobEvent{message: itemName + ": " + message, current: int(completed.Load()), total: total})
				}
				err := run(ctx, job.value, progress)
				processed.Add(1)
				current := int(completed.Add(1))
				if err != nil {
					failureMu.Lock()
					result.failures = append(result.failures, itemName+": "+err.Error())
					failureMu.Unlock()
					emitEvent(projectJobEvent{message: itemName + ": " + err.Error(), step: true, current: current, total: total})
				} else {
					emitEvent(projectJobEvent{message: itemName, step: true, current: current, total: total})
				}
			}
		}()
	}
	canceled := false
	for _, item := range items {
		select {
		case jobs <- batchItem{value: item}:
		case <-ctx.Done():
			canceled = true
		}
		if canceled {
			break
		}
	}
	close(jobs)
	workers.Wait()
	result.processed = int(processed.Load())
	if emit == nil || ctx.Err() != nil {
		return result
	}
	if len(result.failures) > 0 {
		emit(projectJobEvent{
			message: fmt.Sprintf("%s finished: %d/%d failed", kind, len(result.failures), total),
			err:     errors.New(strings.Join(result.failures, "; ")), done: true, current: total, total: total,
		})
		return result
	}
	emit(projectJobEvent{message: completeMessage, done: true, current: total, total: total})
	return result
}

// runProjectBatch owns per-project progress and error aggregation. Individual
// workflows supply only their project operation and message adapter.
func runProjectBatch(ctx context.Context, projects []ui.Project, kind, completeMessage string, run func(context.Context, ui.Project, func(string)) error, emit func(projectJobEvent)) {
	runBatchConcurrent(ctx, projects, 4, func(project ui.Project) string { return project.Name }, kind, completeMessage, run, emit)
}
