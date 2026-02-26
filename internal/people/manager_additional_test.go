package people

import (
	"testing"

	peopleapi "google.golang.org/api/people/v1"
)

// TestConvertPerson_EdgeCases tests edge cases for person conversion
func TestConvertPerson_EdgeCases(t *testing.T) {
	t.Run("nil person", func(t *testing.T) {
		got := convertPerson(nil)
		if got.ResourceName != "" || got.Etag != "" {
			t.Error("expected empty fields for nil person")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertPerson(&peopleapi.Person{})
		if got.ResourceName != "" || got.Etag != "" {
			t.Error("expected empty fields for empty person")
		}
	})

	t.Run("person with nil names", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c1",
			Names:        nil,
		}
		got := convertPerson(p)
		if got.ResourceName != "people/c1" {
			t.Errorf("expected resource name, got %s", got.ResourceName)
		}
		if got.DisplayName != "" {
			t.Errorf("expected empty display name, got %s", got.DisplayName)
		}
	})

	t.Run("person with empty names", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c1",
			Names:        []*peopleapi.Name{},
		}
		got := convertPerson(p)
		if got.DisplayName != "" {
			t.Errorf("expected empty display name, got %s", got.DisplayName)
		}
	})

	t.Run("person with nil emails", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName:   "people/c1",
			Names:          []*peopleapi.Name{{DisplayName: "Test"}},
			EmailAddresses: nil,
		}
		got := convertPerson(p)
		if got.DisplayName != "Test" {
			t.Errorf("expected Test, got %s", got.DisplayName)
		}
		if got.Emails != nil {
			t.Error("expected nil emails")
		}
	})

	t.Run("person with nil phones", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c1",
			PhoneNumbers: nil,
		}
		got := convertPerson(p)
		if got.Phones != nil {
			t.Error("expected nil phones")
		}
	})

	t.Run("person with nil organizations", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName:  "people/c1",
			Organizations: nil,
		}
		got := convertPerson(p)
		if got.Organizations != nil {
			t.Error("expected nil organizations")
		}
	})

	t.Run("person with nil addresses", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c1",
			Addresses:    nil,
		}
		got := convertPerson(p)
		if got.Addresses != nil {
			t.Error("expected nil addresses")
		}
	})

	t.Run("email with nil metadata", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "test@example.com", Type: "work", Metadata: nil},
			},
		}
		got := convertPerson(p)
		if len(got.Emails) != 1 {
			t.Fatalf("expected 1 email, got %d", len(got.Emails))
		}
		if got.Emails[0].Primary {
			t.Error("expected email to not be primary when metadata is nil")
		}
	})

	t.Run("phone with nil metadata", func(t *testing.T) {
		p := &peopleapi.Person{
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+123", Type: "mobile", Metadata: nil},
			},
		}
		got := convertPerson(p)
		if len(got.Phones) != 1 {
			t.Fatalf("expected 1 phone, got %d", len(got.Phones))
		}
		if got.Phones[0].Primary {
			t.Error("expected phone to not be primary when metadata is nil")
		}
	})

	t.Run("multiple emails with mixed primary", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "primary@example.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
				{Value: "secondary@example.com", Type: "home"},
				{Value: "other@example.com", Type: "other", Metadata: &peopleapi.FieldMetadata{Primary: false}},
			},
		}
		got := convertPerson(p)
		if len(got.Emails) != 3 {
			t.Fatalf("expected 3 emails, got %d", len(got.Emails))
		}
		if !got.Emails[0].Primary {
			t.Error("expected first email to be primary")
		}
		if got.Emails[1].Primary {
			t.Error("expected second email to not be primary")
		}
		if got.Emails[2].Primary {
			t.Error("expected third email to not be primary")
		}
	})

	t.Run("full person with all fields", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c123",
			Etag:         "abc123",
			Names: []*peopleapi.Name{{
				DisplayName: "John Doe",
				GivenName:   "John",
				FamilyName:  "Doe",
			}},
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "john@example.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
			},
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+1-234-567", Type: "mobile", Metadata: &peopleapi.FieldMetadata{Primary: true}},
			},
			Organizations: []*peopleapi.Organization{
				{Name: "Acme Corp", Title: "Engineer", Department: "Engineering"},
			},
			Addresses: []*peopleapi.Address{
				{FormattedValue: "123 Main St", Type: "home", City: "Springfield", Region: "IL", Country: "US"},
			},
		}
		got := convertPerson(p)
		if got.ResourceName != "people/c123" {
			t.Errorf("expected resource name people/c123, got %s", got.ResourceName)
		}
		if got.Etag != "abc123" {
			t.Errorf("expected etag abc123, got %s", got.Etag)
		}
		if got.DisplayName != "John Doe" {
			t.Errorf("expected John Doe, got %s", got.DisplayName)
		}
		if got.GivenName != "John" {
			t.Errorf("expected John, got %s", got.GivenName)
		}
		if got.FamilyName != "Doe" {
			t.Errorf("expected Doe, got %s", got.FamilyName)
		}
		if len(got.Emails) != 1 || got.Emails[0].Value != "john@example.com" {
			t.Error("email mismatch")
		}
		if len(got.Phones) != 1 || got.Phones[0].Value != "+1-234-567" {
			t.Error("phone mismatch")
		}
		if len(got.Organizations) != 1 || got.Organizations[0].Name != "Acme Corp" {
			t.Error("organization mismatch")
		}
		if len(got.Addresses) != 1 || got.Addresses[0].City != "Springfield" {
			t.Error("address mismatch")
		}
	})
}

