package types

import "testing"

// Mock implementation of TableRenderer for testing
type mockTableRenderer struct {
	headers      []string
	rows         [][]string
	emptyMessage string
}

func (m *mockTableRenderer) Headers() []string {
	return m.headers
}

func (m *mockTableRenderer) Rows() [][]string {
	return m.rows
}

func (m *mockTableRenderer) EmptyMessage() string {
	return m.emptyMessage
}

// Mock implementation of TableRenderable for testing
type mockTableRenderable struct {
	renderer TableRenderer
}

func (m *mockTableRenderable) AsTableRenderer() TableRenderer {
	return m.renderer
}

func TestTableRenderer_Interface(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{"ID", "Name", "Type"},
		rows:         [][]string{{"1", "file.txt", "text/plain"}},
		emptyMessage: "No files found",
	}

	// Test that mock implements TableRenderer interface
	var _ TableRenderer = renderer

	// Test Headers method
	headers := renderer.Headers()
	if len(headers) != 3 {
		t.Errorf("Headers length = %d, want 3", len(headers))
	}
	if headers[0] != "ID" {
		t.Errorf("Headers[0] = %s, want ID", headers[0])
	}

	// Test Rows method
	rows := renderer.Rows()
	if len(rows) != 1 {
		t.Errorf("Rows length = %d, want 1", len(rows))
	}
	if len(rows[0]) != 3 {
		t.Errorf("Rows[0] length = %d, want 3", len(rows[0]))
	}

	// Test EmptyMessage method
	msg := renderer.EmptyMessage()
	if msg != "No files found" {
		t.Errorf("EmptyMessage = %s, want 'No files found'", msg)
	}
}

func TestTableRenderable_Interface(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{"Column1", "Column2"},
		rows:         [][]string{{"A", "B"}, {"C", "D"}},
		emptyMessage: "Empty",
	}

	renderable := &mockTableRenderable{
		renderer: renderer,
	}

	// Test that mock implements TableRenderable interface
	var _ TableRenderable = renderable

	// Test AsTableRenderer method
	tableRenderer := renderable.AsTableRenderer()
	if tableRenderer == nil {
		t.Fatal("AsTableRenderer returned nil")
	}

	// Verify we get back the expected renderer
	headers := tableRenderer.Headers()
	if len(headers) != 2 {
		t.Errorf("Headers length = %d, want 2", len(headers))
	}

	rows := tableRenderer.Rows()
	if len(rows) != 2 {
		t.Errorf("Rows length = %d, want 2", len(rows))
	}
}

func TestTableRenderer_EmptyTable(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{"ID", "Name"},
		rows:         [][]string{},
		emptyMessage: "No data available",
	}

	headers := renderer.Headers()
	if len(headers) != 2 {
		t.Errorf("Headers length = %d, want 2", len(headers))
	}

	rows := renderer.Rows()
	if len(rows) != 0 {
		t.Errorf("Rows length = %d, want 0", len(rows))
	}

	msg := renderer.EmptyMessage()
	if msg != "No data available" {
		t.Errorf("EmptyMessage = %s, want 'No data available'", msg)
	}
}

func TestTableRenderer_MultipleRows(t *testing.T) {
	renderer := &mockTableRenderer{
		headers: []string{"ID", "Name", "Size"},
		rows: [][]string{
			{"1", "file1.txt", "100"},
			{"2", "file2.txt", "200"},
			{"3", "file3.txt", "300"},
		},
		emptyMessage: "No files",
	}

	rows := renderer.Rows()
	if len(rows) != 3 {
		t.Errorf("Rows length = %d, want 3", len(rows))
	}

	// Verify each row has correct number of columns
	for i, row := range rows {
		if len(row) != 3 {
			t.Errorf("Row %d length = %d, want 3", i, len(row))
		}
	}

	// Verify first row values
	if rows[0][0] != "1" {
		t.Errorf("rows[0][0] = %s, want 1", rows[0][0])
	}
	if rows[0][1] != "file1.txt" {
		t.Errorf("rows[0][1] = %s, want file1.txt", rows[0][1])
	}
}

func TestTableRenderer_NilRows(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{"ID", "Name"},
		rows:         nil,
		emptyMessage: "No data",
	}

	rows := renderer.Rows()
	if rows != nil {
		t.Errorf("Rows should be nil, got %v", rows)
	}

	// Empty message should still work
	msg := renderer.EmptyMessage()
	if msg != "No data" {
		t.Errorf("EmptyMessage = %s, want 'No data'", msg)
	}
}

func TestTableRenderer_EmptyHeaders(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{},
		rows:         [][]string{{"data"}},
		emptyMessage: "Empty",
	}

	headers := renderer.Headers()
	if len(headers) != 0 {
		t.Errorf("Headers length = %d, want 0", len(headers))
	}
}

func TestTableRenderable_MultipleConversions(t *testing.T) {
	renderer := &mockTableRenderer{
		headers:      []string{"Col1"},
		rows:         [][]string{{"Value"}},
		emptyMessage: "Empty",
	}

	renderable := &mockTableRenderable{renderer: renderer}

	// Call AsTableRenderer multiple times
	r1 := renderable.AsTableRenderer()
	r2 := renderable.AsTableRenderer()

	// Both should return the same renderer
	if r1 != r2 {
		t.Error("AsTableRenderer should return the same renderer instance")
	}
}

func TestTableRenderer_SpecialCharacters(t *testing.T) {
	renderer := &mockTableRenderer{
		headers: []string{"ID", "Name with spaces", "Description\nMultiline"},
		rows: [][]string{
			{"1", "file with\ttabs", "description\nwith\nnewlines"},
			{"2", "file with 'quotes'", `description with "double quotes"`},
		},
		emptyMessage: "No files found",
	}

	headers := renderer.Headers()
	if len(headers) != 3 {
		t.Errorf("Headers length = %d, want 3", len(headers))
	}

	rows := renderer.Rows()
	if len(rows) != 2 {
		t.Errorf("Rows length = %d, want 2", len(rows))
	}

	// Verify special characters are preserved
	if rows[0][1] != "file with\ttabs" {
		t.Errorf("Special characters not preserved in rows[0][1]")
	}
}

func TestTableRenderer_LargeTable(t *testing.T) {
	// Create a table with many rows
	rows := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = []string{"id", "name", "value"}
	}

	renderer := &mockTableRenderer{
		headers:      []string{"ID", "Name", "Value"},
		rows:         rows,
		emptyMessage: "No data",
	}

	result := renderer.Rows()
	if len(result) != 1000 {
		t.Errorf("Rows length = %d, want 1000", len(result))
	}
}
