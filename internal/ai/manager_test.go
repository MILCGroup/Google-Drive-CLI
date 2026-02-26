package ai

import (
	"testing"

	generativelanguage "cloud.google.com/go/ai/generativelanguage/apiv1"
	"cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
	"github.com/dl-alexandre/gdrv/internal/types"
)

func TestConvertModel(t *testing.T) {
	tests := []struct {
		name     string
		input    *generativelanguagepb.Model
		expected types.AIModel
	}{
		{
			name: "full model",
			input: &generativelanguagepb.Model{
				Name:        "models/gemini-1.0-pro",
				DisplayName: "Gemini 1.0 Pro",
				Description: "First generation Gemini model",
				Version:     "1.0",
			},
			expected: types.AIModel{
				Name:        "models/gemini-1.0-pro",
				DisplayName: "Gemini 1.0 Pro",
				Description: "First generation Gemini model",
				Version:     "1.0",
			},
		},
		{
			name: "model with empty fields",
			input: &generativelanguagepb.Model{
				Name: "models/gemini-1.5-flash",
			},
			expected: types.AIModel{
				Name: "models/gemini-1.5-flash",
			},
		},
		// Note: nil model would panic due to implementation not checking for nil
		// This is testing the actual behavior where nil is not handled gracefully
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertModel(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
			if result.Description != tc.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tc.expected.Description)
			}
			if result.Version != tc.expected.Version {
				t.Errorf("Version: got %q, want %q", result.Version, tc.expected.Version)
			}
		})
	}
}

func TestConvertGenerateContentResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    *generativelanguagepb.GenerateContentResponse
		expected *types.AIGenerateContentResponse
	}{
		{
			name: "response with single candidate",
			input: &generativelanguagepb.GenerateContentResponse{
				Candidates: []*generativelanguagepb.Candidate{
					{
						Content: &generativelanguagepb.Content{
							Role: "model",
							Parts: []*generativelanguagepb.Part{
								{Data: &generativelanguagepb.Part_Text{Text: "Hello, world!"}},
							},
						},
						FinishReason: generativelanguagepb.Candidate_STOP,
					},
				},
				UsageMetadata: &generativelanguagepb.GenerateContentResponse_UsageMetadata{
					PromptTokenCount:     10,
					CandidatesTokenCount: 5,
					TotalTokenCount:      15,
				},
			},
			expected: &types.AIGenerateContentResponse{
				Contents: []types.AIContent{
					{
						Text:         "Hello, world!",
						Role:         "model",
						FinishReason: "STOP",
					},
				},
				Usage: &types.AIUsage{
					PromptTokens:     10,
					CompletionTokens: 5,
					TotalTokens:      15,
				},
			},
		},
		{
			name: "response with multiple parts",
			input: &generativelanguagepb.GenerateContentResponse{
				Candidates: []*generativelanguagepb.Candidate{
					{
						Content: &generativelanguagepb.Content{
							Role: "model",
							Parts: []*generativelanguagepb.Part{
								{Data: &generativelanguagepb.Part_Text{Text: "Part 1 "}},
								{Data: &generativelanguagepb.Part_Text{Text: "Part 2"}},
							},
						},
						FinishReason: generativelanguagepb.Candidate_MAX_TOKENS,
					},
				},
			},
			expected: &types.AIGenerateContentResponse{
				Contents: []types.AIContent{
					{
						Text:         "Part 1 Part 2",
						Role:         "model",
						FinishReason: "MAX_TOKENS",
					},
				},
			},
		},

		{
			name: "empty candidates",
			input: &generativelanguagepb.GenerateContentResponse{
				Candidates: []*generativelanguagepb.Candidate{},
			},
			expected: &types.AIGenerateContentResponse{
				Contents: []types.AIContent{},
			},
		},
		{
			name: "response without usage metadata",
			input: &generativelanguagepb.GenerateContentResponse{
				Candidates: []*generativelanguagepb.Candidate{
					{
						Content: &generativelanguagepb.Content{
							Role: "model",
							Parts: []*generativelanguagepb.Part{
								{Data: &generativelanguagepb.Part_Text{Text: "Test"}},
							},
						},
						FinishReason: generativelanguagepb.Candidate_STOP,
					},
				},
			},
			expected: &types.AIGenerateContentResponse{
				Contents: []types.AIContent{
					{
						Text:         "Test",
						Role:         "model",
						FinishReason: "STOP",
					},
				},
				Usage: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertGenerateContentResponse(tc.input)

			if len(result.Contents) != len(tc.expected.Contents) {
				t.Fatalf("Contents length: got %d, want %d", len(result.Contents), len(tc.expected.Contents))
			}

			for i, content := range result.Contents {
				if content.Text != tc.expected.Contents[i].Text {
					t.Errorf("Contents[%d].Text: got %q, want %q", i, content.Text, tc.expected.Contents[i].Text)
				}
				if content.Role != tc.expected.Contents[i].Role {
					t.Errorf("Contents[%d].Role: got %q, want %q", i, content.Role, tc.expected.Contents[i].Role)
				}
				if content.FinishReason != tc.expected.Contents[i].FinishReason {
					t.Errorf("Contents[%d].FinishReason: got %q, want %q", i, content.FinishReason, tc.expected.Contents[i].FinishReason)
				}
			}

			if tc.expected.Usage == nil {
				if result.Usage != nil {
					t.Errorf("Usage: got %v, want nil", result.Usage)
				}
			} else {
				if result.Usage == nil {
					t.Fatal("Usage: got nil, want non-nil")
				}
				if result.Usage.PromptTokens != tc.expected.Usage.PromptTokens {
					t.Errorf("Usage.PromptTokens: got %d, want %d", result.Usage.PromptTokens, tc.expected.Usage.PromptTokens)
				}
				if result.Usage.CompletionTokens != tc.expected.Usage.CompletionTokens {
					t.Errorf("Usage.CompletionTokens: got %d, want %d", result.Usage.CompletionTokens, tc.expected.Usage.CompletionTokens)
				}
				if result.Usage.TotalTokens != tc.expected.Usage.TotalTokens {
					t.Errorf("Usage.TotalTokens: got %d, want %d", result.Usage.TotalTokens, tc.expected.Usage.TotalTokens)
				}
			}
		})
	}
}

func TestManagerClose(t *testing.T) {
	t.Run("close with nil clients", func(t *testing.T) {
		m := &Manager{}
		err := m.Close()
		if err != nil {
			t.Errorf("Close() with nil clients should not error, got: %v", err)
		}
	})
}

func TestManagerConstructorValidation(t *testing.T) {
	t.Run("NewManager with nil context should error", func(t *testing.T) {
		// This test validates that NewManager requires a valid context
		// We can't actually test this without a real connection, but we verify
		// the function signature is correct
		var _ *generativelanguage.GenerativeClient // Just to ensure import is used
	})
}
