package types

import (
	"fmt"
	"strings"
)

type GmailMessage struct {
	ID           string   `json:"id"`
	ThreadID     string   `json:"threadId"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	Subject      string   `json:"subject"`
	Date         string   `json:"date"`
	Snippet      string   `json:"snippet"`
	LabelIDs     []string `json:"labelIds"`
	SizeEstimate int64    `json:"sizeEstimate"`
}

func (m *GmailMessage) Headers() []string {
	return []string{"ID", "From", "To", "Subject", "Date"}
}

func (m *GmailMessage) Rows() [][]string {
	return [][]string{{
		m.ID,
		m.From,
		m.To,
		m.Subject,
		m.Date,
	}}
}

func (m *GmailMessage) EmptyMessage() string {
	return "No message found"
}

type GmailMessageList struct {
	Messages           []GmailMessage `json:"messages"`
	ResultSizeEstimate int64          `json:"resultSizeEstimate"`
}

func (l *GmailMessageList) Headers() []string {
	return []string{"ID", "From", "Subject", "Date", "Labels"}
}

func (l *GmailMessageList) Rows() [][]string {
	rows := make([][]string, len(l.Messages))
	for i, msg := range l.Messages {
		rows[i] = []string{
			msg.ID,
			msg.From,
			msg.Subject,
			msg.Date,
			strings.Join(msg.LabelIDs, ", "),
		}
	}
	return rows
}

func (l *GmailMessageList) EmptyMessage() string {
	return "No messages found"
}

type GmailThread struct {
	ID       string         `json:"id"`
	Snippet  string         `json:"snippet"`
	Messages []GmailMessage `json:"messages"`
}

func (t *GmailThread) Headers() []string {
	return []string{"ID", "Messages", "Snippet"}
}

func (t *GmailThread) Rows() [][]string {
	return [][]string{{
		t.ID,
		fmt.Sprintf("%d", len(t.Messages)),
		t.Snippet,
	}}
}

func (t *GmailThread) EmptyMessage() string {
	return "No thread found"
}

type GmailDraft struct {
	ID      string       `json:"id"`
	Message GmailMessage `json:"message"`
}

func (d *GmailDraft) Headers() []string {
	return []string{"ID", "To", "Subject"}
}

func (d *GmailDraft) Rows() [][]string {
	return [][]string{{
		d.ID,
		d.Message.To,
		d.Message.Subject,
	}}
}

func (d *GmailDraft) EmptyMessage() string {
	return "No draft found"
}

type GmailDraftList struct {
	Drafts []GmailDraft `json:"drafts"`
}

func (l *GmailDraftList) Headers() []string {
	return []string{"ID", "To", "Subject"}
}

func (l *GmailDraftList) Rows() [][]string {
	rows := make([][]string, len(l.Drafts))
	for i, draft := range l.Drafts {
		rows[i] = []string{
			draft.ID,
			draft.Message.To,
			draft.Message.Subject,
		}
	}
	return rows
}

func (l *GmailDraftList) EmptyMessage() string {
	return "No drafts found"
}

type GmailLabel struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	MessagesTotal  int64  `json:"messagesTotal"`
	MessagesUnread int64  `json:"messagesUnread"`
	ThreadsTotal   int64  `json:"threadsTotal"`
	ThreadsUnread  int64  `json:"threadsUnread"`
}

func (l *GmailLabel) Headers() []string {
	return []string{"ID", "Name", "Type", "Total", "Unread"}
}

func (l *GmailLabel) Rows() [][]string {
	return [][]string{{
		l.ID,
		l.Name,
		l.Type,
		fmt.Sprintf("%d", l.MessagesTotal),
		fmt.Sprintf("%d", l.MessagesUnread),
	}}
}

func (l *GmailLabel) EmptyMessage() string {
	return "No label found"
}

type GmailLabelList struct {
	Labels []GmailLabel `json:"labels"`
}

func (l *GmailLabelList) Headers() []string {
	return []string{"ID", "Name", "Type", "Total", "Unread"}
}

func (l *GmailLabelList) Rows() [][]string {
	rows := make([][]string, len(l.Labels))
	for i, label := range l.Labels {
		rows[i] = []string{
			label.ID,
			label.Name,
			label.Type,
			fmt.Sprintf("%d", label.MessagesTotal),
			fmt.Sprintf("%d", label.MessagesUnread),
		}
	}
	return rows
}

func (l *GmailLabelList) EmptyMessage() string {
	return "No labels found"
}

type GmailFilter struct {
	ID                   string   `json:"id"`
	CriteriaFrom         string   `json:"criteriaFrom"`
	CriteriaTo           string   `json:"criteriaTo"`
	CriteriaSubject      string   `json:"criteriaSubject"`
	CriteriaQuery        string   `json:"criteriaQuery"`
	CriteriaHasAttachment bool    `json:"criteriaHasAttachment"`
	ActionAddLabelIDs    []string `json:"actionAddLabelIds"`
	ActionRemoveLabelIDs []string `json:"actionRemoveLabelIds"`
	ActionForward        string   `json:"actionForward"`
}

func (f *GmailFilter) Headers() []string {
	return []string{"ID", "From", "To", "Subject", "Query"}
}

func (f *GmailFilter) Rows() [][]string {
	return [][]string{{
		f.ID,
		f.CriteriaFrom,
		f.CriteriaTo,
		f.CriteriaSubject,
		f.CriteriaQuery,
	}}
}

func (f *GmailFilter) EmptyMessage() string {
	return "No filter found"
}

type GmailFilterList struct {
	Filters []GmailFilter `json:"filters"`
}

func (l *GmailFilterList) Headers() []string {
	return []string{"ID", "From", "To", "Subject", "Query"}
}

func (l *GmailFilterList) Rows() [][]string {
	rows := make([][]string, len(l.Filters))
	for i, filter := range l.Filters {
		rows[i] = []string{
			filter.ID,
			filter.CriteriaFrom,
			filter.CriteriaTo,
			filter.CriteriaSubject,
			filter.CriteriaQuery,
		}
	}
	return rows
}

func (l *GmailFilterList) EmptyMessage() string {
	return "No filters found"
}

type GmailVacationSettings struct {
	EnableAutoReply       bool   `json:"enableAutoReply"`
	ResponseSubject       string `json:"responseSubject"`
	ResponseBodyPlainText string `json:"responseBodyPlainText"`
	ResponseBodyHtml      string `json:"responseBodyHtml"`
	StartTime             int64  `json:"startTime"`
	EndTime               int64  `json:"endTime"`
}

func (v *GmailVacationSettings) Headers() []string {
	return []string{"Enabled", "Subject", "Start", "End"}
}

func (v *GmailVacationSettings) Rows() [][]string {
	return [][]string{{
		fmt.Sprintf("%t", v.EnableAutoReply),
		v.ResponseSubject,
		fmt.Sprintf("%d", v.StartTime),
		fmt.Sprintf("%d", v.EndTime),
	}}
}

func (v *GmailVacationSettings) EmptyMessage() string {
	return "No vacation settings found"
}

type GmailSendAs struct {
	SendAsEmail    string `json:"sendAsEmail"`
	DisplayName    string `json:"displayName"`
	ReplyToAddress string `json:"replyToAddress"`
	IsPrimary      bool   `json:"isPrimary"`
	IsDefault      bool   `json:"isDefault"`
}

func (s *GmailSendAs) Headers() []string {
	return []string{"Email", "Display Name", "Reply To", "Primary", "Default"}
}

func (s *GmailSendAs) Rows() [][]string {
	return [][]string{{
		s.SendAsEmail,
		s.DisplayName,
		s.ReplyToAddress,
		fmt.Sprintf("%t", s.IsPrimary),
		fmt.Sprintf("%t", s.IsDefault),
	}}
}

func (s *GmailSendAs) EmptyMessage() string {
	return "No send-as alias found"
}

type GmailSendAsList struct {
	SendAs []GmailSendAs `json:"sendAs"`
}

func (l *GmailSendAsList) Headers() []string {
	return []string{"Email", "Display Name", "Reply To", "Primary", "Default"}
}

func (l *GmailSendAsList) Rows() [][]string {
	rows := make([][]string, len(l.SendAs))
	for i, sa := range l.SendAs {
		rows[i] = []string{
			sa.SendAsEmail,
			sa.DisplayName,
			sa.ReplyToAddress,
			fmt.Sprintf("%t", sa.IsPrimary),
			fmt.Sprintf("%t", sa.IsDefault),
		}
	}
	return rows
}

func (l *GmailSendAsList) EmptyMessage() string {
	return "No send-as aliases found"
}

type GmailSendResult struct {
	ID       string   `json:"id"`
	ThreadID string   `json:"threadId"`
	LabelIDs []string `json:"labelIds"`
}

func (r *GmailSendResult) Headers() []string {
	return []string{"ID", "Thread ID", "Labels"}
}

func (r *GmailSendResult) Rows() [][]string {
	return [][]string{{
		r.ID,
		r.ThreadID,
		strings.Join(r.LabelIDs, ", "),
	}}
}

func (r *GmailSendResult) EmptyMessage() string {
	return "No send result available"
}

type GmailAttachment struct {
	MessageID    string `json:"messageId"`
	AttachmentID string `json:"attachmentId"`
	Filename     string `json:"filename"`
	MimeType     string `json:"mimeType"`
	Size         int64  `json:"size"`
}

func (a *GmailAttachment) Headers() []string {
	return []string{"Message ID", "Attachment ID", "Filename", "Type", "Size"}
}

func (a *GmailAttachment) Rows() [][]string {
	return [][]string{{
		a.MessageID,
		a.AttachmentID,
		a.Filename,
		a.MimeType,
		fmt.Sprintf("%d", a.Size),
	}}
}

func (a *GmailAttachment) EmptyMessage() string {
	return "No attachment found"
}

type GmailBatchResult struct {
	Count  int    `json:"count"`
	Action string `json:"action"`
}

func (r *GmailBatchResult) Headers() []string {
	return []string{"Action", "Count"}
}

func (r *GmailBatchResult) Rows() [][]string {
	return [][]string{{
		r.Action,
		fmt.Sprintf("%d", r.Count),
	}}
}

func (r *GmailBatchResult) EmptyMessage() string {
	return "No batch result available"
}
