package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	gmail "google.golang.org/api/gmail/v1"
)

// Manager wraps the Gmail API with business logic and retry handling.
type Manager struct {
	client  *api.Client
	service *gmail.Service
}

// NewManager creates a new Gmail Manager.
func NewManager(client *api.Client, service *gmail.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// Search lists messages matching a Gmail search query.
// Returns message stubs (ID + ThreadID only) plus a nextPageToken for pagination.
func (m *Manager) Search(ctx context.Context, reqCtx *types.RequestContext, query string, maxResults int64, pageToken string) (*types.GmailMessageList, string, error) {
	call := m.service.Users.Messages.List("me")
	if query != "" {
		call = call.Q(query)
	}
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.ListMessagesResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	messages := make([]types.GmailMessage, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = types.GmailMessage{
			ID:       msg.Id,
			ThreadID: msg.ThreadId,
		}
	}

	return &types.GmailMessageList{
		Messages:           messages,
		ResultSizeEstimate: int64(result.ResultSizeEstimate),
	}, result.NextPageToken, nil
}

// GetMessage retrieves a single message by ID.
// Format may be "full", "metadata", or "minimal".
func (m *Manager) GetMessage(ctx context.Context, reqCtx *types.RequestContext, messageID, format string) (*types.GmailMessage, error) {
	call := m.service.Users.Messages.Get("me", messageID)
	if format != "" {
		call = call.Format(format)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertMessage(result), nil
}

// GetThread retrieves a thread and all its messages.
func (m *Manager) GetThread(ctx context.Context, reqCtx *types.RequestContext, threadID, format string) (*types.GmailThread, error) {
	call := m.service.Users.Threads.Get("me", threadID)
	if format != "" {
		call = call.Format(format)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Thread, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	messages := make([]types.GmailMessage, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = *convertMessage(msg)
	}

	return &types.GmailThread{
		ID:       result.Id,
		Snippet:  result.Snippet,
		Messages: messages,
	}, nil
}

// Send composes and sends an email message.
// If inReplyTo is set, the message is threaded as a reply with In-Reply-To and References headers.
func (m *Manager) Send(ctx context.Context, reqCtx *types.RequestContext, to, subject, body, cc, bcc, htmlBody, inReplyTo, threadID string) (*types.GmailSendResult, error) {
	raw := composeMessage("me", to, cc, bcc, subject, body, htmlBody, inReplyTo)
	encoded := base64.URLEncoding.EncodeToString(raw)

	msg := &gmail.Message{
		Raw: encoded,
	}
	if inReplyTo != "" && threadID != "" {
		msg.ThreadId = threadID
	}

	call := m.service.Users.Messages.Send("me", msg)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.GmailSendResult{
		ID:       result.Id,
		ThreadID: result.ThreadId,
		LabelIDs: result.LabelIds,
	}, nil
}

// ListDrafts lists the user's drafts.
func (m *Manager) ListDrafts(ctx context.Context, reqCtx *types.RequestContext, maxResults int64, pageToken string) (*types.GmailDraftList, string, error) {
	call := m.service.Users.Drafts.List("me")
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.ListDraftsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	drafts := make([]types.GmailDraft, len(result.Drafts))
	for i, d := range result.Drafts {
		drafts[i] = types.GmailDraft{
			ID: d.Id,
		}
		if d.Message != nil {
			drafts[i].Message = *convertMessage(d.Message)
		}
	}

	return &types.GmailDraftList{
		Drafts: drafts,
	}, result.NextPageToken, nil
}

// GetDraft retrieves a single draft by ID.
func (m *Manager) GetDraft(ctx context.Context, reqCtx *types.RequestContext, draftID string) (*types.GmailDraft, error) {
	call := m.service.Users.Drafts.Get("me", draftID)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Draft, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	draft := &types.GmailDraft{
		ID: result.Id,
	}
	if result.Message != nil {
		draft.Message = *convertMessage(result.Message)
	}

	return draft, nil
}

// CreateDraft creates a new draft message.
func (m *Manager) CreateDraft(ctx context.Context, reqCtx *types.RequestContext, to, subject, body string) (*types.GmailDraft, error) {
	raw := composeMessage("me", to, "", "", subject, body, "", "")
	encoded := base64.URLEncoding.EncodeToString(raw)

	draft := &gmail.Draft{
		Message: &gmail.Message{
			Raw: encoded,
		},
	}
	call := m.service.Users.Drafts.Create("me", draft)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Draft, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	d := &types.GmailDraft{
		ID: result.Id,
	}
	if result.Message != nil {
		d.Message = *convertMessage(result.Message)
	}

	return d, nil
}

// UpdateDraft replaces the content of an existing draft.
func (m *Manager) UpdateDraft(ctx context.Context, reqCtx *types.RequestContext, draftID, to, subject, body string) (*types.GmailDraft, error) {
	raw := composeMessage("me", to, "", "", subject, body, "", "")
	encoded := base64.URLEncoding.EncodeToString(raw)

	draft := &gmail.Draft{
		Message: &gmail.Message{
			Raw: encoded,
		},
	}
	call := m.service.Users.Drafts.Update("me", draftID, draft)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Draft, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	d := &types.GmailDraft{
		ID: result.Id,
	}
	if result.Message != nil {
		d.Message = *convertMessage(result.Message)
	}

	return d, nil
}

// DeleteDraft permanently deletes a draft. This cannot be undone.
func (m *Manager) DeleteDraft(ctx context.Context, reqCtx *types.RequestContext, draftID string) error {
	call := m.service.Users.Drafts.Delete("me", draftID)

	// Wrap void API call so it returns a value compatible with ExecuteWithRetry.
	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Draft, error) {
		return nil, call.Do()
	})
	return err
}

// SendDraft sends an existing draft.
func (m *Manager) SendDraft(ctx context.Context, reqCtx *types.RequestContext, draftID string) (*types.GmailSendResult, error) {
	draft := &gmail.Draft{
		Id: draftID,
	}
	call := m.service.Users.Drafts.Send("me", draft)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.GmailSendResult{
		ID:       result.Id,
		ThreadID: result.ThreadId,
		LabelIDs: result.LabelIds,
	}, nil
}

// ListLabels returns all labels for the authenticated user.
func (m *Manager) ListLabels(ctx context.Context, reqCtx *types.RequestContext) (*types.GmailLabelList, error) {
	call := m.service.Users.Labels.List("me")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.ListLabelsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	labels := make([]types.GmailLabel, len(result.Labels))
	for i, l := range result.Labels {
		labels[i] = convertLabel(l)
	}

	return &types.GmailLabelList{
		Labels: labels,
	}, nil
}

// CreateLabel creates a new user label.
func (m *Manager) CreateLabel(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.GmailLabel, error) {
	label := &gmail.Label{
		Name:                    name,
		LabelListVisibility:     "labelShow",
		MessageListVisibility:   "show",
	}
	call := m.service.Users.Labels.Create("me", label)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Label, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	converted := convertLabel(result)
	return &converted, nil
}

// DeleteLabel permanently deletes a label. Built-in labels cannot be deleted.
func (m *Manager) DeleteLabel(ctx context.Context, reqCtx *types.RequestContext, labelID string) error {
	call := m.service.Users.Labels.Delete("me", labelID)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Label, error) {
		return nil, call.Do()
	})
	return err
}

