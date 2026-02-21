package people

import (
	"context"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/types"
	people "google.golang.org/api/people/v1"
)

const defaultPersonFields = "names,emailAddresses,phoneNumbers,organizations,addresses"
const otherContactReadMask = "names,emailAddresses,phoneNumbers"
const directoryReadMask = "names,emailAddresses,phoneNumbers,organizations"

// Manager provides operations for the Google People/Contacts API.
type Manager struct {
	client  *api.Client
	service *people.Service
}

// NewManager creates a new People API manager.
func NewManager(client *api.Client, service *people.Service) *Manager {
	return &Manager{client: client, service: service}
}

// ListContacts lists the authenticated user's contacts.
func (m *Manager) ListContacts(ctx context.Context, reqCtx *types.RequestContext, pageSize int64, pageToken, sortOrder string) (*types.ContactList, string, error) {
	call := m.service.People.Connections.List("people/me").
		PersonFields(defaultPersonFields)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if sortOrder != "" {
		call = call.SortOrder(sortOrder)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.ListConnectionsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	contacts := make([]types.Contact, 0, len(result.Connections))
	for _, p := range result.Connections {
		contacts = append(contacts, convertPerson(p))
	}

	return &types.ContactList{
		Contacts:    contacts,
		TotalPeople: int(result.TotalPeople),
		TotalItems:  int(result.TotalItems),
	}, result.NextPageToken, nil
}

// SearchContacts searches the authenticated user's contacts.
func (m *Manager) SearchContacts(ctx context.Context, reqCtx *types.RequestContext, query string, pageSize int64) (*types.ContactList, error) {
	call := m.service.People.SearchContacts().
		Query(query).
		ReadMask(defaultPersonFields)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.SearchResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]types.Contact, 0, len(result.Results))
	for _, r := range result.Results {
		if r.Person != nil {
			contacts = append(contacts, convertPerson(r.Person))
		}
	}

	return &types.ContactList{
		Contacts: contacts,
	}, nil
}

// GetContact retrieves a single contact by resource name.
func (m *Manager) GetContact(ctx context.Context, reqCtx *types.RequestContext, resourceName string) (*types.Contact, error) {
	call := m.service.People.Get(resourceName).
		PersonFields(defaultPersonFields)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.Person, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	contact := convertPerson(result)
	return &contact, nil
}

