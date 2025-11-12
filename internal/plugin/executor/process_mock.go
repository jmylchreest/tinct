package executor

import (
	"context"
	"errors"
	"io"
	"time"
)

// MockProcessRunner is a mock implementation of ProcessRunner for testing.
type MockProcessRunner struct {
	// RunFunc allows tests to provide custom behavior
	RunFunc func(ctx context.Context, path string, args []string, stdin io.Reader) (stdout, stderr []byte, err error)

	// Delay simulates slow process execution
	Delay time.Duration

	// ShouldTimeout if true, will block until context is cancelled
	ShouldTimeout bool

	// CallCount tracks how many times Run was called
	CallCount int

	// LastPath stores the last path passed to Run
	LastPath string

	// LastArgs stores the last args passed to Run
	LastArgs []string
}

// Run executes the mock behavior.
func (m *MockProcessRunner) Run(ctx context.Context, path string, args []string, stdin io.Reader) ([]byte, []byte, error) {
	m.CallCount++
	m.LastPath = path
	m.LastArgs = args

	// Simulate timeout behavior
	if m.ShouldTimeout {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}

	// Simulate delay
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}

	// Use custom function if provided
	if m.RunFunc != nil {
		return m.RunFunc(ctx, path, args, stdin)
	}

	// Default: return empty success
	return []byte("{}"), nil, nil
}

// NewMockProcessRunner creates a new mock process runner.
func NewMockProcessRunner() *MockProcessRunner {
	return &MockProcessRunner{}
}

// NewTimeoutMockProcessRunner creates a mock that simulates a timeout.
func NewTimeoutMockProcessRunner() *MockProcessRunner {
	return &MockProcessRunner{
		ShouldTimeout: true,
	}
}

// NewDelayMockProcessRunner creates a mock that simulates a slow process.
func NewDelayMockProcessRunner(delay time.Duration) *MockProcessRunner {
	return &MockProcessRunner{
		Delay: delay,
	}
}

// NewErrorMockProcessRunner creates a mock that returns an error.
func NewErrorMockProcessRunner(errMsg string) *MockProcessRunner {
	return &MockProcessRunner{
		RunFunc: func(ctx context.Context, path string, args []string, stdin io.Reader) ([]byte, []byte, error) {
			return nil, []byte(errMsg), errors.New(errMsg)
		},
	}
}

// NewSuccessMockProcessRunner creates a mock that returns successful JSON output.
func NewSuccessMockProcessRunner(stdout []byte) *MockProcessRunner {
	return &MockProcessRunner{
		RunFunc: func(ctx context.Context, path string, args []string, stdin io.Reader) ([]byte, []byte, error) {
			return stdout, nil, nil
		},
	}
}
