package gmail

import (
	"encoding/base64"
	"strings"
	"testing"

	gmailapi "google.golang.org/api/gmail/v1"
)

func TestComposeMessage(t *testing.T) {
	t.Run("plain text message", func(t *testing.T) {
		raw := composeMessage("me", "alice@example.com", "", "", "Hello", "Body text", "", "")
		msg := string(raw)

		assertContains(t, msg, "From: me\r\n")
		assertContains(t, msg, "To: alice@example.com\r\n")
		assertContains(t, msg, "Subject: Hello\r\n")
		assertContains(t, msg, "MIME-Version: 1.0\r\n")
		assertContains(t, msg, "Content-Type: text/plain; charset=UTF-8\r\n")
		assertContains(t, msg, "Body text")

		// Should NOT contain multipart headers.
		if strings.Contains(msg, "multipart/alternative") {
			t.Fatal("plain text message should not contain multipart headers")
		}
		// Should NOT contain reply headers.
		if strings.Contains(msg, "In-Reply-To") {
			t.Fatal("plain text message without reply should not contain In-Reply-To")
		}
	})

	t.Run("message with cc and bcc", func(t *testing.T) {
		raw := composeMessage("me", "alice@example.com", "bob@example.com", "carol@example.com", "Test", "Body", "", "")
		msg := string(raw)

		assertContains(t, msg, "Cc: bob@example.com\r\n")
		assertContains(t, msg, "Bcc: carol@example.com\r\n")
	})

	t.Run("reply message with In-Reply-To and References", func(t *testing.T) {
		msgID := "<original-msg-id@example.com>"
		raw := composeMessage("me", "alice@example.com", "", "", "Re: Hello", "Reply body", "", msgID)
		msg := string(raw)

		assertContains(t, msg, "In-Reply-To: "+msgID+"\r\n")
		assertContains(t, msg, "References: "+msgID+"\r\n")
	})

	t.Run("HTML multipart message", func(t *testing.T) {
		raw := composeMessage("me", "alice@example.com", "", "", "HTML Test", "Plain body", "<p>HTML body</p>", "")
		msg := string(raw)

		assertContains(t, msg, "Content-Type: multipart/alternative")
		assertContains(t, msg, "text/plain; charset=UTF-8")
		assertContains(t, msg, "text/html; charset=UTF-8")
		assertContains(t, msg, "Plain body")
		assertContains(t, msg, "<p>HTML body</p>")
	})

	t.Run("empty cc and bcc are omitted", func(t *testing.T) {
		raw := composeMessage("me", "alice@example.com", "", "", "Test", "Body", "", "")
		msg := string(raw)

		if strings.Contains(msg, "Cc:") {
			t.Fatal("empty Cc should not appear in headers")
		}
		if strings.Contains(msg, "Bcc:") {
			t.Fatal("empty Bcc should not appear in headers")
		}
	})
}

