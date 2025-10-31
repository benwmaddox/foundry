package foundry

import (
	"errors"
	"fmt"
	"sync"
)

// ForEachParallel applies fn to every item using up to workers goroutines. It
// returns an error if workers is less than one, fn is nil, or if fn panics.
func ForEachParallel[T any](items []T, workers int, fn func(T)) error {
	if fn == nil {
		return errors.New("foundry: fn is nil")
	}
	if workers <= 0 {
		return fmt.Errorf("foundry: workers must be positive (got %d)", workers)
	}
	if len(items) == 0 {
		return nil
	}
	if workers > len(items) {
		workers = len(items)
	}

	jobs := make(chan T)
	var wg sync.WaitGroup
	wg.Add(workers)

	var once sync.Once
	var firstErr error
	signalError := func(err error) {
		once.Do(func() {
			firstErr = err
		})
	}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for item := range jobs {
				func(it T) {
					defer func() {
						if r := recover(); r != nil {
							signalError(fmt.Errorf("foundry: panic in parallel worker: %v", r))
						}
					}()
					fn(it)
				}(item)
			}
		}()
	}

	for _, item := range items {
		if firstErr != nil {
			break
		}
		jobs <- item
	}
	close(jobs)
	wg.Wait()

	return firstErr
}
