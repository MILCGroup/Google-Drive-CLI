package chat

import (
	"context"
	"fmt"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/api/chat/v1"
)

// Manager handles Google Chat API operations
type Manager struct {
	client  *api.Client
	service *chat.Service
}

// NewManager creates a new Chat manager
func NewManager(client *api.Client, service *chat.Service) *Manager {
	return &Manager{
		client:  client,
		service: service,
	}
}

// ListSpaces lists all spaces the user has access to
func (m *Manager) ListSpaces(ctx context.Context, reqCtx *types.RequestContext, pageSize int, pageToken string) (*types.ChatSpacesListResponse, error) {
	call := m.service.Spaces.List()
	if pageSize > 0 {
		call = call.PageSize(int64(pageSize))
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.ListSpacesResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	spaces := make([]types.ChatSpace, 0, len(result.Spaces))
	for _, s := range result.Spaces {
		spaces = append(spaces, convertSpace(s))
	}

	return &types.ChatSpacesListResponse{
		Spaces:        spaces,
		NextPageToken: result.NextPageToken,
	}, nil
}

// GetSpace gets details about a specific space
func (m *Manager) GetSpace(ctx context.Context, reqCtx *types.RequestContext, spaceID string) (*types.ChatSpace, error) {
	name := formatSpaceName(spaceID)
	call := m.service.Spaces.Get(name)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Space, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	space := convertSpace(result)
	return &space, nil
}

// CreateSpace creates a new space
func (m *Manager) CreateSpace(ctx context.Context, reqCtx *types.RequestContext, displayName, spaceType string, externalUserAllowed bool) (*types.ChatCreateSpaceResponse, error) {
	space := &chat.Space{
		DisplayName:         displayName,
		ExternalUserAllowed: externalUserAllowed,
	}

	// Set space type if provided
	if spaceType != "" {
		space.SpaceType = spaceType
	}

	call := m.service.Spaces.Create(space)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Space, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.ChatCreateSpaceResponse{
		ID:   result.Name,
		Name: result.DisplayName,
		Type: result.SpaceType,
	}, nil
}

// DeleteSpace deletes a space
func (m *Manager) DeleteSpace(ctx context.Context, reqCtx *types.RequestContext, spaceID string) error {
	name := formatSpaceName(spaceID)
	call := m.service.Spaces.Delete(name)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Empty, error) {
		return call.Do()
	})
	return err
}

// ListMessages lists messages in a space
func (m *Manager) ListMessages(ctx context.Context, reqCtx *types.RequestContext, spaceID string, pageSize int, pageToken, filter string) (*types.ChatMessagesListResponse, error) {
	parent := formatSpaceName(spaceID)
	call := m.service.Spaces.Messages.List(parent)
	if pageSize > 0 {
		call = call.PageSize(int64(pageSize))
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if filter != "" {
		call = call.Filter(filter)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.ListMessagesResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	messages := make([]types.ChatMessage, 0, len(result.Messages))
	for _, msg := range result.Messages {
		messages = append(messages, convertMessage(msg))
	}

	return &types.ChatMessagesListResponse{
		Messages:      messages,
		NextPageToken: result.NextPageToken,
	}, nil
}

// GetMessage gets a specific message
func (m *Manager) GetMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID string) (*types.ChatMessage, error) {
	name := formatMessageName(spaceID, messageID)
	call := m.service.Spaces.Messages.Get(name)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	msg := convertMessage(result)
	return &msg, nil
}

// CreateMessage creates a new message in a space
func (m *Manager) CreateMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, text, threadID string) (*types.ChatCreateMessageResponse, error) {
	parent := formatSpaceName(spaceID)
	message := &chat.Message{
		Text: text,
	}

	// If thread ID is provided, set up thread
	if threadID != "" {
		message.Thread = &chat.Thread{
			Name: fmt.Sprintf("%s/threads/%s", parent, threadID),
		}
	}

	call := m.service.Spaces.Messages.Create(parent, message)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	return &types.ChatCreateMessageResponse{
		ID:         result.Name,
		SpaceID:    parent,
		ThreadID:   getThreadIDFromName(result.Thread),
		Text:       result.Text,
		CreateTime: result.CreateTime,
	}, nil
}

// UpdateMessage updates an existing message
func (m *Manager) UpdateMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID, text string) (*types.ChatMessage, error) {
	name := formatMessageName(spaceID, messageID)
	message := &chat.Message{
		Text: text,
	}

	call := m.service.Spaces.Messages.Patch(name, message)
	call = call.UpdateMask("text")

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Message, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	msg := convertMessage(result)
	return &msg, nil
}

// DeleteMessage deletes a message
func (m *Manager) DeleteMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID string) error {
	name := formatMessageName(spaceID, messageID)
	call := m.service.Spaces.Messages.Delete(name)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Empty, error) {
		return call.Do()
	})
	return err
}

