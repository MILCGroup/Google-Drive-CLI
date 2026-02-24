package groups

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/types"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

// Manager handles Cloud Identity Groups API operations.
type Manager struct {
	client  *api.Client
	service *cloudidentity.Service
}

// NewManager creates a new Cloud Identity Groups manager.
func NewManager(client *api.Client, service *cloudidentity.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// ListGroups lists Cloud Identity groups under the given parent.
// Parent must be in the format "customers/{customerId}".
// Returns the group list, the next page token (empty if no more pages), and any error.
func (m *Manager) ListGroups(ctx context.Context, reqCtx *types.RequestContext, parent string, pageSize int64, pageToken string) (*types.CloudIdentityGroupList, string, error) {
	call := m.service.Groups.List().Parent(parent)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*cloudidentity.ListGroupsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	groups := make([]types.CloudIdentityGroup, len(result.Groups))
	for i, g := range result.Groups {
		groups[i] = convertGroup(g)
	}

	return &types.CloudIdentityGroupList{Groups: groups}, result.NextPageToken, nil
}

// ListMembers lists memberships of a Cloud Identity group.
// groupName must be in the format "groups/{groupId}".
// Returns the member list, the next page token (empty if no more pages), and any error.
func (m *Manager) ListMembers(ctx context.Context, reqCtx *types.RequestContext, groupName string, pageSize int64, pageToken string) (*types.CloudIdentityMemberList, string, error) {
	call := m.service.Groups.Memberships.List(groupName)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*cloudidentity.ListMembershipsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, "", err
	}

	members := make([]types.CloudIdentityMember, len(result.Memberships))
	for i, membership := range result.Memberships {
		members[i] = convertMember(membership)
	}

	return &types.CloudIdentityMemberList{Members: members}, result.NextPageToken, nil
}

// convertGroup maps a Cloud Identity Group to the internal type.
func convertGroup(g *cloudidentity.Group) types.CloudIdentityGroup {
	if g == nil {
		return types.CloudIdentityGroup{}
	}

	group := types.CloudIdentityGroup{
		Name:        g.Name,
		DisplayName: g.DisplayName,
		Description: g.Description,
		CreateTime:  g.CreateTime,
		UpdateTime:  g.UpdateTime,
	}

	if g.GroupKey != nil {
		group.GroupKeyID = g.GroupKey.Id
		group.GroupKeyNamespace = g.GroupKey.Namespace
	}

	if g.Labels != nil {
		group.Labels = make(map[string]string, len(g.Labels))
		for k, v := range g.Labels {
			group.Labels[k] = v
		}
	}

	return group
}

// convertMember maps a Cloud Identity Membership to the internal type.
func convertMember(m *cloudidentity.Membership) types.CloudIdentityMember {
	if m == nil {
		return types.CloudIdentityMember{}
	}

	member := types.CloudIdentityMember{
		Name:       m.Name,
		CreateTime: m.CreateTime,
	}

	if m.PreferredMemberKey != nil {
		member.PreferredMemberKeyID = m.PreferredMemberKey.Id
		member.PreferredMemberKeyNamespace = m.PreferredMemberKey.Namespace
	}

	if len(m.Roles) > 0 {
		member.Roles = make([]types.MemberRole, len(m.Roles))
		for i, role := range m.Roles {
			member.Roles[i] = types.MemberRole{Name: role.Name}
		}
	}

	return member
}
