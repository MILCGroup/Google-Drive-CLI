package forms

import (
	"testing"

	formsapi "google.golang.org/api/forms/v1"
)

func TestConvertForm(t *testing.T) {
	t.Run("nil form", func(t *testing.T) {
		got := convertForm(nil)
		if got.FormID != "" || got.Title != "" || len(got.Items) != 0 {
			t.Fatalf("expected empty form, got %+v", got)
		}
	})

	t.Run("form without info", func(t *testing.T) {
		got := convertForm(&formsapi.Form{
			FormId:     "f1",
			RevisionId: "rev1",
		})
		if got.FormID != "f1" {
			t.Fatalf("expected formId f1, got %q", got.FormID)
		}
		if got.RevisionID != "rev1" {
			t.Fatalf("expected revisionId rev1, got %q", got.RevisionID)
		}
		if got.Title != "" {
			t.Fatalf("expected empty title, got %q", got.Title)
		}
	})

	t.Run("form with info and items", func(t *testing.T) {
		got := convertForm(&formsapi.Form{
			FormId:        "f2",
			RevisionId:    "rev2",
			ResponderUri:  "https://docs.google.com/forms/d/e/f2/viewform",
			LinkedSheetId: "sheet1",
			Info: &formsapi.Info{
				Title:         "Survey",
				Description:   "A survey",
				DocumentTitle: "Survey Doc",
			},
			Items: []*formsapi.Item{
				{
					ItemId:      "item1",
					Title:       "Your name",
					Description: "Enter your full name",
					QuestionItem: &formsapi.QuestionItem{
						Question: &formsapi.Question{
							TextQuestion: &formsapi.TextQuestion{},
							Required:     true,
						},
					},
				},
				{
					ItemId: "item2",
					Title:  "Favorite color",
					QuestionItem: &formsapi.QuestionItem{
						Question: &formsapi.Question{
							ChoiceQuestion: &formsapi.ChoiceQuestion{},
						},
					},
				},
			},
		})

		if got.FormID != "f2" {
			t.Fatalf("expected formId f2, got %q", got.FormID)
		}
		if got.Title != "Survey" {
			t.Fatalf("expected title Survey, got %q", got.Title)
		}
		if got.Description != "A survey" {
			t.Fatalf("expected description, got %q", got.Description)
		}
		if got.DocumentTitle != "Survey Doc" {
			t.Fatalf("expected documentTitle, got %q", got.DocumentTitle)
		}
		if got.ResponderURI != "https://docs.google.com/forms/d/e/f2/viewform" {
			t.Fatalf("expected responderUri, got %q", got.ResponderURI)
		}
		if got.LinkedSheetID != "sheet1" {
			t.Fatalf("expected linkedSheetId, got %q", got.LinkedSheetID)
		}
		if len(got.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got.Items))
		}
		if got.Items[0].QuestionType != "TEXT" {
			t.Fatalf("expected TEXT question type, got %q", got.Items[0].QuestionType)
		}
		if !got.Items[0].Required {
			t.Fatalf("expected item1 to be required")
		}
		if got.Items[1].QuestionType != "CHOICE" {
			t.Fatalf("expected CHOICE question type, got %q", got.Items[1].QuestionType)
		}
	})

	t.Run("form with nil items slice", func(t *testing.T) {
		got := convertForm(&formsapi.Form{
			FormId: "f3",
			Info:   &formsapi.Info{Title: "Empty Form"},
		})
		if got.Items != nil {
			t.Fatalf("expected nil items, got %v", got.Items)
		}
	})
}

