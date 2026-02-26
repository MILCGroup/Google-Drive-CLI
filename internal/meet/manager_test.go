package meet

import (
	"testing"
	"time"

	"cloud.google.com/go/apps/meet/apiv2/meetpb"
	"github.com/dl-alexandre/gdrv/internal/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConvertSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    *meetpb.Space
		expected types.MeetSpace
	}{
		{
			name: "full space",
			input: &meetpb.Space{
				Name:        "spaces/abc123",
				MeetingUri:  "https://meet.google.com/abc-defg-hij",
				MeetingCode: "abc-defg-hij",
				Config: &meetpb.SpaceConfig{
					AccessType: meetpb.SpaceConfig_OPEN,
				},
				ActiveConference: &meetpb.ActiveConference{
					ConferenceRecord: "conferenceRecords/xyz789",
				},
			},
			expected: types.MeetSpace{
				Name:             "spaces/abc123",
				MeetingUri:       "https://meet.google.com/abc-defg-hij",
				MeetingCode:      "abc-defg-hij",
				Config:           &types.MeetSpaceConfig{AccessType: "OPEN"},
				ActiveConference: "conferenceRecords/xyz789",
			},
		},
		{
			name: "space without config",
			input: &meetpb.Space{
				Name:        "spaces/def456",
				MeetingUri:  "https://meet.google.com/def-ghij-klm",
				MeetingCode: "def-ghij-klm",
			},
			expected: types.MeetSpace{
				Name:        "spaces/def456",
				MeetingUri:  "https://meet.google.com/def-ghij-klm",
				MeetingCode: "def-ghij-klm",
			},
		},

		{
			name: "empty space",
			input: &meetpb.Space{
				Name: "",
			},
			expected: types.MeetSpace{
				Name: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertSpace(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.MeetingUri != tc.expected.MeetingUri {
				t.Errorf("MeetingUri: got %q, want %q", result.MeetingUri, tc.expected.MeetingUri)
			}
			if result.MeetingCode != tc.expected.MeetingCode {
				t.Errorf("MeetingCode: got %q, want %q", result.MeetingCode, tc.expected.MeetingCode)
			}
			if tc.expected.Config != nil {
				if result.Config == nil {
					t.Fatal("Config: got nil, want non-nil")
				}
				if result.Config.AccessType != tc.expected.Config.AccessType {
					t.Errorf("Config.AccessType: got %q, want %q", result.Config.AccessType, tc.expected.Config.AccessType)
				}
			}
			if result.ActiveConference != tc.expected.ActiveConference {
				t.Errorf("ActiveConference: got %q, want %q", result.ActiveConference, tc.expected.ActiveConference)
			}
		})
	}
}

func TestConvertConferenceRecord(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	expireTime := time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *meetpb.ConferenceRecord
		expected types.MeetConferenceRecord
	}{
		{
			name: "full record",
			input: &meetpb.ConferenceRecord{
				Name:       "conferenceRecords/xyz789",
				Space:      "spaces/abc123",
				StartTime:  timestamppb.New(startTime),
				EndTime:    timestamppb.New(endTime),
				ExpireTime: timestamppb.New(expireTime),
			},
			expected: types.MeetConferenceRecord{
				Name:       "conferenceRecords/xyz789",
				Space:      "spaces/abc123",
				StartTime:  "2024-01-15 10:00:00",
				EndTime:    "2024-01-15 11:00:00",
				ExpireTime: "2024-04-15 10:00:00",
			},
		},
		{
			name: "record without times",
			input: &meetpb.ConferenceRecord{
				Name:  "conferenceRecords/abc456",
				Space: "spaces/def789",
			},
			expected: types.MeetConferenceRecord{
				Name:  "conferenceRecords/abc456",
				Space: "spaces/def789",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertConferenceRecord(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Space != tc.expected.Space {
				t.Errorf("Space: got %q, want %q", result.Space, tc.expected.Space)
			}
			if result.StartTime != tc.expected.StartTime {
				t.Errorf("StartTime: got %q, want %q", result.StartTime, tc.expected.StartTime)
			}
			if result.EndTime != tc.expected.EndTime {
				t.Errorf("EndTime: got %q, want %q", result.EndTime, tc.expected.EndTime)
			}
			if result.ExpireTime != tc.expected.ExpireTime {
				t.Errorf("ExpireTime: got %q, want %q", result.ExpireTime, tc.expected.ExpireTime)
			}
		})
	}
}

func TestConvertParticipant(t *testing.T) {
	startTime := time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 10, 55, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *meetpb.Participant
		expected types.MeetParticipant
	}{
		{
			name: "signed-in user",
			input: &meetpb.Participant{
				Name: "conferenceRecords/xyz789/participants/participant123",
				User: &meetpb.Participant_SignedinUser{
					SignedinUser: &meetpb.SignedinUser{
						DisplayName: "John Doe",
						User:        "users/john@example.com",
					},
				},
				EarliestStartTime: timestamppb.New(startTime),
				LatestEndTime:     timestamppb.New(endTime),
			},
			expected: types.MeetParticipant{
				Name:              "conferenceRecords/xyz789/participants/participant123",
				DisplayName:       "John Doe",
				Email:             "users/john@example.com",
				EarliestStartTime: "2024-01-15 10:05:00",
				LatestEndTime:     "2024-01-15 10:55:00",
			},
		},
		{
			name: "anonymous user",
			input: &meetpb.Participant{
				Name: "conferenceRecords/xyz789/participants/participant456",
				User: &meetpb.Participant_AnonymousUser{
					AnonymousUser: &meetpb.AnonymousUser{
						DisplayName: "Anonymous",
					},
				},
			},
			expected: types.MeetParticipant{
				Name:        "conferenceRecords/xyz789/participants/participant456",
				DisplayName: "Anonymous",
			},
		},
		{
			name: "phone user",
			input: &meetpb.Participant{
				Name: "conferenceRecords/xyz789/participants/participant789",
				User: &meetpb.Participant_PhoneUser{
					PhoneUser: &meetpb.PhoneUser{
						DisplayName: "+1-555-1234",
					},
				},
			},
			expected: types.MeetParticipant{
				Name:        "conferenceRecords/xyz789/participants/participant789",
				DisplayName: "+1-555-1234",
				PhoneNumber: "Phone",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertParticipant(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
			if result.Email != tc.expected.Email {
				t.Errorf("Email: got %q, want %q", result.Email, tc.expected.Email)
			}
			if result.PhoneNumber != tc.expected.PhoneNumber {
				t.Errorf("PhoneNumber: got %q, want %q", result.PhoneNumber, tc.expected.PhoneNumber)
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
