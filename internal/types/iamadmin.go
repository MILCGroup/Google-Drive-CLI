package types

import "fmt"

// ServiceAccount represents an IAM service account
type ServiceAccount struct {
	Name           string `json:"name"`
	ProjectId      string `json:"projectId,omitempty"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	Description    string `json:"description,omitempty"`
	Disabled       bool   `json:"disabled"`
	Oauth2ClientId string `json:"oauth2ClientId,omitempty"`
}

func (s *ServiceAccount) Headers() []string {
	return []string{"Account Name", "Email", "Display Name", "Disabled"}
}

func (s *ServiceAccount) Rows() [][]string {
	return [][]string{{
		truncateID(s.Name, 40),
		s.Email,
		s.DisplayName,
		fmt.Sprintf("%v", s.Disabled),
	}}
}

func (s *ServiceAccount) EmptyMessage() string {
	return "No service account information available"
}

// ServiceAccountKey represents a service account key
type ServiceAccountKey struct {
	Name         string `json:"name"`
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`
	KeyOrigin    string `json:"keyOrigin,omitempty"`
	KeyType      string `json:"keyType,omitempty"`
	ValidAfter   string `json:"validAfter,omitempty"`
	ValidBefore  string `json:"validBefore,omitempty"`
}

func (k *ServiceAccountKey) Headers() []string {
	return []string{"Key Name", "Algorithm", "Type", "Valid After", "Valid Before"}
}

func (k *ServiceAccountKey) Rows() [][]string {
	return [][]string{{
		truncateID(k.Name, 40),
		k.KeyAlgorithm,
		k.KeyType,
		k.ValidAfter,
		k.ValidBefore,
	}}
}

func (k *ServiceAccountKey) EmptyMessage() string {
	return "No key information available"
}

// IAMRole represents an IAM role
type IAMRole struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Stage       string `json:"stage,omitempty"`
}

func (r *IAMRole) Headers() []string {
	return []string{"Role Name", "Title", "Stage"}
}

func (r *IAMRole) Rows() [][]string {
	return [][]string{{
		truncateID(r.Name, 40),
		r.Title,
		r.Stage,
	}}
}

func (r *IAMRole) EmptyMessage() string {
	return "No role information available"
}

// ServiceAccountsListResponse represents a list of service accounts
type ServiceAccountsListResponse struct {
	Accounts      []ServiceAccount `json:"accounts"`
	NextPageToken string           `json:"nextPageToken,omitempty"`
}

func (r *ServiceAccountsListResponse) Headers() []string {
	return []string{"Account Name", "Email", "Display Name", "Disabled"}
}

func (r *ServiceAccountsListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Accounts))
	for i, account := range r.Accounts {
		rows[i] = []string{
			truncateID(account.Name, 40),
			account.Email,
			account.DisplayName,
			fmt.Sprintf("%v", account.Disabled),
		}
	}
	return rows
}

func (r *ServiceAccountsListResponse) EmptyMessage() string {
	return "No service accounts found"
}

// ServiceAccountKeysListResponse represents a list of service account keys
type ServiceAccountKeysListResponse struct {
	Keys []ServiceAccountKey `json:"keys"`
}

func (r *ServiceAccountKeysListResponse) Headers() []string {
	return []string{"Key Name", "Algorithm", "Type", "Valid After", "Valid Before"}
}

func (r *ServiceAccountKeysListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Keys))
	for i, key := range r.Keys {
		rows[i] = []string{
			truncateID(key.Name, 40),
			key.KeyAlgorithm,
			key.KeyType,
			key.ValidAfter,
			key.ValidBefore,
		}
	}
	return rows
}

func (r *ServiceAccountKeysListResponse) EmptyMessage() string {
	return "No keys found"
}

// IAMRolesListResponse represents a list of IAM roles
type IAMRolesListResponse struct {
	Roles         []IAMRole `json:"roles"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

func (r *IAMRolesListResponse) Headers() []string {
	return []string{"Role Name", "Title", "Stage"}
}

func (r *IAMRolesListResponse) Rows() [][]string {
	rows := make([][]string, len(r.Roles))
	for i, role := range r.Roles {
		rows[i] = []string{
			truncateID(role.Name, 40),
			role.Title,
			role.Stage,
		}
	}
	return rows
}

func (r *IAMRolesListResponse) EmptyMessage() string {
	return "No roles found"
}
