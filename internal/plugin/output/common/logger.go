// Package common provides shared utilities for output plugins.
package common

import (
	"fmt"
	"io"
)

// VerboseLogger implements a simple logger that writes to an io.Writer.
// This is used by output plugins that need to log verbose information
// during template processing or generation.
type VerboseLogger struct {
	out io.Writer
}

// NewVerboseLogger creates a new VerboseLogger that writes to the given writer.
func NewVerboseLogger(out io.Writer) *VerboseLogger {
	return &VerboseLogger{out: out}
}

// Printf writes a formatted message to the logger's output writer.
// A newline is automatically appended to the format string.
func (l *VerboseLogger) Printf(format string, v ...any) {
	fmt.Fprintf(l.out, format+"\n", v...)
}
