package ai

import (
	"context"
	"fmt"
	"io"

	generativelanguage "cloud.google.com/go/ai/generativelanguage/apiv1"
	"cloud.google.com/go/ai/generativelanguage/apiv1/generativelanguagepb"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Manager handles Google Generative Language (Gemini) API operations via gRPC
type Manager struct {
	genClient   *generativelanguage.GenerativeClient
	modelClient *generativelanguage.ModelClient
}

// NewManager creates a new AI manager with gRPC clients
func NewManager(ctx context.Context, opts ...option.ClientOption) (*Manager, error) {
	genClient, err := generativelanguage.NewGenerativeClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create generative client: %w", err)
	}

	modelClient, err := generativelanguage.NewModelClient(ctx, opts...)
	if err != nil {
		genClient.Close()
		return nil, fmt.Errorf("failed to create model client: %w", err)
	}

	return &Manager{
		genClient:   genClient,
		modelClient: modelClient,
	}, nil
}

// Close closes the gRPC connections
func (m *Manager) Close() error {
	var genErr, modelErr error
	if m.genClient != nil {
		genErr = m.genClient.Close()
	}
	if m.modelClient != nil {
		modelErr = m.modelClient.Close()
	}
	if genErr != nil {
		return genErr
	}
	return modelErr
}

// ListModels lists available Gemini models
func (m *Manager) ListModels(ctx context.Context, reqCtx *types.RequestContext, pageSize int32, pageToken string) (*types.AIModelsListResponse, error) {
	req := &generativelanguagepb.ListModelsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	}

	iter := m.modelClient.ListModels(ctx, req)
	var models []types.AIModel
	for {
		model, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}
		models = append(models, convertModel(model))
	}

	return &types.AIModelsListResponse{
		Models: models,
	}, nil
}

// GetModel gets details about a specific model
func (m *Manager) GetModel(ctx context.Context, reqCtx *types.RequestContext, modelName string) (*types.AIModel, error) {
	req := &generativelanguagepb.GetModelRequest{
		Name: modelName,
	}

	resp, err := m.modelClient.GetModel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	model := convertModel(resp)
	return &model, nil
}

// GenerateContent generates content using a Gemini model
func (m *Manager) GenerateContent(ctx context.Context, reqCtx *types.RequestContext, modelName, prompt string, temperature float32) (*types.AIGenerateContentResponse, error) {
	req := &generativelanguagepb.GenerateContentRequest{
		Model: modelName,
		Contents: []*generativelanguagepb.Content{
			{
				Role: "user",
				Parts: []*generativelanguagepb.Part{
					{
						Data: &generativelanguagepb.Part_Text{
							Text: prompt,
						},
					},
				},
			},
		},
		GenerationConfig: &generativelanguagepb.GenerationConfig{
			Temperature: &temperature,
		},
	}

	resp, err := m.genClient.GenerateContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return convertGenerateContentResponse(resp), nil
}

// StreamGenerateContent streams generated content from a Gemini model
func (m *Manager) StreamGenerateContent(ctx context.Context, reqCtx *types.RequestContext, modelName, prompt string, temperature float32) (<-chan *types.AIContent, <-chan error) {
	contentChan := make(chan *types.AIContent)
	errChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errChan)

		req := &generativelanguagepb.GenerateContentRequest{
			Model: modelName,
			Contents: []*generativelanguagepb.Content{
				{
					Role: "user",
					Parts: []*generativelanguagepb.Part{
						{
							Data: &generativelanguagepb.Part_Text{
								Text: prompt,
							},
						},
					},
				},
			},
			GenerationConfig: &generativelanguagepb.GenerationConfig{
				Temperature: &temperature,
			},
		}

		stream, err := m.genClient.StreamGenerateContent(ctx, req)
		if err != nil {
			errChan <- fmt.Errorf("failed to start stream: %w", err)
			return
		}

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("stream error: %w", err)
				return
			}

			for _, candidate := range resp.Candidates {
				for _, part := range candidate.Content.Parts {
					if text := part.GetText(); text != "" {
						contentChan <- &types.AIContent{
							Text: text,
							Role: candidate.Content.Role,
						}
					}
				}
			}
		}
	}()

	return contentChan, errChan
}

func convertModel(model *generativelanguagepb.Model) types.AIModel {
	return types.AIModel{
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Version:     model.Version,
	}
}

func convertGenerateContentResponse(resp *generativelanguagepb.GenerateContentResponse) *types.AIGenerateContentResponse {
	contents := make([]types.AIContent, 0)
	for _, candidate := range resp.Candidates {
		var text string
		for _, part := range candidate.Content.Parts {
			if t := part.GetText(); t != "" {
				text += t
			}
		}
		contents = append(contents, types.AIContent{
			Text:         text,
			Role:         candidate.Content.Role,
			FinishReason: candidate.FinishReason.String(),
		})
	}

	var usage *types.AIUsage
	if resp.UsageMetadata != nil {
		usage = &types.AIUsage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}

	return &types.AIGenerateContentResponse{
		Contents: contents,
		Usage:    usage,
	}
}
