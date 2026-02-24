package cli

import (
	"context"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	chatmgr "github.com/milcgroup/gdrv/internal/chat"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	chatapi "google.golang.org/api/chat/v1"
)

type ChatCmd struct {
	Spaces   ChatSpacesCmd   `cmd:"" help:"Manage Google Chat spaces"`
	Messages ChatMessagesCmd `cmd:"" help:"Manage Google Chat messages"`
	Members  ChatMembersCmd  `cmd:"" help:"Manage Google Chat members"`
}

type ChatSpacesCmd struct {
	List   ChatSpacesListCmd   `cmd:"" help:"List all spaces"`
	Get    ChatSpacesGetCmd    `cmd:"" help:"Get details about a specific space"`
	Create ChatSpacesCreateCmd `cmd:"" help:"Create a new space"`
	Delete ChatSpacesDeleteCmd `cmd:"" help:"Delete a space"`
}

type ChatMessagesCmd struct {
	List   ChatMessagesListCmd   `cmd:"" help:"List messages in a space"`
	Get    ChatMessagesGetCmd    `cmd:"" help:"Get a specific message"`
	Create ChatMessagesCreateCmd `cmd:"" help:"Create a message in a space"`
	Update ChatMessagesUpdateCmd `cmd:"" help:"Update a message"`
	Delete ChatMessagesDeleteCmd `cmd:"" help:"Delete a message"`
}

type ChatMembersCmd struct {
	List   ChatMembersListCmd   `cmd:"" help:"List members of a space"`
	Get    ChatMembersGetCmd    `cmd:"" help:"Get member details"`
	Create ChatMembersCreateCmd `cmd:"" help:"Add a member to a space"`
	Delete ChatMembersDeleteCmd `cmd:"" help:"Remove a member from a space"`
}

// Spaces command structs

type ChatSpacesListCmd struct {
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type ChatSpacesGetCmd struct {
	SpaceID string `arg:"" name:"space-id" help:"Space ID"`
}

type ChatSpacesCreateCmd struct {
	DisplayName   string `help:"Display name for the space" name:"display-name"`
	Type          string `help:"Space type (SPACE or GROUP_CHAT)" default:"SPACE" name:"type"`
	ExternalUsers bool   `help:"Allow external users" name:"external-users"`
}

type ChatSpacesDeleteCmd struct {
	SpaceID string `arg:"" name:"space-id" help:"Space ID"`
}

// Messages command structs

type ChatMessagesListCmd struct {
	SpaceID   string `arg:"" name:"space-id" help:"Space ID"`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Filter    string `help:"Filter for messages" name:"filter"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type ChatMessagesGetCmd struct {
	SpaceID   string `arg:"" name:"space-id" help:"Space ID"`
	MessageID string `arg:"" name:"message-id" help:"Message ID"`
}

type ChatMessagesCreateCmd struct {
	SpaceID  string `arg:"" name:"space-id" help:"Space ID"`
	Text     string `help:"Message text (required)" name:"text"`
	ThreadID string `help:"Thread ID to reply in" name:"thread"`
}

type ChatMessagesUpdateCmd struct {
	SpaceID   string `arg:"" name:"space-id" help:"Space ID"`
	MessageID string `arg:"" name:"message-id" help:"Message ID"`
	Text      string `help:"New message text (required)" name:"text"`
}

type ChatMessagesDeleteCmd struct {
	SpaceID   string `arg:"" name:"space-id" help:"Space ID"`
	MessageID string `arg:"" name:"message-id" help:"Message ID"`
}

// Members command structs

type ChatMembersListCmd struct {
	SpaceID   string `arg:"" name:"space-id" help:"Space ID"`
	Limit     int    `help:"Maximum results per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type ChatMembersGetCmd struct {
	SpaceID  string `arg:"" name:"space-id" help:"Space ID"`
	MemberID string `arg:"" name:"member-id" help:"Member ID"`
}

type ChatMembersCreateCmd struct {
	SpaceID string `arg:"" name:"space-id" help:"Space ID"`
	Email   string `help:"Member email (required)" name:"email"`
	Role    string `help:"Member role (MEMBER or MANAGER)" default:"MEMBER" name:"role"`
}

type ChatMembersDeleteCmd struct {
	SpaceID  string `arg:"" name:"space-id" help:"Space ID"`
	MemberID string `arg:"" name:"member-id" help:"Member ID"`
}

// Run methods for Spaces

func (cmd *ChatSpacesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allSpaces []types.ChatSpace
		pageToken := ""
		for {
			result, err := mgr.ListSpaces(ctx, reqCtx, cmd.Limit, pageToken)
			if err != nil {
				return handleCLIError(out, "chat.spaces.list", err)
			}
			allSpaces = append(allSpaces, result.Spaces...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("chat.spaces.list", allSpaces)
		}
		return out.WriteSuccess("chat.spaces.list", map[string]interface{}{
			"spaces": allSpaces,
		})
	}

	result, err := mgr.ListSpaces(ctx, reqCtx, cmd.Limit, cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "chat.spaces.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.spaces.list", result.Spaces)
	}
	return out.WriteSuccess("chat.spaces.list", result)
}