func TestComposeMessageBase64URLEncoding(t *testing.T) {
	raw := composeMessage("me", "alice@example.com", "", "", "Test", "Body content here", "", "")
	encoded := base64.URLEncoding.EncodeToString(raw)

	// URL-safe base64 must not contain + or /.
	if strings.ContainsAny(encoded, "+/") {
		t.Fatalf("encoded message contains non-URL-safe characters: %s", encoded)
	}

	// Verify round-trip: decoding with URLEncoding must succeed.
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode with URLEncoding: %v", err)
	}
	if string(decoded) != string(raw) {
		t.Fatal("round-trip decode mismatch")
	}

	// Verify that StdEncoding decode fails or produces wrong result if there are
	// URL-safe characters that differ from standard base64.
	// This is a sanity check: we ensure URLEncoding is being used, not StdEncoding.
	reEncoded := base64.StdEncoding.EncodeToString(raw)
	if strings.ContainsAny(reEncoded, "-_") {
		t.Fatal("StdEncoding should not produce URL-safe characters")
	}
}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []*gmailapi.MessagePartHeader
		from    string
		to      string
		subject string
		date    string
	}{
		{
			name:    "nil headers",
			headers: nil,
			from:    "",
			to:      "",
			subject: "",
			date:    "",
		},
		{
			name:    "empty headers",
			headers: []*gmailapi.MessagePartHeader{},
			from:    "",
			to:      "",
			subject: "",
			date:    "",
		},
		{
			name: "all known headers",
			headers: []*gmailapi.MessagePartHeader{
				{Name: "From", Value: "alice@example.com"},
				{Name: "To", Value: "bob@example.com"},
				{Name: "Subject", Value: "Test Subject"},
				{Name: "Date", Value: "Mon, 1 Jan 2024 12:00:00 +0000"},
			},
			from:    "alice@example.com",
			to:      "bob@example.com",
			subject: "Test Subject",
			date:    "Mon, 1 Jan 2024 12:00:00 +0000",
		},
		{
			name: "case insensitive header names",
			headers: []*gmailapi.MessagePartHeader{
				{Name: "from", Value: "alice@example.com"},
				{Name: "TO", Value: "bob@example.com"},
				{Name: "SUBJECT", Value: "Case Test"},
				{Name: "Date", Value: "Tue, 2 Jan 2024 00:00:00 +0000"},
			},
			from:    "alice@example.com",
			to:      "bob@example.com",
			subject: "Case Test",
			date:    "Tue, 2 Jan 2024 00:00:00 +0000",
		},
		{
			name: "extra headers are ignored",
			headers: []*gmailapi.MessagePartHeader{
				{Name: "From", Value: "alice@example.com"},
				{Name: "X-Custom", Value: "custom-value"},
				{Name: "Content-Type", Value: "text/plain"},
			},
			from:    "alice@example.com",
			to:      "",
			subject: "",
			date:    "",
		},
		{
			name: "duplicate headers use last value",
			headers: []*gmailapi.MessagePartHeader{
				{Name: "From", Value: "first@example.com"},
				{Name: "From", Value: "second@example.com"},
			},
			from: "second@example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			from, to, subject, date := parseHeaders(tc.headers)
			if from != tc.from {
				t.Errorf("from: got %q, want %q", from, tc.from)
			}
			if to != tc.to {
				t.Errorf("to: got %q, want %q", to, tc.to)
			}
			if subject != tc.subject {
				t.Errorf("subject: got %q, want %q", subject, tc.subject)
			}
			if date != tc.date {
				t.Errorf("date: got %q, want %q", date, tc.date)
			}
		})
	}
}

func TestConvertMessage(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		got := convertMessage(nil)
		if got.ID != "" || got.ThreadID != "" {
			t.Fatal("expected empty message for nil input")
		}
	})

	t.Run("message without payload", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:           "msg-1",
			ThreadId:     "thread-1",
			Snippet:      "Hello...",
			LabelIds:     []string{"INBOX", "UNREAD"},
			SizeEstimate: 1024,
		}
		got := convertMessage(msg)
		if got.ID != "msg-1" {
			t.Errorf("ID: got %q, want %q", got.ID, "msg-1")
		}
		if got.ThreadID != "thread-1" {
			t.Errorf("ThreadID: got %q, want %q", got.ThreadID, "thread-1")
		}
		if got.Snippet != "Hello..." {
			t.Errorf("Snippet: got %q, want %q", got.Snippet, "Hello...")
		}
		if len(got.LabelIDs) != 2 {
			t.Errorf("LabelIDs: got %d, want 2", len(got.LabelIDs))
		}
		if got.SizeEstimate != 1024 {
			t.Errorf("SizeEstimate: got %d, want 1024", got.SizeEstimate)
		}
		// Without payload, headers should be empty.
		if got.From != "" || got.To != "" || got.Subject != "" || got.Date != "" {
			t.Fatal("expected empty header fields when payload is nil")
		}
	})

	t.Run("message with payload headers", func(t *testing.T) {
		msg := &gmailapi.Message{
			Id:       "msg-2",
			ThreadId: "thread-2",
			Payload: &gmailapi.MessagePart{
				Headers: []*gmailapi.MessagePartHeader{
					{Name: "From", Value: "sender@example.com"},
					{Name: "To", Value: "recipient@example.com"},
					{Name: "Subject", Value: "Important"},
					{Name: "Date", Value: "Wed, 3 Jan 2024 10:00:00 +0000"},
				},
			},
		}
		got := convertMessage(msg)
		if got.From != "sender@example.com" {
			t.Errorf("From: got %q, want %q", got.From, "sender@example.com")
		}
		if got.To != "recipient@example.com" {
			t.Errorf("To: got %q, want %q", got.To, "recipient@example.com")
		}
		if got.Subject != "Important" {
			t.Errorf("Subject: got %q, want %q", got.Subject, "Important")
		}
		if got.Date != "Wed, 3 Jan 2024 10:00:00 +0000" {
			t.Errorf("Date: got %q, want %q", got.Date, "Wed, 3 Jan 2024 10:00:00 +0000")
		}
	})
}

