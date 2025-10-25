// Package cli provides command-line interface utilities.
package cli

import (
	"strings"
)

// Table represents a simple table formatter with dynamic column widths.
type Table struct {
	headers []string
	rows    [][]string
	padding int
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	return &Table{
		headers: headers,
		rows:    make([][]string, 0),
		padding: 2, // 2 spaces between columns
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(row []string) {
	if len(row) != len(t.headers) {
		// Pad or truncate to match header count
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

	// Calculate column widths
	colWidths := make([]int, len(t.headers))
	for i, h := range t.headers {
		colWidths[i] = len(h)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var result strings.Builder

	// Format header
	headerParts := make([]string, len(t.headers))
	for i, h := range t.headers {
		headerParts[i] = padRight(h, colWidths[i])
	}
	result.WriteString(strings.Join(headerParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")

	// Format separator
	sepParts := make([]string, len(t.headers))
	for i, w := range colWidths {
		sepParts[i] = strings.Repeat("-", w)
	}
	result.WriteString(strings.Join(sepParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")

	// Format data rows
	for _, row := range t.rows {
		rowParts := make([]string, len(t.headers))
		for i, cell := range row {
			if i < len(colWidths) {
				rowParts[i] = padRight(cell, colWidths[i])
			}
		}
		result.WriteString(strings.Join(rowParts, strings.Repeat(" ", t.padding)))
		result.WriteString("\n")
	}

	return result.String()
}

// padRight pads a string with spaces on the right to reach the desired width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
