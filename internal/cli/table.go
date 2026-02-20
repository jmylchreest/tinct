// Package cli provides command-line interface utilities.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

// Table represents a simple table formatter with dynamic column widths.
type Table struct {
	headers               []string
	rows                  [][]string
	padding               int
	maxWidths             map[int]int // Maximum width per column index (0 = no limit)
	terminalAwareCol      int         // Column index to size based on terminal width (-1 = none)
	terminalAwareMinW     int         // Minimum width for terminal-aware column
	terminalWidthOverride int         // Override terminal width for testing

	// Live mode support
	live      bool           // If true, updates in-place; if false, static output
	isTTY     bool           // Whether the output is a TTY (supports ANSI codes)
	batching  bool           // If true, defer rendering until EndBatch() is called
	rowIDs    map[string]int // Maps row ID to row index (for live mode)
	rowOrder  []string       // Preserves insertion order of row IDs
	writer    io.Writer      // Output writer (defaults to os.Stderr for live mode)
	rendered  bool           // Whether table has been rendered
	lineCount int            // Number of lines in last render
	mu        sync.Mutex     // Protects concurrent access in live mode
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	return &Table{
		headers:           headers,
		rows:              make([][]string, 0),
		padding:           2, // 2 spaces between columns
		maxWidths:         make(map[int]int),
		terminalAwareCol:  -1, // Disabled by default
		terminalAwareMinW: 0,
		live:              false,
		rowIDs:            make(map[string]int),
		rowOrder:          make([]string, 0),
		writer:            os.Stderr,
	}
}

// SetLive enables or disables live updating mode.
// In live mode, the table collects updates and renders the final state at Finish().
// Intermediate updates are not rendered to avoid terminal compatibility issues.
func (t *Table) SetLive(live bool) *Table {
	t.live = live
	// Check if writer is a TTY (supports ANSI escape codes)
	if live {
		t.isTTY = isWriterTTY(t.writer)
		// In live mode, always batch updates until Finish() is called
		t.batching = true
	}
	return t
}

// WithWriter sets the output writer for the table.
func (t *Table) WithWriter(w io.Writer) *Table {
	t.writer = w
	// Re-check TTY status if live mode is enabled
	if t.live {
		t.isTTY = isWriterTTY(w)
	}
	return t
}

// StartBatch begins a batch operation where rendering is deferred.
// Call EndBatch() to render the accumulated changes.
func (t *Table) StartBatch() *Table {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.batching = true
	return t
}

// EndBatch ends a batch operation and renders the table.
func (t *Table) EndBatch() *Table {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.batching = false
	if t.live {
		t.renderLive()
	}
	return t
}

// SetColumnMaxWidth sets a maximum width for a specific column.
// Text longer than this will be wrapped to multiple lines.
func (t *Table) SetColumnMaxWidth(colIndex, maxWidth int) {
	t.maxWidths[colIndex] = maxWidth
}

// EnableTerminalAwareWidth enables terminal-aware width calculation for a column.
// The specified column will size to fit available terminal width (after other columns).
// minWidth specifies the minimum width for the column.
func (t *Table) EnableTerminalAwareWidth(colIndex, minWidth int) {
	t.terminalAwareCol = colIndex
	t.terminalAwareMinW = minWidth
}

// AddRow adds a row to the table.
// For static tables, just adds the row.
// For live tables, you should use AddRowWithID instead.
func (t *Table) AddRow(row []string) {
	if t.live {
		t.mu.Lock()
		defer t.mu.Unlock()
	}

	normalizedRow := t.normalizeRow(row)
	t.rows = append(t.rows, normalizedRow)

	if t.live && !t.batching {
		t.renderLive()
	}
}

// AddRowWithID adds or updates a row in the table with a unique identifier.
// This is the preferred method for live tables.
// id is a unique identifier for the row.
// row is the row data matching the headers.
func (t *Table) AddRowWithID(id string, row []string) {
	if t.live {
		t.mu.Lock()
		defer t.mu.Unlock()
	}

	normalizedRow := t.normalizeRow(row)

	if idx, exists := t.rowIDs[id]; exists {
		// Update existing row
		t.rows[idx] = normalizedRow
	} else {
		// Add new row
		t.rowIDs[id] = len(t.rows)
		t.rowOrder = append(t.rowOrder, id)
		t.rows = append(t.rows, normalizedRow)
	}

	if t.live && !t.batching {
		t.renderLive()
	}
}

