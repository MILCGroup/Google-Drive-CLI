package people

import (
	"testing"

	peopleapi "google.golang.org/api/people/v1"
)

func TestConstants(t *testing.T) {
	t.Run("defaultPersonFields contains required fields", func(t *testing.T) {
		expected := "names,emailAddresses,phoneNumbers,organizations,addresses"
		if defaultPersonFields != expected {
			t.Fatalf("expected defaultPersonFields %q, got %q", expected, defaultPersonFields)
		}
	})

	t.Run("otherContactReadMask contains required fields", func(t *testing.T) {
		expected := "names,emailAddresses,phoneNumbers"
		if otherContactReadMask != expected {
			t.Fatalf("expected otherContactReadMask %q, got %q", expected, otherContactReadMask)
		}
	})

	t.Run("directoryReadMask contains required fields", func(t *testing.T) {
		expected := "names,emailAddresses,phoneNumbers,organizations"
		if directoryReadMask != expected {
			t.Fatalf("expected directoryReadMask %q, got %q", expected, directoryReadMask)
		}
	})
}

func TestConvertPerson(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := convertPerson(nil)
		if got.ResourceName != "" || got.Etag != "" || got.DisplayName != "" {
			t.Fatalf("expected empty contact for nil input")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertPerson(&peopleapi.Person{})
		if got.ResourceName != "" || got.Etag != "" {
			t.Fatalf("expected empty contact for empty person")
		}
	})

	t.Run("etag preservation", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c123",
			Etag:         "abc123etag",
		}
		got := convertPerson(p)
		if got.Etag != "abc123etag" {
			t.Fatalf("expected etag %q, got %q", "abc123etag", got.Etag)
		}
		if got.ResourceName != "people/c123" {
			t.Fatalf("expected resourceName %q, got %q", "people/c123", got.ResourceName)
		}
	})

	t.Run("names mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c1",
			Names: []*peopleapi.Name{{
				DisplayName: "John Doe",
				GivenName:   "John",
				FamilyName:  "Doe",
			}},
		}
		got := convertPerson(p)
		if got.DisplayName != "John Doe" {
			t.Fatalf("expected displayName %q, got %q", "John Doe", got.DisplayName)
		}
		if got.GivenName != "John" {
			t.Fatalf("expected givenName %q, got %q", "John", got.GivenName)
		}
		if got.FamilyName != "Doe" {
			t.Fatalf("expected familyName %q, got %q", "Doe", got.FamilyName)
		}
	})

	t.Run("emails mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "john@example.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
				{Value: "john.doe@home.com", Type: "home"},
			},
		}
		got := convertPerson(p)
		if len(got.Emails) != 2 {
			t.Fatalf("expected 2 emails, got %d", len(got.Emails))
		}
		if got.Emails[0].Value != "john@example.com" {
			t.Fatalf("expected first email %q, got %q", "john@example.com", got.Emails[0].Value)
		}
		if got.Emails[0].Type != "work" {
			t.Fatalf("expected first email type %q, got %q", "work", got.Emails[0].Type)
		}
		if !got.Emails[0].Primary {
			t.Fatalf("expected first email to be primary")
		}
		if got.Emails[1].Value != "john.doe@home.com" {
			t.Fatalf("expected second email %q, got %q", "john.doe@home.com", got.Emails[1].Value)
		}
		if got.Emails[1].Primary {
			t.Fatalf("expected second email to not be primary")
		}
	})

	t.Run("phones mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+1234567890", Type: "mobile", Metadata: &peopleapi.FieldMetadata{Primary: true}},
				{Value: "+0987654321", Type: "work"},
			},
		}
		got := convertPerson(p)
		if len(got.Phones) != 2 {
			t.Fatalf("expected 2 phones, got %d", len(got.Phones))
		}
		if got.Phones[0].Value != "+1234567890" {
			t.Fatalf("expected first phone %q, got %q", "+1234567890", got.Phones[0].Value)
		}
		if !got.Phones[0].Primary {
			t.Fatalf("expected first phone to be primary")
		}
		if got.Phones[1].Value != "+0987654321" {
			t.Fatalf("expected second phone %q, got %q", "+0987654321", got.Phones[1].Value)
		}
		if got.Phones[1].Primary {
			t.Fatalf("expected second phone to not be primary")
		}
	})

	t.Run("organizations mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			Organizations: []*peopleapi.Organization{
				{Name: "Acme Corp", Title: "Engineer", Department: "R&D"},
			},
		}
		got := convertPerson(p)
		if len(got.Organizations) != 1 {
			t.Fatalf("expected 1 org, got %d", len(got.Organizations))
		}
		if got.Organizations[0].Name != "Acme Corp" {
			t.Fatalf("expected org name %q, got %q", "Acme Corp", got.Organizations[0].Name)
		}
		if got.Organizations[0].Title != "Engineer" {
			t.Fatalf("expected org title %q, got %q", "Engineer", got.Organizations[0].Title)
		}
		if got.Organizations[0].Department != "R&D" {
			t.Fatalf("expected org department %q, got %q", "R&D", got.Organizations[0].Department)
		}
	})

	t.Run("addresses mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			Addresses: []*peopleapi.Address{
				{FormattedValue: "123 Main St", Type: "home", City: "Springfield", Region: "IL", Country: "US"},
			},
		}
		got := convertPerson(p)
		if len(got.Addresses) != 1 {
			t.Fatalf("expected 1 address, got %d", len(got.Addresses))
		}
		if got.Addresses[0].FormattedValue != "123 Main St" {
			t.Fatalf("expected formatted value %q, got %q", "123 Main St", got.Addresses[0].FormattedValue)
		}
		if got.Addresses[0].City != "Springfield" {
			t.Fatalf("expected city %q, got %q", "Springfield", got.Addresses[0].City)
		}
		if got.Addresses[0].Region != "IL" {
			t.Fatalf("expected region %q, got %q", "IL", got.Addresses[0].Region)
		}
		if got.Addresses[0].Country != "US" {
			t.Fatalf("expected country %q, got %q", "US", got.Addresses[0].Country)
		}
	})

	t.Run("full person", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/c42",
			Etag:         "etag-42",
			Names: []*peopleapi.Name{{
				DisplayName: "Jane Smith",
				GivenName:   "Jane",
				FamilyName:  "Smith",
			}},
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "jane@example.com", Type: "work"},
			},
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+555-1234", Type: "mobile"},
			},
			Organizations: []*peopleapi.Organization{
				{Name: "Widgets Inc", Title: "CTO", Department: "Engineering"},
			},
			Addresses: []*peopleapi.Address{
				{FormattedValue: "456 Oak Ave", Type: "work", City: "Denver", Region: "CO", Country: "US"},
			},
		}
		got := convertPerson(p)
		if got.ResourceName != "people/c42" {
			t.Fatalf("expected resourceName %q, got %q", "people/c42", got.ResourceName)
		}
		if got.Etag != "etag-42" {
			t.Fatalf("expected etag %q, got %q", "etag-42", got.Etag)
		}
		if got.DisplayName != "Jane Smith" {
			t.Fatalf("expected displayName %q, got %q", "Jane Smith", got.DisplayName)
		}
		if len(got.Emails) != 1 || got.Emails[0].Value != "jane@example.com" {
			t.Fatalf("expected email mapping")
		}
		if len(got.Phones) != 1 || got.Phones[0].Value != "+555-1234" {
			t.Fatalf("expected phone mapping")
		}
		if len(got.Organizations) != 1 || got.Organizations[0].Name != "Widgets Inc" {
			t.Fatalf("expected org mapping")
		}
		if len(got.Addresses) != 1 || got.Addresses[0].City != "Denver" {
			t.Fatalf("expected address mapping")
		}
	})

	t.Run("multiple names uses first", func(t *testing.T) {
		p := &peopleapi.Person{
			Names: []*peopleapi.Name{
				{DisplayName: "Primary Name", GivenName: "Primary", FamilyName: "Name"},
				{DisplayName: "Secondary Name", GivenName: "Secondary", FamilyName: "Name"},
			},
		}
		got := convertPerson(p)
		if got.DisplayName != "Primary Name" {
			t.Fatalf("expected first name to be used, got %q", got.DisplayName)
		}
		if got.GivenName != "Primary" {
			t.Fatalf("expected first given name %q, got %q", "Primary", got.GivenName)
		}
	})

	t.Run("email without metadata", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "test@example.com", Type: "other"},
			},
		}
		got := convertPerson(p)
		if got.Emails[0].Primary {
			t.Fatalf("expected email without metadata to not be primary")
		}
	})
}

