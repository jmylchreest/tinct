// Package progress provides a terminal progress bar.
package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

// isTerminal checks if the writer is connected to a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd())) // #nosec G115 -- file descriptors are small positive integers, safe to convert uintptr→int
	}
	return false
}

// ProgressBar displays a progress bar with percentage.
type ProgressBar struct {
	total   int64
	current int64
	width   int
	prefix  string
	writer  io.Writer
	mu      sync.Mutex
	lastLen int
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(total int64, prefix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  40,
		prefix: prefix,
		writer: os.Stderr,
	}
}

// Set updates the current progress value.
func (p *ProgressBar) Set(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.render()
}

// Finish completes the progress bar and shows 100%.
func (p *ProgressBar) Finish(finalMessage string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	p.render()

	// Clear the progress bar line only if terminal.
	if isTerminal(p.writer) {
		fmt.Fprintf(p.writer, "\r%s\r", strings.Repeat(" ", p.lastLen))
	}

	if finalMessage != "" {
		fmt.Fprintf(p.writer, "%s\n", finalMessage)
	}
}

// render draws the progress bar.
func (p *ProgressBar) render() {
	var percent float64
	if p.total > 0 {
		percent = float64(p.current) / float64(p.total) * 100
	}

	filled := min(int(float64(p.width)*(float64(p.current)/float64(p.total))), p.width)

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	var sizeStr string
	if p.total > 0 {
		sizeStr = fmt.Sprintf(" %s/%s", formatBytes(p.current), formatBytes(p.total))
	}

	if !isTerminal(p.writer) {
		return
	}

	line := fmt.Sprintf("\r%s [%s] %.1f%%%s", p.prefix, bar, percent, sizeStr)
	p.lastLen = len(line)
	fmt.Fprint(p.writer, line)
}

// formatBytes formats bytes into human-readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
