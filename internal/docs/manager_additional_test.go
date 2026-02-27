package docs

import (
	"strings"
	"testing"

	docsapi "google.golang.org/api/docs/v1"
)

// TestExtractTextFromBody_EdgeCases tests edge cases for text extraction
func TestExtractTextFromBody_EdgeCases(t *testing.T) {
	t.Run("nil body", func(t *testing.T) {
		got := extractTextFromBody(nil)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("nil content", func(t *testing.T) {
		got := extractTextFromBody(&docsapi.Body{Content: nil})
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("empty content", func(t *testing.T) {
		got := extractTextFromBody(&docsapi.Body{Content: []*docsapi.StructuralElement{}})
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("paragraph with nil elements", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: &docsapi.Paragraph{Elements: nil}},
		}}
		got := extractTextFromBody(body)
		if got != "" {
			t.Errorf("expected empty string for nil paragraph elements, got %q", got)
		}
	})

	t.Run("paragraph with empty elements", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{}}},
		}}
		got := extractTextFromBody(body)
		if got != "" {
			t.Errorf("expected empty string for empty paragraph elements, got %q", got)
		}
	})

	t.Run("nil structural element", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			nil,
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: &docsapi.TextRun{Content: "Hello"}},
			}}},
		}}
		got := extractTextFromBody(body)
		if got != "Hello" {
			t.Errorf("expected Hello, got %q", got)
		}
	})

	t.Run("element with nil paragraph", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: nil},
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: &docsapi.TextRun{Content: "World"}},
			}}},
		}}
		got := extractTextFromBody(body)
		if got != "World" {
			t.Errorf("expected World, got %q", got)
		}
	})

	t.Run("text run with empty content", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: &docsapi.TextRun{Content: ""}},
				{TextRun: &docsapi.TextRun{Content: "Text"}},
			}}},
		}}
		got := extractTextFromBody(body)
		if got != "Text" {
			t.Errorf("expected Text, got %q", got)
		}
	})

	t.Run("nil text run", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: nil},
				{TextRun: &docsapi.TextRun{Content: "AfterNil"}},
			}}},
		}}
		got := extractTextFromBody(body)
		if got != "AfterNil" {
			t.Errorf("expected AfterNil, got %q", got)
		}
	})

	t.Run("table with nil rows", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: nil}},
		}}
		got := extractTextFromBody(body)
		if got != "" {
			t.Errorf("expected empty string for nil table rows, got %q", got)
		}
	})

	t.Run("table with empty rows", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{}}},
		}}
		got := extractTextFromBody(body)
		if got != "" {
			t.Errorf("expected empty string for empty table rows, got %q", got)
		}
	})

	t.Run("table with nil cells", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{
				{TableCells: nil},
			}}},
		}}
		got := extractTextFromBody(body)
		// Should have tab and newline even with nil cells
		if !strings.Contains(got, "\n") {
			t.Errorf("expected newline in table output, got %q", got)
		}
	})

	t.Run("table with empty cells", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{
				{TableCells: []*docsapi.TableCell{}},
			}}},
		}}
		got := extractTextFromBody(body)
		if !strings.Contains(got, "\n") {
			t.Errorf("expected newline in table output, got %q", got)
		}
	})

	t.Run("multiple section breaks", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: &docsapi.TextRun{Content: "First"}},
			}}},
			{SectionBreak: &docsapi.SectionBreak{}},
			{SectionBreak: &docsapi.SectionBreak{}},
			{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
				{TextRun: &docsapi.TextRun{Content: "Second"}},
			}}},
		}}
		got := extractTextFromBody(body)
		if !strings.Contains(got, "First") || !strings.Contains(got, "Second") {
			t.Errorf("expected First and Second, got %q", got)
		}
		// Should have two section break newlines
		if !strings.Contains(got, "\n\n\n\n") {
			t.Errorf("expected multiple newlines from section breaks, got %q", got)
		}
	})

	t.Run("complex nested table", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{
				{
					TableCells: []*docsapi.TableCell{
						{
							Content: []*docsapi.StructuralElement{
								{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
									{TextRun: &docsapi.TextRun{Content: "A1"}},
								}}},
							},
						},
						{
							Content: []*docsapi.StructuralElement{
								{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
									{TextRun: &docsapi.TextRun{Content: "B1"}},
								}}},
							},
						},
					},
				},
				{
					TableCells: []*docsapi.TableCell{
						{
							Content: []*docsapi.StructuralElement{
								{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
									{TextRun: &docsapi.TextRun{Content: "A2"}},
								}}},
							},
						},
						{
							Content: []*docsapi.StructuralElement{
								{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
									{TextRun: &docsapi.TextRun{Content: "B2"}},
								}}},
							},
						},
					},
				},
			}}},
		}}
		got := extractTextFromBody(body)
		if !strings.Contains(got, "A1") || !strings.Contains(got, "B1") {
			t.Errorf("expected A1 and B1 in first row, got %q", got)
		}
		if !strings.Contains(got, "A2") || !strings.Contains(got, "B2") {
			t.Errorf("expected A2 and B2 in second row, got %q", got)
		}
		// Check for tabs between cells and newlines between rows
		if !strings.Contains(got, "\t") {
			t.Errorf("expected tabs between cells, got %q", got)
		}
		if !strings.Contains(got, "\n") {
			t.Errorf("expected newlines between rows, got %q", got)
		}
	})
}

