package chat

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/types"
)

// ChatTransport defines the interface for Chat API operations
// This abstraction allows switching between REST and gRPC implementations
type ChatTransport interface {
	// Transport metadata
	TransportType() TransportType
	Close() error

	// Spaces operations
	ListSpaces(ctx context.Context, reqCtx *types.RequestContext, pageSize int, pageToken string) (*types.ChatSpacesListResponse, error)
	GetSpace(ctx context.Context, reqCtx *types.RequestContext, spaceID string) (*types.ChatSpace, error)
	CreateSpace(ctx context.Context, reqCtx *types.RequestContext, displayName, spaceType string, externalUserAllowed bool) (*types.ChatCreateSpaceResponse, error)
	DeleteSpace(ctx context.Context, reqCtx *types.RequestContext, spaceID string) error

	// Messages operations
	ListMessages(ctx context.Context, reqCtx *types.RequestContext, spaceID string, pageSize int, pageToken, filter string) (*types.ChatMessagesListResponse, error)
	GetMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID string) (*types.ChatMessage, error)
	CreateMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, text, threadID string) (*types.ChatCreateMessageResponse, error)
	UpdateMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID, text string) (*types.ChatMessage, error)
	DeleteMessage(ctx context.Context, reqCtx *types.RequestContext, spaceID, messageID string) error

	// Members operations
	ListMembers(ctx context.Context, reqCtx *types.RequestContext, spaceID string, pageSize int, pageToken string) (*types.ChatMembersListResponse, error)
	GetMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, memberID string) (*types.ChatMember, error)
	CreateMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, email, role string) (*types.ChatMember, error)
	DeleteMember(ctx context.Context, reqCtx *types.RequestContext, spaceID, memberID string) error
}

// TransportType represents the type of transport
type TransportType string

const (
	// TransportREST uses the REST/JSON API
	TransportREST TransportType = "rest"
	// TransportGRPC uses the gRPC API
	TransportGRPC TransportType = "grpc"
)
