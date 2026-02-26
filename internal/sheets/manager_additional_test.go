package sheets

import (
	"testing"

	sheetsapi "google.golang.org/api/sheets/v4"
)

// TestConvertSpreadsheet_EdgeCases tests edge cases for spreadsheet conversion
func TestConvertSpreadsheet_EdgeCases(t *testing.T) {
	t.Run("nil spreadsheet returns empty struct", func(t *testing.T) {
		got := convertSpreadsheet(nil)
		if got == nil {
			t.Fatal("expected non-nil result")
		}
		if got.ID != "" || got.Title != "" || got.SheetCount != 0 {
			t.Error("expected all zero values for nil input")
		}
		if got.Sheets != nil {
			t.Error("expected nil Sheets for nil input")
		}
	})

	t.Run("empty sheets slice", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Properties:    &sheetsapi.SpreadsheetProperties{Title: "Test"},
			Sheets:        []*sheetsapi.Sheet{},
		})
		if got.SheetCount != 0 {
			t.Errorf("expected sheet count 0, got %d", got.SheetCount)
		}
		if len(got.Sheets) != 0 {
			t.Errorf("expected empty sheets slice, got %d", len(got.Sheets))
		}
	})

	t.Run("sheet without properties", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Sheets:        []*sheetsapi.Sheet{{}, {}},
		})
		if got.SheetCount != 2 {
			t.Errorf("expected sheet count 2, got %d", got.SheetCount)
		}
		if len(got.Sheets) != 2 {
			t.Errorf("expected 2 sheets, got %d", len(got.Sheets))
		}
		// Verify zero values
		if got.Sheets[0].ID != 0 || got.Sheets[1].ID != 0 {
			t.Error("expected zero IDs for sheets without properties")
		}
	})

	t.Run("spreadsheet with nil sheets", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Properties:    &sheetsapi.SpreadsheetProperties{Title: "NoSheets"},
			Sheets:        nil,
		})
		if got.SheetCount != 0 {
			t.Errorf("expected sheet count 0 for nil Sheets, got %d", got.SheetCount)
		}
		if got.Sheets != nil {
			t.Errorf("expected nil Sheets slice, got %v", got.Sheets)
		}
	})

	t.Run("nil properties handled", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Properties:    nil,
			Sheets:        nil,
		})
		if got.ID != "s1" {
			t.Errorf("expected ID s1, got %s", got.ID)
		}
		if got.Title != "" || got.Locale != "" || got.TimeZone != "" {
			t.Error("expected empty property values when Properties is nil")
		}
	})

	t.Run("spreadsheet with empty ID", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "",
			Properties:    &sheetsapi.SpreadsheetProperties{Title: "EmptyID"},
		})
		if got.ID != "" {
			t.Errorf("expected empty ID, got %s", got.ID)
		}
		if got.Title != "EmptyID" {
			t.Errorf("expected title EmptyID, got %s", got.Title)
		}
	})

	t.Run("multiple sheets with mixed properties", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Properties: &sheetsapi.SpreadsheetProperties{
				Title:    "MultiSheet",
				Locale:   "fr",
				TimeZone: "Europe/Paris",
			},
			Sheets: []*sheetsapi.Sheet{
				{
					Properties: &sheetsapi.SheetProperties{
						SheetId: 1,
						Title:   "First",
						Index:   0,
					},
				},
				{
					Properties: &sheetsapi.SheetProperties{
						SheetId:   2,
						Title:     "Second",
						Index:     1,
						SheetType: "GRID",
					},
				},
				{}, // Sheet without properties
			},
		})

		if got.Title != "MultiSheet" {
			t.Errorf("expected title MultiSheet, got %s", got.Title)
		}
		if got.Locale != "fr" {
			t.Errorf("expected locale fr, got %s", got.Locale)
		}
		if got.SheetCount != 3 {
			t.Errorf("expected sheet count 3, got %d", got.SheetCount)
		}
		if len(got.Sheets) != 3 {
			t.Errorf("expected 3 sheets, got %d", len(got.Sheets))
		}
		if got.Sheets[0].Title != "First" || got.Sheets[0].ID != 1 {
			t.Errorf("unexpected first sheet: %+v", got.Sheets[0])
		}
		if got.Sheets[1].Title != "Second" || got.Sheets[1].Type != "GRID" {
			t.Errorf("unexpected second sheet: %+v", got.Sheets[1])
		}
		if got.Sheets[2].ID != 0 {
			t.Errorf("expected third sheet to have zero ID, got %d", got.Sheets[2].ID)
		}
	})
}

