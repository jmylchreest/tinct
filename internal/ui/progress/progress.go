// Package progress provides terminal progress indicators (spinners, bars, status).
package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Spinner characters for different animation styles.
var (
	SpinnerDots    = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine    = []string{"-", "\\", "|", "/"}
	SpinnerArrow   = []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	SpinnerCircle  = []string{"◐", "◓", "◑", "◒"}
	SpinnerDefault = SpinnerDots
)

// isTerminal checks if the writer is connected to a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Spinner displays an animated spinner with a message.
type Spinner struct {
	message  string
	frames   []string
	interval time.Duration
	writer   io.Writer
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	active   bool
}

// NewSpinner creates a new spinner with the default animation.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   SpinnerDefault,
		interval: 80 * time.Millisecond,
		writer:   os.Stderr,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// WithFrames sets custom spinner frames.
func (s *Spinner) WithFrames(frames []string) *Spinner {
	s.frames = frames
	return s
}

// WithInterval sets the animation interval.
func (s *Spinner) WithInterval(d time.Duration) *Spinner {
	s.interval = d
	return s
}

// WithWriter sets the output writer.
func (s *Spinner) WithWriter(w io.Writer) *Spinner {
	s.writer = w
	return s
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		defer close(s.done)

		// Only show animation if writing to a terminal
		isTerm := isTerminal(s.writer)

		frame := 0
		for {
			select {
			case <-s.stop:
				// Clear the line only if terminal
				if isTerm {
					fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", len(s.message)+3))
				}
				return
			case <-ticker.C:
				// Only render animation if terminal
				if isTerm {
					fmt.Fprintf(s.writer, "\r%s %s", s.frames[frame], s.message)
					frame = (frame + 1) % len(s.frames)
				}
			}
		}
	}()
}

// UpdateMessage changes the spinner message while it's running.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Stop stops the spinner and optionally shows a final message.
func (s *Spinner) Stop(finalMessage string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	close(s.stop)
	<-s.done

	if finalMessage != "" {
		fmt.Fprintf(s.writer, "%s\n", finalMessage)
	}

	s.mu.Lock()
	s.active = false
	s.mu.Unlock()
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

// WithWidth sets the progress bar width.
func (p *ProgressBar) WithWidth(width int) *ProgressBar {
	p.width = width
	return p
}

// WithWriter sets the output writer.
func (p *ProgressBar) WithWriter(w io.Writer) *ProgressBar {
	p.writer = w
	return p
}

// Set updates the current progress value.
func (p *ProgressBar) Set(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.render()
}

// Add increments the progress by delta.
func (p *ProgressBar) Add(delta int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += delta
	if p.current > p.total {
		p.current = p.total
	}
	p.render()
}

// Finish completes the progress bar and shows 100%.
func (p *ProgressBar) Finish(finalMessage string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	p.render()

	// Clear the progress bar line only if terminal
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

	filled := int(float64(p.width) * (float64(p.current) / float64(p.total)))
	if filled > p.width {
		filled = p.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	var sizeStr string
	if p.total > 0 {
		sizeStr = fmt.Sprintf(" %s/%s", formatBytes(p.current), formatBytes(p.total))
	}

	// Only render if writing to a terminal
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

// Status represents a simple status line that can be updated.
type Status struct {
	message string
	writer  io.Writer
	mu      sync.Mutex
	lastLen int
}

// NewStatus creates a new status line.
func NewStatus(message string) *Status {
	return &Status{
		message: message,
		writer:  os.Stderr,
	}
}

// WithWriter sets the output writer.
func (s *Status) WithWriter(w io.Writer) *Status {
	s.writer = w
	return s
}

// Update changes the status message.
func (s *Status) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.message = message

	// Only render if writing to a terminal
	if isTerminal(s.writer) {
		// Clear previous line and write new message
		fmt.Fprintf(s.writer, "\r%s\r%s", strings.Repeat(" ", s.lastLen), message)
		s.lastLen = len(message)
	}
}

// Clear removes the status line.
func (s *Status) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Only clear if writing to a terminal
	if isTerminal(s.writer) {
		fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", s.lastLen))
		s.lastLen = 0
	}
}

// Finish clears the status line and optionally prints a final message.
func (s *Status) Finish(finalMessage string) {
	s.Clear()
	if finalMessage != "" {
		fmt.Fprintln(s.writer, finalMessage)
	}
}