// ListFilters returns all filters for the authenticated user.
func (m *Manager) ListFilters(ctx context.Context, reqCtx *types.RequestContext) (*types.GmailFilterList, error) {
	call := m.service.Users.Settings.Filters.List("me")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.ListFiltersResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	filters := make([]types.GmailFilter, len(result.Filter))
	for i, f := range result.Filter {
		filters[i] = convertFilter(f)
	}

	return &types.GmailFilterList{
		Filters: filters,
	}, nil
}

// CreateFilter creates a new email filter.
func (m *Manager) CreateFilter(ctx context.Context, reqCtx *types.RequestContext, from, to, subject, query string, hasAttachment bool, addLabelIDs, removeLabelIDs []string, forward string) (*types.GmailFilter, error) {
	filter := &gmail.Filter{
		Criteria: &gmail.FilterCriteria{
			From:          from,
			To:            to,
			Subject:       subject,
			Query:         query,
			HasAttachment: hasAttachment,
		},
		Action: &gmail.FilterAction{
			AddLabelIds:    addLabelIDs,
			RemoveLabelIds: removeLabelIDs,
			Forward:        forward,
		},
	}
	call := m.service.Users.Settings.Filters.Create("me", filter)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Filter, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	converted := convertFilter(result)
	return &converted, nil
}

// DeleteFilter permanently deletes a filter.
func (m *Manager) DeleteFilter(ctx context.Context, reqCtx *types.RequestContext, filterID string) error {
	call := m.service.Users.Settings.Filters.Delete("me", filterID)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Filter, error) {
		return nil, call.Do()
	})
	return err
}

