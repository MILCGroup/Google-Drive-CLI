package slides

import (
	"testing"

	slidesapi "google.golang.org/api/slides/v1"
)

// TestConvertPresentation_EdgeCases tests edge cases for presentation conversion
func TestConvertPresentation_EdgeCases(t *testing.T) {
	t.Run("nil presentation", func(t *testing.T) {
		got := convertPresentation(nil)
		if got == nil {
			t.Fatal("expected non-nil result")
		}
		if got.PresentationID != "" || got.Title != "" || got.SlideCount != 0 {
			t.Error("expected all zero values for nil input")
		}
		if got.Slides != nil {
			t.Error("expected nil Slides for nil input")
		}
	})

	t.Run("empty presentation", func(t *testing.T) {
		got := convertPresentation(&slidesapi.Presentation{})
		if got.PresentationID != "" {
			t.Errorf("expected empty ID, got %s", got.PresentationID)
		}
		if got.Title != "" {
			t.Errorf("expected empty title, got %s", got.Title)
		}
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides, got %d", got.SlideCount)
		}
		if len(got.Slides) != 0 {
			t.Errorf("expected empty slides slice, got %d", len(got.Slides))
		}
	})

	t.Run("presentation with nil slides", func(t *testing.T) {
		got := convertPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Title:          "Test",
			Slides:         nil,
		})
		if got.PresentationID != "p1" {
			t.Errorf("expected ID p1, got %s", got.PresentationID)
		}
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides for nil Slides, got %d", got.SlideCount)
		}
	})

	t.Run("presentation with empty slides slice", func(t *testing.T) {
		got := convertPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Title:          "Empty",
			Slides:         []*slidesapi.Page{},
		})
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides for empty slice, got %d", got.SlideCount)
		}
		if len(got.Slides) != 0 {
			t.Errorf("expected empty slides, got %d", len(got.Slides))
		}
	})

	t.Run("multiple slides", func(t *testing.T) {
		got := convertPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Title:          "MultiSlide",
			Slides: []*slidesapi.Page{
				{ObjectId: "slide1"},
				{ObjectId: "slide2"},
				{ObjectId: "slide3"},
			},
		})
		if got.SlideCount != 3 {
			t.Errorf("expected 3 slides, got %d", got.SlideCount)
		}
		if len(got.Slides) != 3 {
			t.Errorf("expected 3 slides in result, got %d", len(got.Slides))
		}
		if got.Slides[0].ObjectID != "slide1" {
			t.Errorf("expected slide1, got %s", got.Slides[0].ObjectID)
		}
		if got.Slides[2].ObjectID != "slide3" {
			t.Errorf("expected slide3, got %s", got.Slides[2].ObjectID)
		}
	})

	t.Run("slides with empty object IDs", func(t *testing.T) {
		got := convertPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{ObjectId: ""},
				{ObjectId: ""},
			},
		})
		if got.SlideCount != 2 {
			t.Errorf("expected 2 slides, got %d", got.SlideCount)
		}
		if got.Slides[0].ObjectID != "" || got.Slides[1].ObjectID != "" {
			t.Error("expected empty object IDs")
		}
	})
}

