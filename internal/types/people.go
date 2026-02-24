package types

// ContactEmail represents an email address on a contact.
type ContactEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ContactPhone represents a phone number on a contact.
type ContactPhone struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ContactOrg represents an organization entry on a contact.
type ContactOrg struct {
	Name       string `json:"name,omitempty"`
	Title      string `json:"title,omitempty"`
	Department string `json:"department,omitempty"`
}

// ContactAddress represents a physical address on a contact.
type ContactAddress struct {
	FormattedValue string `json:"formattedValue,omitempty"`
	Type           string `json:"type,omitempty"`
	City           string `json:"city,omitempty"`
	Region         string `json:"region,omitempty"`
	Country        string `json:"country,omitempty"`
}

// Contact represents a single Google People contact.
type Contact struct {
	ResourceName  string           `json:"resourceName"`
	Etag          string           `json:"etag"`
	DisplayName   string           `json:"displayName,omitempty"`
	GivenName     string           `json:"givenName,omitempty"`
	FamilyName    string           `json:"familyName,omitempty"`
	Emails        []ContactEmail   `json:"emails,omitempty"`
	Phones        []ContactPhone   `json:"phones,omitempty"`
	Organizations []ContactOrg     `json:"organizations,omitempty"`
	Addresses     []ContactAddress `json:"addresses,omitempty"`
}

func (c *Contact) Headers() []string {
	return []string{"Name", "Email", "Phone", "Organization"}
}

func (c *Contact) Rows() [][]string {
	email := ""
	if len(c.Emails) > 0 {
		email = c.Emails[0].Value
	}
	phone := ""
	if len(c.Phones) > 0 {
		phone = c.Phones[0].Value
	}
	org := ""
	if len(c.Organizations) > 0 {
		org = c.Organizations[0].Name
	}
	return [][]string{{
		c.DisplayName,
		email,
		phone,
		org,
	}}
}

func (c *Contact) EmptyMessage() string {
	return "No contact found"
}

// ContactList represents a list of Google People contacts.
type ContactList struct {
	Contacts    []Contact `json:"contacts"`
	TotalPeople int       `json:"totalPeople,omitempty"`
	TotalItems  int       `json:"totalItems,omitempty"`
}

func (cl *ContactList) Headers() []string {
	return []string{"Name", "Email", "Phone", "Organization"}
}

func (cl *ContactList) Rows() [][]string {
	rows := make([][]string, len(cl.Contacts))
	for i, c := range cl.Contacts {
		email := ""
		if len(c.Emails) > 0 {
			email = c.Emails[0].Value
		}
		phone := ""
		if len(c.Phones) > 0 {
			phone = c.Phones[0].Value
		}
		org := ""
		if len(c.Organizations) > 0 {
			org = c.Organizations[0].Name
		}
		rows[i] = []string{
			c.DisplayName,
			email,
			phone,
			org,
		}
	}
	return rows
}

func (cl *ContactList) EmptyMessage() string {
	return "No contacts found"
}

// OtherContact represents a contact from the "Other contacts" section.
type OtherContact struct {
	ResourceName string         `json:"resourceName"`
	DisplayName  string         `json:"displayName,omitempty"`
	Emails       []ContactEmail `json:"emails,omitempty"`
	Phones       []ContactPhone `json:"phones,omitempty"`
}

func (oc *OtherContact) Headers() []string {
	return []string{"Name", "Email", "Phone"}
}

func (oc *OtherContact) Rows() [][]string {
	email := ""
	if len(oc.Emails) > 0 {
		email = oc.Emails[0].Value
	}
	phone := ""
	if len(oc.Phones) > 0 {
		phone = oc.Phones[0].Value
	}
	return [][]string{{
		oc.DisplayName,
		email,
		phone,
	}}
}

func (oc *OtherContact) EmptyMessage() string {
	return "No other contact found"
}

// OtherContactList represents a list of "Other contacts".
type OtherContactList struct {
	Contacts []OtherContact `json:"contacts"`
}

func (ol *OtherContactList) Headers() []string {
	return []string{"Name", "Email", "Phone"}
}

func (ol *OtherContactList) Rows() [][]string {
	rows := make([][]string, len(ol.Contacts))
	for i, oc := range ol.Contacts {
		email := ""
		if len(oc.Emails) > 0 {
			email = oc.Emails[0].Value
		}
		phone := ""
		if len(oc.Phones) > 0 {
			phone = oc.Phones[0].Value
		}
		rows[i] = []string{
			oc.DisplayName,
			email,
			phone,
		}
	}
	return rows
}

func (ol *OtherContactList) EmptyMessage() string {
	return "No other contacts found"
}

// DirectoryPerson represents a person from the directory (domain).
type DirectoryPerson struct {
	ResourceName  string         `json:"resourceName"`
	DisplayName   string         `json:"displayName,omitempty"`
	Emails        []ContactEmail `json:"emails,omitempty"`
	Phones        []ContactPhone `json:"phones,omitempty"`
	Organizations []ContactOrg   `json:"organizations,omitempty"`
}

func (dp *DirectoryPerson) Headers() []string {
	return []string{"Name", "Email", "Phone", "Organization"}
}

func (dp *DirectoryPerson) Rows() [][]string {
	email := ""
	if len(dp.Emails) > 0 {
		email = dp.Emails[0].Value
	}
	phone := ""
	if len(dp.Phones) > 0 {
		phone = dp.Phones[0].Value
	}
	org := ""
	if len(dp.Organizations) > 0 {
		org = dp.Organizations[0].Name
	}
	return [][]string{{
		dp.DisplayName,
		email,
		phone,
		org,
	}}
}

func (dp *DirectoryPerson) EmptyMessage() string {
	return "No directory person found"
}

// DirectoryPersonList represents a list of directory people.
type DirectoryPersonList struct {
	People []DirectoryPerson `json:"people"`
}

func (dl *DirectoryPersonList) Headers() []string {
	return []string{"Name", "Email", "Phone", "Organization"}
}

func (dl *DirectoryPersonList) Rows() [][]string {
	rows := make([][]string, len(dl.People))
	for i, dp := range dl.People {
		email := ""
		if len(dp.Emails) > 0 {
			email = dp.Emails[0].Value
		}
		phone := ""
		if len(dp.Phones) > 0 {
			phone = dp.Phones[0].Value
		}
		org := ""
		if len(dp.Organizations) > 0 {
			org = dp.Organizations[0].Name
		}
		rows[i] = []string{
			dp.DisplayName,
			email,
			phone,
			org,
		}
	}
	return rows
}

func (dl *DirectoryPersonList) EmptyMessage() string {
	return "No directory people found"
}

// ContactResult represents the result of a contact mutation (create, update, delete).
type ContactResult struct {
	ResourceName string `json:"resourceName"`
	DisplayName  string `json:"displayName,omitempty"`
}

func (cr *ContactResult) Headers() []string {
	return []string{"Resource Name", "Display Name"}
}

func (cr *ContactResult) Rows() [][]string {
	return [][]string{{
		cr.ResourceName,
		cr.DisplayName,
	}}
}

func (cr *ContactResult) EmptyMessage() string {
	return "No result"
}