// TestConvertSpreadsheet_TypeVariations tests various sheet types
func TestConvertSpreadsheet_TypeVariations(t *testing.T) {
	sheetTypes := []string{"GRID", "OBJECT", "CHART", "", "DUMMY"}
	for _, sheetType := range sheetTypes {
		t.Run("sheet type "+sheetType, func(t *testing.T) {
			got := convertSpreadsheet(&sheetsapi.Spreadsheet{
				SpreadsheetId: "s1",
				Sheets: []*sheetsapi.Sheet{{
					Properties: &sheetsapi.SheetProperties{
						SheetId:   1,
						Title:     "Test",
						SheetType: sheetType,
					},
				}},
			})
			if got.Sheets[0].Type != sheetType {
				t.Errorf("expected type %q, got %q", sheetType, got.Sheets[0].Type)
			}
		})
	}
}

// TestConvertSpreadsheet_NumericEdgeCases tests numeric edge cases
func TestConvertSpreadsheet_NumericEdgeCases(t *testing.T) {
	t.Run("negative index values", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Sheets: []*sheetsapi.Sheet{{
				Properties: &sheetsapi.SheetProperties{
					SheetId: -1,
					Index:   -5,
					Title:   "Negative",
				},
			}},
		})
		if got.Sheets[0].ID != -1 {
			t.Errorf("expected ID -1, got %d", got.Sheets[0].ID)
		}
		if got.Sheets[0].Index != -5 {
			t.Errorf("expected index -5, got %d", got.Sheets[0].Index)
		}
	})

	t.Run("large sheet ID", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Sheets: []*sheetsapi.Sheet{{
				Properties: &sheetsapi.SheetProperties{
					SheetId: 999999999,
					Title:   "LargeID",
				},
			}},
		})
		if got.Sheets[0].ID != 999999999 {
			t.Errorf("expected large ID, got %d", got.Sheets[0].ID)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		got := convertSpreadsheet(&sheetsapi.Spreadsheet{
			SpreadsheetId: "s1",
			Sheets: []*sheetsapi.Sheet{{
				Properties: &sheetsapi.SheetProperties{
					SheetId:   0,
					Index:     0,
					Title:     "Zero",
					SheetType: "",
				},
			}},
		})
		if got.Sheets[0].ID != 0 {
			t.Errorf("expected ID 0, got %d", got.Sheets[0].ID)
		}
		if got.Sheets[0].Index != 0 {
			t.Errorf("expected index 0, got %d", got.Sheets[0].Index)
		}
	})
}

// BenchmarkConvertSpreadsheet benchmarks the conversion function
func BenchmarkConvertSpreadsheet(b *testing.B) {
	spreadsheet := &sheetsapi.Spreadsheet{
		SpreadsheetId: "s1",
		Properties: &sheetsapi.SpreadsheetProperties{
			Title:    "Benchmark",
			Locale:   "en",
			TimeZone: "UTC",
		},
		Sheets: []*sheetsapi.Sheet{
			{
				Properties: &sheetsapi.SheetProperties{
					SheetId:   1,
					Title:     "Sheet1",
					Index:     0,
					SheetType: "GRID",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSpreadsheet(spreadsheet)
	}
}

func BenchmarkConvertSpreadsheetNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSpreadsheet(nil)
	}
}

func BenchmarkConvertSpreadsheetManySheets(b *testing.B) {
	sheets := make([]*sheetsapi.Sheet, 100)
	for i := 0; i < 100; i++ {
		sheets[i] = &sheetsapi.Sheet{
			Properties: &sheetsapi.SheetProperties{
				SheetId:   int64(i),
				Title:     "Sheet" + string(rune('0'+i%10)),
				Index:     int64(i),
				SheetType: "GRID",
			},
		}
	}
	spreadsheet := &sheetsapi.Spreadsheet{
		SpreadsheetId: "s1",
		Properties:    &sheetsapi.SpreadsheetProperties{Title: "ManySheets"},
		Sheets:        sheets,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSpreadsheet(spreadsheet)
	}
}