// UpdateRow updates specific columns in an existing row identified by ID.
// columns is a map of column name -> new value.
func (t *Table) UpdateRow(id string, columns map[string]string) {
	if t.live {
		t.mu.Lock()
		defer t.mu.Unlock()
	}

	idx, exists := t.rowIDs[id]
	if !exists {
		// Create new row if it doesn't exist
		newRow := make([]string, len(t.headers))
		for i, header := range t.headers {
			if val, ok := columns[header]; ok {
				newRow[i] = val
			}
		}
		t.rowIDs[id] = len(t.rows)
		t.rowOrder = append(t.rowOrder, id)
		t.rows = append(t.rows, newRow)
	} else {
		// Update existing row
		for i, header := range t.headers {
			if val, ok := columns[header]; ok {
				t.rows[idx][i] = val
			}
		}
	}

	if t.live && !t.batching {
		t.renderLive()
	}
}

// normalizeRow pads or truncates a row to match header count.
func (t *Table) normalizeRow(row []string) []string {
	if len(row) == len(t.headers) {
		return row
	}

	newRow := make([]string, len(t.headers))
	copy(newRow, row)
	return newRow
}

// Render formats and returns the table as a string (for static mode).
// For live mode, use Finish() to complete the table.
func (t *Table) Render() string {
	if t.live {
		// For live mode, just ensure it's rendered once more
		t.mu.Lock()
		defer t.mu.Unlock()
		t.renderLive()
		return ""
	}

	if len(t.headers) == 0 {
		return ""
	}

	// First pass: wrap cells with existing constraints
	wrappedRows := t.wrapAllCells()

	// Calculate column widths (includes terminal-aware adjustments)
	colWidths := t.calculateColumnWidths(wrappedRows)

	// Second pass: re-wrap cells if terminal-aware width changed things
	if t.terminalAwareCol >= 0 {
		wrappedRows = t.wrapAllCells()
	}

	// Build the table string
	return t.buildTableString(wrappedRows, colWidths)
}

// renderLive renders the table in live mode (must be called with lock held).
func (t *Table) renderLive() {
	if len(t.headers) == 0 {
		return
	}

	// If not a TTY, only render once (no cursor control available)
	if !t.isTTY {
		if !t.rendered {
			t.renderStatic()
			t.rendered = true
		}
		return
	}

	// Move cursor up to overwrite previous render (if already rendered)
	if t.rendered {
		fmt.Fprintf(t.writer, "\033[%dA", t.lineCount)
	}

	// For live mode, we use simpler rendering without text wrapping
	// to avoid complexity with cursor positioning
	colWidths := t.calculateSimpleColumnWidths()

	lineCount := 0

	// Render header
	headerParts := make([]string, len(t.headers))
	separatorParts := make([]string, len(t.headers))
	for i, header := range t.headers {
		headerParts[i] = padRight(header, colWidths[i])
		separatorParts[i] = strings.Repeat("-", colWidths[i])
	}

	// Use \033[K to clear to end of line for clean overwrites
	fmt.Fprintf(t.writer, "\r%s\033[K\n", strings.Join(headerParts, strings.Repeat(" ", t.padding)))
	lineCount++
	fmt.Fprintf(t.writer, "\r%s\033[K\n", strings.Join(separatorParts, strings.Repeat(" ", t.padding)))
	lineCount++

	// Render rows
	for _, row := range t.rows {
		rowParts := make([]string, len(t.headers))
		for i, cell := range row {
			rowParts[i] = padRight(cell, colWidths[i])
		}
		fmt.Fprintf(t.writer, "\r%s\033[K\n", strings.Join(rowParts, strings.Repeat(" ", t.padding)))
		lineCount++
	}

	// Clear any extra lines from previous render
	if t.rendered {
		for i := lineCount; i < t.lineCount; i++ {
			fmt.Fprintf(t.writer, "\r\033[K\n")
			lineCount++
		}
	}

	t.lineCount = lineCount
	t.rendered = true
}