// CreateContact creates a new contact with the given details.
func (m *Manager) CreateContact(ctx context.Context, reqCtx *types.RequestContext, givenName, familyName string, emails, phones []string) (*types.ContactResult, error) {
	person := &people.Person{
		Names: []*people.Name{{
			GivenName:  givenName,
			FamilyName: familyName,
		}},
	}

	if len(emails) > 0 {
		person.EmailAddresses = make([]*people.EmailAddress, len(emails))
		for i, e := range emails {
			person.EmailAddresses[i] = &people.EmailAddress{Value: e}
		}
	}

	if len(phones) > 0 {
		person.PhoneNumbers = make([]*people.PhoneNumber, len(phones))
		for i, p := range phones {
			person.PhoneNumbers[i] = &people.PhoneNumber{Value: p}
		}
	}

	call := m.service.People.CreateContact(person)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.Person, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	displayName := ""
	if len(result.Names) > 0 {
		displayName = result.Names[0].DisplayName
	}

	return &types.ContactResult{
		ResourceName: result.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// UpdateContact updates an existing contact. Etag is required. Name fields
// are optional -- pass nil to leave unchanged.
func (m *Manager) UpdateContact(ctx context.Context, reqCtx *types.RequestContext, resourceName, etag string, givenName, familyName *string, emails, phones []string) (*types.ContactResult, error) {
	person := &people.Person{
		Etag: etag,
	}

	var updateFields []string

	if givenName != nil || familyName != nil {
		name := &people.Name{}
		if givenName != nil {
			name.GivenName = *givenName
		}
		if familyName != nil {
			name.FamilyName = *familyName
		}
		person.Names = []*people.Name{name}
		updateFields = append(updateFields, "names")
	}

	if emails != nil {
		person.EmailAddresses = make([]*people.EmailAddress, len(emails))
		for i, e := range emails {
			person.EmailAddresses[i] = &people.EmailAddress{Value: e}
		}
		updateFields = append(updateFields, "emailAddresses")
	}

	if phones != nil {
		person.PhoneNumbers = make([]*people.PhoneNumber, len(phones))
		for i, p := range phones {
			person.PhoneNumbers[i] = &people.PhoneNumber{Value: p}
		}
		updateFields = append(updateFields, "phoneNumbers")
	}

	call := m.service.People.UpdateContact(resourceName, person).
		UpdatePersonFields(strings.Join(updateFields, ","))

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.Person, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	displayName := ""
	if len(result.Names) > 0 {
		displayName = result.Names[0].DisplayName
	}

	return &types.ContactResult{
		ResourceName: result.ResourceName,
		DisplayName:  displayName,
	}, nil
}

// DeleteContact deletes a contact by resource name.
func (m *Manager) DeleteContact(ctx context.Context, reqCtx *types.RequestContext, resourceName string) error {
	call := m.service.People.DeleteContact(resourceName)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (struct{}, error) {
		_, doErr := call.Do()
		return struct{}{}, doErr
	})
	return err
}

// ListOtherContacts lists contacts in the "Other contacts" section.
func (m *Manager) ListOtherContacts(ctx context.Context, reqCtx *types.RequestContext, pageSize int64, pageToken string) (*types.OtherContactList, string, error) {
	call := m.service.OtherContacts.List().
		ReadMask(otherContactReadMask)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.ListOtherContactsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	contacts := make([]types.OtherContact, 0, len(result.OtherContacts))
	for _, p := range result.OtherContacts {
		contacts = append(contacts, convertOtherContact(p))
	}

	return &types.OtherContactList{
		Contacts: contacts,
	}, result.NextPageToken, nil
}

// SearchOtherContacts searches the "Other contacts" section.
func (m *Manager) SearchOtherContacts(ctx context.Context, reqCtx *types.RequestContext, query string, pageSize int64) (*types.OtherContactList, error) {
	call := m.service.OtherContacts.Search().
		Query(query).
		ReadMask(otherContactReadMask)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.SearchResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]types.OtherContact, 0, len(result.Results))
	for _, r := range result.Results {
		if r.Person != nil {
			contacts = append(contacts, convertOtherContact(r.Person))
		}
	}

	return &types.OtherContactList{
		Contacts: contacts,
	}, nil
}

// ListDirectory lists people in the domain directory.
// Note: the ListDirectoryPeople API does not support a query parameter;
// the query argument is accepted for interface consistency but is unused.
func (m *Manager) ListDirectory(ctx context.Context, reqCtx *types.RequestContext, pageSize int64, pageToken, _ string) (*types.DirectoryPersonList, string, error) {
	call := m.service.People.ListDirectoryPeople().
		ReadMask(directoryReadMask).
		Sources("DIRECTORY_SOURCE_TYPE_DOMAIN_PROFILE")
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.ListDirectoryPeopleResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	people := make([]types.DirectoryPerson, 0, len(result.People))
	for _, p := range result.People {
		people = append(people, convertDirectoryPerson(p))
	}

	return &types.DirectoryPersonList{
		People: people,
	}, result.NextPageToken, nil
}

// SearchDirectory searches the domain directory for people.
func (m *Manager) SearchDirectory(ctx context.Context, reqCtx *types.RequestContext, query string, pageSize int64) (*types.DirectoryPersonList, error) {
	call := m.service.People.SearchDirectoryPeople().
		Query(query).
		ReadMask(directoryReadMask).
		Sources("DIRECTORY_SOURCE_TYPE_DOMAIN_PROFILE")
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*people.SearchDirectoryPeopleResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	dirPeople := make([]types.DirectoryPerson, 0, len(result.People))
	for _, p := range result.People {
		dirPeople = append(dirPeople, convertDirectoryPerson(p))
	}

	return &types.DirectoryPersonList{
		People: dirPeople,
	}, nil
}

