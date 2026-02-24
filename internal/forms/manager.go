package forms

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/types"
	forms "google.golang.org/api/forms/v1"
)

// Manager handles Google Forms API operations.
type Manager struct {
	client  *api.Client
	service *forms.Service
}

// NewManager creates a new Forms manager.
func NewManager(client *api.Client, service *forms.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// GetForm retrieves a form by ID including its items and questions.
func (m *Manager) GetForm(ctx context.Context, reqCtx *types.RequestContext, formID string) (*types.Form, error) {
	call := m.service.Forms.Get(formID)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*forms.Form, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertForm(result), nil
}

// ListResponses lists responses for a form. Answers are keyed by questionId,
// not by question title. Returns the response list, the next page token, and
// any error.
func (m *Manager) ListResponses(ctx context.Context, reqCtx *types.RequestContext, formID string, pageSize int64, pageToken string) (*types.FormResponseList, string, error) {
	call := m.service.Forms.Responses.List(formID)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*forms.ListFormResponsesResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	responses := make([]types.FormResponse, len(result.Responses))
	for i, r := range result.Responses {
		responses[i] = convertResponse(r)
	}

	return &types.FormResponseList{Responses: responses}, result.NextPageToken, nil
}

// CreateForm creates a new form with the given title.
//
// NOTE: Forms created after March 31, 2026 are unpublished by default.
// Callers may need to publish the form separately via SetPublishSettings.
func (m *Manager) CreateForm(ctx context.Context, reqCtx *types.RequestContext, title string) (*types.FormCreateResult, error) {
	form := &forms.Form{
		Info: &forms.Info{
			Title: title,
		},
	}

	call := m.service.Forms.Create(form)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*forms.Form, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	formTitle := ""
	if result.Info != nil {
		formTitle = result.Info.Title
	}

	return &types.FormCreateResult{
		FormID:       result.FormId,
		Title:        formTitle,
		ResponderURI: result.ResponderUri,
	}, nil
}

// convertForm maps a Google Forms API Form to the internal types.Form.
func convertForm(f *forms.Form) *types.Form {
	if f == nil {
		return &types.Form{}
	}

	result := &types.Form{
		FormID:        f.FormId,
		RevisionID:    f.RevisionId,
		ResponderURI:  f.ResponderUri,
		LinkedSheetID: f.LinkedSheetId,
	}

	if f.Info != nil {
		result.Title = f.Info.Title
		result.Description = f.Info.Description
		result.DocumentTitle = f.Info.DocumentTitle
	}

	if f.Items != nil {
		result.Items = make([]types.FormItem, len(f.Items))
		for i, item := range f.Items {
			result.Items[i] = convertFormItem(item)
		}
	}

	return result
}

// convertFormItem extracts the question type from a form item.
// It checks QuestionItem first, then QuestionGroupItem, and defaults to
// an empty question type for non-question items (e.g. text, image, video).
func convertFormItem(item *forms.Item) types.FormItem {
	if item == nil {
		return types.FormItem{}
	}

	fi := types.FormItem{
		ItemID:      item.ItemId,
		Title:       item.Title,
		Description: item.Description,
	}

	if item.QuestionItem != nil && item.QuestionItem.Question != nil {
		fi.QuestionType = questionType(item.QuestionItem.Question)
		fi.Required = item.QuestionItem.Question.Required
	} else if item.QuestionGroupItem != nil && len(item.QuestionGroupItem.Questions) > 0 {
		fi.QuestionType = questionType(item.QuestionGroupItem.Questions[0])
		fi.Required = item.QuestionGroupItem.Questions[0].Required
	}

	return fi
}

// questionType determines the question kind string from a Question struct.
func questionType(q *forms.Question) string {
	if q == nil {
		return ""
	}
	switch {
	case q.ChoiceQuestion != nil:
		return "CHOICE"
	case q.TextQuestion != nil:
		return "TEXT"
	case q.ScaleQuestion != nil:
		return "SCALE"
	case q.DateQuestion != nil:
		return "DATE"
	case q.TimeQuestion != nil:
		return "TIME"
	case q.FileUploadQuestion != nil:
		return "FILE_UPLOAD"
	case q.RowQuestion != nil:
		return "ROW"
	case q.RatingQuestion != nil:
		return "RATING"
	default:
		return ""
	}
}

// convertResponse maps a Google Forms API FormResponse to the internal
// types.FormResponse. Answers are keyed by questionId, not by question title.
func convertResponse(r *forms.FormResponse) types.FormResponse {
	if r == nil {
		return types.FormResponse{}
	}

	resp := types.FormResponse{
		ResponseID:        r.ResponseId,
		CreateTime:        r.CreateTime,
		LastSubmittedTime: r.LastSubmittedTime,
		RespondentEmail:   r.RespondentEmail,
	}

	if r.Answers != nil {
		resp.Answers = make(map[string]types.FormAnswer, len(r.Answers))
		for qID, answer := range r.Answers {
			fa := types.FormAnswer{
				QuestionID: answer.QuestionId,
			}

			if answer.TextAnswers != nil {
				fa.TextAnswers = make([]string, len(answer.TextAnswers.Answers))
				for i, ta := range answer.TextAnswers.Answers {
					fa.TextAnswers[i] = ta.Value
				}
			}

			if answer.FileUploadAnswers != nil {
				fa.FileUploadAnswers = make([]string, len(answer.FileUploadAnswers.Answers))
				for i, fua := range answer.FileUploadAnswers.Answers {
					fa.FileUploadAnswers[i] = fua.FileId
				}
			}

			resp.Answers[qID] = fa
		}
	}

	return resp
}
