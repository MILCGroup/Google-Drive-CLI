package cli

import (
	"context"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	chatmgr "github.com/dl-alexandre/gdrv/internal/chat"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	"github.com/spf13/cobra"
	chatapi "google.golang.org/api/chat/v1"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Google Chat operations",
	Long:  "Commands for managing Google Chat spaces, messages, and members",
}

// Spaces commands
var chatSpacesCmd = &cobra.Command{
	Use:   "spaces",
	Short: "Manage Chat spaces",
}

var chatSpacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Chat spaces",
	RunE:  runChatSpacesList,
}

var chatSpacesGetCmd = &cobra.Command{
	Use:   "get <space-id>",
	Short: "Get space details",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatSpacesGet,
}

var chatSpacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new space",
	RunE:  runChatSpacesCreate,
}

var chatSpacesDeleteCmd = &cobra.Command{
	Use:   "delete <space-id>",
	Short: "Delete a space",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatSpacesDelete,
}

// Messages commands
var chatMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Manage Chat messages",
}

var chatMessagesListCmd = &cobra.Command{
	Use:   "list <space-id>",
	Short: "List messages in a space",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatMessagesList,
}

var chatMessagesGetCmd = &cobra.Command{
	Use:   "get <space-id> <message-id>",
	Short: "Get a specific message",
	Args:  cobra.ExactArgs(2),
	RunE:  runChatMessagesGet,
}

var chatMessagesCreateCmd = &cobra.Command{
	Use:   "create <space-id>",
	Short: "Create a message in a space",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatMessagesCreate,
}

var chatMessagesUpdateCmd = &cobra.Command{
	Use:   "update <space-id> <message-id>",
	Short: "Update a message",
	Args:  cobra.ExactArgs(2),
	RunE:  runChatMessagesUpdate,
}

var chatMessagesDeleteCmd = &cobra.Command{
	Use:   "delete <space-id> <message-id>",
	Short: "Delete a message",
	Args:  cobra.ExactArgs(2),
	RunE:  runChatMessagesDelete,
}

// Members commands
var chatMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Manage Chat space members",
}

var chatMembersListCmd = &cobra.Command{
	Use:   "list <space-id>",
	Short: "List members of a space",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatMembersList,
}

var chatMembersGetCmd = &cobra.Command{
	Use:   "get <space-id> <member-id>",
	Short: "Get member details",
	Args:  cobra.ExactArgs(2),
	RunE:  runChatMembersGet,
}

var chatMembersCreateCmd = &cobra.Command{
	Use:   "create <space-id>",
	Short: "Add a member to a space",
	Args:  cobra.ExactArgs(1),
	RunE:  runChatMembersCreate,
}

var chatMembersDeleteCmd = &cobra.Command{
	Use:   "delete <space-id> <member-id>",
	Short: "Remove a member from a space",
	Args:  cobra.ExactArgs(2),
	RunE:  runChatMembersDelete,
}

// Flags
var (
	chatSpaceType          string
	chatSpaceDisplayName   string
	chatSpaceExternalUsers bool
	chatMessageText        string
	chatMessageThreadID    string
	chatMessageFilter      string
	chatMemberEmail        string
	chatMemberRole         string
	chatListLimit          int
	chatListPageToken      string
	chatListPaginate       bool
)

