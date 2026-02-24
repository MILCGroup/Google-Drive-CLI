package cli

import (
	"context"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	"github.com/milcgroup/gdrv/internal/permissions"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
)

// PermissionsCmd is the top-level permissions command group.
type PermissionsCmd struct {
	List       PermListCmd       `cmd:"" help:"List permissions"`
	Create     PermCreateCmd     `cmd:"" help:"Create a permission"`
	Update     PermUpdateCmd     `cmd:"" help:"Update a permission"`
	Remove     PermRemoveCmd     `cmd:"" help:"Remove a permission"`
	CreateLink PermCreateLinkCmd `cmd:"create-link" help:"Create a public link"`
	Audit      PermAuditCmd      `cmd:"" help:"Audit permissions"`
	Analyze    PermAnalyzeCmd    `cmd:"" help:"Analyze folder permissions"`
	Report     PermReportCmd     `cmd:"" help:"Generate permission report"`
	Bulk       PermBulkCmd       `cmd:"" help:"Bulk permission operations"`
	Search     PermSearchCmd     `cmd:"" help:"Search permissions"`
}

type PermAuditCmd struct {
	Public         PermAuditPublicCmd         `cmd:"" help:"Audit public files"`
	External       PermAuditExternalCmd       `cmd:"" help:"Audit external shares"`
	AnyoneWithLink PermAuditAnyoneWithLinkCmd `cmd:"anyone-with-link" help:"Audit anyone-with-link files"`
	User           PermAuditUserCmd           `cmd:"" help:"Audit user access"`
}

type PermBulkCmd struct {
	RemovePublic PermBulkRemovePublicCmd `cmd:"remove-public" help:"Bulk remove public access"`
	UpdateRole   PermBulkUpdateRoleCmd   `cmd:"update-role" help:"Bulk change role"`
}

// --- Leaf command structs ---

type PermListCmd struct {
	FileID string `arg:"" name:"file-id" help:"File ID or path"`
}

type PermCreateCmd struct {
	FileID            string `arg:"" name:"file-id" help:"File ID or path"`
	Type              string `help:"Permission type (user, group, domain, anyone)" name:"type" required:""`
	Role              string `help:"Permission role (reader, commenter, writer, organizer)" name:"role" required:""`
	Email             string `help:"Email address (for user/group type)" name:"email"`
	Domain            string `help:"Domain (for domain type)" name:"domain"`
	SendNotification  bool   `help:"Send email notification" name:"send-notification" default:"true"`
	Message           string `help:"Custom email message" name:"message"`
	TransferOwnership bool   `help:"Transfer ownership (requires owner role)" name:"transfer-ownership"`
	AllowDiscovery    bool   `help:"Allow file discovery (for anyone type)" name:"allow-discovery"`
}

type PermUpdateCmd struct {
	FileID       string `arg:"" name:"file-id" help:"File ID"`
	PermissionID string `arg:"" name:"permission-id" help:"Permission ID"`
	Role         string `help:"New role" name:"role" required:""`
}

type PermRemoveCmd struct {
	FileID       string `arg:"" name:"file-id" help:"File ID"`
	PermissionID string `arg:"" name:"permission-id" help:"Permission ID"`
}

type PermCreateLinkCmd struct {
	FileID         string `arg:"" name:"file-id" help:"File ID or path"`
	Role           string `help:"Permission role (reader, commenter, writer)" name:"role" default:"reader"`
	AllowDiscovery bool   `help:"Allow file discovery in search" name:"allow-discovery"`
}

type PermAuditPublicCmd struct {
	FolderID           string `help:"Limit audit to specific folder" name:"folder-id"`
	Recursive          bool   `help:"Include subfolders" name:"recursive"`
	IncludePermissions bool   `help:"Include full permission details" name:"include-permissions"`
}

type PermAuditExternalCmd struct {
	FolderID           string `help:"Limit audit to specific folder" name:"folder-id"`
	Recursive          bool   `help:"Include subfolders" name:"recursive"`
	InternalDomain     string `help:"Internal domain (required)" name:"internal-domain" required:""`
	IncludePermissions bool   `help:"Include full permission details" name:"include-permissions"`
}

type PermAuditAnyoneWithLinkCmd struct {
	FolderID           string `help:"Limit audit to specific folder" name:"folder-id"`
	Recursive          bool   `help:"Include subfolders" name:"recursive"`
	IncludePermissions bool   `help:"Include full permission details" name:"include-permissions"`
}

type PermAuditUserCmd struct {
	Email              string `arg:"" name:"email" help:"User email address"`
	FolderID           string `help:"Limit audit to specific folder" name:"folder-id"`
	Recursive          bool   `help:"Include subfolders" name:"recursive"`
	IncludePermissions bool   `help:"Include full permission details" name:"include-permissions"`
}

type PermAnalyzeCmd struct {
	FolderID       string `arg:"" name:"folder-id" help:"Folder ID"`
	Recursive      bool   `help:"Analyze subfolders recursively" name:"recursive"`
	MaxDepth       int    `help:"Maximum recursion depth (0 = unlimited)" name:"max-depth"`
	IncludeDetails bool   `help:"Include detailed file lists" name:"include-details"`
	InternalDomain string `help:"Internal domain for external detection" name:"internal-domain"`
}