// TestConvertOtherContact_EdgeCases tests edge cases for other contact conversion
func TestConvertOtherContact_EdgeCases(t *testing.T) {
	t.Run("nil person", func(t *testing.T) {
		got := convertOtherContact(nil)
		if got.ResourceName != "" {
			t.Error("expected empty resource name for nil person")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertOtherContact(&peopleapi.Person{})
		if got.ResourceName != "" {
			t.Error("expected empty resource name for empty person")
		}
	})

	t.Run("other contact with nil names", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "otherContacts/c1",
			Names:        nil,
		}
		got := convertOtherContact(p)
		if got.ResourceName != "otherContacts/c1" {
			t.Errorf("expected resource name, got %s", got.ResourceName)
		}
		if got.DisplayName != "" {
			t.Errorf("expected empty display name, got %s", got.DisplayName)
		}
	})

	t.Run("other contact with nil emails", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName:   "otherContacts/c1",
			EmailAddresses: nil,
		}
		got := convertOtherContact(p)
		if got.Emails != nil {
			t.Error("expected nil emails")
		}
	})

	t.Run("other contact with nil phones", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "otherContacts/c1",
			PhoneNumbers: nil,
		}
		got := convertOtherContact(p)
		if got.Phones != nil {
			t.Error("expected nil phones")
		}
	})

	t.Run("full other contact", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "otherContacts/c1",
			Names: []*peopleapi.Name{{
				DisplayName: "Other Contact",
			}},
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "other@example.com", Type: "work"},
			},
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+987-654-3210", Type: "home"},
			},
		}
		got := convertOtherContact(p)
		if got.ResourceName != "otherContacts/c1" {
			t.Errorf("expected otherContacts/c1, got %s", got.ResourceName)
		}
		if got.DisplayName != "Other Contact" {
			t.Errorf("expected Other Contact, got %s", got.DisplayName)
		}
		if len(got.Emails) != 1 || got.Emails[0].Value != "other@example.com" {
			t.Error("email mismatch")
		}
		if len(got.Phones) != 1 || got.Phones[0].Value != "+987-654-3210" {
			t.Error("phone mismatch")
		}
	})
}

// TestConvertDirectoryPerson_EdgeCases tests edge cases for directory person conversion
func TestConvertDirectoryPerson_EdgeCases(t *testing.T) {
	t.Run("nil person", func(t *testing.T) {
		got := convertDirectoryPerson(nil)
		if got.ResourceName != "" {
			t.Error("expected empty resource name for nil person")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertDirectoryPerson(&peopleapi.Person{})
		if got.ResourceName != "" {
			t.Error("expected empty resource name for empty person")
		}
	})

	t.Run("directory person with nil names", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d1",
			Names:        nil,
		}
		got := convertDirectoryPerson(p)
		if got.ResourceName != "people/d1" {
			t.Errorf("expected resource name, got %s", got.ResourceName)
		}
		if got.DisplayName != "" {
			t.Errorf("expected empty display name, got %s", got.DisplayName)
		}
	})

	t.Run("directory person with nil emails", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName:   "people/d1",
			EmailAddresses: nil,
		}
		got := convertDirectoryPerson(p)
		if got.Emails != nil {
			t.Error("expected nil emails")
		}
	})

	t.Run("directory person with nil phones", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d1",
			PhoneNumbers: nil,
		}
		got := convertDirectoryPerson(p)
		if got.Phones != nil {
			t.Error("expected nil phones")
		}
	})

	t.Run("directory person with nil organizations", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName:  "people/d1",
			Organizations: nil,
		}
		got := convertDirectoryPerson(p)
		if got.Organizations != nil {
			t.Error("expected nil organizations")
		}
	})

	t.Run("full directory person", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d1",
			Names: []*peopleapi.Name{{
				DisplayName: "Domain User",
			}},
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "user@domain.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
			},
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+555-555-5555", Type: "work"},
			},
			Organizations: []*peopleapi.Organization{
				{Name: "BigCorp", Title: "Director", Department: "Sales"},
			},
		}
		got := convertDirectoryPerson(p)
		if got.ResourceName != "people/d1" {
			t.Errorf("expected people/d1, got %s", got.ResourceName)
		}
		if got.DisplayName != "Domain User" {
			t.Errorf("expected Domain User, got %s", got.DisplayName)
		}
		if len(got.Emails) != 1 || !got.Emails[0].Primary {
			t.Error("expected primary email")
		}
		if len(got.Phones) != 1 || got.Phones[0].Value != "+555-555-5555" {
			t.Error("phone mismatch")
		}
		if len(got.Organizations) != 1 || got.Organizations[0].Title != "Director" {
			t.Error("organization mismatch")
		}
	})
}

// Benchmarks
func BenchmarkConvertPerson(b *testing.B) {
	p := &peopleapi.Person{
		ResourceName: "people/c1",
		Etag:         "etag123",
		Names: []*peopleapi.Name{{
			DisplayName: "Test User",
			GivenName:   "Test",
			FamilyName:  "User",
		}},
		EmailAddresses: []*peopleapi.EmailAddress{
			{Value: "test@example.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
		},
		PhoneNumbers: []*peopleapi.PhoneNumber{
			{Value: "+123", Type: "mobile"},
		},
		Organizations: []*peopleapi.Organization{
			{Name: "Corp", Title: "Dev"},
		},
		Addresses: []*peopleapi.Address{
			{FormattedValue: "123 Main St", City: "City"},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertPerson(p)
	}
}

func BenchmarkConvertPersonNil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertPerson(nil)
	}
}