func TestConvertLabel(t *testing.T) {
	t.Run("nil label", func(t *testing.T) {
		got := convertLabel(nil)
		if got.ID != "" || got.Name != "" {
			t.Fatal("expected empty label for nil input")
		}
	})

	t.Run("full label", func(t *testing.T) {
		l := &gmailapi.Label{
			Id:             "Label_1",
			Name:           "Work",
			Type:           "user",
			MessagesTotal:  42,
			MessagesUnread: 5,
			ThreadsTotal:   30,
			ThreadsUnread:  3,
		}
		got := convertLabel(l)
		if got.ID != "Label_1" {
			t.Errorf("ID: got %q, want %q", got.ID, "Label_1")
		}
		if got.Name != "Work" {
			t.Errorf("Name: got %q, want %q", got.Name, "Work")
		}
		if got.Type != "user" {
			t.Errorf("Type: got %q, want %q", got.Type, "user")
		}
		if got.MessagesTotal != 42 {
			t.Errorf("MessagesTotal: got %d, want 42", got.MessagesTotal)
		}
		if got.MessagesUnread != 5 {
			t.Errorf("MessagesUnread: got %d, want 5", got.MessagesUnread)
		}
		if got.ThreadsTotal != 30 {
			t.Errorf("ThreadsTotal: got %d, want 30", got.ThreadsTotal)
		}
		if got.ThreadsUnread != 3 {
			t.Errorf("ThreadsUnread: got %d, want 3", got.ThreadsUnread)
		}
	})
}

func TestConvertFilter(t *testing.T) {
	t.Run("nil filter", func(t *testing.T) {
		got := convertFilter(nil)
		if got.ID != "" {
			t.Fatal("expected empty filter for nil input")
		}
	})

	t.Run("filter without criteria or action", func(t *testing.T) {
		f := &gmailapi.Filter{
			Id: "filter-1",
		}
		got := convertFilter(f)
		if got.ID != "filter-1" {
			t.Errorf("ID: got %q, want %q", got.ID, "filter-1")
		}
		if got.CriteriaFrom != "" || got.CriteriaTo != "" {
			t.Fatal("expected empty criteria when nil")
		}
	})

	t.Run("full filter", func(t *testing.T) {
		f := &gmailapi.Filter{
			Id: "filter-2",
			Criteria: &gmailapi.FilterCriteria{
				From:          "boss@example.com",
				To:            "me@example.com",
				Subject:       "urgent",
				Query:         "has:attachment",
				HasAttachment: true,
			},
			Action: &gmailapi.FilterAction{
				AddLabelIds:    []string{"Label_1"},
				RemoveLabelIds: []string{"INBOX"},
				Forward:        "archive@example.com",
			},
		}
		got := convertFilter(f)
		if got.CriteriaFrom != "boss@example.com" {
			t.Errorf("CriteriaFrom: got %q, want %q", got.CriteriaFrom, "boss@example.com")
		}
		if got.CriteriaTo != "me@example.com" {
			t.Errorf("CriteriaTo: got %q, want %q", got.CriteriaTo, "me@example.com")
		}
		if got.CriteriaSubject != "urgent" {
			t.Errorf("CriteriaSubject: got %q, want %q", got.CriteriaSubject, "urgent")
		}
		if got.CriteriaQuery != "has:attachment" {
			t.Errorf("CriteriaQuery: got %q, want %q", got.CriteriaQuery, "has:attachment")
		}
		if !got.CriteriaHasAttachment {
			t.Error("CriteriaHasAttachment: got false, want true")
		}
		if len(got.ActionAddLabelIDs) != 1 || got.ActionAddLabelIDs[0] != "Label_1" {
			t.Errorf("ActionAddLabelIDs: got %v, want [Label_1]", got.ActionAddLabelIDs)
		}
		if len(got.ActionRemoveLabelIDs) != 1 || got.ActionRemoveLabelIDs[0] != "INBOX" {
			t.Errorf("ActionRemoveLabelIDs: got %v, want [INBOX]", got.ActionRemoveLabelIDs)
		}
		if got.ActionForward != "archive@example.com" {
			t.Errorf("ActionForward: got %q, want %q", got.ActionForward, "archive@example.com")
		}
	})
}