type PermReportCmd struct {
	FileID         string `arg:"" name:"file-id" help:"File ID or folder ID"`
	InternalDomain string `help:"Internal domain for external detection" name:"internal-domain"`
}

type PermBulkRemovePublicCmd struct {
	FolderID        string `help:"Folder to operate on (required)" name:"folder-id" required:""`
	Recursive       bool   `help:"Include subfolders" name:"recursive"`
	MaxFiles        int    `help:"Maximum files to process (0 = unlimited)" name:"max-files"`
	ContinueOnError bool   `help:"Continue if individual operations fail" name:"continue-on-error"`
}

type PermBulkUpdateRoleCmd struct {
	FolderID        string `help:"Folder to operate on (required)" name:"folder-id" required:""`
	Recursive       bool   `help:"Include subfolders" name:"recursive"`
	FromRole        string `help:"Source role (required)" name:"from-role" required:""`
	ToRole          string `help:"Target role (required)" name:"to-role" required:""`
	MaxFiles        int    `help:"Maximum files to process (0 = unlimited)" name:"max-files"`
	ContinueOnError bool   `help:"Continue if individual operations fail" name:"continue-on-error"`
}

type PermSearchCmd struct {
	Email     string `help:"Search by email address" name:"email"`
	Role      string `help:"Search by role" name:"role"`
	FolderID  string `help:"Limit search to specific folder" name:"folder-id"`
	Recursive bool   `help:"Include subfolders" name:"recursive"`
}

// --- Helper ---

func getPermissionManager(flags types.GlobalFlags) (*permissions.Manager, error) {
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
	return permissions.NewManager(client), nil
}

// --- Run methods ---

func (cmd *PermListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permission.list", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	result, err := mgr.List(context.Background(), reqCtx, cmd.FileID, permissions.ListOptions{})
	if err != nil {
		return handleCLIError(writer, "permission.list", err)
	}

	return writer.WriteSuccess("permission.list", result)
}

func (cmd *PermCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Validate type
	validTypes := map[string]bool{"user": true, "group": true, "domain": true, "anyone": true}
	if !validTypes[cmd.Type] {
		return writer.WriteError("permissions.create", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Invalid permission type. Must be one of: user, group, domain, anyone").Build())
	}

	// Validate role
	validRoles := map[string]bool{"reader": true, "commenter": true, "writer": true, "organizer": true, "owner": true}
	if !validRoles[cmd.Role] {
		return writer.WriteError("permissions.create", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Invalid permission role. Must be one of: reader, commenter, writer, organizer, owner").Build())
	}

	// Validate email for user/group type
	if (cmd.Type == "user" || cmd.Type == "group") && cmd.Email == "" {
		return writer.WriteError("permissions.create", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Email address is required for user or group permission type").Build())
	}

	// Validate domain for domain type
	if cmd.Type == "domain" && cmd.Domain == "" {
		return writer.WriteError("permissions.create", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Domain is required for domain permission type").Build())
	}

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.create", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	opts := permissions.CreateOptions{
		Type:                  cmd.Type,
		Role:                  cmd.Role,
		EmailAddress:          cmd.Email,
		Domain:                cmd.Domain,
		SendNotificationEmail: cmd.SendNotification,
		EmailMessage:          cmd.Message,
		TransferOwnership:     cmd.TransferOwnership,
		AllowFileDiscovery:    cmd.AllowDiscovery,
	}

	result, err := mgr.Create(context.Background(), reqCtx, cmd.FileID, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.create", err)
	}

	return writer.WriteSuccess("permissions.create", result)
}

func (cmd *PermUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	// Validate role
	validRoles := map[string]bool{"reader": true, "commenter": true, "writer": true, "organizer": true, "owner": true}
	if !validRoles[cmd.Role] {
		return writer.WriteError("permission.update", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Invalid permission role. Must be one of: reader, commenter, writer, organizer, owner").Build())
	}

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permission.update", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	result, err := mgr.Update(context.Background(), reqCtx, cmd.FileID, cmd.PermissionID, permissions.UpdateOptions{Role: cmd.Role})
	if err != nil {
		return handleCLIError(writer, "permission.update", err)
	}

	return writer.WriteSuccess("permission.update", result)
}

func (cmd *PermRemoveCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permission.remove", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	err = mgr.Delete(context.Background(), reqCtx, cmd.FileID, cmd.PermissionID, permissions.DeleteOptions{})
	if err != nil {
		return handleCLIError(writer, "permission.remove", err)
	}

	return writer.WriteSuccess("permission.remove", map[string]interface{}{
		"deleted":      true,
		"fileId":       cmd.FileID,
		"permissionId": cmd.PermissionID,
	})
}

func (cmd *PermCreateLinkCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	validRoles := map[string]bool{"reader": true, "commenter": true, "writer": true}
	if !validRoles[cmd.Role] {
		return writer.WriteError("permission.create-link", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Invalid permission role for public link. Must be one of: reader, commenter, writer").Build())
	}

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permission.create-link", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	result, err := mgr.CreatePublicLink(context.Background(), reqCtx, cmd.FileID, cmd.Role, cmd.AllowDiscovery)
	if err != nil {
		return handleCLIError(writer, "permission.create-link", err)
	}

	return writer.WriteSuccess("permission.create-link", result)
}

func (cmd *PermAuditPublicCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.public", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.AuditOptions{
		FolderID:           cmd.FolderID,
		Recursive:          cmd.Recursive,
		IncludePermissions: cmd.IncludePermissions,
	}

	result, err := mgr.AuditPublic(context.Background(), reqCtx, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.public", err)
	}

	return writer.WriteSuccess("permissions.audit.public", result)
}

func (cmd *PermAuditExternalCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.external", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.AuditOptions{
		FolderID:           cmd.FolderID,
		Recursive:          cmd.Recursive,
		InternalDomain:     cmd.InternalDomain,
		IncludePermissions: cmd.IncludePermissions,
	}

	result, err := mgr.AuditExternal(context.Background(), reqCtx, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.external", err)
	}

	return writer.WriteSuccess("permissions.audit.external", result)
}

func (cmd *PermAuditAnyoneWithLinkCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.anyone-with-link", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.AuditOptions{
		FolderID:           cmd.FolderID,
		Recursive:          cmd.Recursive,
		IncludePermissions: cmd.IncludePermissions,
	}

	result, err := mgr.AuditAnyoneWithLink(context.Background(), reqCtx, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.anyone-with-link", err)
	}

	return writer.WriteSuccess("permissions.audit.anyone-with-link", result)
}

func (cmd *PermAuditUserCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.user", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.AuditOptions{
		FolderID:           cmd.FolderID,
		Recursive:          cmd.Recursive,
		IncludePermissions: cmd.IncludePermissions,
	}

	result, err := mgr.AuditUser(context.Background(), reqCtx, cmd.Email, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.audit.user", err)
	}

	return writer.WriteSuccess("permissions.audit.user", result)
}

func (cmd *PermAnalyzeCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.analyze", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.AnalyzeOptions{
		Recursive:      cmd.Recursive,
		MaxDepth:       cmd.MaxDepth,
		IncludeDetails: cmd.IncludeDetails,
		InternalDomain: cmd.InternalDomain,
	}

	result, err := mgr.AnalyzeFolder(context.Background(), reqCtx, cmd.FolderID, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.analyze", err)
	}

	return writer.WriteSuccess("permissions.analyze", result)
}

