// Package cli provides command-line interface utilities.
package cli

import (
	"os"
	"strings"

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
	}
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
func (t *Table) AddRow(row []string) {
	if len(row) != len(t.headers) {
		// Pad or truncate to match header count.
		newRow := make([]string, len(t.headers))
		copy(newRow, row)
		for i := len(row); i < len(t.headers); i++ {
			newRow[i] = ""
		}
		t.rows = append(t.rows, newRow)
	} else {
		t.rows = append(t.rows, row)
	}
}

// Render formats and returns the table as a string.
func (t *Table) Render() string {
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
