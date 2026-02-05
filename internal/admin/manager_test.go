package admin

import (
	"testing"

	adminapi "google.golang.org/api/admin/directory/v1"
)

func TestConvertUsers(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := convertUsers(nil); len(got) != 0 {
			t.Fatalf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("empty users", func(t *testing.T) {
		if got := convertUsers(&adminapi.Users{Users: []*adminapi.User{}}); len(got) != 0 {
			t.Fatalf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("missing name", func(t *testing.T) {
		users := &adminapi.Users{Users: []*adminapi.User{{Id: "1", PrimaryEmail: "a@example.com"}}}
		got := convertUsers(users)
		if got[0].Name.FullName != "" || got[0].Name.GivenName != "" || got[0].Name.FamilyName != "" {
			t.Fatalf("expected empty name fields")
		}
	})

	t.Run("full fields", func(t *testing.T) {
		users := &adminapi.Users{Users: []*adminapi.User{{
			Id:           "1",
			PrimaryEmail: "a@example.com",
			Name:         &adminapi.UserName{FullName: "A B", GivenName: "A", FamilyName: "B"},
			IsAdmin:      true,
			Suspended:    true,
		}}}
		got := convertUsers(users)
		if got[0].PrimaryEmail != "a@example.com" {
			t.Fatalf("expected email to match")
		}
		if got[0].Name.FullName != "A B" {
			t.Fatalf("expected full name to match")
		}
		if !got[0].IsAdmin || !got[0].Suspended {
			t.Fatalf("expected admin and suspended to be true")
		}
	})
}

func TestConvertGroups(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := convertGroups(nil); len(got) != 0 {
			t.Fatalf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("valid member count", func(t *testing.T) {
		groups := &adminapi.Groups{Groups: []*adminapi.Group{{
			Id:                 "1",
			Email:              "g@example.com",
			Name:               "Group",
			DirectMembersCount: 42,
		}}}
		got := convertGroups(groups)
		if got[0].DirectMembersCount != 42 {
			t.Fatalf("expected count 42, got %d", got[0].DirectMembersCount)
		}
	})
}

func TestConvertUser(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := convertUser(nil)
		if got.ID != "" || got.PrimaryEmail != "" {
			t.Fatalf("expected empty user")
		}
	})

	t.Run("with name", func(t *testing.T) {
		user := &adminapi.User{
			Id:           "1",
			PrimaryEmail: "a@example.com",
			Name:         &adminapi.UserName{FullName: "A B", GivenName: "A", FamilyName: "B"},
		}
		got := convertUser(user)
		if got.Name.FullName != "A B" || got.Name.GivenName != "A" || got.Name.FamilyName != "B" {
			t.Fatalf("expected name fields")
		}
	})
}

func TestConvertGroup(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := convertGroup(nil)
		if got.ID != "" || got.Email != "" {
			t.Fatalf("expected empty group")
		}
	})

	t.Run("with count", func(t *testing.T) {
		group := &adminapi.Group{
			Id:                 "1",
			Email:              "g@example.com",
			Name:               "Group",
			DirectMembersCount: 3,
		}
		got := convertGroup(group)
		if got.DirectMembersCount != 3 {
			t.Fatalf("expected count 3")
		}
	})
}

func TestBoolPtr(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		ptr := boolPtr(true)
		if ptr == nil {
			t.Fatal("expected non-nil pointer")
		}
		if !*ptr {
			t.Fatal("expected true value")
		}
	})

	t.Run("false", func(t *testing.T) {
		ptr := boolPtr(false)
		if ptr == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *ptr {
			t.Fatal("expected false value")
		}
	})
}

func TestConvertMembers(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := convertMembers(nil); len(got) != 0 {
			t.Fatalf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("empty members", func(t *testing.T) {
		if got := convertMembers(&adminapi.Members{Members: []*adminapi.Member{}}); len(got) != 0 {
			t.Fatalf("expected empty slice, got %d", len(got))
		}
	})

	t.Run("multiple members", func(t *testing.T) {
		members := &adminapi.Members{Members: []*adminapi.Member{
			{Id: "1", Email: "a@example.com", Role: "MEMBER", Type: "USER"},
			{Id: "2", Email: "b@example.com", Role: "OWNER", Type: "GROUP"},
		}}
		got := convertMembers(members)
		if len(got) != 2 {
			t.Fatalf("expected 2 members, got %d", len(got))
		}
		if got[0].Email != "a@example.com" {
			t.Fatalf("expected first member email to match")
		}
		if got[1].Role != "OWNER" {
			t.Fatalf("expected second member role to be OWNER")
		}
	})
}

func TestConvertMember(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := convertMember(nil)
		if got.ID != "" || got.Email != "" {
			t.Fatalf("expected empty member")
		}
	})

	t.Run("full member", func(t *testing.T) {
		member := &adminapi.Member{
			Id:    "1",
			Email: "a@example.com",
			Role:  "MEMBER",
			Type:  "USER",
		}
		got := convertMember(member)
		if got.ID != "1" {
			t.Fatalf("expected ID 1, got %s", got.ID)
		}
		if got.Email != "a@example.com" {
			t.Fatalf("expected email a@example.com, got %s", got.Email)
		}
		if got.Role != "MEMBER" {
			t.Fatalf("expected role MEMBER, got %s", got.Role)
		}
		if got.Type != "USER" {
			t.Fatalf("expected type USER, got %s", got.Type)
		}
	})
}
