package foundry

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	stepLoggerMu sync.RWMutex
	stepLogger   = log.New(os.Stdout, "[foundry] ", log.LstdFlags)
)

// WithStep logs the beginning and completion of a named operation. It always
// executes fn and returns its error, ensuring consistent timing output.
func WithStep(name string, fn func() error) error {
	if name == "" {
		return errors.New("foundry: step name is empty")
	}
	if fn == nil {
		return errors.New("foundry: fn is nil")
	}

	logf("start %s", name)
	start := time.Now()
	err := fn()
	duration := time.Since(start).Round(time.Millisecond)
	if err != nil {
		logf("error %s (%s): %v", name, duration, err)
		return err
	}
	logf("done %s (%s)", name, duration)
	return nil
}

// SetStepLogger allows tests to redirect WithStep logging. Production callers
// do not generally need this.
func SetStepLogger(l *log.Logger) {
	stepLoggerMu.Lock()
	defer stepLoggerMu.Unlock()
	if l == nil {
		stepLogger = log.New(os.Stdout, "[foundry] ", log.LstdFlags)
		return
	}
	stepLogger = l
}

func logf(format string, args ...any) {
	stepLoggerMu.RLock()
	defer stepLoggerMu.RUnlock()
	_ = stepLogger.Output(2, fmt.Sprintf(format, args...))
}