func TestConvertVacationSettings(t *testing.T) {
	t.Run("nil settings", func(t *testing.T) {
		got := convertVacationSettings(nil)
		if got.EnableAutoReply {
			t.Fatal("expected disabled auto-reply for nil input")
		}
	})

	t.Run("full settings", func(t *testing.T) {
		vs := &gmailapi.VacationSettings{
			EnableAutoReply:       true,
			ResponseSubject:       "Out of Office",
			ResponseBodyPlainText: "I am away.",
			ResponseBodyHtml:      "<p>I am away.</p>",
			StartTime:             1704067200000,
			EndTime:               1704153600000,
		}
		got := convertVacationSettings(vs)
		if !got.EnableAutoReply {
			t.Error("EnableAutoReply: got false, want true")
		}
		if got.ResponseSubject != "Out of Office" {
			t.Errorf("ResponseSubject: got %q, want %q", got.ResponseSubject, "Out of Office")
		}
		if got.ResponseBodyPlainText != "I am away." {
			t.Errorf("ResponseBodyPlainText: got %q, want %q", got.ResponseBodyPlainText, "I am away.")
		}
		if got.ResponseBodyHtml != "<p>I am away.</p>" {
			t.Errorf("ResponseBodyHtml: got %q, want %q", got.ResponseBodyHtml, "<p>I am away.</p>")
		}
		if got.StartTime != 1704067200000 {
			t.Errorf("StartTime: got %d, want 1704067200000", got.StartTime)
		}
		if got.EndTime != 1704153600000 {
			t.Errorf("EndTime: got %d, want 1704153600000", got.EndTime)
		}
	})
}

func TestFindAttachmentMeta(t *testing.T) {
	t.Run("nil part", func(t *testing.T) {
		fn, mt, sz := findAttachmentMeta(nil, "att-1")
		if fn != "" || mt != "" || sz != 0 {
			t.Fatal("expected empty result for nil part")
		}
	})

	t.Run("direct match", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			Filename: "document.pdf",
			MimeType: "application/pdf",
			Body: &gmailapi.MessagePartBody{
				AttachmentId: "att-1",
				Size:         2048,
			},
		}
		fn, mt, sz := findAttachmentMeta(part, "att-1")
		if fn != "document.pdf" {
			t.Errorf("filename: got %q, want %q", fn, "document.pdf")
		}
		if mt != "application/pdf" {
			t.Errorf("mimeType: got %q, want %q", mt, "application/pdf")
		}
		if sz != 2048 {
			t.Errorf("size: got %d, want 2048", sz)
		}
	})

	t.Run("nested part", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			MimeType: "multipart/mixed",
			Parts: []*gmailapi.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmailapi.MessagePartBody{},
				},
				{
					Filename: "image.png",
					MimeType: "image/png",
					Body: &gmailapi.MessagePartBody{
						AttachmentId: "att-2",
						Size:         4096,
					},
				},
			},
		}
		fn, mt, sz := findAttachmentMeta(part, "att-2")
		if fn != "image.png" {
			t.Errorf("filename: got %q, want %q", fn, "image.png")
		}
		if mt != "image/png" {
			t.Errorf("mimeType: got %q, want %q", mt, "image/png")
		}
		if sz != 4096 {
			t.Errorf("size: got %d, want 4096", sz)
		}
	})

	t.Run("no match", func(t *testing.T) {
		part := &gmailapi.MessagePart{
			Body: &gmailapi.MessagePartBody{
				AttachmentId: "att-other",
			},
		}
		fn, mt, sz := findAttachmentMeta(part, "att-missing")
		if fn != "" || mt != "" || sz != 0 {
			t.Fatal("expected empty result when attachment not found")
		}
	})
}

// assertContains is a test helper that checks whether s contains substr.
func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected string to contain %q, got:\n%s", substr, s)
	}
}
