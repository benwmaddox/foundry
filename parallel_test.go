package foundry

import (
	"fmt"
	"sync"
	"testing"
)

func TestForEachParallel(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}
	var (
		mu   sync.Mutex
		sum  int
		calls []int
	)

	err := ForEachParallel(items, 3, func(v int) {
		mu.Lock()
		sum += v
		calls = append(calls, v)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("ForEachParallel returned error: %v", err)
	}

	if sum != 21 {
		t.Fatalf("sum got %d want 21", sum)
	}
	if len(calls) != len(items) {
		t.Fatalf("expected %d calls, got %d", len(items), len(calls))
	}
}

func TestForEachParallelPanic(t *testing.T) {
	items := []int{1, 2, 3}
	err := ForEachParallel(items, 2, func(v int) {
		if v == 2 {
			panic("boom")
		}
	})
	if err == nil || err.Error() != "foundry: panic in parallel worker: boom" {
		t.Fatalf("expected panic error, got %v", err)
	}
}

func TestForEachParallelInvalidArgs(t *testing.T) {
	err := ForEachParallel([]int{1}, 0, func(i int) {})
	if err == nil {
		t.Fatalf("expected error for workers <= 0")
	}

	err = ForEachParallel([]int{1}, 1, nil)
	if err == nil {
		t.Fatalf("expected error for nil fn")
	}

	err = ForEachParallel([]int{}, 1, func(i int) {})
	if err != nil {
		t.Fatalf("expected no error for empty slice, got %v", err)
	}

	if err := ForEachParallel([]int{1}, 1, func(i int) {
		panic(fmt.Errorf("wrapped"))
	}); err == nil {
		t.Fatalf("expected panic to be reported")
	}
}
