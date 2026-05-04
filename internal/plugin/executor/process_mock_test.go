package executor

import (
	"context"
	"io"
)

// MockProcessRunner is a test-only ProcessRunner that returns predetermined
// behaviour. Lives in a _test.go file so it doesn't ship with production
// builds.
type MockProcessRunner struct {
	ShouldTimeout bool
	CallCount     int
	LastPath      string
	LastArgs      []string
}

// Run executes the mock behaviour.
func (m *MockProcessRunner) Run(ctx context.Context, path string, args []string, _ io.Reader) (stdout, stderr []byte, err error) {
	m.CallCount++
	m.LastPath = path
	m.LastArgs = args

	if m.ShouldTimeout {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}
	return []byte("{}"), nil, nil
}

// NewTimeoutMockProcessRunner creates a mock that simulates a timeout by
// blocking until the supplied context is cancelled. The only constructor
// the test suite still uses; the prior NewMockProcessRunner /
// NewDelayMockProcessRunner / NewErrorMockProcessRunner /
// NewSuccessMockProcessRunner helpers were unused and dropped with the
// rest of the dead-code sweep.
func NewTimeoutMockProcessRunner() *MockProcessRunner {
	return &MockProcessRunner{ShouldTimeout: true}
}
