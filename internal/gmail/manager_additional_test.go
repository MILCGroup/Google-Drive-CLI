package gmail

import (
	"encoding/base64"
	"strings"
	"testing"

	gmailapi "google.golang.org/api/gmail/v1"
)

// TestComposeMessage_EdgeCases tests edge cases for message composition
func TestComposeMessage_EdgeCases(t *testing.T) {
	t.Run("empty subject and body", func(t *testing.T) {
		raw := composeMessage("from@example.com", "to@example.com", "", "", "", "", "", "")
		msg := string(raw)
		if !strings.Contains(msg, "From: from@example.com") {
			t.Error("expected From header")
		}
		if !strings.Contains(msg, "Subject: ") {
			t.Error("expected Subject header (even if empty)")
		}
	})

	t.Run("special characters in content", func(t *testing.T) {
		raw := composeMessage("from@example.com", "to@example.com", "", "", "Subject", "Line1\nLine2\tTabbed", "", "")
		msg := string(raw)
		if !strings.Contains(msg, "Line1\nLine2\tTabbed") {
			t.Error("expected special characters preserved in body")
		}
	})

	t.Run("unicode in body", func(t *testing.T) {
		raw := composeMessage("from@example.com", "to@example.com", "", "", "Subject", "Hello 世界", "", "")
		msg := string(raw)
		if !strings.Contains(msg, "Hello 世界") {
			t.Error("expected unicode preserved in body")
		}
	})

	t.Run("html with special characters", func(t *testing.T) {
		html := "<p>Hello &amp; Welcome</p>"
		raw := composeMessage("from@example.com", "to@example.com", "", "", "Subject", "", html, "")
		msg := string(raw)
		if !strings.Contains(msg, html) {
			t.Error("expected HTML special characters preserved")
		}
	})

	t.Run("long subject line", func(t *testing.T) {
		longSubject := strings.Repeat("A", 200)
		raw := composeMessage("from@example.com", "to@example.com", "", "", longSubject, "Body", "", "")
		msg := string(raw)
		if !strings.Contains(msg, "Subject: "+longSubject) {
			t.Error("expected long subject preserved")
		}
	})

	t.Run("multiple recipients in headers", func(t *testing.T) {
		raw := composeMessage("sender@example.com", "rec1@example.com,rec2@example.com", "cc1@example.com,cc2@example.com", "bcc@example.com", "Subject", "Body", "", "")
		msg := string(raw)
		if !strings.Contains(msg, "To: rec1@example.com,rec2@example.com") {
			t.Error("expected multiple To recipients")
		}
		if !strings.Contains(msg, "Cc: cc1@example.com,cc2@example.com") {
			t.Error("expected multiple Cc recipients")
		}
		if !strings.Contains(msg, "Bcc: bcc@example.com") {
			t.Error("expected Bcc recipient")
		}
	})
}

// TestParseHeaders_EdgeCases tests edge cases for header parsing
func TestParseHeaders_EdgeCases(t *testing.T) {
	t.Run("headers with empty values", func(t *testing.T) {
		headers := []*gmailapi.MessagePartHeader{
			{Name: "From", Value: ""},
			{Name: "To", Value: ""},
			{Name: "Subject", Value: ""},
			{Name: "Date", Value: ""},
		}
		from, to, subject, date := parseHeaders(headers)
		if from != "" || to != "" || subject != "" || date != "" {
			t.Error("expected all empty values")
		}
	})

	t.Run("headers with whitespace values", func(t *testing.T) {
		headers := []*gmailapi.MessagePartHeader{
			{Name: "From", Value: "   "},
			{Name: "Subject", Value: "  Subject  "},
		}
		from, _, subject, _ := parseHeaders(headers)
		if from != "   " {
			t.Errorf("expected whitespace preserved in From: %q", from)
		}
		if subject != "  Subject  " {
			t.Errorf("expected whitespace preserved in Subject: %q", subject)
		}
	})

	t.Run("mixed case header names", func(t *testing.T) {
		headers := []*gmailapi.MessagePartHeader{
			{Name: "fRoM", Value: "sender@example.com"},
			{Name: "tO", Value: "recipient@example.com"},
			{Name: "SuBjEcT", Value: "Test"},
			{Name: "dAtE", Value: "Mon, 1 Jan 2024"},
		}
		from, to, subject, date := parseHeaders(headers)
		if from != "sender@example.com" {
			t.Errorf("expected from to work with mixed case: %q", from)
		}
		if to != "recipient@example.com" {
			t.Errorf("expected to to work with mixed case: %q", to)
		}
		if subject != "Test" {
			t.Errorf("expected subject to work with mixed case: %q", subject)
		}
		if date != "Mon, 1 Jan 2024" {
			t.Errorf("expected date to work with mixed case: %q", date)
		}
	})

	t.Run("unknown headers", func(t *testing.T) {
		headers := []*gmailapi.MessagePartHeader{
			{Name: "X-Custom-Header", Value: "custom-value"},
			{Name: "Received", Value: "from server"},
			{Name: "Content-Type", Value: "text/plain"},
		}
		from, to, subject, date := parseHeaders(headers)
		if from != "" || to != "" || subject != "" || date != "" {
			t.Error("expected empty values for unknown headers")
		}
	})

	t.Run("very long header values", func(t *testing.T) {
		longValue := strings.Repeat("A", 1000)
		headers := []*gmailapi.MessagePartHeader{
			{Name: "From", Value: longValue},
		}
		from, _, _, _ := parseHeaders(headers)
		if from != longValue {
			t.Error("expected long value preserved")
		}
	})
}