// renderStatic renders the table once without cursor control (for non-TTY output).
func (t *Table) renderStatic() {
	colWidths := t.calculateSimpleColumnWidths()

	// Render header
	headerParts := make([]string, len(t.headers))
	separatorParts := make([]string, len(t.headers))
	for i, header := range t.headers {
		headerParts[i] = padRight(header, colWidths[i])
		separatorParts[i] = strings.Repeat("-", colWidths[i])
	}

	fmt.Fprintf(t.writer, "%s\n", strings.Join(headerParts, strings.Repeat(" ", t.padding)))
	fmt.Fprintf(t.writer, "%s\n", strings.Join(separatorParts, strings.Repeat(" ", t.padding)))

	// Render rows
	for _, row := range t.rows {
		rowParts := make([]string, len(t.headers))
		for i, cell := range row {
			rowParts[i] = padRight(cell, colWidths[i])
		}
		fmt.Fprintf(t.writer, "%s\n", strings.Join(rowParts, strings.Repeat(" ", t.padding)))
	}
}

// calculateSimpleColumnWidths calculates column widths without text wrapping.
func (t *Table) calculateSimpleColumnWidths() []int {
	colWidths := make([]int, len(t.headers))

	// Start with header widths
	for i, header := range t.headers {
		colWidths[i] = len(header)
	}

	// Update based on row content
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	return colWidths
}