func TestConvertOtherContact(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := convertOtherContact(nil)
		if got.ResourceName != "" || got.DisplayName != "" {
			t.Fatalf("expected empty other contact for nil input")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertOtherContact(&peopleapi.Person{})
		if got.ResourceName != "" {
			t.Fatalf("expected empty resource name")
		}
	})

	t.Run("names mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "otherContacts/c1",
			Names: []*peopleapi.Name{{
				DisplayName: "Other Person",
			}},
		}
		got := convertOtherContact(p)
		if got.ResourceName != "otherContacts/c1" {
			t.Fatalf("expected resourceName %q, got %q", "otherContacts/c1", got.ResourceName)
		}
		if got.DisplayName != "Other Person" {
			t.Fatalf("expected displayName %q, got %q", "Other Person", got.DisplayName)
		}
	})

	t.Run("emails mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "other@example.com", Type: "work"},
			},
		}
		got := convertOtherContact(p)
		if len(got.Emails) != 1 {
			t.Fatalf("expected 1 email, got %d", len(got.Emails))
		}
		if got.Emails[0].Value != "other@example.com" {
			t.Fatalf("expected email %q, got %q", "other@example.com", got.Emails[0].Value)
		}
	})

	t.Run("phones mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+111-222-3333", Type: "home", Metadata: &peopleapi.FieldMetadata{Primary: true}},
			},
		}
		got := convertOtherContact(p)
		if len(got.Phones) != 1 {
			t.Fatalf("expected 1 phone, got %d", len(got.Phones))
		}
		if got.Phones[0].Value != "+111-222-3333" {
			t.Fatalf("expected phone %q, got %q", "+111-222-3333", got.Phones[0].Value)
		}
		if !got.Phones[0].Primary {
			t.Fatalf("expected phone to be primary")
		}
	})

	t.Run("no organizations or addresses", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "otherContacts/c2",
			Organizations: []*peopleapi.Organization{
				{Name: "Should Not Appear"},
			},
			Addresses: []*peopleapi.Address{
				{City: "Ignored"},
			},
		}
		got := convertOtherContact(p)
		// OtherContact type does not have Organizations or Addresses fields
		if got.ResourceName != "otherContacts/c2" {
			t.Fatalf("expected resourceName %q, got %q", "otherContacts/c2", got.ResourceName)
		}
	})
}