// TestConvertMessage_EdgeCases tests edge cases for message conversion
func TestConvertMessage_EdgeCases(t *testing.T) {
	t.Run("message with nil label IDs", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:       "msg1",
			ThreadId: "thread1",
			LabelIds: nil,
		}
		got := convertMessage(msg)
		if got.LabelIDs != nil {
			t.Error("expected nil LabelIDs")
		}
	})

	t.Run("message with empty label IDs", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:       "msg1",
			ThreadId: "thread1",
			LabelIds: []string{},
		}
		got := convertMessage(msg)
		if len(got.LabelIDs) != 0 {
			t.Errorf("expected empty LabelIDs, got %d", len(got.LabelIDs))
		}
	})

	t.Run("message with many labels", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:       "msg1",
			ThreadId: "thread1",
			LabelIds: []string{"INBOX", "UNREAD", "IMPORTANT", "CATEGORY_PERSONAL", "Label_1"},
		}
		got := convertMessage(msg)
		if len(got.LabelIDs) != 5 {
			t.Errorf("expected 5 labels, got %d", len(got.LabelIDs))
		}
	})

	t.Run("message with nil payload headers", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:      "msg1",
			Payload: &gmailapi.MessagePart{Headers: nil},
		}
		got := convertMessage(msg)
		if got.From != "" || got.To != "" || got.Subject != "" || got.Date != "" {
			t.Error("expected empty header fields when payload headers nil")
		}
	})

	t.Run("message with empty payload headers", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:      "msg1",
			Payload: &gmailapi.MessagePart{Headers: []*gmailapi.MessagePartHeader{}},
		}
		got := convertMessage(msg)
		if got.From != "" || got.To != "" || got.Subject != "" || got.Date != "" {
			t.Error("expected empty header fields when payload headers empty")
		}
	})

	t.Run("message with large size", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:           "msg1",
			SizeEstimate: 10485760, // 10MB
		}
		got := convertMessage(msg)
		if got.SizeEstimate != 10485760 {
			t.Errorf("expected size 10485760, got %d", got.SizeEstimate)
		}
	})

	t.Run("message with empty thread ID", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:       "msg1",
			ThreadId: "",
		}
		got := convertMessage(msg)
		if got.ThreadID != "" {
			t.Errorf("expected empty thread ID, got %s", got.ThreadID)
		}
	})
}

// TestConvertLabel_EdgeCases tests edge cases for label conversion
func TestConvertLabel_EdgeCases(t *testing.T) {
	t.Run("system label", func(t *testing.T) {
		label := &gmailapi.Label{
			Id:   "INBOX",
			Name: "INBOX",
			Type: "system",
		}
		got := convertLabel(label)
		if got.ID != "INBOX" {
			t.Errorf("expected INBOX ID, got %s", got.ID)
		}
		if got.Type != "system" {
			t.Errorf("expected system type, got %s", got.Type)
		}
	})

	t.Run("label with zero counts", func(t *testing.T) {
		label := &gmailapi.Label{
			Id:             "Label_1",
			Name:           "EmptyLabel",
			Type:           "user",
			MessagesTotal:  0,
			MessagesUnread: 0,
			ThreadsTotal:   0,
			ThreadsUnread:  0,
		}
		got := convertLabel(label)
		if got.MessagesTotal != 0 {
			t.Errorf("expected 0 messages, got %d", got.MessagesTotal)
		}
	})

	t.Run("label with large counts", func(t *testing.T) {
		label := &gmailapi.Label{
			Id:             "Label_1",
			Name:           "BigLabel",
			Type:           "user",
			MessagesTotal:  999999,
			MessagesUnread: 500000,
			ThreadsTotal:   100000,
			ThreadsUnread:  50000,
		}
		got := convertLabel(label)
		if got.MessagesTotal != 999999 {
			t.Errorf("expected 999999 messages, got %d", got.MessagesTotal)
		}
	})
}