func (cmd *ChatSpacesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetSpace(ctx, reqCtx, cmd.SpaceID)
	if err != nil {
		return handleCLIError(out, "chat.spaces.get", err)
	}

	return out.WriteSuccess("chat.spaces.get", result)
}

func (cmd *ChatSpacesCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if cmd.DisplayName == "" {
		return out.WriteError("chat.spaces.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--display-name is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateSpace(ctx, reqCtx, cmd.DisplayName, cmd.Type, cmd.ExternalUsers)
	if err != nil {
		return handleCLIError(out, "chat.spaces.create", err)
	}

	return out.WriteSuccess("chat.spaces.create", result)
}

func (cmd *ChatSpacesDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteSpace(ctx, reqCtx, cmd.SpaceID); err != nil {
		return handleCLIError(out, "chat.spaces.delete", err)
	}

	return out.WriteSuccess("chat.spaces.delete", map[string]string{
		"spaceId": cmd.SpaceID,
		"status":  "deleted",
	})
}

// Run methods for Messages

func (cmd *ChatMessagesListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allMessages []types.ChatMessage
		pageToken := ""
		for {
			result, err := mgr.ListMessages(ctx, reqCtx, cmd.SpaceID, cmd.Limit, pageToken, cmd.Filter)
			if err != nil {
				return handleCLIError(out, "chat.messages.list", err)
			}
			allMessages = append(allMessages, result.Messages...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("chat.messages.list", allMessages)
		}
		return out.WriteSuccess("chat.messages.list", map[string]interface{}{
			"messages": allMessages,
		})
	}

	result, err := mgr.ListMessages(ctx, reqCtx, cmd.SpaceID, cmd.Limit, cmd.PageToken, cmd.Filter)
	if err != nil {
		return handleCLIError(out, "chat.messages.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.messages.list", result.Messages)
	}
	return out.WriteSuccess("chat.messages.list", result)
}

func (cmd *ChatMessagesGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetMessage(ctx, reqCtx, cmd.SpaceID, cmd.MessageID)
	if err != nil {
		return handleCLIError(out, "chat.messages.get", err)
	}

	return out.WriteSuccess("chat.messages.get", result)
}

func (cmd *ChatMessagesCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if cmd.Text == "" {
		return out.WriteError("chat.messages.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--text is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateMessage(ctx, reqCtx, cmd.SpaceID, cmd.Text, cmd.ThreadID)
	if err != nil {
		return handleCLIError(out, "chat.messages.create", err)
	}

	return out.WriteSuccess("chat.messages.create", result)
}

func (cmd *ChatMessagesUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if cmd.Text == "" {
		return out.WriteError("chat.messages.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--text is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.UpdateMessage(ctx, reqCtx, cmd.SpaceID, cmd.MessageID, cmd.Text)
	if err != nil {
		return handleCLIError(out, "chat.messages.update", err)
	}

	return out.WriteSuccess("chat.messages.update", result)
}

func (cmd *ChatMessagesDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteMessage(ctx, reqCtx, cmd.SpaceID, cmd.MessageID); err != nil {
		return handleCLIError(out, "chat.messages.delete", err)
	}

	return out.WriteSuccess("chat.messages.delete", map[string]string{
		"spaceId":   cmd.SpaceID,
		"messageId": cmd.MessageID,
		"status":    "deleted",
	})
}

// Run methods for Members

func (cmd *ChatMembersListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allMembers []types.ChatMember
		pageToken := ""
		for {
			result, err := mgr.ListMembers(ctx, reqCtx, cmd.SpaceID, cmd.Limit, pageToken)
			if err != nil {
				return handleCLIError(out, "chat.members.list", err)
			}
			allMembers = append(allMembers, result.Members...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("chat.members.list", allMembers)
		}
		return out.WriteSuccess("chat.members.list", map[string]interface{}{
			"members": allMembers,
		})
	}

	result, err := mgr.ListMembers(ctx, reqCtx, cmd.SpaceID, cmd.Limit, cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "chat.members.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.members.list", result.Members)
	}
	return out.WriteSuccess("chat.members.list", result)
}

func (cmd *ChatMembersGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetMember(ctx, reqCtx, cmd.SpaceID, cmd.MemberID)
	if err != nil {
		return handleCLIError(out, "chat.members.get", err)
	}

	return out.WriteSuccess("chat.members.get", result)
}

func (cmd *ChatMembersCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if cmd.Email == "" {
		return out.WriteError("chat.members.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--email is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateMember(ctx, reqCtx, cmd.SpaceID, cmd.Email, cmd.Role)
	if err != nil {
		return handleCLIError(out, "chat.members.create", err)
	}

	return out.WriteSuccess("chat.members.create", result)
}

func (cmd *ChatMembersDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteMember(ctx, reqCtx, cmd.SpaceID, cmd.MemberID); err != nil {
		return handleCLIError(out, "chat.members.delete", err)
	}

	return out.WriteSuccess("chat.members.delete", map[string]string{
		"spaceId":  cmd.SpaceID,
		"memberId": cmd.MemberID,
		"status":   "removed",
	})
}

// Helper function to get Chat service
func getChatService(ctx context.Context, flags types.GlobalFlags) (*chatapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceChat); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetChatService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return svc, client, reqCtx, nil
}
