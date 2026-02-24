package types

import "fmt"

type FormItem struct {
	ItemID       string `json:"itemId"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	QuestionType string `json:"questionType,omitempty"`
	Required     bool   `json:"required"`
}

type Form struct {
	FormID        string     `json:"formId"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	DocumentTitle string     `json:"documentTitle,omitempty"`
	RevisionID    string     `json:"revisionId,omitempty"`
	ResponderURI  string     `json:"responderUri,omitempty"`
	LinkedSheetID string     `json:"linkedSheetId,omitempty"`
	Items         []FormItem `json:"items,omitempty"`
}

func (f *Form) Headers() []string {
	return []string{"ID", "Title", "Items", "Responder URL"}
}

func (f *Form) Rows() [][]string {
	return [][]string{{
		f.FormID,
		f.Title,
		fmt.Sprintf("%d", len(f.Items)),
		f.ResponderURI,
	}}
}

func (f *Form) EmptyMessage() string {
	return "No form found"
}

type FormAnswer struct {
	QuestionID        string   `json:"questionId"`
	TextAnswers       []string `json:"textAnswers,omitempty"`
	FileUploadAnswers []string `json:"fileUploadAnswers,omitempty"`
	ChoiceAnswers     []string `json:"choiceAnswers,omitempty"`
}

type FormResponse struct {
	ResponseID        string                `json:"responseId"`
	CreateTime        string                `json:"createTime,omitempty"`
	LastSubmittedTime string                `json:"lastSubmittedTime,omitempty"`
	RespondentEmail   string                `json:"respondentEmail,omitempty"`
	Answers           map[string]FormAnswer `json:"answers,omitempty"`
}

func (r *FormResponse) Headers() []string {
	return []string{"Response ID", "Email", "Submitted"}
}

func (r *FormResponse) Rows() [][]string {
	return [][]string{{
		r.ResponseID,
		r.RespondentEmail,
		r.LastSubmittedTime,
	}}
}

func (r *FormResponse) EmptyMessage() string {
	return "No response found"
}

type FormResponseList struct {
	Responses []FormResponse `json:"responses,omitempty"`
}

func (l *FormResponseList) Headers() []string {
	return []string{"Response ID", "Email", "Submitted"}
}

func (l *FormResponseList) Rows() [][]string {
	rows := make([][]string, len(l.Responses))
	for i, r := range l.Responses {
		rows[i] = []string{
			r.ResponseID,
			r.RespondentEmail,
			r.LastSubmittedTime,
		}
	}
	return rows
}

func (l *FormResponseList) EmptyMessage() string {
	return "No responses found"
}

type FormCreateResult struct {
	FormID       string `json:"formId"`
	Title        string `json:"title"`
	ResponderURI string `json:"responderUri,omitempty"`
}

func (r *FormCreateResult) Headers() []string {
	return []string{"ID", "Title", "Responder URL"}
}

func (r *FormCreateResult) Rows() [][]string {
	return [][]string{{
		r.FormID,
		r.Title,
		r.ResponderURI,
	}}
}

func (r *FormCreateResult) EmptyMessage() string {
	return "No result"
}