// TestExtractTextFromPresentation_EdgeCases tests text extraction edge cases
func TestExtractTextFromPresentation_EdgeCases(t *testing.T) {
	t.Run("nil presentation", func(t *testing.T) {
		got := extractTextFromPresentation(nil)
		if got == nil {
			t.Fatal("expected non-nil result")
		}
		if got.PresentationID != "" || got.Title != "" || got.SlideCount != 0 {
			t.Error("expected all zero values for nil input")
		}
		if got.TextBySlide != nil {
			t.Error("expected nil TextBySlide for nil input")
		}
	})

	t.Run("empty presentation", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{})
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides, got %d", got.SlideCount)
		}
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected empty text, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("nil slides slice", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides:         nil,
		})
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides for nil Slides, got %d", got.SlideCount)
		}
	})

	t.Run("empty slides slice", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides:         []*slidesapi.Page{},
		})
		if got.SlideCount != 0 {
			t.Errorf("expected 0 slides, got %d", got.SlideCount)
		}
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected empty text, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("slide with no page elements", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{ObjectId: "s1"},
			},
		})
		if got.SlideCount != 1 {
			t.Errorf("expected 1 slide, got %d", got.SlideCount)
		}
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("slide with nil page elements", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{ObjectId: "s1", PageElements: nil},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text for nil PageElements, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("page element with no shape", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{ObjectId: "e1"}, // No shape
					},
				},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text for element without shape, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("shape with nil text", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape:    &slidesapi.Shape{Text: nil},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text for nil text, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("shape with nil text elements", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape:    &slidesapi.Shape{Text: &slidesapi.TextContent{TextElements: nil}},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text for nil text elements, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("shape with empty text elements", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape:    &slidesapi.Shape{Text: &slidesapi.TextContent{TextElements: []*slidesapi.TextElement{}}},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected no text for empty text elements, got %d elements", len(got.TextBySlide))
		}
	})

	t.Run("multiple slides with text", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Title:          "Multi",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "slide1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "shape1",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: &slidesapi.TextRun{Content: "Slide1Text"}},
									},
								},
							},
						},
					},
				},
				{
					ObjectId: "slide2",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "shape2",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: &slidesapi.TextRun{Content: "Slide2Text"}},
									},
								},
							},
						},
					},
				},
			},
		})
		if got.SlideCount != 2 {
			t.Errorf("expected 2 slides, got %d", got.SlideCount)
		}
		if len(got.TextBySlide) != 2 {
			t.Errorf("expected 2 text entries, got %d", len(got.TextBySlide))
		}
		if got.TextBySlide[0].SlideIndex != 1 {
			t.Errorf("expected first slide index 1, got %d", got.TextBySlide[0].SlideIndex)
		}
		if got.TextBySlide[1].SlideIndex != 2 {
			t.Errorf("expected second slide index 2, got %d", got.TextBySlide[1].SlideIndex)
		}
		if got.TextBySlide[0].Text != "Slide1Text" {
			t.Errorf("expected Slide1Text, got %s", got.TextBySlide[0].Text)
		}
		if got.TextBySlide[1].Text != "Slide2Text" {
			t.Errorf("expected Slide2Text, got %s", got.TextBySlide[1].Text)
		}
	})

	t.Run("text element with nil text run", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: nil},
										{TextRun: &slidesapi.TextRun{Content: "AfterNil"}},
									},
								},
							},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 1 {
			t.Errorf("expected 1 text entry, got %d", len(got.TextBySlide))
		}
		if got.TextBySlide[0].Text != "AfterNil" {
			t.Errorf("expected AfterNil, got %s", got.TextBySlide[0].Text)
		}
	})

	t.Run("empty text content in text run", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: &slidesapi.TextRun{Content: ""}},
										{TextRun: &slidesapi.TextRun{Content: "Nonempty"}},
									},
								},
							},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 1 {
			t.Errorf("expected 1 text entry, got %d", len(got.TextBySlide))
		}
		if got.TextBySlide[0].Text != "Nonempty" {
			t.Errorf("expected Nonempty (empty strings not included), got %q", got.TextBySlide[0].Text)
		}
	})

	t.Run("all empty text runs produce no entry", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: &slidesapi.TextRun{Content: ""}},
										{TextRun: &slidesapi.TextRun{Content: ""}},
									},
								},
							},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 0 {
			t.Errorf("expected 0 text entries for all empty text, got %d", len(got.TextBySlide))
		}
	})

	t.Run("multiple text elements in same shape", func(t *testing.T) {
		got := extractTextFromPresentation(&slidesapi.Presentation{
			PresentationId: "p1",
			Slides: []*slidesapi.Page{
				{
					ObjectId: "s1",
					PageElements: []*slidesapi.PageElement{
						{
							ObjectId: "e1",
							Shape: &slidesapi.Shape{
								Text: &slidesapi.TextContent{
									TextElements: []*slidesapi.TextElement{
										{TextRun: &slidesapi.TextRun{Content: "Hello "}},
										{TextRun: &slidesapi.TextRun{Content: "World"}},
										{TextRun: &slidesapi.TextRun{Content: "!"}},
									},
								},
							},
						},
					},
				},
			},
		})
		if len(got.TextBySlide) != 1 {
			t.Errorf("expected 1 text entry, got %d", len(got.TextBySlide))
		}
		if got.TextBySlide[0].Text != "Hello World!" {
			t.Errorf("expected concatenated text, got %q", got.TextBySlide[0].Text)
		}
	})
}

