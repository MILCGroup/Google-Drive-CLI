package iamadmin

import (
	"context"
	"errors"
	"fmt"

	iam "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Manager handles IAM Admin API operations via gRPC
type Manager struct {
	client *iam.IamClient
}

// NewManager creates a new IAM Admin manager
func NewManager(ctx context.Context, opts ...option.ClientOption) (*Manager, error) {
	client, err := iam.NewIamClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM admin client: %w", err)
	}

	return &Manager{
		client: client,
	}, nil
}

// Close closes the client
func (m *Manager) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// ListServiceAccounts lists service accounts in a project
func (m *Manager) ListServiceAccounts(ctx context.Context, reqCtx *types.RequestContext, projectID, pageToken string, pageSize int32) (*types.ServiceAccountsListResponse, error) {
	req := &adminpb.ListServiceAccountsRequest{
		Name:      fmt.Sprintf("projects/%s", projectID),
		PageToken: pageToken,
		PageSize:  pageSize,
	}

	iter := m.client.ListServiceAccounts(ctx, req)
	var accounts []types.ServiceAccount
	for {
		account, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list service accounts: %w", err)
		}
		accounts = append(accounts, convertServiceAccount(account))
	}

	return &types.ServiceAccountsListResponse{
		Accounts: accounts,
	}, nil
}

// GetServiceAccount gets a specific service account
func (m *Manager) GetServiceAccount(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.ServiceAccount, error) {
	req := &adminpb.GetServiceAccountRequest{
		Name: name,
	}

	resp, err := m.client.GetServiceAccount(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}

	account := convertServiceAccount(resp)
	return &account, nil
}

// CreateServiceAccount creates a new service account
func (m *Manager) CreateServiceAccount(ctx context.Context, reqCtx *types.RequestContext, projectID, accountID, displayName, description string) (*types.ServiceAccount, error) {
	req := &adminpb.CreateServiceAccountRequest{
		Name:      fmt.Sprintf("projects/%s", projectID),
		AccountId: accountID,
		ServiceAccount: &adminpb.ServiceAccount{
			DisplayName: displayName,
			Description: description,
		},
	}

	resp, err := m.client.CreateServiceAccount(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create service account: %w", err)
	}

	account := convertServiceAccount(resp)
	return &account, nil
}

// DeleteServiceAccount deletes a service account
func (m *Manager) DeleteServiceAccount(ctx context.Context, reqCtx *types.RequestContext, name string) error {
	req := &adminpb.DeleteServiceAccountRequest{
		Name: name,
	}

	if err := m.client.DeleteServiceAccount(ctx, req); err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	return nil
}

// ListServiceAccountKeys lists keys for a service account
func (m *Manager) ListServiceAccountKeys(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.ServiceAccountKeysListResponse, error) {
	req := &adminpb.ListServiceAccountKeysRequest{
		Name: name,
	}

	resp, err := m.client.ListServiceAccountKeys(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list service account keys: %w", err)
	}

	var keys []types.ServiceAccountKey
	for _, key := range resp.Keys {
		keys = append(keys, convertServiceAccountKey(key))
	}

	return &types.ServiceAccountKeysListResponse{
		Keys: keys,
	}, nil
}

// DeleteServiceAccountKey deletes a service account key
func (m *Manager) DeleteServiceAccountKey(ctx context.Context, reqCtx *types.RequestContext, name string) error {
	req := &adminpb.DeleteServiceAccountKeyRequest{
		Name: name,
	}

	if err := m.client.DeleteServiceAccountKey(ctx, req); err != nil {
		return fmt.Errorf("failed to delete service account key: %w", err)
	}

	return nil
}

// ListRoles lists IAM roles
func (m *Manager) ListRoles(ctx context.Context, reqCtx *types.RequestContext, parent, pageToken string, pageSize int32) (*types.IAMRolesListResponse, error) {
	req := &adminpb.ListRolesRequest{
		Parent:    parent,
		PageToken: pageToken,
		PageSize:  pageSize,
	}

	resp, err := m.client.ListRoles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	var roles []types.IAMRole
	for _, role := range resp.Roles {
		roles = append(roles, convertRole(role))
	}

	return &types.IAMRolesListResponse{
		Roles:         roles,
		NextPageToken: resp.NextPageToken,
	}, nil
}

// GetRole gets a specific IAM role
func (m *Manager) GetRole(ctx context.Context, reqCtx *types.RequestContext, name string) (*types.IAMRole, error) {
	req := &adminpb.GetRoleRequest{
		Name: name,
	}

	resp, err := m.client.GetRole(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role := convertRole(resp)
	return &role, nil
}

func convertServiceAccount(account *adminpb.ServiceAccount) types.ServiceAccount {
	return types.ServiceAccount{
		Name:           account.Name,
		ProjectId:      account.ProjectId,
		Email:          account.Email,
		DisplayName:    account.DisplayName,
		Description:    account.Description,
		Disabled:       account.Disabled,
		Oauth2ClientId: account.Oauth2ClientId,
	}
}

func convertServiceAccountKey(key *adminpb.ServiceAccountKey) types.ServiceAccountKey {
	k := types.ServiceAccountKey{
		Name:         key.Name,
		KeyAlgorithm: key.KeyAlgorithm.String(),
		KeyOrigin:    key.KeyOrigin.String(),
		KeyType:      key.KeyType.String(),
	}
	if key.ValidAfterTime != nil {
		k.ValidAfter = key.ValidAfterTime.AsTime().Format("2006-01-02 15:04:05")
	}
	if key.ValidBeforeTime != nil {
		k.ValidBefore = key.ValidBeforeTime.AsTime().Format("2006-01-02 15:04:05")
	}
	return k
}

func convertRole(role *adminpb.Role) types.IAMRole {
	return types.IAMRole{
		Name:        role.Name,
		Title:       role.Title,
		Description: role.Description,
		Stage:       role.Stage.String(),
	}
}