func TestConvertFormItem(t *testing.T) {
	tests := []struct {
		name             string
		item             *formsapi.Item
		wantItemID       string
		wantTitle        string
		wantDescription  string
		wantQuestionType string
		wantRequired     bool
	}{
		{
			name:         "nil item",
			item:         nil,
			wantItemID:   "",
			wantTitle:    "",
			wantRequired: false,
		},
		{
			name: "text question",
			item: &formsapi.Item{
				ItemId:      "q1",
				Title:       "Name",
				Description: "Your name",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						TextQuestion: &formsapi.TextQuestion{},
						Required:     true,
					},
				},
			},
			wantItemID:       "q1",
			wantTitle:        "Name",
			wantDescription:  "Your name",
			wantQuestionType: "TEXT",
			wantRequired:     true,
		},
		{
			name: "choice question",
			item: &formsapi.Item{
				ItemId: "q2",
				Title:  "Color",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						ChoiceQuestion: &formsapi.ChoiceQuestion{},
					},
				},
			},
			wantItemID:       "q2",
			wantTitle:        "Color",
			wantQuestionType: "CHOICE",
		},
		{
			name: "scale question",
			item: &formsapi.Item{
				ItemId: "q3",
				Title:  "Rating",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						ScaleQuestion: &formsapi.ScaleQuestion{},
					},
				},
			},
			wantItemID:       "q3",
			wantTitle:        "Rating",
			wantQuestionType: "SCALE",
		},
		{
			name: "date question",
			item: &formsapi.Item{
				ItemId: "q4",
				Title:  "Birthday",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						DateQuestion: &formsapi.DateQuestion{},
					},
				},
			},
			wantItemID:       "q4",
			wantTitle:        "Birthday",
			wantQuestionType: "DATE",
		},
		{
			name: "time question",
			item: &formsapi.Item{
				ItemId: "q5",
				Title:  "Preferred time",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						TimeQuestion: &formsapi.TimeQuestion{},
					},
				},
			},
			wantItemID:       "q5",
			wantTitle:        "Preferred time",
			wantQuestionType: "TIME",
		},
		{
			name: "file upload question",
			item: &formsapi.Item{
				ItemId: "q6",
				Title:  "Upload resume",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						FileUploadQuestion: &formsapi.FileUploadQuestion{},
					},
				},
			},
			wantItemID:       "q6",
			wantTitle:        "Upload resume",
			wantQuestionType: "FILE_UPLOAD",
		},
		{
			name: "rating question",
			item: &formsapi.Item{
				ItemId: "q7",
				Title:  "Rate us",
				QuestionItem: &formsapi.QuestionItem{
					Question: &formsapi.Question{
						RatingQuestion: &formsapi.RatingQuestion{},
					},
				},
			},
			wantItemID:       "q7",
			wantTitle:        "Rate us",
			wantQuestionType: "RATING",
		},
		{
			name: "question group item with row question",
			item: &formsapi.Item{
				ItemId: "qg1",
				Title:  "Grid question",
				QuestionGroupItem: &formsapi.QuestionGroupItem{
					Questions: []*formsapi.Question{
						{
							RowQuestion: &formsapi.RowQuestion{},
							Required:    true,
						},
					},
				},
			},
			wantItemID:       "qg1",
			wantTitle:        "Grid question",
			wantQuestionType: "ROW",
			wantRequired:     true,
		},
		{
			name: "question group item with empty questions",
			item: &formsapi.Item{
				ItemId: "qg2",
				Title:  "Empty group",
				QuestionGroupItem: &formsapi.QuestionGroupItem{
					Questions: []*formsapi.Question{},
				},
			},
			wantItemID:       "qg2",
			wantTitle:        "Empty group",
			wantQuestionType: "",
		},
		{
			name: "non-question item (text item)",
			item: &formsapi.Item{
				ItemId:   "t1",
				Title:    "Section header",
				TextItem: &formsapi.TextItem{},
			},
			wantItemID:       "t1",
			wantTitle:        "Section header",
			wantQuestionType: "",
		},
		{
			name: "question item with nil question",
			item: &formsapi.Item{
				ItemId:       "q8",
				Title:        "Broken",
				QuestionItem: &formsapi.QuestionItem{},
			},
			wantItemID:       "q8",
			wantTitle:        "Broken",
			wantQuestionType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertFormItem(tt.item)
			if got.ItemID != tt.wantItemID {
				t.Errorf("ItemID = %q, want %q", got.ItemID, tt.wantItemID)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", got.Title, tt.wantTitle)
			}
			if got.Description != tt.wantDescription {
				t.Errorf("Description = %q, want %q", got.Description, tt.wantDescription)
			}
			if got.QuestionType != tt.wantQuestionType {
				t.Errorf("QuestionType = %q, want %q", got.QuestionType, tt.wantQuestionType)
			}
			if got.Required != tt.wantRequired {
				t.Errorf("Required = %v, want %v", got.Required, tt.wantRequired)
			}
		})
	}
}