// GetVacation retrieves the current vacation responder settings.
func (m *Manager) GetVacation(ctx context.Context, reqCtx *types.RequestContext) (*types.GmailVacationSettings, error) {
	call := m.service.Users.Settings.GetVacation("me")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.VacationSettings, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertVacationSettings(result), nil
}

// SetVacation updates the vacation responder settings.
func (m *Manager) SetVacation(ctx context.Context, reqCtx *types.RequestContext, settings *types.GmailVacationSettings) (*types.GmailVacationSettings, error) {
	vs := &gmail.VacationSettings{
		EnableAutoReply:       settings.EnableAutoReply,
		ResponseSubject:       settings.ResponseSubject,
		ResponseBodyPlainText: settings.ResponseBodyPlainText,
		ResponseBodyHtml:      settings.ResponseBodyHtml,
		StartTime:             settings.StartTime,
		EndTime:               settings.EndTime,
	}
	call := m.service.Users.Settings.UpdateVacation("me", vs)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.VacationSettings, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return convertVacationSettings(result), nil
}

// ListSendAs returns all send-as aliases for the authenticated user.
func (m *Manager) ListSendAs(ctx context.Context, reqCtx *types.RequestContext) (*types.GmailSendAsList, error) {
	call := m.service.Users.Settings.SendAs.List("me")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.ListSendAsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	aliases := make([]types.GmailSendAs, len(result.SendAs))
	for i, sa := range result.SendAs {
		aliases[i] = types.GmailSendAs{
			SendAsEmail:    sa.SendAsEmail,
			DisplayName:    sa.DisplayName,
			ReplyToAddress: sa.ReplyToAddress,
			IsPrimary:      sa.IsPrimary,
			IsDefault:      sa.IsDefault,
		}
	}

	return &types.GmailSendAsList{
		SendAs: aliases,
	}, nil
}

// BatchDelete permanently deletes multiple messages by ID.
func (m *Manager) BatchDelete(ctx context.Context, reqCtx *types.RequestContext, messageIDs []string) (*types.GmailBatchResult, error) {
	req := &gmail.BatchDeleteMessagesRequest{
		Ids: messageIDs,
	}
	call := m.service.Users.Messages.BatchDelete("me", req)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return nil, call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.GmailBatchResult{
		Count:  len(messageIDs),
		Action: "delete",
	}, nil
}

// BatchModify adds or removes labels from multiple messages.
func (m *Manager) BatchModify(ctx context.Context, reqCtx *types.RequestContext, messageIDs, addLabelIDs, removeLabelIDs []string) (*types.GmailBatchResult, error) {
	req := &gmail.BatchModifyMessagesRequest{
		Ids:            messageIDs,
		AddLabelIds:    addLabelIDs,
		RemoveLabelIds: removeLabelIDs,
	}
	call := m.service.Users.Messages.BatchModify("me", req)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return nil, call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.GmailBatchResult{
		Count:  len(messageIDs),
		Action: "modify",
	}, nil
}

// GetAttachment retrieves an attachment's raw bytes and metadata.
func (m *Manager) GetAttachment(ctx context.Context, reqCtx *types.RequestContext, messageID, attachmentID string) ([]byte, *types.GmailAttachment, error) {
	// First get the message to find attachment metadata (filename, mimeType, size).
	msgCall := m.service.Users.Messages.Get("me", messageID)
	msg, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.Message, error) {
		return msgCall.Do()
	})
	if err != nil {
		return nil, nil, err
	}

	filename, mimeType, size := findAttachmentMeta(msg.Payload, attachmentID)

	// Now fetch the actual attachment data.
	attCall := m.service.Users.Messages.Attachments.Get("me", messageID, attachmentID)
	att, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*gmail.MessagePartBody, error) {
		return attCall.Do()
	})
	if err != nil {
		return nil, nil, err
	}

	data, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode attachment data: %w", err)
	}

	meta := &types.GmailAttachment{
		MessageID:    messageID,
		AttachmentID: attachmentID,
		Filename:     filename,
		MimeType:     mimeType,
		Size:         int64(size),
	}

	return data, meta, nil
}

// --- Helper functions ---

