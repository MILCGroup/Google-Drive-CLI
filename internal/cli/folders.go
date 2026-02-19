package cli

import (
	"context"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/folders"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	"github.com/spf13/cobra"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Folder operations",
	Long:  "Commands for managing folders in Google Drive",
}

var folderCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new folder",
	Long:  "Create a new folder in Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderCreate,
}

var folderListCmd = &cobra.Command{
	Use:   "list <folder-id>",
	Short: "List folder contents",
	Long:  "List contents of a folder in Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderList,
}

var folderDeleteCmd = &cobra.Command{
	Use:   "delete <folder-id>",
	Short: "Delete a folder",
	Long:  "Delete a folder from Google Drive",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderDelete,
}

var folderMoveCmd = &cobra.Command{
	Use:   "move <folder-id> <new-parent-id>",
	Short: "Move a folder",
	Long:  "Move a folder to a new parent folder",
	Args:  cobra.ExactArgs(2),
	RunE:  runFolderMove,
}

var folderGetCmd = &cobra.Command{
	Use:   "get <folder-id>",
	Short: "Get folder metadata",
	Long:  "Retrieve metadata for a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderGet,
}

// Flags
var (
	folderParentID  string
	folderRecursive bool
	folderPageSize  int
	folderPageToken string
	folderFields    string
	folderPaginate  bool
)

func init() {
	rootCmd.AddCommand(foldersCmd)

	foldersCmd.AddCommand(folderCreateCmd)
	foldersCmd.AddCommand(folderListCmd)
	foldersCmd.AddCommand(folderDeleteCmd)
	foldersCmd.AddCommand(folderMoveCmd)
	foldersCmd.AddCommand(folderGetCmd)

	// Create flags
	folderCreateCmd.Flags().StringVar(&folderParentID, "parent", "", "Parent folder ID")

	// List flags
	folderListCmd.Flags().IntVar(&folderPageSize, "page-size", 100, "Number of items per page")
	folderListCmd.Flags().StringVar(&folderPageToken, "page-token", "", "Page token for pagination")
	folderListCmd.Flags().BoolVar(&folderPaginate, "paginate", false, "Automatically fetch all pages")

	// Delete flags
	folderDeleteCmd.Flags().BoolVar(&folderRecursive, "recursive", false, "Delete folder contents recursively")

	// Get flags
	folderGetCmd.Flags().StringVar(&folderFields, "fields", "", "Fields to retrieve (comma-separated)")
}

func getFolderManager() (*folders.Manager, error) {
	flags := GetGlobalFlags()

	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)
	creds, err := authMgr.LoadCredentials(flags.Profile)
	if err != nil {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeAuthRequired,
			"Authentication required. Run 'gdrv auth login' first.").Build())
	}

	service, err := authMgr.GetDriveService(context.Background(), creds)
	if err != nil {
		return nil, err
	}

	client := api.NewClient(service, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	return folders.NewManager(client), nil
}

func runFolderCreate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager()
	if err != nil {
		return handleCLIError(writer, "folder.create", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	name := args[0]

	result, err := mgr.Create(context.Background(), reqCtx, name, folderParentID)
	if err != nil {
		return handleCLIError(writer, "folder.create", err)
	}

	return writer.WriteSuccess("folder.create", result)
}

func runFolderList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager()
	if err != nil {
		return handleCLIError(writer, "folder.list", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	folderID := args[0]

	// If --paginate flag is set, fetch all pages
	if folderPaginate {
		var allFiles []*types.DriveFile
		pageToken := folderPageToken
		for {
			result, err := mgr.List(context.Background(), reqCtx, folderID, folderPageSize, pageToken)
			if err != nil {
				return handleCLIError(writer, "folder.list", err)
			}
			allFiles = append(allFiles, result.Files...)
			if result.NextPageToken == "" {
				break
			}
			pageToken = result.NextPageToken
		}
		return writer.WriteSuccess("folder.list", map[string]interface{}{
			"files": allFiles,
		})
	}

	result, err := mgr.List(context.Background(), reqCtx, folderID, folderPageSize, folderPageToken)
	if err != nil {
		return handleCLIError(writer, "folder.list", err)
	}

	return writer.WriteSuccess("folder.list", result)
}

func runFolderDelete(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager()
	if err != nil {
		return handleCLIError(writer, "folder.delete", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	folderID := args[0]

	err = mgr.Delete(context.Background(), reqCtx, folderID, folderRecursive)
	if err != nil {
		return handleCLIError(writer, "folder.delete", err)
	}

	return writer.WriteSuccess("folder.delete", map[string]interface{}{
		"deleted":   true,
		"folderId":  folderID,
		"recursive": folderRecursive,
	})
}

func runFolderMove(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager()
	if err != nil {
		return handleCLIError(writer, "folder.move", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeMutation)
	folderID := args[0]
	newParentID := args[1]

	result, err := mgr.Move(context.Background(), reqCtx, folderID, newParentID)
	if err != nil {
		return handleCLIError(writer, "folder.move", err)
	}

	return writer.WriteSuccess("folder.move", result)
}

func runFolderGet(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getFolderManager()
	if err != nil {
		return handleCLIError(writer, "folder.get", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeGetByID)
	folderID := args[0]

	result, err := mgr.Get(context.Background(), reqCtx, folderID, folderFields)
	if err != nil {
		return handleCLIError(writer, "folder.get", err)
	}

	return writer.WriteSuccess("folder.get", result)
}