func TestConvertResponse(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		got := convertResponse(nil)
		if got.ResponseID != "" || got.Answers != nil {
			t.Fatalf("expected empty response, got %+v", got)
		}
	})

	t.Run("response without answers", func(t *testing.T) {
		got := convertResponse(&formsapi.FormResponse{
			ResponseId:        "r1",
			CreateTime:        "2026-01-15T10:00:00Z",
			LastSubmittedTime: "2026-01-15T10:05:00Z",
			RespondentEmail:   "user@example.com",
		})
		if got.ResponseID != "r1" {
			t.Fatalf("expected responseId r1, got %q", got.ResponseID)
		}
		if got.CreateTime != "2026-01-15T10:00:00Z" {
			t.Fatalf("expected createTime, got %q", got.CreateTime)
		}
		if got.LastSubmittedTime != "2026-01-15T10:05:00Z" {
			t.Fatalf("expected lastSubmittedTime, got %q", got.LastSubmittedTime)
		}
		if got.RespondentEmail != "user@example.com" {
			t.Fatalf("expected email, got %q", got.RespondentEmail)
		}
		if got.Answers != nil {
			t.Fatalf("expected nil answers, got %v", got.Answers)
		}
	})

	t.Run("response with text answers", func(t *testing.T) {
		got := convertResponse(&formsapi.FormResponse{
			ResponseId: "r2",
			Answers: map[string]formsapi.Answer{
				"qid1": {
					QuestionId: "qid1",
					TextAnswers: &formsapi.TextAnswers{
						Answers: []*formsapi.TextAnswer{
							{Value: "Alice"},
						},
					},
				},
				"qid2": {
					QuestionId: "qid2",
					TextAnswers: &formsapi.TextAnswers{
						Answers: []*formsapi.TextAnswer{
							{Value: "Red"},
							{Value: "Blue"},
						},
					},
				},
			},
		})
		if len(got.Answers) != 2 {
			t.Fatalf("expected 2 answers, got %d", len(got.Answers))
		}

		a1, ok := got.Answers["qid1"]
		if !ok {
			t.Fatalf("expected answer for qid1")
		}
		if a1.QuestionID != "qid1" {
			t.Fatalf("expected questionId qid1, got %q", a1.QuestionID)
		}
		if len(a1.TextAnswers) != 1 || a1.TextAnswers[0] != "Alice" {
			t.Fatalf("expected text answer Alice, got %v", a1.TextAnswers)
		}

		a2, ok := got.Answers["qid2"]
		if !ok {
			t.Fatalf("expected answer for qid2")
		}
		if len(a2.TextAnswers) != 2 {
			t.Fatalf("expected 2 text answers, got %d", len(a2.TextAnswers))
		}
		if a2.TextAnswers[0] != "Red" || a2.TextAnswers[1] != "Blue" {
			t.Fatalf("expected Red and Blue, got %v", a2.TextAnswers)
		}
	})

	t.Run("response with file upload answers", func(t *testing.T) {
		got := convertResponse(&formsapi.FormResponse{
			ResponseId: "r3",
			Answers: map[string]formsapi.Answer{
				"qid3": {
					QuestionId: "qid3",
					FileUploadAnswers: &formsapi.FileUploadAnswers{
						Answers: []*formsapi.FileUploadAnswer{
							{FileId: "file-abc"},
							{FileId: "file-def"},
						},
					},
				},
			},
		})
		a3, ok := got.Answers["qid3"]
		if !ok {
			t.Fatalf("expected answer for qid3")
		}
		if len(a3.FileUploadAnswers) != 2 {
			t.Fatalf("expected 2 file upload answers, got %d", len(a3.FileUploadAnswers))
		}
		if a3.FileUploadAnswers[0] != "file-abc" || a3.FileUploadAnswers[1] != "file-def" {
			t.Fatalf("expected file IDs, got %v", a3.FileUploadAnswers)
		}
	})

	t.Run("response with mixed answer types", func(t *testing.T) {
		got := convertResponse(&formsapi.FormResponse{
			ResponseId: "r4",
			Answers: map[string]formsapi.Answer{
				"qid4": {
					QuestionId: "qid4",
					TextAnswers: &formsapi.TextAnswers{
						Answers: []*formsapi.TextAnswer{
							{Value: "Answer text"},
						},
					},
					FileUploadAnswers: &formsapi.FileUploadAnswers{
						Answers: []*formsapi.FileUploadAnswer{
							{FileId: "file-ghi"},
						},
					},
				},
			},
		})
		a4 := got.Answers["qid4"]
		if len(a4.TextAnswers) != 1 || a4.TextAnswers[0] != "Answer text" {
			t.Fatalf("expected text answer, got %v", a4.TextAnswers)
		}
		if len(a4.FileUploadAnswers) != 1 || a4.FileUploadAnswers[0] != "file-ghi" {
			t.Fatalf("expected file upload answer, got %v", a4.FileUploadAnswers)
		}
	})
}
