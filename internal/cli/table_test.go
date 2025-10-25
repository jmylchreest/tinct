// Package cli provides command-line interface utilities.
package cli

import (
	"strings"
	"testing"
)

func TestNewTable(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	table := NewTable(headers)

	if table == nil {
		t.Fatal("NewTable returned nil")
	}

	if len(table.headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(table.headers))
	}

	if table.padding != 2 {
		t.Errorf("Expected padding of 2, got %d", table.padding)
	}
}

func TestTableAddRow(t *testing.T) {
	table := NewTable([]string{"Name", "Age"})

	// Add matching row
	table.AddRow([]string{"Alice", "30"})
	if len(table.rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(table.rows))
	}

	// Add row with fewer columns (should be padded)
	table.AddRow([]string{"Bob"})
	if len(table.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.rows))
	}
	if len(table.rows[1]) != 2 {
		t.Errorf("Expected row to be padded to 2 columns, got %d", len(table.rows[1]))
	}
	if table.rows[1][1] != "" {
		t.Errorf("Expected empty string for padded column, got %q", table.rows[1][1])
	}

	// Add row with more columns (should be truncated)
	table.AddRow([]string{"Charlie", "25", "Extra"})
	if len(table.rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(table.rows))
	}
	if len(table.rows[2]) != 2 {
		t.Errorf("Expected row to be truncated to 2 columns, got %d", len(table.rows[2]))
	}
}

func TestTableRender(t *testing.T) {
	table := NewTable([]string{"Name", "Age", "City"})
	table.AddRow([]string{"Alice", "30", "New York"})
	table.AddRow([]string{"Bob", "25", "LA"})

	output := table.Render()

	// Check that output contains headers
	if !strings.Contains(output, "Name") {
		t.Error("Output should contain 'Name' header")
	}
	if !strings.Contains(output, "Age") {
		t.Error("Output should contain 'Age' header")
	}
	if !strings.Contains(output, "City") {
		t.Error("Output should contain 'City' header")
	}

	// Check that output contains data
	if !strings.Contains(output, "Alice") {
		t.Error("Output should contain 'Alice'")
	}
	if !strings.Contains(output, "Bob") {
		t.Error("Output should contain 'Bob'")
	}
	if !strings.Contains(output, "New York") {
		t.Error("Output should contain 'New York'")
	}

	// Check for separator line (should contain dashes)
	lines := strings.Split(output, "\n")
	if len(lines) < 4 { // header + separator + 2 data rows + trailing newline
		t.Errorf("Expected at least 4 lines, got %d", len(lines))
	}

	// Second line should be separator with dashes
	if !strings.Contains(lines[1], "---") {
		t.Errorf("Expected separator line with dashes, got: %q", lines[1])
	}
}

func TestTableRenderEmpty(t *testing.T) {
	// Empty table (no headers)
	table := &Table{
		headers: []string{},
		rows:    make([][]string, 0),
		padding: 2,
	}

	output := table.Render()
	if output != "" {
		t.Errorf("Expected empty string for empty table, got: %q", output)
	}
}

func TestTableRenderNoRows(t *testing.T) {
	// Table with headers but no rows
	table := NewTable([]string{"Column1", "Column2"})

	output := table.Render()

	// Should still render headers and separator
	if !strings.Contains(output, "Column1") {
		t.Error("Output should contain headers even without rows")
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Error("Expected at least header and separator lines")
	}
}

func TestTableColumnAlignment(t *testing.T) {
	table := NewTable([]string{"Short", "Very Long Header", "Mid"})
	table.AddRow([]string{"A", "B", "C"})
	table.AddRow([]string{"123456789", "X", "Test"})

	output := table.Render()
	lines := strings.Split(output, "\n")

	if len(lines) < 4 {
		t.Fatalf("Expected at least 4 lines, got %d", len(lines))
	}

	// Check that columns are aligned (all rows should have same positions)
	// The "Very Long Header" should determine the width of that column
	headerLine := lines[0]
	separatorLine := lines[1]

	// Separator should have dashes matching column widths
	if len(separatorLine) != len(headerLine) {
		t.Errorf("Separator length (%d) should match header length (%d)", len(separatorLine), len(headerLine))
	}
}

func TestTableWithSpecialCharacters(t *testing.T) {
	table := NewTable([]string{"Name", "Symbol"})
	table.AddRow([]string{"Test", "→ →"})
	table.AddRow([]string{"Special", "★ ☆"})

	output := table.Render()

	if !strings.Contains(output, "→") {
		t.Error("Output should contain special character →")
	}
	if !strings.Contains(output, "★") {
		t.Error("Output should contain special character ★")
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"test", 10, "test      "},
		{"hello", 5, "hello"},
		{"world", 3, "world"}, // Width less than string length
		{"", 5, "     "},
		{"x", 1, "x"},
	}

	for _, tt := range tests {
		result := padRight(tt.input, tt.width)
		if result != tt.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
		}
	}
}

func TestTableMultipleColumns(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2", "Col3", "Col4", "Col5"})
	table.AddRow([]string{"A", "B", "C", "D", "E"})
	table.AddRow([]string{"1", "2", "3", "4", "5"})

	output := table.Render()

	// Verify all columns are present
	for _, col := range []string{"Col1", "Col2", "Col3", "Col4", "Col5"} {
		if !strings.Contains(output, col) {
			t.Errorf("Output should contain column %s", col)
		}
	}

	// Verify all data is present
	for _, val := range []string{"A", "B", "C", "D", "E", "1", "2", "3", "4", "5"} {
		if !strings.Contains(output, val) {
			t.Errorf("Output should contain value %s", val)
		}
	}
}

func TestTableConsistentSpacing(t *testing.T) {
	table := NewTable([]string{"Name", "Value"})
	table.AddRow([]string{"Short", "1"})
	table.AddRow([]string{"VeryLongName", "2"})

	output := table.Render()

	// Check that all expected content is present
	if !strings.Contains(output, "Name") {
		t.Error("Output should contain 'Name' header")
	}
	if !strings.Contains(output, "Value") {
		t.Error("Output should contain 'Value' header")
	}
	if !strings.Contains(output, "Short") {
		t.Error("Output should contain 'Short'")
	}
	if !strings.Contains(output, "VeryLongName") {
		t.Error("Output should contain 'VeryLongName'")
	}
	if !strings.Contains(output, "1") {
		t.Error("Output should contain '1'")
	}
	if !strings.Contains(output, "2") {
		t.Error("Output should contain '2'")
	}

	// Check that we have the expected structure (header, separator, rows)
	lines := strings.Split(output, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if line != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines < 4 {
		t.Errorf("Expected at least 4 non-empty lines (header, separator, 2 data rows), got %d", nonEmptyLines)
	}
}