func init() {
	// Spaces flags
	chatSpacesListCmd.Flags().IntVar(&chatListLimit, "limit", 100, "Maximum results per page")
	chatSpacesListCmd.Flags().StringVar(&chatListPageToken, "page-token", "", "Page token for pagination")
	chatSpacesListCmd.Flags().BoolVar(&chatListPaginate, "paginate", false, "Automatically fetch all pages")

	chatSpacesCreateCmd.Flags().StringVar(&chatSpaceDisplayName, "display-name", "", "Display name for the space")
	chatSpacesCreateCmd.Flags().StringVar(&chatSpaceType, "type", "SPACE", "Space type (SPACE or GROUP_CHAT)")
	chatSpacesCreateCmd.Flags().BoolVar(&chatSpaceExternalUsers, "external-users", false, "Allow external users")

	// Messages flags
	chatMessagesListCmd.Flags().IntVar(&chatListLimit, "limit", 100, "Maximum results per page")
	chatMessagesListCmd.Flags().StringVar(&chatListPageToken, "page-token", "", "Page token for pagination")
	chatMessagesListCmd.Flags().StringVar(&chatMessageFilter, "filter", "", "Filter for messages")
	chatMessagesListCmd.Flags().BoolVar(&chatListPaginate, "paginate", false, "Automatically fetch all pages")

	chatMessagesCreateCmd.Flags().StringVar(&chatMessageText, "text", "", "Message text (required)")
	chatMessagesCreateCmd.Flags().StringVar(&chatMessageThreadID, "thread", "", "Thread ID to reply in")

	chatMessagesUpdateCmd.Flags().StringVar(&chatMessageText, "text", "", "New message text (required)")

	// Members flags
	chatMembersListCmd.Flags().IntVar(&chatListLimit, "limit", 100, "Maximum results per page")
	chatMembersListCmd.Flags().StringVar(&chatListPageToken, "page-token", "", "Page token for pagination")
	chatMembersListCmd.Flags().BoolVar(&chatListPaginate, "paginate", false, "Automatically fetch all pages")

	chatMembersCreateCmd.Flags().StringVar(&chatMemberEmail, "email", "", "Member email (required)")
	chatMembersCreateCmd.Flags().StringVar(&chatMemberRole, "role", "MEMBER", "Member role (MEMBER or MANAGER)")

	// Add subcommands
	chatSpacesCmd.AddCommand(chatSpacesListCmd)
	chatSpacesCmd.AddCommand(chatSpacesGetCmd)
	chatSpacesCmd.AddCommand(chatSpacesCreateCmd)
	chatSpacesCmd.AddCommand(chatSpacesDeleteCmd)

	chatMessagesCmd.AddCommand(chatMessagesListCmd)
	chatMessagesCmd.AddCommand(chatMessagesGetCmd)
	chatMessagesCmd.AddCommand(chatMessagesCreateCmd)
	chatMessagesCmd.AddCommand(chatMessagesUpdateCmd)
	chatMessagesCmd.AddCommand(chatMessagesDeleteCmd)

	chatMembersCmd.AddCommand(chatMembersListCmd)
	chatMembersCmd.AddCommand(chatMembersGetCmd)
	chatMembersCmd.AddCommand(chatMembersCreateCmd)
	chatMembersCmd.AddCommand(chatMembersDeleteCmd)

	chatCmd.AddCommand(chatSpacesCmd)
	chatCmd.AddCommand(chatMessagesCmd)
	chatCmd.AddCommand(chatMembersCmd)

	rootCmd.AddCommand(chatCmd)
}

// Spaces handlers

