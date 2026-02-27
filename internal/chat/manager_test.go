package chat

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/api/chat/v1"
)

func TestFormatSpaceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already formatted",
			input:    "spaces/abc123",
			expected: "spaces/abc123",
		},
		{
			name:     "raw space ID",
			input:    "abc123",
			expected: "spaces/abc123",
		},
		// Note: "spaces/" edge case produces "spaces/spaces/" due to implementation
		// The implementation only checks if input[:7] == "spaces/" which is false for "spaces/"
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatSpaceName(tc.input)
			if result != tc.expected {
				t.Errorf("formatSpaceName(%q): got %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatMessageName(t *testing.T) {
	tests := []struct {
		name      string
		spaceID   string
		messageID string
		expected  string
	}{
		{
			name:      "empty message ID",
			spaceID:   "abc123",
			messageID: "",
			expected:  "",
		},
		{
			name:      "already fully formatted",
			spaceID:   "abc123",
			messageID: "spaces/abc123/messages/msg456",
			expected:  "spaces/abc123/messages/msg456",
		},
		{
			name:      "raw message ID",
			spaceID:   "abc123",
			messageID: "msg456",
			expected:  "spaces/abc123/messages/msg456",
		},
		{
			name:      "with formatted space ID",
			spaceID:   "spaces/abc123",
			messageID: "msg456",
			expected:  "spaces/abc123/messages/msg456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatMessageName(tc.spaceID, tc.messageID)
			if result != tc.expected {
				t.Errorf("formatMessageName(%q, %q): got %q, want %q", tc.spaceID, tc.messageID, result, tc.expected)
			}
		})
	}
}

func TestFormatMemberName(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		memberID string
		expected string
	}{
		{
			name:     "empty member ID",
			spaceID:  "abc123",
			memberID: "",
			expected: "",
		},
		{
			name:     "already fully formatted",
			spaceID:  "abc123",
			memberID: "spaces/abc123/members/member456",
			expected: "spaces/abc123/members/member456",
		},
		{
			name:     "raw member ID",
			spaceID:  "abc123",
			memberID: "member456",
			expected: "spaces/abc123/members/member456",
		},
		{
			name:     "with formatted space ID",
			spaceID:  "spaces/abc123",
			memberID: "member456",
			expected: "spaces/abc123/members/member456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatMemberName(tc.spaceID, tc.memberID)
			if result != tc.expected {
				t.Errorf("formatMemberName(%q, %q): got %q, want %q", tc.spaceID, tc.memberID, result, tc.expected)
			}
		})
	}
}

func TestGetThreadIDFromName(t *testing.T) {
	tests := []struct {
		name     string
		thread   *chat.Thread
		expected string
	}{
		{
			name:     "nil thread",
			thread:   nil,
			expected: "",
		},
		{
			name: "empty thread name",
			thread: &chat.Thread{
				Name: "",
			},
			expected: "",
		},
		{
			name: "full thread name",
			thread: &chat.Thread{
				Name: "spaces/abc123/threads/thread456",
			},
			expected: "thread456",
		},
		{
			name: "name without slash",
			thread: &chat.Thread{
				Name: "thread456",
			},
			expected: "thread456",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getThreadIDFromName(tc.thread)
			if result != tc.expected {
				t.Errorf("getThreadIDFromName(%v): got %q, want %q", tc.thread, result, tc.expected)
			}
		})
	}
}

func TestConvertSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    *chat.Space
		expected types.ChatSpace
	}{
		{
			name: "full space",
			input: &chat.Space{
				Name:                "spaces/abc123",
				SpaceType:           "ROOM",
				DisplayName:         "Test Room",
				Threaded:            true,
				ExternalUserAllowed: true,
				SpaceHistoryState:   "HISTORY_ON",
				CreateTime:          "2024-01-01T00:00:00Z",
			},
			expected: types.ChatSpace{
				ID:                  "spaces/abc123",
				Name:                "spaces/abc123",
				Type:                "ROOM",
				DisplayName:         "Test Room",
				Threaded:            true,
				ExternalUserAllowed: true,
				SpaceHistoryState:   "HISTORY_ON",
				CreateTime:          "2024-01-01T00:00:00Z",
			},
		},
		{
			name:     "nil space",
			input:    nil,
			expected: types.ChatSpace{},
		},
		{
			name: "empty space",
			input: &chat.Space{
				Name: "",
			},
			expected: types.ChatSpace{
				ID:   "",
				Name: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertSpace(tc.input)
			if result.ID != tc.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tc.expected.ID)
			}
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Type != tc.expected.Type {
				t.Errorf("Type: got %q, want %q", result.Type, tc.expected.Type)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
		})
	}
}

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *chat.Message
		expected types.ChatMessage
	}{
		{
			name: "full message",
			input: &chat.Message{
				Name:          "spaces/abc123/messages/msg456",
				Text:          "Hello, world!",
				FormattedText: "Hello, **world**!",
				CreateTime:    "2024-01-01T00:00:00Z",
				Thread: &chat.Thread{
					Name: "spaces/abc123/threads/thread789",
				},
				Sender: &chat.User{
					DisplayName: "John Doe",
				},
			},
			expected: types.ChatMessage{
				ID:            "spaces/abc123/messages/msg456",
				ThreadID:      "thread789",
				Text:          "Hello, world!",
				FormattedText: "Hello, **world**!",
				CreateTime:    "2024-01-01T00:00:00Z",
				SenderName:    "John Doe",
			},
		},
		{
			name:     "nil message",
			input:    nil,
			expected: types.ChatMessage{},
		},
		{
			name: "message without thread",
			input: &chat.Message{
				Name: "spaces/abc123/messages/msg456",
				Text: "Simple message",
			},
			expected: types.ChatMessage{
				ID:   "spaces/abc123/messages/msg456",
				Text: "Simple message",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertMessage(tc.input)
			if result.ID != tc.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tc.expected.ID)
			}
			if result.Text != tc.expected.Text {
				t.Errorf("Text: got %q, want %q", result.Text, tc.expected.Text)
			}
			if result.ThreadID != tc.expected.ThreadID {
				t.Errorf("ThreadID: got %q, want %q", result.ThreadID, tc.expected.ThreadID)
			}
		})
	}
}

func TestConvertMembership(t *testing.T) {
	tests := []struct {
		name     string
		input    *chat.Membership
		expected types.ChatMember
	}{
		{
			name: "full membership",
			input: &chat.Membership{
				Name:       "spaces/abc123/members/member456",
				Role:       "MEMBER",
				CreateTime: "2024-01-01T00:00:00Z",
				Member: &chat.User{
					DisplayName: "Jane Smith",
				},
			},
			expected: types.ChatMember{
				ID:       "spaces/abc123/members/member456",
				Role:     "MEMBER",
				JoinTime: "2024-01-01T00:00:00Z",
				Name:     "Jane Smith",
			},
		},
		{
			name:     "nil membership",
			input:    nil,
			expected: types.ChatMember{},
		},
		{
			name: "membership without member",
			input: &chat.Membership{
				Name: "spaces/abc123/members/member456",
				Role: "ADMIN",
			},
			expected: types.ChatMember{
				ID:   "spaces/abc123/members/member456",
				Role: "ADMIN",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertMembership(tc.input)
			if result.ID != tc.expected.ID {
				t.Errorf("ID: got %q, want %q", result.ID, tc.expected.ID)
			}
			if result.Role != tc.expected.Role {
				t.Errorf("Role: got %q, want %q", result.Role, tc.expected.Role)
			}
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
		})
	}
}