// TestExtractTextFromElement_EdgeCases tests edge cases for element extraction
func TestExtractTextFromElement_EdgeCases(t *testing.T) {
	t.Run("nil element", func(t *testing.T) {
		var text strings.Builder
		extractTextFromElement(nil, &text)
		if text.String() != "" {
			t.Errorf("expected empty string for nil element, got %q", text.String())
		}
	})

	t.Run("element with only section break", func(t *testing.T) {
		var text strings.Builder
		element := &docsapi.StructuralElement{
			SectionBreak: &docsapi.SectionBreak{},
		}
		extractTextFromElement(element, &text)
		if text.String() != "\n\n" {
			t.Errorf("expected \\n\\n from section break, got %q", text.String())
		}
	})

	t.Run("element with all nil fields", func(t *testing.T) {
		var text strings.Builder
		element := &docsapi.StructuralElement{
			Paragraph:    nil,
			Table:        nil,
			SectionBreak: nil,
		}
		extractTextFromElement(element, &text)
		if text.String() != "" {
			t.Errorf("expected empty string when all fields nil, got %q", text.String())
		}
	})

	t.Run("nested table cell with nil content", func(t *testing.T) {
		body := &docsapi.Body{Content: []*docsapi.StructuralElement{
			{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{
				{
					TableCells: []*docsapi.TableCell{
						{Content: nil},
						{Content: []*docsapi.StructuralElement{
							{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
								{TextRun: &docsapi.TextRun{Content: "ValidCell"}},
							}}},
						}},
					},
				},
			}}},
		}}
		got := extractTextFromBody(body)
		if !strings.Contains(got, "ValidCell") {
			t.Errorf("expected ValidCell, got %q", got)
		}
	})
}

// TestCountWords_EdgeCases tests edge cases for word counting
func TestCountWords_EdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		if got := countWords(""); got != 0 {
			t.Errorf("countWords(\"\") = %d, want 0", got)
		}
	})

	t.Run("only whitespace", func(t *testing.T) {
		tests := []string{
			"   ",
			"\t",
			"\n",
			"\t\n  \r",
			"     ",
		}
		for _, test := range tests {
			if got := countWords(test); got != 0 {
				t.Errorf("countWords(%q) = %d, want 0", test, got)
			}
		}
	})

	t.Run("single word", func(t *testing.T) {
		if got := countWords("Hello"); got != 1 {
			t.Errorf("countWords(\"Hello\") = %d, want 1", got)
		}
	})

	t.Run("two words", func(t *testing.T) {
		if got := countWords("Hello World"); got != 2 {
			t.Errorf("countWords(\"Hello World\") = %d, want 2", got)
		}
	})

	t.Run("multiple spaces between words", func(t *testing.T) {
		if got := countWords("Hello   World"); got != 2 {
			t.Errorf("countWords(\"Hello   World\") = %d, want 2", got)
		}
	})

	t.Run("tabs and newlines as delimiters", func(t *testing.T) {
		if got := countWords("Hello\tWorld\nFoo"); got != 3 {
			t.Errorf("countWords with tabs/newlines = %d, want 3", got)
		}
	})

	t.Run("punctuation attached to words", func(t *testing.T) {
		// Fields splits on any whitespace, so "Hello," counts as one word
		if got := countWords("Hello, World!"); got != 2 {
			t.Errorf("countWords(\"Hello, World!\") = %d, want 2", got)
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		// Unicode characters count as words if not whitespace
		if got := countWords("Hello 世界"); got != 2 {
			t.Errorf("countWords unicode = %d, want 2", got)
		}
	})

	t.Run("very long text", func(t *testing.T) {
		longText := strings.Repeat("word ", 1000)
		if got := countWords(longText); got != 1000 {
			t.Errorf("countWords long text = %d, want 1000", got)
		}
	})

	t.Run("leading and trailing whitespace", func(t *testing.T) {
		if got := countWords("  Hello World  "); got != 2 {
			t.Errorf("countWords with surrounding whitespace = %d, want 2", got)
		}
	})
}

// BenchmarkExtractTextFromBody benchmarks text extraction
func BenchmarkExtractTextFromBody(b *testing.B) {
	body := &docsapi.Body{Content: []*docsapi.StructuralElement{
		{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
			{TextRun: &docsapi.TextRun{Content: "Hello World"}},
		}}},
		{Table: &docsapi.Table{TableRows: []*docsapi.TableRow{
			{TableCells: []*docsapi.TableCell{
				{Content: []*docsapi.StructuralElement{
					{Paragraph: &docsapi.Paragraph{Elements: []*docsapi.ParagraphElement{
						{TextRun: &docsapi.TextRun{Content: "Cell1"}},
					}}},
				}},
			}},
		}}},
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractTextFromBody(body)
	}
}

func BenchmarkExtractTextFromBodyNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractTextFromBody(nil)
	}
}

func BenchmarkCountWords(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = countWords(text)
	}
}

func BenchmarkCountWordsLong(b *testing.B) {
	text := strings.Repeat("word ", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = countWords(text)
	}
}