// TestExtractTextFromShape_EdgeCases tests shape text extraction edge cases
func TestExtractTextFromShape_EdgeCases(t *testing.T) {
	t.Run("nil content", func(t *testing.T) {
		got := extractTextFromShape(nil)
		if got != "" {
			t.Errorf("expected empty string for nil, got %q", got)
		}
	})

	t.Run("nil text elements", func(t *testing.T) {
		got := extractTextFromShape(&slidesapi.TextContent{TextElements: nil})
		if got != "" {
			t.Errorf("expected empty string for nil elements, got %q", got)
		}
	})

	t.Run("empty text elements", func(t *testing.T) {
		got := extractTextFromShape(&slidesapi.TextContent{TextElements: []*slidesapi.TextElement{}})
		if got != "" {
			t.Errorf("expected empty string for empty elements, got %q", got)
		}
	})

	t.Run("text element with nil text run", func(t *testing.T) {
		content := &slidesapi.TextContent{
			TextElements: []*slidesapi.TextElement{
				{TextRun: nil},
				{TextRun: &slidesapi.TextRun{Content: "After"}},
			},
		}
		got := extractTextFromShape(content)
		if got != "After" {
			t.Errorf("expected After, got %q", got)
		}
	})

	t.Run("text run with empty content", func(t *testing.T) {
		content := &slidesapi.TextContent{
			TextElements: []*slidesapi.TextElement{
				{TextRun: &slidesapi.TextRun{Content: ""}},
				{TextRun: &slidesapi.TextRun{Content: "X"}},
			},
		}
		got := extractTextFromShape(content)
		if got != "X" {
			t.Errorf("expected X (empty not included), got %q", got)
		}
	})

	t.Run("all empty text runs", func(t *testing.T) {
		content := &slidesapi.TextContent{
			TextElements: []*slidesapi.TextElement{
				{TextRun: &slidesapi.TextRun{Content: ""}},
				{TextRun: &slidesapi.TextRun{Content: ""}},
			},
		}
		got := extractTextFromShape(content)
		if got != "" {
			t.Errorf("expected empty string for all empty, got %q", got)
		}
	})

	t.Run("single character texts", func(t *testing.T) {
		content := &slidesapi.TextContent{
			TextElements: []*slidesapi.TextElement{
				{TextRun: &slidesapi.TextRun{Content: "A"}},
				{TextRun: &slidesapi.TextRun{Content: "B"}},
				{TextRun: &slidesapi.TextRun{Content: "C"}},
			},
		}
		got := extractTextFromShape(content)
		if got != "ABC" {
			t.Errorf("expected ABC, got %q", got)
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		content := &slidesapi.TextContent{
			TextElements: []*slidesapi.TextElement{
				{TextRun: &slidesapi.TextRun{Content: "Hello "}},
				{TextRun: &slidesapi.TextRun{Content: "世界"}},
			},
		}
		got := extractTextFromShape(content)
		if got != "Hello 世界" {
			t.Errorf("expected unicode text, got %q", got)
		}
	})
}

// BenchmarkConvertPresentation benchmarks presentation conversion
func BenchmarkConvertPresentation(b *testing.B) {
	pres := &slidesapi.Presentation{
		PresentationId: "p1",
		Title:          "Benchmark",
		Slides: []*slidesapi.Page{
			{ObjectId: "slide1"},
			{ObjectId: "slide2"},
			{ObjectId: "slide3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertPresentation(pres)
	}
}

func BenchmarkConvertPresentationNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertPresentation(nil)
	}
}

func BenchmarkConvertPresentationManySlides(b *testing.B) {
	slides := make([]*slidesapi.Page, 100)
	for i := 0; i < 100; i++ {
		slides[i] = &slidesapi.Page{ObjectId: string(rune('a' + i%26))}
	}
	pres := &slidesapi.Presentation{
		PresentationId: "p1",
		Title:          "ManySlides",
		Slides:         slides,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertPresentation(pres)
	}
}

// BenchmarkExtractTextFromPresentation benchmarks text extraction
func BenchmarkExtractTextFromPresentation(b *testing.B) {
	pres := &slidesapi.Presentation{
		PresentationId: "p1",
		Slides: []*slidesapi.Page{
			{
				ObjectId: "s1",
				PageElements: []*slidesapi.PageElement{
					{
						ObjectId: "e1",
						Shape: &slidesapi.Shape{
							Text: &slidesapi.TextContent{
								TextElements: []*slidesapi.TextElement{
									{TextRun: &slidesapi.TextRun{Content: "Hello World"}},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractTextFromPresentation(pres)
	}
}