func runChatSpacesList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if chatListPaginate {
		var allSpaces []types.ChatSpace
		pageToken := ""
		for {
			result, err := mgr.ListSpaces(ctx, reqCtx, chatListLimit, pageToken)
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

	result, err := mgr.ListSpaces(ctx, reqCtx, chatListLimit, chatListPageToken)
	if err != nil {
		return handleCLIError(out, "chat.spaces.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.spaces.list", result.Spaces)
	}
	return out.WriteSuccess("chat.spaces.list", result)
}

func runChatSpacesGet(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetSpace(ctx, reqCtx, args[0])
	if err != nil {
		return handleCLIError(out, "chat.spaces.get", err)
	}

	return out.WriteSuccess("chat.spaces.get", result)
}

func runChatSpacesCreate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if chatSpaceDisplayName == "" {
		return out.WriteError("chat.spaces.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--display-name is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateSpace(ctx, reqCtx, chatSpaceDisplayName, chatSpaceType, chatSpaceExternalUsers)
	if err != nil {
		return handleCLIError(out, "chat.spaces.create", err)
	}

	return out.WriteSuccess("chat.spaces.create", result)
}

func runChatSpacesDelete(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.spaces.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteSpace(ctx, reqCtx, args[0]); err != nil {
		return handleCLIError(out, "chat.spaces.delete", err)
	}

	return out.WriteSuccess("chat.spaces.delete", map[string]string{
		"spaceId": args[0],
		"status":  "deleted",
	})
}

// Messages handlers

func runChatMessagesList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if chatListPaginate {
		var allMessages []types.ChatMessage
		pageToken := ""
		for {
			result, err := mgr.ListMessages(ctx, reqCtx, args[0], chatListLimit, pageToken, chatMessageFilter)
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

	result, err := mgr.ListMessages(ctx, reqCtx, args[0], chatListLimit, chatListPageToken, chatMessageFilter)
	if err != nil {
		return handleCLIError(out, "chat.messages.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.messages.list", result.Messages)
	}
	return out.WriteSuccess("chat.messages.list", result)
}

func runChatMessagesGet(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetMessage(ctx, reqCtx, args[0], args[1])
	if err != nil {
		return handleCLIError(out, "chat.messages.get", err)
	}

	return out.WriteSuccess("chat.messages.get", result)
}

func runChatMessagesCreate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if chatMessageText == "" {
		return out.WriteError("chat.messages.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--text is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateMessage(ctx, reqCtx, args[0], chatMessageText, chatMessageThreadID)
	if err != nil {
		return handleCLIError(out, "chat.messages.create", err)
	}

	return out.WriteSuccess("chat.messages.create", result)
}

func runChatMessagesUpdate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if chatMessageText == "" {
		return out.WriteError("chat.messages.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--text is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.UpdateMessage(ctx, reqCtx, args[0], args[1], chatMessageText)
	if err != nil {
		return handleCLIError(out, "chat.messages.update", err)
	}

	return out.WriteSuccess("chat.messages.update", result)
}

func runChatMessagesDelete(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.messages.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteMessage(ctx, reqCtx, args[0], args[1]); err != nil {
		return handleCLIError(out, "chat.messages.delete", err)
	}

	return out.WriteSuccess("chat.messages.delete", map[string]string{
		"spaceId":   args[0],
		"messageId": args[1],
		"status":    "deleted",
	})
}

// Members handlers

func runChatMembersList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if chatListPaginate {
		var allMembers []types.ChatMember
		pageToken := ""
		for {
			result, err := mgr.ListMembers(ctx, reqCtx, args[0], chatListLimit, pageToken)
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

	result, err := mgr.ListMembers(ctx, reqCtx, args[0], chatListLimit, chatListPageToken)
	if err != nil {
		return handleCLIError(out, "chat.members.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("chat.members.list", result.Members)
	}
	return out.WriteSuccess("chat.members.list", result)
}

func runChatMembersGet(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetMember(ctx, reqCtx, args[0], args[1])
	if err != nil {
		return handleCLIError(out, "chat.members.get", err)
	}

	return out.WriteSuccess("chat.members.get", result)
}

func runChatMembersCreate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	if chatMemberEmail == "" {
		return out.WriteError("chat.members.create", utils.NewCLIError(utils.ErrCodeInvalidArgument, "--email is required").Build())
	}

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	result, err := mgr.CreateMember(ctx, reqCtx, args[0], chatMemberEmail, chatMemberRole)
	if err != nil {
		return handleCLIError(out, "chat.members.create", err)
	}

	return out.WriteSuccess("chat.members.create", result)
}

func runChatMembersDelete(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	ctx := context.Background()

	svc, client, reqCtx, err := getChatService(ctx, flags)
	if err != nil {
		return out.WriteError("chat.members.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	mgr := chatmgr.NewManager(client, svc)
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteMember(ctx, reqCtx, args[0], args[1]); err != nil {
		return handleCLIError(out, "chat.members.delete", err)
	}

	return out.WriteSuccess("chat.members.delete", map[string]string{
		"spaceId":  args[0],
		"memberId": args[1],
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
