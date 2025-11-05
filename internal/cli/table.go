// Package cli provides command-line interface utilities.
package cli

import (
	"strings"
)

// Table represents a simple table formatter with dynamic column widths.
type Table struct {
	headers   []string
	rows      [][]string
	padding   int
	maxWidths map[int]int // Maximum width per column index (0 = no limit)
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	return &Table{
		headers:   headers,
		rows:      make([][]string, 0),
		padding:   2, // 2 spaces between columns
		maxWidths: make(map[int]int),
	}
}

// SetColumnMaxWidth sets a maximum width for a specific column.
// Text longer than this will be wrapped to multiple lines.
func (t *Table) SetColumnMaxWidth(colIndex int, maxWidth int) {
	t.maxWidths[colIndex] = maxWidth
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

	// Wrap cells that exceed max width.
	wrappedRows := make([][][]string, len(t.rows))
	for rowIdx, row := range t.rows {
		wrappedRows[rowIdx] = make([][]string, len(row))
		for colIdx, cell := range row {
			if maxWidth, hasLimit := t.maxWidths[colIdx]; hasLimit && maxWidth > 0 {
				wrappedRows[rowIdx][colIdx] = wrapText(cell, maxWidth)
			} else {
				wrappedRows[rowIdx][colIdx] = []string{cell}
			}
		}
	}

	// Calculate column widths (respecting max widths).
	colWidths := make([]int, len(t.headers))
	for i, h := range t.headers {
		colWidths[i] = len(h)
	}

	for _, wrappedRow := range wrappedRows {
		for i, wrappedCell := range wrappedRow {
			if i < len(colWidths) {
				for _, line := range wrappedCell {
					if len(line) > colWidths[i] {
						if maxWidth, hasLimit := t.maxWidths[i]; hasLimit && maxWidth > 0 {
							if maxWidth > colWidths[i] {
								colWidths[i] = maxWidth
							}
						} else {
							colWidths[i] = len(line)
						}
					}
				}
			}
		}
	}

	var result strings.Builder

	// Format header.
	headerParts := make([]string, len(t.headers))
	for i, h := range t.headers {
		headerParts[i] = padRight(h, colWidths[i])
	}
	result.WriteString(strings.Join(headerParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")

	// Format separator.
	sepParts := make([]string, len(t.headers))
	for i, w := range colWidths {
		sepParts[i] = strings.Repeat("-", w)
	}
	result.WriteString(strings.Join(sepParts, strings.Repeat(" ", t.padding)))
	result.WriteString("\n")

	// Format data rows (with wrapping support).
	for _, wrappedRow := range wrappedRows {
		// Find max lines in this row.
		maxLines := 1
		for _, wrappedCell := range wrappedRow {
			if len(wrappedCell) > maxLines {
				maxLines = len(wrappedCell)
			}
		}

		// Print each line of the row.
		for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
			rowParts := make([]string, len(t.headers))
			for colIdx := range t.headers {
				if colIdx < len(wrappedRow) && lineIdx < len(wrappedRow[colIdx]) {
					rowParts[colIdx] = padRight(wrappedRow[colIdx][lineIdx], colWidths[colIdx])
				} else {
					rowParts[colIdx] = padRight("", colWidths[colIdx])
				}
			}
			result.WriteString(strings.Join(rowParts, strings.Repeat(" ", t.padding)))
			result.WriteString("\n")
		}
	}

	return result.String()
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
				currentLine = ""
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