func TestConvertDirectoryPerson(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := convertDirectoryPerson(nil)
		if got.ResourceName != "" || got.DisplayName != "" {
			t.Fatalf("expected empty directory person for nil input")
		}
	})

	t.Run("empty person", func(t *testing.T) {
		got := convertDirectoryPerson(&peopleapi.Person{})
		if got.ResourceName != "" {
			t.Fatalf("expected empty resource name")
		}
	})

	t.Run("names mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d1",
			Names: []*peopleapi.Name{{
				DisplayName: "Domain User",
			}},
		}
		got := convertDirectoryPerson(p)
		if got.ResourceName != "people/d1" {
			t.Fatalf("expected resourceName %q, got %q", "people/d1", got.ResourceName)
		}
		if got.DisplayName != "Domain User" {
			t.Fatalf("expected displayName %q, got %q", "Domain User", got.DisplayName)
		}
	})

	t.Run("emails mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "user@corp.com", Type: "work"},
			},
		}
		got := convertDirectoryPerson(p)
		if len(got.Emails) != 1 || got.Emails[0].Value != "user@corp.com" {
			t.Fatalf("expected email mapping")
		}
	})

	t.Run("phones mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+999-888-7777", Type: "work"},
			},
		}
		got := convertDirectoryPerson(p)
		if len(got.Phones) != 1 || got.Phones[0].Value != "+999-888-7777" {
			t.Fatalf("expected phone mapping")
		}
	})

	t.Run("organizations mapping", func(t *testing.T) {
		p := &peopleapi.Person{
			Organizations: []*peopleapi.Organization{
				{Name: "BigCorp", Title: "Director", Department: "Sales"},
			},
		}
		got := convertDirectoryPerson(p)
		if len(got.Organizations) != 1 {
			t.Fatalf("expected 1 org, got %d", len(got.Organizations))
		}
		if got.Organizations[0].Name != "BigCorp" {
			t.Fatalf("expected org name %q, got %q", "BigCorp", got.Organizations[0].Name)
		}
		if got.Organizations[0].Title != "Director" {
			t.Fatalf("expected org title %q, got %q", "Director", got.Organizations[0].Title)
		}
		if got.Organizations[0].Department != "Sales" {
			t.Fatalf("expected org department %q, got %q", "Sales", got.Organizations[0].Department)
		}
	})

	t.Run("no addresses field", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d2",
			Addresses: []*peopleapi.Address{
				{City: "Ignored"},
			},
		}
		got := convertDirectoryPerson(p)
		// DirectoryPerson type does not have Addresses field
		if got.ResourceName != "people/d2" {
			t.Fatalf("expected resourceName %q, got %q", "people/d2", got.ResourceName)
		}
	})

	t.Run("full directory person", func(t *testing.T) {
		p := &peopleapi.Person{
			ResourceName: "people/d42",
			Names: []*peopleapi.Name{{
				DisplayName: "Full User",
			}},
			EmailAddresses: []*peopleapi.EmailAddress{
				{Value: "full@corp.com", Type: "work", Metadata: &peopleapi.FieldMetadata{Primary: true}},
			},
			PhoneNumbers: []*peopleapi.PhoneNumber{
				{Value: "+123", Type: "mobile"},
			},
			Organizations: []*peopleapi.Organization{
				{Name: "FullCorp", Title: "VP", Department: "Eng"},
			},
		}
		got := convertDirectoryPerson(p)
		if got.ResourceName != "people/d42" {
			t.Fatalf("expected resourceName %q, got %q", "people/d42", got.ResourceName)
		}
		if got.DisplayName != "Full User" {
			t.Fatalf("expected displayName %q, got %q", "Full User", got.DisplayName)
		}
		if len(got.Emails) != 1 || !got.Emails[0].Primary {
			t.Fatalf("expected primary email")
		}
		if len(got.Phones) != 1 || got.Phones[0].Value != "+123" {
			t.Fatalf("expected phone mapping")
		}
		if len(got.Organizations) != 1 || got.Organizations[0].Name != "FullCorp" {
			t.Fatalf("expected org mapping")
		}
	})
}

func TestNewManager(t *testing.T) {
	t.Run("creates non-nil manager", func(t *testing.T) {
		m := NewManager(nil, nil)
		if m == nil {
			t.Fatal("expected non-nil manager")
		}
	})
}
