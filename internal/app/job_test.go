package app

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dnwSilver/tld/internal/ui"
)

func TestProjectBatchContinuesAfterOneFailure(t *testing.T) {
	projects := []ui.Project{{ID: 1, Name: "one"}, {ID: 2, Name: "two"}, {ID: 3, Name: "three"}}
	visited := make([]int64, 0, len(projects))
	events := make([]projectJobEvent, 0)
	var mu sync.Mutex
	runProjectBatch(context.Background(), projects, "Sync", "Sync complete", func(_ context.Context, project ui.Project, progress func(string)) error {
		mu.Lock()
		visited = append(visited, project.ID)
		mu.Unlock()
		progress("reading")
		if project.ID == 2 {
			return errors.New("denied")
		}
		return nil
	}, func(event projectJobEvent) {
		events = append(events, event)
	})
	if len(visited) != 3 {
		t.Fatalf("visited = %#v", visited)
	}
	steps := 0
	for _, event := range events {
		if event.step {
			steps++
		}
	}
	final := events[len(events)-1]
	if steps != 3 || !final.done || final.current != 3 || final.total != 3 || final.err == nil || !strings.Contains(final.err.Error(), "two: denied") {
		t.Fatalf("events = %#v", events)
	}
}

func TestRunBatchConcurrentRespectsLimit(t *testing.T) {
	var active atomic.Int64
	var peak atomic.Int64
	items := make([]int, 20)
	runBatchConcurrent(context.Background(), items, 4, func(int) string { return "item" }, "test", "done", func(context.Context, int, func(string)) error {
		current := active.Add(1)
		for {
			old := peak.Load()
			if current <= old || peak.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		active.Add(-1)
		return nil
	}, nil)
	if got := peak.Load(); got < 2 || got > 4 {
		t.Fatalf("peak concurrency = %d, want 2..4", got)
	}
}

func TestPolicyPinsJobCanBeCanceled(t *testing.T) {
	m := newModel(nil)
	m.screen = ui.ScreenPolicies
	m.policyUpdateStatus = ui.SettingsStatus{Running: true}
	ctx, cancel := context.WithCancel(context.Background())
	m.policyPinsCancel = cancel
	cmd, ok := m.cancelScreenJob()
	if !ok || cmd == nil || m.policyUpdateStatus.Running || m.policyUpdateStatus.Message != "Pins canceled" {
		t.Fatalf("cancel result: ok=%v status=%#v", ok, m.policyUpdateStatus)
	}
	if ctx.Err() != context.Canceled {
		t.Fatalf("job context = %v", ctx.Err())
	}
}

func TestBatchResultCountsProcessedItemsAndFailures(t *testing.T) {
	result := runBatch(context.Background(), []string{"first", "second", "third"}, func(item string) string { return item }, "Pins", "done", func(_ context.Context, item string, _ func(string)) error {
		if item == "second" {
			return errors.New("registry timeout")
		}
		return nil
	}, nil)
	if result.processed != 3 || len(result.failures) != 1 || result.failures[0] != "second: registry timeout" {
		t.Fatalf("result = %#v", result)
	}
}