// convertPerson maps a People API Person to a domain Contact.
func convertPerson(p *people.Person) types.Contact {
	if p == nil {
		return types.Contact{}
	}

	c := types.Contact{
		ResourceName: p.ResourceName,
		Etag:         p.Etag,
	}

	if len(p.Names) > 0 {
		c.DisplayName = p.Names[0].DisplayName
		c.GivenName = p.Names[0].GivenName
		c.FamilyName = p.Names[0].FamilyName
	}

	if len(p.EmailAddresses) > 0 {
		c.Emails = make([]types.ContactEmail, len(p.EmailAddresses))
		for i, e := range p.EmailAddresses {
			c.Emails[i] = types.ContactEmail{
				Value: e.Value,
				Type:  e.Type,
			}
			if e.Metadata != nil && e.Metadata.Primary {
				c.Emails[i].Primary = true
			}
		}
	}

	if len(p.PhoneNumbers) > 0 {
		c.Phones = make([]types.ContactPhone, len(p.PhoneNumbers))
		for i, ph := range p.PhoneNumbers {
			c.Phones[i] = types.ContactPhone{
				Value: ph.Value,
				Type:  ph.Type,
			}
			if ph.Metadata != nil && ph.Metadata.Primary {
				c.Phones[i].Primary = true
			}
		}
	}

	if len(p.Organizations) > 0 {
		c.Organizations = make([]types.ContactOrg, len(p.Organizations))
		for i, o := range p.Organizations {
			c.Organizations[i] = types.ContactOrg{
				Name:       o.Name,
				Title:      o.Title,
				Department: o.Department,
			}
		}
	}

	if len(p.Addresses) > 0 {
		c.Addresses = make([]types.ContactAddress, len(p.Addresses))
		for i, a := range p.Addresses {
			c.Addresses[i] = types.ContactAddress{
				FormattedValue: a.FormattedValue,
				Type:           a.Type,
				City:           a.City,
				Region:         a.Region,
				Country:        a.Country,
			}
		}
	}

	return c
}

// convertOtherContact maps a People API Person to a domain OtherContact.
func convertOtherContact(p *people.Person) types.OtherContact {
	if p == nil {
		return types.OtherContact{}
	}

	oc := types.OtherContact{
		ResourceName: p.ResourceName,
	}

	if len(p.Names) > 0 {
		oc.DisplayName = p.Names[0].DisplayName
	}

	if len(p.EmailAddresses) > 0 {
		oc.Emails = make([]types.ContactEmail, len(p.EmailAddresses))
		for i, e := range p.EmailAddresses {
			oc.Emails[i] = types.ContactEmail{
				Value: e.Value,
				Type:  e.Type,
			}
			if e.Metadata != nil && e.Metadata.Primary {
				oc.Emails[i].Primary = true
			}
		}
	}

	if len(p.PhoneNumbers) > 0 {
		oc.Phones = make([]types.ContactPhone, len(p.PhoneNumbers))
		for i, ph := range p.PhoneNumbers {
			oc.Phones[i] = types.ContactPhone{
				Value: ph.Value,
				Type:  ph.Type,
			}
			if ph.Metadata != nil && ph.Metadata.Primary {
				oc.Phones[i].Primary = true
			}
		}
	}

	return oc
}

// convertDirectoryPerson maps a People API Person to a domain DirectoryPerson.
func convertDirectoryPerson(p *people.Person) types.DirectoryPerson {
	if p == nil {
		return types.DirectoryPerson{}
	}

	dp := types.DirectoryPerson{
		ResourceName: p.ResourceName,
	}

	if len(p.Names) > 0 {
		dp.DisplayName = p.Names[0].DisplayName
	}

	if len(p.EmailAddresses) > 0 {
		dp.Emails = make([]types.ContactEmail, len(p.EmailAddresses))
		for i, e := range p.EmailAddresses {
			dp.Emails[i] = types.ContactEmail{
				Value: e.Value,
				Type:  e.Type,
			}
			if e.Metadata != nil && e.Metadata.Primary {
				dp.Emails[i].Primary = true
			}
		}
	}

	if len(p.PhoneNumbers) > 0 {
		dp.Phones = make([]types.ContactPhone, len(p.PhoneNumbers))
		for i, ph := range p.PhoneNumbers {
			dp.Phones[i] = types.ContactPhone{
				Value: ph.Value,
				Type:  ph.Type,
			}
			if ph.Metadata != nil && ph.Metadata.Primary {
				dp.Phones[i].Primary = true
			}
		}
	}

	if len(p.Organizations) > 0 {
		dp.Organizations = make([]types.ContactOrg, len(p.Organizations))
		for i, o := range p.Organizations {
			dp.Organizations[i] = types.ContactOrg{
				Name:       o.Name,
				Title:      o.Title,
				Department: o.Department,
			}
		}
	}

	return dp
}
