package groups

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/types"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

func TestConvertGroup(t *testing.T) {
	tests := []struct {
		name  string
		input *cloudidentity.Group
		check func(t *testing.T, got types.CloudIdentityGroup)
	}{
		{
			name:  "nil input returns zero value",
			input: nil,
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.Name != "" || got.DisplayName != "" || got.GroupKeyID != "" {
					t.Fatal("expected zero-value group for nil input")
				}
			},
		},
		{
			name: "basic fields mapped correctly",
			input: &cloudidentity.Group{
				Name:        "groups/abc123",
				DisplayName: "Engineering",
				Description: "Engineering team",
				CreateTime:  "2025-01-15T10:30:00Z",
				UpdateTime:  "2025-06-20T14:00:00Z",
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.Name != "groups/abc123" {
					t.Fatalf("expected Name groups/abc123, got %s", got.Name)
				}
				if got.DisplayName != "Engineering" {
					t.Fatalf("expected DisplayName Engineering, got %s", got.DisplayName)
				}
				if got.Description != "Engineering team" {
					t.Fatalf("expected Description Engineering team, got %s", got.Description)
				}
				if got.CreateTime != "2025-01-15T10:30:00Z" {
					t.Fatalf("expected CreateTime 2025-01-15T10:30:00Z, got %s", got.CreateTime)
				}
				if got.UpdateTime != "2025-06-20T14:00:00Z" {
					t.Fatalf("expected UpdateTime 2025-06-20T14:00:00Z, got %s", got.UpdateTime)
				}
			},
		},
		{
			name: "group key with id and namespace",
			input: &cloudidentity.Group{
				Name: "groups/def456",
				GroupKey: &cloudidentity.EntityKey{
					Id:        "eng@example.com",
					Namespace: "identitysources/example",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.GroupKeyID != "eng@example.com" {
					t.Fatalf("expected GroupKeyID eng@example.com, got %s", got.GroupKeyID)
				}
				if got.GroupKeyNamespace != "identitysources/example" {
					t.Fatalf("expected GroupKeyNamespace identitysources/example, got %s", got.GroupKeyNamespace)
				}
			},
		},
		{
			name: "group key with id only",
			input: &cloudidentity.Group{
				Name: "groups/ghi789",
				GroupKey: &cloudidentity.EntityKey{
					Id: "team@example.com",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.GroupKeyID != "team@example.com" {
					t.Fatalf("expected GroupKeyID team@example.com, got %s", got.GroupKeyID)
				}
				if got.GroupKeyNamespace != "" {
					t.Fatalf("expected empty GroupKeyNamespace, got %s", got.GroupKeyNamespace)
				}
			},
		},
		{
			name: "nil group key leaves fields empty",
			input: &cloudidentity.Group{
				Name:     "groups/no-key",
				GroupKey: nil,
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.GroupKeyID != "" || got.GroupKeyNamespace != "" {
					t.Fatal("expected empty group key fields when GroupKey is nil")
				}
			},
		},
		{
			name: "labels mapped correctly",
			input: &cloudidentity.Group{
				Name: "groups/labeled",
				Labels: map[string]string{
					"cloudidentity.googleapis.com/groups.discussion_forum": "",
					"custom-label": "custom-value",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.Labels == nil {
					t.Fatal("expected non-nil labels map")
				}
				if len(got.Labels) != 2 {
					t.Fatalf("expected 2 labels, got %d", len(got.Labels))
				}
				if _, ok := got.Labels["cloudidentity.googleapis.com/groups.discussion_forum"]; !ok {
					t.Fatal("expected discussion_forum label key")
				}
				if got.Labels["custom-label"] != "custom-value" {
					t.Fatalf("expected custom-label value custom-value, got %s", got.Labels["custom-label"])
				}
			},
		},
		{
			name: "nil labels stays nil",
			input: &cloudidentity.Group{
				Name:   "groups/no-labels",
				Labels: nil,
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.Labels != nil {
					t.Fatalf("expected nil labels, got %v", got.Labels)
				}
			},
		},
		{
			name: "all fields together",
			input: &cloudidentity.Group{
				Name:        "groups/full",
				DisplayName: "Full Group",
				Description: "A fully populated group",
				CreateTime:  "2024-03-01T00:00:00Z",
				UpdateTime:  "2025-12-01T00:00:00Z",
				GroupKey: &cloudidentity.EntityKey{
					Id:        "full@example.com",
					Namespace: "identitysources/full",
				},
				Labels: map[string]string{
					"cloudidentity.googleapis.com/groups.discussion_forum": "",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityGroup) {
				if got.Name != "groups/full" {
					t.Fatalf("expected Name groups/full, got %s", got.Name)
				}
				if got.DisplayName != "Full Group" {
					t.Fatalf("expected DisplayName Full Group, got %s", got.DisplayName)
				}
				if got.Description != "A fully populated group" {
					t.Fatalf("expected Description, got %s", got.Description)
				}
				if got.GroupKeyID != "full@example.com" {
					t.Fatalf("expected GroupKeyID full@example.com, got %s", got.GroupKeyID)
				}
				if got.GroupKeyNamespace != "identitysources/full" {
					t.Fatalf("expected GroupKeyNamespace identitysources/full, got %s", got.GroupKeyNamespace)
				}
				if len(got.Labels) != 1 {
					t.Fatalf("expected 1 label, got %d", len(got.Labels))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertGroup(tt.input)
			tt.check(t, got)
		})
	}
}

func TestConvertMember(t *testing.T) {
	tests := []struct {
		name  string
		input *cloudidentity.Membership
		check func(t *testing.T, got types.CloudIdentityMember)
	}{
		{
			name:  "nil input returns zero value",
			input: nil,
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.Name != "" || got.PreferredMemberKeyID != "" {
					t.Fatal("expected zero-value member for nil input")
				}
				if got.Roles != nil {
					t.Fatal("expected nil roles for nil input")
				}
			},
		},
		{
			name: "basic fields mapped correctly",
			input: &cloudidentity.Membership{
				Name:       "groups/abc123/memberships/mem1",
				CreateTime: "2025-02-10T08:00:00Z",
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.Name != "groups/abc123/memberships/mem1" {
					t.Fatalf("expected Name groups/abc123/memberships/mem1, got %s", got.Name)
				}
				if got.CreateTime != "2025-02-10T08:00:00Z" {
					t.Fatalf("expected CreateTime 2025-02-10T08:00:00Z, got %s", got.CreateTime)
				}
			},
		},
		{
			name: "preferred member key with id and namespace",
			input: &cloudidentity.Membership{
				Name: "groups/abc123/memberships/mem2",
				PreferredMemberKey: &cloudidentity.EntityKey{
					Id:        "user@example.com",
					Namespace: "identitysources/example",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.PreferredMemberKeyID != "user@example.com" {
					t.Fatalf("expected PreferredMemberKeyID user@example.com, got %s", got.PreferredMemberKeyID)
				}
				if got.PreferredMemberKeyNamespace != "identitysources/example" {
					t.Fatalf("expected PreferredMemberKeyNamespace identitysources/example, got %s", got.PreferredMemberKeyNamespace)
				}
			},
		},
		{
			name: "preferred member key with id only",
			input: &cloudidentity.Membership{
				Name: "groups/abc123/memberships/mem3",
				PreferredMemberKey: &cloudidentity.EntityKey{
					Id: "admin@example.com",
				},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.PreferredMemberKeyID != "admin@example.com" {
					t.Fatalf("expected PreferredMemberKeyID admin@example.com, got %s", got.PreferredMemberKeyID)
				}
				if got.PreferredMemberKeyNamespace != "" {
					t.Fatalf("expected empty PreferredMemberKeyNamespace, got %s", got.PreferredMemberKeyNamespace)
				}
			},
		},
		{
			name: "nil preferred member key leaves fields empty",
			input: &cloudidentity.Membership{
				Name:               "groups/abc123/memberships/mem4",
				PreferredMemberKey: nil,
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.PreferredMemberKeyID != "" || got.PreferredMemberKeyNamespace != "" {
					t.Fatal("expected empty member key fields when PreferredMemberKey is nil")
				}
			},
		},
		{
			name: "single role",
			input: &cloudidentity.Membership{
				Name: "groups/abc123/memberships/mem5",
				Roles: []*cloudidentity.MembershipRole{
					{Name: "MEMBER"},
				},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if len(got.Roles) != 1 {
					t.Fatalf("expected 1 role, got %d", len(got.Roles))
				}
				if got.Roles[0].Name != "MEMBER" {
					t.Fatalf("expected role name MEMBER, got %s", got.Roles[0].Name)
				}
			},
		},
		{
			name: "multiple roles",
			input: &cloudidentity.Membership{
				Name: "groups/abc123/memberships/mem6",
				Roles: []*cloudidentity.MembershipRole{
					{Name: "OWNER"},
					{Name: "MEMBER"},
					{Name: "MANAGER"},
				},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if len(got.Roles) != 3 {
					t.Fatalf("expected 3 roles, got %d", len(got.Roles))
				}
				expected := []string{"OWNER", "MEMBER", "MANAGER"}
				for i, exp := range expected {
					if got.Roles[i].Name != exp {
						t.Fatalf("expected role[%d] %s, got %s", i, exp, got.Roles[i].Name)
					}
				}
			},
		},
		{
			name: "empty roles slice",
			input: &cloudidentity.Membership{
				Name:  "groups/abc123/memberships/mem7",
				Roles: []*cloudidentity.MembershipRole{},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.Roles != nil {
					t.Fatalf("expected nil roles for empty input slice, got %v", got.Roles)
				}
			},
		},
		{
			name: "all fields together",
			input: &cloudidentity.Membership{
				Name:       "groups/full/memberships/complete",
				CreateTime: "2025-05-20T12:00:00Z",
				PreferredMemberKey: &cloudidentity.EntityKey{
					Id:        "complete@example.com",
					Namespace: "identitysources/complete",
				},
				Roles: []*cloudidentity.MembershipRole{
					{Name: "OWNER"},
					{Name: "MEMBER"},
				},
			},
			check: func(t *testing.T, got types.CloudIdentityMember) {
				if got.Name != "groups/full/memberships/complete" {
					t.Fatalf("expected Name groups/full/memberships/complete, got %s", got.Name)
				}
				if got.CreateTime != "2025-05-20T12:00:00Z" {
					t.Fatalf("expected CreateTime 2025-05-20T12:00:00Z, got %s", got.CreateTime)
				}
				if got.PreferredMemberKeyID != "complete@example.com" {
					t.Fatalf("expected PreferredMemberKeyID complete@example.com, got %s", got.PreferredMemberKeyID)
				}
				if got.PreferredMemberKeyNamespace != "identitysources/complete" {
					t.Fatalf("expected PreferredMemberKeyNamespace identitysources/complete, got %s", got.PreferredMemberKeyNamespace)
				}
				if len(got.Roles) != 2 {
					t.Fatalf("expected 2 roles, got %d", len(got.Roles))
				}
				if got.Roles[0].Name != "OWNER" {
					t.Fatalf("expected first role OWNER, got %s", got.Roles[0].Name)
				}
				if got.Roles[1].Name != "MEMBER" {
					t.Fatalf("expected second role MEMBER, got %s", got.Roles[1].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMember(tt.input)
			tt.check(t, got)
		})
	}
}