func (cmd *PermReportCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.report", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)

	result, err := mgr.GenerateReport(context.Background(), reqCtx, cmd.FileID, cmd.InternalDomain)
	if err != nil {
		return handleCLIError(writer, "permissions.report", err)
	}

	return writer.WriteSuccess("permissions.report", result)
}

func (cmd *PermBulkRemovePublicCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.bulk.remove-public", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.BulkOptions{
		FolderID:        cmd.FolderID,
		Recursive:       cmd.Recursive,
		DryRun:          flags.DryRun,
		MaxFiles:        cmd.MaxFiles,
		ContinueOnError: cmd.ContinueOnError,
	}

	result, err := mgr.BulkRemovePublic(context.Background(), reqCtx, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.bulk.remove-public", err)
	}

	return writer.WriteSuccess("permissions.bulk.remove-public", result)
}

func (cmd *PermBulkUpdateRoleCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.bulk.update-role", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	opts := types.BulkOptions{
		FolderID:        cmd.FolderID,
		Recursive:       cmd.Recursive,
		DryRun:          flags.DryRun,
		MaxFiles:        cmd.MaxFiles,
		ContinueOnError: cmd.ContinueOnError,
	}

	result, err := mgr.BulkUpdateRole(context.Background(), reqCtx, cmd.FromRole, cmd.ToRole, opts)
	if err != nil {
		return handleCLIError(writer, "permissions.bulk.update-role", err)
	}

	return writer.WriteSuccess("permissions.bulk.update-role", result)
}

func (cmd *PermSearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	writer := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	if cmd.Email == "" && cmd.Role == "" {
		return writer.WriteError("permissions.search", utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Either --email or --role must be specified").Build())
	}

	mgr, err := getPermissionManager(flags)
	if err != nil {
		return handleCLIError(writer, "permissions.search", err)
	}

	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypePermissionOp)
	searchOpts := types.SearchOptions{
		Email:     cmd.Email,
		Role:      cmd.Role,
		FolderID:  cmd.FolderID,
		Recursive: cmd.Recursive,
	}

	var result *types.AuditResult
	if cmd.Email != "" {
		result, err = mgr.SearchByEmail(context.Background(), reqCtx, searchOpts)
	} else {
		result, err = mgr.SearchByRole(context.Background(), reqCtx, searchOpts)
	}

	if err != nil {
		return handleCLIError(writer, "permissions.search", err)
	}

	return writer.WriteSuccess("permissions.search", result)
}