// Finish completes the table rendering.
// For live mode, renders the final state (all updates were collected without rendering).
// For static mode, does nothing (use Render() instead).
func (t *Table) Finish() {
	if !t.live {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Render the final state - use static rendering to avoid terminal compatibility issues
	if !t.rendered {
		t.renderStatic()
		t.rendered = true
	}
}

// wrapAllCells wraps text in all cells according to max width constraints.
func (t *Table) wrapAllCells() [][][]string {
	wrappedRows := make([][][]string, len(t.rows))
	for rowIdx, row := range t.rows {
		wrappedRows[rowIdx] = make([][]string, len(row))
		for colIdx, cell := range row {
			wrappedRows[rowIdx][colIdx] = t.wrapCell(cell, colIdx)
		}
	}
	return wrappedRows
}

// wrapCell wraps a single cell's text based on column max width.
func (t *Table) wrapCell(cell string, colIdx int) []string {
	if maxWidth, hasLimit := t.maxWidths[colIdx]; hasLimit && maxWidth > 0 {
		return wrapText(cell, maxWidth)
	}
	return []string{cell}
}

// calculateColumnWidths determines the width of each column.
func (t *Table) calculateColumnWidths(wrappedRows [][][]string) []int {
	colWidths := t.getHeaderWidths()
	t.updateWidthsFromContent(colWidths, wrappedRows)

	// Apply terminal-aware width if configured
	if t.terminalAwareCol >= 0 && t.terminalAwareCol < len(colWidths) {
		t.applyTerminalAwareWidth(colWidths)
	}

	return colWidths
}

// getHeaderWidths returns initial column widths based on header lengths.
func (t *Table) getHeaderWidths() []int {
	colWidths := make([]int, len(t.headers))
	for i, h := range t.headers {
		colWidths[i] = len(h)
	}
	return colWidths
}

// updateWidthsFromContent adjusts column widths based on wrapped cell content.
func (t *Table) updateWidthsFromContent(colWidths []int, wrappedRows [][][]string) {
	for _, wrappedRow := range wrappedRows {
		for colIdx, wrappedCell := range wrappedRow {
			if colIdx >= len(colWidths) {
				continue
			}
			t.updateColumnWidth(colWidths, colIdx, wrappedCell)
		}
	}
}

// updateColumnWidth updates a single column's width based on cell content.
func (t *Table) updateColumnWidth(colWidths []int, colIdx int, wrappedCell []string) {
	for _, line := range wrappedCell {
		if len(line) <= colWidths[colIdx] {
			continue
		}

		// Check if we have a max width constraint
		if maxWidth, hasLimit := t.maxWidths[colIdx]; hasLimit && maxWidth > 0 {
			if maxWidth > colWidths[colIdx] {
				colWidths[colIdx] = maxWidth
			}
		} else {
			colWidths[colIdx] = len(line)
		}
	}
}

// buildTableString constructs the final table string with headers, separator, and rows.
func (t *Table) buildTableString(wrappedRows [][][]string, colWidths []int) string {
	var result strings.Builder

	t.writeHeader(&result, colWidths)
	t.writeSeparator(&result, colWidths)
	t.writeRows(&result, wrappedRows, colWidths)

	return result.String()
}

// writeHeader writes the table header line.
func (t *Table) writeHeader(result *strings.Builder, colWidths []int) {
	headerParts := make([]string, len(t.headers))
	for i, h := range t.headers {
		headerParts[i] = padRight(h, colWidths[i])
	}
	result.WriteString(strings.Join(headerParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")
}

// writeSeparator writes the separator line between header and data.
func (t *Table) writeSeparator(result *strings.Builder, colWidths []int) {
	sepParts := make([]string, len(t.headers))
	for i, w := range colWidths {
		sepParts[i] = strings.Repeat("-", w)
	}
	result.WriteString(strings.Join(sepParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")
}

// writeRows writes all data rows with multi-line support.
func (t *Table) writeRows(result *strings.Builder, wrappedRows [][][]string, colWidths []int) {
	for _, wrappedRow := range wrappedRows {
		t.writeMultiLineRow(result, wrappedRow, colWidths)
	}
}

// writeMultiLineRow writes a single row that may span multiple lines.
func (t *Table) writeMultiLineRow(result *strings.Builder, wrappedRow [][]string, colWidths []int) {
	maxLines := t.getMaxLines(wrappedRow)

	for lineIdx := range maxLines {
		t.writeRowLine(result, wrappedRow, colWidths, lineIdx)
	}
}

// getMaxLines returns the maximum number of lines needed for a wrapped row.
func (t *Table) getMaxLines(wrappedRow [][]string) int {
	maxLines := 1
	for _, wrappedCell := range wrappedRow {
		if len(wrappedCell) > maxLines {
			maxLines = len(wrappedCell)
		}
	}
	return maxLines
}

// writeRowLine writes a single line of a potentially multi-line row.
func (t *Table) writeRowLine(result *strings.Builder, wrappedRow [][]string, colWidths []int, lineIdx int) {
	rowParts := make([]string, len(t.headers))
	for colIdx := range t.headers {
		rowParts[colIdx] = t.getCellLine(wrappedRow, colIdx, lineIdx, colWidths[colIdx])
	}
	result.WriteString(strings.Join(rowParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")
}

// getCellLine gets a specific line from a wrapped cell, or empty string if line doesn't exist.
func (t *Table) getCellLine(wrappedRow [][]string, colIdx, lineIdx, colWidth int) string {
	if colIdx < len(wrappedRow) && lineIdx < len(wrappedRow[colIdx]) {
		return padRight(wrappedRow[colIdx][lineIdx], colWidth)
	}
	return padRight("", colWidth)
}

// padRight pads a string with spaces on the right to reach the desired width.
// If the string is already longer than or equal to the width, it is returned unchanged.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// wrapText wraps text to fit within the specified width, breaking at word boundaries.
func wrapText(text string, width int) []string {
	if width <= 0 || len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	currentLine := ""
	for _, word := range words {
		// If the word itself is longer than width, break it.
		if len(word) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			// Split long word across multiple lines.
			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}
			currentLine = word
			continue
		}

		// Try adding word to current line.
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			// Word doesn't fit, start new line.
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	// Add remaining text.
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// getTerminalWidth returns the width of the terminal in characters.
// Returns 0 if unable to determine (e.g., not a TTY).
func (t *Table) getTerminalWidth() int {
	if t.terminalWidthOverride > 0 {
		return t.terminalWidthOverride
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0
	}
	return width
}

// applyTerminalAwareWidth adjusts the terminal-aware column to fit available space.
func (t *Table) applyTerminalAwareWidth(colWidths []int) {
	termWidth := t.getTerminalWidth()
	if termWidth <= 0 {
		// Can't determine terminal width, use minimum or existing width
		if colWidths[t.terminalAwareCol] < t.terminalAwareMinW {
			t.maxWidths[t.terminalAwareCol] = t.terminalAwareMinW
			colWidths[t.terminalAwareCol] = t.terminalAwareMinW
		}
		return
	}

	// Calculate space used by other columns
	usedWidth := 0
	for i, width := range colWidths {
		if i != t.terminalAwareCol {
			usedWidth += width
		}
	}

	// Add padding between columns (n-1 gaps)
	usedWidth += (len(colWidths) - 1) * t.padding

	// Calculate available width for terminal-aware column
	availableWidth := max(
		// Ensure we don't go below minimum width
		termWidth-usedWidth, t.terminalAwareMinW)

	// Set the max width for wrapping and update column width
	t.maxWidths[t.terminalAwareCol] = availableWidth
	colWidths[t.terminalAwareCol] = availableWidth
}

// isWriterTTY checks if a writer is a terminal (TTY) that supports ANSI escape codes.
func isWriterTTY(w io.Writer) bool {
	// Check if the writer is a file with a valid file descriptor
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