// composeMessage builds an RFC 2822 email message.
// If htmlBody is non-empty, a multipart/alternative message is produced.
func composeMessage(from, to, cc, bcc, subject, body, htmlBody, inReplyTo string) []byte {
	var buf bytes.Buffer

	// Write headers.
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	if cc != "" {
		fmt.Fprintf(&buf, "Cc: %s\r\n", cc)
	}
	if bcc != "" {
		fmt.Fprintf(&buf, "Bcc: %s\r\n", bcc)
	}
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)

	if inReplyTo != "" {
		fmt.Fprintf(&buf, "In-Reply-To: %s\r\n", inReplyTo)
		fmt.Fprintf(&buf, "References: %s\r\n", inReplyTo)
	}

	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		// Multipart/alternative for HTML + plain text.
		writer := multipart.NewWriter(&buf)
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%s\r\n", writer.Boundary())
		fmt.Fprintf(&buf, "\r\n")

		// Plain text part.
		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textPart, _ := writer.CreatePart(textHeader)
		fmt.Fprint(textPart, body)

		// HTML part.
		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
		htmlPart, _ := writer.CreatePart(htmlHeader)
		fmt.Fprint(htmlPart, htmlBody)

		writer.Close()
	} else {
		// Simple plain text message.
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "\r\n")
		fmt.Fprint(&buf, body)
	}

	return buf.Bytes()
}

// parseHeaders extracts well-known headers (From, To, Subject, Date) from a Gmail message payload.
func parseHeaders(headers []*gmail.MessagePartHeader) (from, to, subject, date string) {
	for _, h := range headers {
		switch strings.ToLower(h.Name) {
		case "from":
			from = h.Value
		case "to":
			to = h.Value
		case "subject":
			subject = h.Value
		case "date":
			date = h.Value
		}
	}
	return
}

// convertMessage converts a Gmail API Message to the domain type.
func convertMessage(msg *gmail.Message) *types.GmailMessage {
	if msg == nil {
		return &types.GmailMessage{}
	}

	result := &types.GmailMessage{
		ID:           msg.Id,
		ThreadID:     msg.ThreadId,
		Snippet:      msg.Snippet,
		LabelIDs:     msg.LabelIds,
		SizeEstimate: msg.SizeEstimate,
	}

	if msg.Payload != nil && msg.Payload.Headers != nil {
		result.From, result.To, result.Subject, result.Date = parseHeaders(msg.Payload.Headers)
	}

	return result
}

// convertLabel converts a Gmail API Label to the domain type.
func convertLabel(l *gmail.Label) types.GmailLabel {
	if l == nil {
		return types.GmailLabel{}
	}
	return types.GmailLabel{
		ID:             l.Id,
		Name:           l.Name,
		Type:           l.Type,
		MessagesTotal:  int64(l.MessagesTotal),
		MessagesUnread: int64(l.MessagesUnread),
		ThreadsTotal:   int64(l.ThreadsTotal),
		ThreadsUnread:  int64(l.ThreadsUnread),
	}
}

// convertFilter converts a Gmail API Filter to the domain type.
func convertFilter(f *gmail.Filter) types.GmailFilter {
	if f == nil {
		return types.GmailFilter{}
	}

	filter := types.GmailFilter{
		ID: f.Id,
	}

	if f.Criteria != nil {
		filter.CriteriaFrom = f.Criteria.From
		filter.CriteriaTo = f.Criteria.To
		filter.CriteriaSubject = f.Criteria.Subject
		filter.CriteriaQuery = f.Criteria.Query
		filter.CriteriaHasAttachment = f.Criteria.HasAttachment
	}

	if f.Action != nil {
		filter.ActionAddLabelIDs = f.Action.AddLabelIds
		filter.ActionRemoveLabelIDs = f.Action.RemoveLabelIds
		filter.ActionForward = f.Action.Forward
	}

	return filter
}

// convertVacationSettings converts a Gmail API VacationSettings to the domain type.
func convertVacationSettings(vs *gmail.VacationSettings) *types.GmailVacationSettings {
	if vs == nil {
		return &types.GmailVacationSettings{}
	}
	return &types.GmailVacationSettings{
		EnableAutoReply:       vs.EnableAutoReply,
		ResponseSubject:       vs.ResponseSubject,
		ResponseBodyPlainText: vs.ResponseBodyPlainText,
		ResponseBodyHtml:      vs.ResponseBodyHtml,
		StartTime:             vs.StartTime,
		EndTime:               vs.EndTime,
	}
}

// findAttachmentMeta walks a message payload tree to find metadata for a given attachment ID.
func findAttachmentMeta(part *gmail.MessagePart, attachmentID string) (filename, mimeType string, size int64) {
	if part == nil {
		return "", "", 0
	}
	if part.Body != nil && part.Body.AttachmentId == attachmentID {
		return part.Filename, part.MimeType, int64(part.Body.Size)
	}
	for _, child := range part.Parts {
		if fn, mt, sz := findAttachmentMeta(child, attachmentID); fn != "" || mt != "" {
			return fn, mt, sz
		}
	}
	return "", "", 0
}