// TestConvertFilter_EdgeCases tests edge cases for filter conversion
func TestConvertFilter_EdgeCases(t *testing.T) {
	t.Run("filter with nil criteria and action", func(t *testing.T) {
		filter := &gmailapi.Filter{
			Id:       "filter1",
			Criteria: nil,
			Action:   nil,
		}
		got := convertFilter(filter)
		if got.ID != "filter1" {
			t.Errorf("expected filter1, got %s", got.ID)
		}
		if got.CriteriaFrom != "" || got.CriteriaTo != "" || got.CriteriaSubject != "" {
			t.Error("expected empty criteria fields")
		}
		if got.ActionAddLabelIDs != nil || got.ActionRemoveLabelIDs != nil {
			t.Error("expected nil action fields")
		}
	})

	t.Run("filter with empty criteria fields", func(t *testing.T) {
		filter := &gmailapi.Filter{
			Id: "filter1",
			Criteria: &gmailapi.FilterCriteria{
				From:          "",
				To:            "",
				Subject:       "",
				Query:         "",
				HasAttachment: false,
			},
			Action: &gmailapi.FilterAction{
				AddLabelIds:    []string{},
				RemoveLabelIds: []string{},
				Forward:        "",
			},
		}
		got := convertFilter(filter)
		if got.CriteriaFrom != "" {
			t.Errorf("expected empty from criteria, got %s", got.CriteriaFrom)
		}
		if len(got.ActionAddLabelIDs) != 0 {
			t.Errorf("expected empty add label IDs, got %d", len(got.ActionAddLabelIDs))
		}
	})

	t.Run("filter with multiple label actions", func(t *testing.T) {
		filter := &gmailapi.Filter{
			Id: "filter1",
			Criteria: &gmailapi.FilterCriteria{
				From: "boss@example.com",
			},
			Action: &gmailapi.FilterAction{
				AddLabelIds:    []string{"Label_1", "Label_2", "Label_3"},
				RemoveLabelIds: []string{"INBOX", "UNREAD"},
			},
		}
		got := convertFilter(filter)
		if len(got.ActionAddLabelIDs) != 3 {
			t.Errorf("expected 3 add labels, got %d", len(got.ActionAddLabelIDs))
		}
		if len(got.ActionRemoveLabelIDs) != 2 {
			t.Errorf("expected 2 remove labels, got %d", len(got.ActionRemoveLabelIDs))
		}
	})
}

// TestConvertVacationSettings_EdgeCases tests edge cases for vacation conversion
func TestConvertVacationSettings_EdgeCases(t *testing.T) {
	t.Run("disabled vacation", func(t *testing.T) {
		settings := &gmailapi.VacationSettings{
			EnableAutoReply:       false,
			ResponseSubject:       "",
			ResponseBodyPlainText: "",
			ResponseBodyHtml:      "",
			StartTime:             0,
			EndTime:               0,
		}
		got := convertVacationSettings(settings)
		if got.EnableAutoReply {
			t.Error("expected auto-reply disabled")
		}
		if got.StartTime != 0 || got.EndTime != 0 {
			t.Error("expected zero timestamps")
		}
	})

	t.Run("vacation with HTML only", func(t *testing.T) {
		settings := &gmailapi.VacationSettings{
			EnableAutoReply:       true,
			ResponseSubject:       "Out of Office",
			ResponseBodyPlainText: "",
			ResponseBodyHtml:      "<h1>I am away</h1>",
		}
		got := convertVacationSettings(settings)
		if got.ResponseBodyPlainText != "" {
			t.Errorf("expected empty plain text, got %s", got.ResponseBodyPlainText)
		}
		if got.ResponseBodyHtml != "<h1>I am away</h1>" {
			t.Errorf("expected HTML body, got %s", got.ResponseBodyHtml)
		}
	})

	t.Run("vacation with distant timestamps", func(t *testing.T) {
		settings := &gmailapi.VacationSettings{
			EnableAutoReply: true,
			StartTime:       1704067200000, // 2024-01-01
			EndTime:         1735689600000, // 2025-01-01
		}
		got := convertVacationSettings(settings)
		if got.StartTime != 1704067200000 {
			t.Errorf("expected start time 1704067200000, got %d", got.StartTime)
		}
		if got.EndTime != 1735689600000 {
			t.Errorf("expected end time 1735689600000, got %d", got.EndTime)
		}
	})
}