// ListMembers lists members of a space
func (m *Manager) ListMembers(ctx context.Context, reqCtx *types.RequestContext, spaceID string, pageSize int, pageToken string) (*types.ChatMembersListResponse, error) {
	parent := formatSpaceName(spaceID)
	call := m.service.Spaces.Members.List(parent)
	if pageSize > 0 {
		call = call.PageSize(int64(pageSize))
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.ListMembershipsResponse, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	members := make([]types.ChatMember, 0, len(result.Memberships))
	for _, membership := range result.Memberships {
		members = append(members, convertMembership(membership))
	}

	return &types.ChatMembersListResponse{
		Members:       members,
		NextPageToken: result.NextPageToken,
	}, nil
}

// GetMember gets details about a specific member
func (m *Manager) GetMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, memberID string) (*types.ChatMember, error) {
	name := formatMemberName(spaceID, memberID)
	call := m.service.Spaces.Members.Get(name)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Membership, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	member := convertMembership(result)
	return &member, nil
}

// CreateMember adds a member to a space
func (m *Manager) CreateMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, email, role string) (*types.ChatMember, error) {
	parent := formatSpaceName(spaceID)

	// Use the email as the member name (users/email@example.com format)
	memberName := fmt.Sprintf("users/%s", email)

	membership := &chat.Membership{
		Role: role,
		Member: &chat.User{
			Name: memberName,
		},
	}

	call := m.service.Spaces.Members.Create(parent, membership)

	result, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Membership, error) {
		return call.Do()
	})
	if err != nil {
		return nil, err
	}

	member := convertMembership(result)
	return &member, nil
}

// DeleteMember removes a member from a space
func (m *Manager) DeleteMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, memberID string) error {
	name := formatMemberName(spaceID, memberID)
	call := m.service.Spaces.Members.Delete(name)

	_, err := api.ExecuteWithRetry(ctx, m.client, reqCtx, func() (*chat.Membership, error) {
		return call.Do()
	})
	return err
}

// Helper functions

func formatSpaceName(spaceID string) string {
	if spaceID == "" {
		return ""
	}
	// If already formatted as spaces/xxx, return as-is
	if len(spaceID) > 7 && spaceID[:7] == "spaces/" {
		return spaceID
	}
	return fmt.Sprintf("spaces/%s", spaceID)
}

func formatMessageName(spaceID, messageID string) string {
	if messageID == "" {
		return ""
	}
	// If already fully formatted, return as-is
	if len(messageID) > 7 && messageID[:7] == "spaces/" {
		return messageID
	}
	space := formatSpaceName(spaceID)
	return fmt.Sprintf("%s/messages/%s", space, messageID)
}

func formatMemberName(spaceID, memberID string) string {
	if memberID == "" {
		return ""
	}
	// If already fully formatted, return as-is
	if len(memberID) > 7 && memberID[:7] == "spaces/" {
		return memberID
	}
	space := formatSpaceName(spaceID)
	return fmt.Sprintf("%s/members/%s", space, memberID)
}

func convertSpace(s *chat.Space) types.ChatSpace {
	if s == nil {
		return types.ChatSpace{}
	}
	return types.ChatSpace{
		ID:                  s.Name,
		Name:                s.Name,
		Type:                s.SpaceType,
		DisplayName:         s.DisplayName,
		Threaded:            s.Threaded,
		ExternalUserAllowed: s.ExternalUserAllowed,
		SpaceHistoryState:   s.SpaceHistoryState,
		CreateTime:          s.CreateTime,
	}
}

func convertMessage(m *chat.Message) types.ChatMessage {
	if m == nil {
		return types.ChatMessage{}
	}

	msg := types.ChatMessage{
		ID:            m.Name,
		SpaceID:       "",
		ThreadID:      getThreadIDFromName(m.Thread),
		Text:          m.Text,
		FormattedText: m.FormattedText,
		CreateTime:    m.CreateTime,
	}

	// Extract sender info if available
	if m.Sender != nil {
		if m.Sender.DisplayName != "" {
			msg.SenderName = m.Sender.DisplayName
		}
	}

	return msg
}

func convertMembership(m *chat.Membership) types.ChatMember {
	if m == nil {
		return types.ChatMember{}
	}

	member := types.ChatMember{
		ID:       m.Name,
		SpaceID:  "",
		Role:     m.Role,
		State:    "",
		JoinTime: m.CreateTime,
	}

	// Extract member info if available
	if m.Member != nil {
		if m.Member.DisplayName != "" {
			member.Name = m.Member.DisplayName
		}
	}

	return member
}

func getThreadIDFromName(thread *chat.Thread) string {
	if thread == nil || thread.Name == "" {
		return ""
	}
	// Extract thread ID from "spaces/xxx/threads/yyy"
	parts := []rune(thread.Name)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '/' {
			return string(parts[i+1:])
		}
	}
	return thread.Name
}
