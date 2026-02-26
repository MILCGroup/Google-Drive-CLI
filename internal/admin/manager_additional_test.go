package admin

import (
	"testing"

	adminapi "google.golang.org/api/admin/directory/v1"
)

// TestConvertUsers_EdgeCases tests edge cases for user list conversion
func TestConvertUsers_EdgeCases(t *testing.T) {
	t.Run("nil users", func(t *testing.T) {
		got := convertUsers(nil)
		if got == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("nil Users field", func(t *testing.T) {
		got := convertUsers(&adminapi.Users{Users: nil})
		if len(got) != 0 {
			t.Errorf("expected empty slice for nil Users, got %d", len(got))
		}
	})

	t.Run("empty Users slice", func(t *testing.T) {
		got := convertUsers(&adminapi.Users{Users: []*adminapi.User{}})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("user with nil Name", func(t *testing.T) {
		users := &adminapi.Users{
			Users: []*adminapi.User{{
				Id:           "1",
				PrimaryEmail: "user@example.com",
				Name:         nil,
			}},
		}
		got := convertUsers(users)
		if len(got) != 1 {
			t.Fatalf("expected 1 user, got %d", len(got))
		}
		if got[0].Name.FullName != "" || got[0].Name.GivenName != "" || got[0].Name.FamilyName != "" {
			t.Error("expected empty name fields when Name is nil")
		}
	})

	t.Run("multiple users with mixed name data", func(t *testing.T) {
		users := &adminapi.Users{
			Users: []*adminapi.User{
				{
					Id:           "1",
					PrimaryEmail: "user1@example.com",
					Name:         &adminapi.UserName{FullName: "User One", GivenName: "User", FamilyName: "One"},
					IsAdmin:      true,
					Suspended:    false,
				},
				{
					Id:           "2",
					PrimaryEmail: "user2@example.com",
					Name:         nil,
					IsAdmin:      false,
					Suspended:    true,
				},
				{
					Id:           "3",
					PrimaryEmail: "user3@example.com",
					Name:         &adminapi.UserName{FullName: "User Three"},
				},
			},
		}
		got := convertUsers(users)
		if len(got) != 3 {
			t.Fatalf("expected 3 users, got %d", len(got))
		}
		if got[0].Name.FullName != "User One" {
			t.Errorf("expected User One, got %s", got[0].Name.FullName)
		}
		if !got[0].IsAdmin {
			t.Error("expected first user to be admin")
		}
		if got[1].Name.FullName != "" {
			t.Errorf("expected empty name for second user, got %s", got[1].Name.FullName)
		}
		if !got[1].Suspended {
			t.Error("expected second user to be suspended")
		}
		if got[2].Name.FullName != "User Three" {
			t.Errorf("expected User Three, got %s", got[2].Name.FullName)
		}
	})
}

// TestConvertGroups_EdgeCases tests edge cases for group list conversion
func TestConvertGroups_EdgeCases(t *testing.T) {
	t.Run("nil groups", func(t *testing.T) {
		got := convertGroups(nil)
		if got == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("nil Groups field", func(t *testing.T) {
		got := convertGroups(&adminapi.Groups{Groups: nil})
		if len(got) != 0 {
			t.Errorf("expected empty slice for nil Groups, got %d", len(got))
		}
	})

	t.Run("empty Groups slice", func(t *testing.T) {
		got := convertGroups(&adminapi.Groups{Groups: []*adminapi.Group{}})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("group with zero member count", func(t *testing.T) {
		groups := &adminapi.Groups{
			Groups: []*adminapi.Group{{
				Id:                 "1",
				Email:              "group@example.com",
				Name:               "Test Group",
				DirectMembersCount: 0,
			}},
		}
		got := convertGroups(groups)
		if len(got) != 1 {
			t.Fatalf("expected 1 group, got %d", len(got))
		}
		if got[0].DirectMembersCount != 0 {
			t.Errorf("expected 0 members, got %d", got[0].DirectMembersCount)
		}
	})

	t.Run("group with large member count", func(t *testing.T) {
		groups := &adminapi.Groups{
			Groups: []*adminapi.Group{{
				Id:                 "1",
				Email:              "large@example.com",
				Name:               "Large Group",
				DirectMembersCount: 99999,
			}},
		}
		got := convertGroups(groups)
		if got[0].DirectMembersCount != 99999 {
			t.Errorf("expected 99999 members, got %d", got[0].DirectMembersCount)
		}
	})

	t.Run("multiple groups", func(t *testing.T) {
		groups := &adminapi.Groups{
			Groups: []*adminapi.Group{
				{Id: "1", Email: "group1@example.com", Name: "Group One", Description: "First group"},
				{Id: "2", Email: "group2@example.com", Name: "Group Two", Description: "Second group"},
				{Id: "3", Email: "group3@example.com", Name: "Group Three"},
			},
		}
		got := convertGroups(groups)
		if len(got) != 3 {
			t.Fatalf("expected 3 groups, got %d", len(got))
		}
		if got[0].Email != "group1@example.com" {
			t.Errorf("expected group1 email, got %s", got[0].Email)
		}
		if got[1].Name != "Group Two" {
			t.Errorf("expected Group Two, got %s", got[1].Name)
		}
		if got[2].Description != "" {
			t.Errorf("expected empty description, got %s", got[2].Description)
		}
	})
}

// TestConvertUser_EdgeCases tests edge cases for user conversion
func TestConvertUser_EdgeCases(t *testing.T) {
	t.Run("nil user", func(t *testing.T) {
		got := convertUser(nil)
		if got.ID != "" || got.PrimaryEmail != "" {
			t.Error("expected empty user for nil input")
		}
	})

	t.Run("empty user", func(t *testing.T) {
		got := convertUser(&adminapi.User{})
		if got.ID != "" || got.PrimaryEmail != "" {
			t.Error("expected empty fields for empty user")
		}
	})

	t.Run("user with partial name", func(t *testing.T) {
		user := &adminapi.User{
			Id:           "1",
			PrimaryEmail: "user@example.com",
			Name: &adminapi.UserName{
				FullName:   "Full Name",
				GivenName:  "",
				FamilyName: "Name",
			},
		}
		got := convertUser(user)
		if got.Name.FullName != "Full Name" {
			t.Errorf("expected Full Name, got %s", got.Name.FullName)
		}
		if got.Name.GivenName != "" {
			t.Errorf("expected empty given name, got %s", got.Name.GivenName)
		}
		if got.Name.FamilyName != "Name" {
			t.Errorf("expected Name, got %s", got.Name.FamilyName)
		}
	})

	t.Run("user with all timestamps", func(t *testing.T) {
		user := &adminapi.User{
			Id:            "1",
			PrimaryEmail:  "user@example.com",
			CreationTime:  "2024-01-01T00:00:00.000Z",
			LastLoginTime: "2024-06-15T12:30:45.000Z",
		}
		got := convertUser(user)
		if got.CreationTime != "2024-01-01T00:00:00.000Z" {
			t.Errorf("unexpected creation time: %s", got.CreationTime)
		}
		if got.LastLoginTime != "2024-06-15T12:30:45.000Z" {
			t.Errorf("unexpected last login time: %s", got.LastLoginTime)
		}
	})

	t.Run("user with delegated admin flag", func(t *testing.T) {
		user := &adminapi.User{
			Id:               "1",
			PrimaryEmail:     "admin@example.com",
			IsAdmin:          true,
			IsDelegatedAdmin: true,
		}
		got := convertUser(user)
		if !got.IsAdmin {
			t.Error("expected IsAdmin to be true")
		}
		if !got.IsDelegatedAdmin {
			t.Error("expected IsDelegatedAdmin to be true")
		}
	})
}

// TestConvertGroup_EdgeCases tests edge cases for group conversion
func TestConvertGroup_EdgeCases(t *testing.T) {
	t.Run("nil group", func(t *testing.T) {
		got := convertGroup(nil)
		if got.ID != "" || got.Email != "" {
			t.Error("expected empty group for nil input")
		}
	})

	t.Run("empty group", func(t *testing.T) {
		got := convertGroup(&adminapi.Group{})
		if got.ID != "" || got.Email != "" || got.Name != "" {
			t.Error("expected empty fields for empty group")
		}
	})

	t.Run("group with admin created flag", func(t *testing.T) {
		group := &adminapi.Group{
			Id:           "1",
			Email:        "group@example.com",
			Name:         "Test Group",
			AdminCreated: true,
		}
		got := convertGroup(group)
		if !got.AdminCreated {
			t.Error("expected AdminCreated to be true")
		}
	})

	t.Run("group without member count", func(t *testing.T) {
		group := &adminapi.Group{
			Id:    "1",
			Email: "group@example.com",
			Name:  "Test Group",
		}
		got := convertGroup(group)
		if got.DirectMembersCount != 0 {
			t.Errorf("expected 0 members when not set, got %d", got.DirectMembersCount)
		}
	})
}

// TestConvertMembers_EdgeCases tests edge cases for member list conversion
func TestConvertMembers_EdgeCases(t *testing.T) {
	t.Run("nil members", func(t *testing.T) {
		got := convertMembers(nil)
		if got == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("nil Members field", func(t *testing.T) {
		got := convertMembers(&adminapi.Members{Members: nil})
		if len(got) != 0 {
			t.Errorf("expected empty slice for nil Members, got %d", len(got))
		}
	})

	t.Run("empty Members slice", func(t *testing.T) {
		got := convertMembers(&adminapi.Members{Members: []*adminapi.Member{}})
		if len(got) != 0 {
			t.Errorf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("member with all fields", func(t *testing.T) {
		members := &adminapi.Members{
			Members: []*adminapi.Member{{
				Id:     "1",
				Email:  "member@example.com",
				Role:   "MANAGER",
				Type:   "USER",
				Status: "ACTIVE",
			}},
		}
		got := convertMembers(members)
		if len(got) != 1 {
			t.Fatalf("expected 1 member, got %d", len(got))
		}
		if got[0].ID != "1" {
			t.Errorf("expected ID 1, got %s", got[0].ID)
		}
		if got[0].Role != "MANAGER" {
			t.Errorf("expected MANAGER role, got %s", got[0].Role)
		}
		if got[0].Status != "ACTIVE" {
			t.Errorf("expected ACTIVE status, got %s", got[0].Status)
		}
	})

	t.Run("multiple members with different types", func(t *testing.T) {
		members := &adminapi.Members{
			Members: []*adminapi.Member{
				{Id: "1", Email: "user1@example.com", Role: "MEMBER", Type: "USER"},
				{Id: "2", Email: "user2@example.com", Role: "OWNER", Type: "USER"},
				{Id: "3", Email: "group@example.com", Role: "MEMBER", Type: "GROUP"},
			},
		}
		got := convertMembers(members)
		if len(got) != 3 {
			t.Fatalf("expected 3 members, got %d", len(got))
		}
		if got[0].Type != "USER" {
			t.Errorf("expected first member type USER, got %s", got[0].Type)
		}
		if got[2].Type != "GROUP" {
			t.Errorf("expected third member type GROUP, got %s", got[2].Type)
		}
	})
}

// TestConvertMember_EdgeCases tests edge cases for member conversion
func TestConvertMember_EdgeCases(t *testing.T) {
	t.Run("nil member", func(t *testing.T) {
		got := convertMember(nil)
		if got.ID != "" || got.Email != "" {
			t.Error("expected empty member for nil input")
		}
	})

	t.Run("empty member", func(t *testing.T) {
		got := convertMember(&adminapi.Member{})
		if got.ID != "" || got.Email != "" || got.Role != "" || got.Type != "" || got.Status != "" {
			t.Error("expected empty fields for empty member")
		}
	})

	t.Run("member with pending status", func(t *testing.T) {
		member := &adminapi.Member{
			Id:     "1",
			Email:  "pending@example.com",
			Role:   "MEMBER",
			Type:   "USER",
			Status: "PENDING",
		}
		got := convertMember(member)
		if got.Status != "PENDING" {
			t.Errorf("expected PENDING status, got %s", got.Status)
		}
	})
}

// Benchmarks
func BenchmarkConvertUsers(b *testing.B) {
	users := &adminapi.Users{
		Users: make([]*adminapi.User, 100),
	}
	for i := 0; i < 100; i++ {
		users.Users[i] = &adminapi.User{
			Id:           string(rune('a' + i%26)),
			PrimaryEmail: "user@example.com",
			Name:         &adminapi.UserName{FullName: "User", GivenName: "U", FamilyName: "User"},
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertUsers(users)
	}
}

func BenchmarkConvertGroups(b *testing.B) {
	groups := &adminapi.Groups{
		Groups: make([]*adminapi.Group, 100),
	}
	for i := 0; i < 100; i++ {
		groups.Groups[i] = &adminapi.Group{
			Id:    string(rune('a' + i%26)),
			Email: "group@example.com",
			Name:  "Group",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertGroups(groups)
	}
}

func BenchmarkConvertMembers(b *testing.B) {
	members := &adminapi.Members{
		Members: make([]*adminapi.Member, 100),
	}
	for i := 0; i < 100; i++ {
		members.Members[i] = &adminapi.Member{
			Id:    string(rune('a' + i%26)),
			Email: "member@example.com",
			Role:  "MEMBER",
			Type:  "USER",
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertMembers(members)
	}
}