// TestFindAttachmentMeta_EdgeCases tests edge cases for attachment metadata
func TestFindAttachmentMeta_EdgeCases(t *testing.T) {
	t.Run("deeply nested part", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			MimeType: "multipart/mixed",
			Parts: []*gmailapi.MessagePart{
				{
					MimeType: "multipart/alternative",
					Parts: []*gmailapi.MessagePart{
						{
							MimeType: "text/plain",
							Body:     &gmailapi.MessagePartBody{},
						},
					},
				},
				{
					MimeType: "multipart/related",
					Parts: []*gmailapi.MessagePart{
						{
							Filename: "deep.pdf",
							MimeType: "application/pdf",
							Body: &gmailapi.MessagePartBody{
								AttachmentId: "deep-attach",
								Size:         1024,
							},
						},
					},
				},
			},
		}
		fn, mt, sz := findAttachmentMeta(part, "deep-attach")
		if fn != "deep.pdf" {
			t.Errorf("expected deep.pdf, got %s", fn)
		}
		if mt != "application/pdf" {
			t.Errorf("expected application/pdf, got %s", mt)
		}
		if sz != 1024 {
			t.Errorf("expected size 1024, got %d", sz)
		}
	})

	t.Run("multiple attachments search for first", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			Parts: []*gmailapi.MessagePart{
				{
					Filename: "first.pdf",
					MimeType: "application/pdf",
					Body: &gmailapi.MessagePartBody{
						AttachmentId: "attach1",
						Size:         100,
					},
				},
				{
					Filename: "second.jpg",
					MimeType: "image/jpeg",
					Body: &gmailapi.MessagePartBody{
						AttachmentId: "attach2",
						Size:         200,
					},
				},
			},
		}
		fn, mimeType, sz := findAttachmentMeta(part, "attach1")
		if fn != "first.pdf" {
			t.Errorf("expected first.pdf, got %s", fn)
		}
		if mimeType != "application/pdf" {
			t.Errorf("expected application/pdf, got %s", mimeType)
		}
		if sz != 100 {
			t.Errorf("expected size 100, got %d", sz)
		}
	})

	t.Run("part with body but no attachment ID", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			Body: &gmailapi.MessagePartBody{
				Data:         base64.URLEncoding.EncodeToString([]byte("body content")),
				AttachmentId: "", // Empty attachment ID
				Size:         100,
			},
		}
		fn, mimeType, sz := findAttachmentMeta(part, "target-id")
		if fn != "" || mimeType != "" || sz != 0 {
			t.Error("expected empty result when attachment ID doesn't match")
		}
	})

	t.Run("attachment with zero size", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			Filename: "empty.txt",
			MimeType: "text/plain",
			Body: &gmailapi.MessagePartBody{
				AttachmentId: "empty-attach",
				Size:         0,
			},
		}
		fn, mimeType, sz := findAttachmentMeta(part, "empty-attach")
		if fn != "empty.txt" {
			t.Errorf("expected empty.txt, got %s", fn)
		}
		if mimeType != "text/plain" {
			t.Errorf("expected text/plain, got %s", mimeType)
		}
		if sz != 0 {
			t.Errorf("expected size 0, got %d", sz)
		}
	})
}

// Benchmarks
func BenchmarkComposeMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = composeMessage("from@example.com", "to@example.com", "", "", "Subject", "Body", "", "")
	}
}

func BenchmarkComposeMessageHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = composeMessage("from@example.com", "to@example.com", "", "", "Subject", "Plain", "<p>HTML</p>", "")
	}
}

func BenchmarkParseHeaders(b *testing.B) {
	headers := []*gmailapi.MessagePartHeader{
		{Name: "From", Value: "sender@example.com"},
		{Name: "To", Value: "recipient@example.com"},
		{Name: "Subject", Value: "Test Subject"},
		{Name: "Date", Value: "Mon, 1 Jan 2024 12:00:00 +0000"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = parseHeaders(headers)
	}
}

func BenchmarkConvertMessage(b *testing.B) {
	msg := &gmailapi.Message{
		Id:           "msg1",
		ThreadId:     "thread1",
		Snippet:      "Hello...",
		LabelIds:     []string{"INBOX", "UNREAD"},
		SizeEstimate: 1024,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertMessage(msg)
	}
}
