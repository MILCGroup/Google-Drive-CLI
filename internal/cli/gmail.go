package cli

import (
	"context"
	"os"

	"github.com/milcgroup/gdrv/internal/api"
	"github.com/milcgroup/gdrv/internal/auth"
	gmailmgr "github.com/milcgroup/gdrv/internal/gmail"
	"github.com/milcgroup/gdrv/internal/types"
	"github.com/milcgroup/gdrv/internal/utils"
	gmailapi "google.golang.org/api/gmail/v1"
)

// ============================================================
// Gmail CLI commands
// ============================================================

type GmailCmd struct {
	Search      GmailSearchCmd      `cmd:"" help:"Search messages"`
	Get         GmailGetCmd         `cmd:"" help:"Get a message"`
	Thread      GmailThreadCmd      `cmd:"" help:"Get a thread"`
	Send        GmailSendCmd        `cmd:"" help:"Send a message"`
	Drafts      GmailDraftsCmd      `cmd:"" help:"Draft operations"`
	Labels      GmailLabelsCmd      `cmd:"" help:"Label operations"`
	Filters     GmailFiltersCmd     `cmd:"" help:"Filter operations"`
	Vacation    GmailVacationCmd    `cmd:"" help:"Vacation responder settings"`
	SendAs      GmailSendAsCmd      `cmd:"send-as" help:"Send-as alias operations"`
	BatchDelete GmailBatchDeleteCmd `cmd:"batch-delete" help:"Batch delete messages"`
	BatchModify GmailBatchModifyCmd `cmd:"batch-modify" help:"Batch modify message labels"`
	Attachment  GmailAttachmentCmd  `cmd:"" help:"Attachment operations"`
}

// --- Top-level command structs ---

type GmailSearchCmd struct {
	Query     string `help:"Gmail search query" name:"query" required:""`
	Limit     int    `help:"Maximum messages to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type GmailGetCmd struct {
	MessageID string `arg:"" name:"message-id" help:"Message ID"`
	Format    string `help:"Response format (full, metadata, minimal)" default:"full" name:"format"`
}

type GmailThreadCmd struct {
	ThreadID string `arg:"" name:"thread-id" help:"Thread ID"`
	Format   string `help:"Response format (full, metadata, minimal)" default:"full" name:"format"`
}

type GmailSendCmd struct {
	To      string `help:"Recipient email address" name:"to" required:""`
	Subject string `help:"Email subject" name:"subject" required:""`
	Body    string `help:"Plain text body" name:"body"`
	HTML    string `help:"HTML body" name:"html"`
	Cc      string `help:"CC recipients" name:"cc"`
	Bcc     string `help:"BCC recipients" name:"bcc"`
	ReplyTo string `help:"Message-ID to reply to (In-Reply-To header)" name:"reply-to"`
	Thread  string `help:"Thread ID for threading replies" name:"thread"`
}

// --- Drafts subcommands ---

type GmailDraftsCmd struct {
	List   GmailDraftsListCmd   `cmd:"" help:"List drafts"`
	Get    GmailDraftsGetCmd    `cmd:"" help:"Get a draft"`
	Create GmailDraftsCreateCmd `cmd:"" help:"Create a draft"`
	Update GmailDraftsUpdateCmd `cmd:"" help:"Update a draft"`
	Delete GmailDraftsDeleteCmd `cmd:"" help:"Delete a draft"`
	Send   GmailDraftsSendCmd   `cmd:"" help:"Send a draft"`
}

type GmailDraftsListCmd struct {
	Limit     int    `help:"Maximum drafts to return per page" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type GmailDraftsGetCmd struct {
	DraftID string `arg:"" name:"draft-id" help:"Draft ID"`
}

type GmailDraftsCreateCmd struct {
	To      string `help:"Recipient email address" name:"to" required:""`
	Subject string `help:"Email subject" name:"subject" required:""`
	Body    string `help:"Plain text body" name:"body"`
}

type GmailDraftsUpdateCmd struct {
	DraftID string `arg:"" name:"draft-id" help:"Draft ID"`
	To      string `help:"Recipient email address" name:"to" required:""`
	Subject string `help:"Email subject" name:"subject" required:""`
	Body    string `help:"Plain text body" name:"body"`
}

type GmailDraftsDeleteCmd struct {
	DraftID string `arg:"" name:"draft-id" help:"Draft ID"`
}

type GmailDraftsSendCmd struct {
	DraftID string `arg:"" name:"draft-id" help:"Draft ID"`
}

// --- Labels subcommands ---

type GmailLabelsCmd struct {
	List   GmailLabelsListCmd   `cmd:"" help:"List labels"`
	Create GmailLabelsCreateCmd `cmd:"" help:"Create a label"`
	Delete GmailLabelsDeleteCmd `cmd:"" help:"Delete a label"`
}

type GmailLabelsListCmd struct{}

type GmailLabelsCreateCmd struct {
	Name string `help:"Label name" name:"name" required:""`
}

type GmailLabelsDeleteCmd struct {
	LabelID string `arg:"" name:"label-id" help:"Label ID"`
}

// --- Filters subcommands ---

type GmailFiltersCmd struct {
	List   GmailFiltersListCmd   `cmd:"" help:"List filters"`
	Create GmailFiltersCreateCmd `cmd:"" help:"Create a filter"`
	Delete GmailFiltersDeleteCmd `cmd:"" help:"Delete a filter"`
}

type GmailFiltersListCmd struct{}

type GmailFiltersCreateCmd struct {
	From          string `help:"Match sender" name:"from"`
	To            string `help:"Match recipient" name:"to"`
	Subject       string `help:"Match subject" name:"subject"`
	Query         string `help:"Match Gmail search query" name:"query"`
	HasAttachment bool   `help:"Match messages with attachments" name:"has-attachment"`
	AddLabels     string `help:"Comma-separated label IDs to add" name:"add-labels"`
	RemoveLabels  string `help:"Comma-separated label IDs to remove" name:"remove-labels"`
	Forward       string `help:"Email address to forward to" name:"forward"`
}

type GmailFiltersDeleteCmd struct {
	FilterID string `arg:"" name:"filter-id" help:"Filter ID"`
}

// --- Vacation subcommands ---

type GmailVacationCmd struct {
	Get GmailVacationGetCmd `cmd:"" help:"Get vacation responder settings"`
	Set GmailVacationSetCmd `cmd:"" help:"Set vacation responder settings"`
}

type GmailVacationGetCmd struct{}

type GmailVacationSetCmd struct {
	Enable   bool   `help:"Enable auto-reply" name:"enable"`
	Subject  string `help:"Response subject" name:"subject"`
	Body     string `help:"Response body (plain text)" name:"body"`
	HTMLBody string `help:"Response body (HTML)" name:"html-body"`
	Start    int64  `help:"Start time (Unix milliseconds)" name:"start"`
	End      int64  `help:"End time (Unix milliseconds)" name:"end"`
}

// --- SendAs subcommands ---

type GmailSendAsCmd struct {
	List GmailSendAsListCmd `cmd:"" help:"List send-as aliases"`
}

type GmailSendAsListCmd struct{}

// --- Batch commands ---

type GmailBatchDeleteCmd struct {
	IDs string `help:"Comma-separated message IDs" name:"ids" required:""`
}

type GmailBatchModifyCmd struct {
	IDs          string `help:"Comma-separated message IDs" name:"ids" required:""`
	AddLabels    string `help:"Comma-separated label IDs to add" name:"add-labels"`
	RemoveLabels string `help:"Comma-separated label IDs to remove" name:"remove-labels"`
}

// --- Attachment subcommands ---

type GmailAttachmentCmd struct {
	Get GmailAttachmentGetCmd `cmd:"" help:"Download an attachment"`
}

type GmailAttachmentGetCmd struct {
	MessageID    string `arg:"" name:"message-id" help:"Message ID"`
	AttachmentID string `help:"Attachment ID" name:"attachment-id" required:""`
	OutputPath   string `help:"Output file path" name:"file-output"`
}

// ============================================================
// Helper: getGmailService
// ============================================================

func getGmailService(ctx context.Context, flags types.GlobalFlags) (*gmailapi.Service, *api.Client, *types.RequestContext, error) {
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServiceGmail); err != nil {
		return nil, nil, nil, err
	}

	svc, err := authMgr.GetGmailService(ctx, creds)
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

// getGmailManager creates a Gmail manager from the service helper.
func getGmailManager(ctx context.Context, flags types.GlobalFlags) (*gmailmgr.Manager, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)

	svc, client, reqCtx, err := getGmailService(ctx, flags)
	if err != nil {
		return nil, nil, out, err
	}

	mgr := gmailmgr.NewManager(client, svc)
	return mgr, reqCtx, out, nil
}

// ============================================================
// Run methods
// ============================================================

func (cmd *GmailSearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.search", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allMessages []types.GmailMessage
		pageToken := cmd.PageToken
		for {
			result, nextToken, err := mgr.Search(ctx, reqCtx, cmd.Query, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "gmail.search", err)
			}
			allMessages = append(allMessages, result.Messages...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("gmail.search", &types.GmailMessageList{
			Messages:           allMessages,
			ResultSizeEstimate: int64(len(allMessages)),
		})
	}

	result, nextPageToken, err := mgr.Search(ctx, reqCtx, cmd.Query, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "gmail.search", err)
	}

	response := map[string]interface{}{
		"messages":           result.Messages,
		"resultSizeEstimate": result.ResultSizeEstimate,
	}
	if nextPageToken != "" {
		response["nextPageToken"] = nextPageToken
	}
	return out.WriteSuccess("gmail.search", response)
}

func (cmd *GmailGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetMessage(ctx, reqCtx, cmd.MessageID, cmd.Format)
	if err != nil {
		return handleCLIError(out, "gmail.get", err)
	}

	return out.WriteSuccess("gmail.get", result)
}

func (cmd *GmailThreadCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.thread", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetThread(ctx, reqCtx, cmd.ThreadID, cmd.Format)
	if err != nil {
		return handleCLIError(out, "gmail.thread", err)
	}

	return out.WriteSuccess("gmail.thread", result)
}

func (cmd *GmailSendCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.send", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.Send(ctx, reqCtx, cmd.To, cmd.Subject, cmd.Body, cmd.Cc, cmd.Bcc, cmd.HTML, cmd.ReplyTo, cmd.Thread)
	if err != nil {
		return handleCLIError(out, "gmail.send", err)
	}

	out.Log("Sent message: %s", result.ID)
	return out.WriteSuccess("gmail.send", result)
}

// --- Drafts ---

func (cmd *GmailDraftsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allDrafts []types.GmailDraft
		pageToken := cmd.PageToken
		for {
			result, nextToken, err := mgr.ListDrafts(ctx, reqCtx, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "gmail.drafts.list", err)
			}
			allDrafts = append(allDrafts, result.Drafts...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		return out.WriteSuccess("gmail.drafts.list", &types.GmailDraftList{
			Drafts: allDrafts,
		})
	}

	result, nextPageToken, err := mgr.ListDrafts(ctx, reqCtx, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.list", err)
	}

	response := map[string]interface{}{
		"drafts": result.Drafts,
	}
	if nextPageToken != "" {
		response["nextPageToken"] = nextPageToken
	}
	return out.WriteSuccess("gmail.drafts.list", response)
}

func (cmd *GmailDraftsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetDraft(ctx, reqCtx, cmd.DraftID)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.get", err)
	}

	return out.WriteSuccess("gmail.drafts.get", result)
}

func (cmd *GmailDraftsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateDraft(ctx, reqCtx, cmd.To, cmd.Subject, cmd.Body)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.create", err)
	}

	out.Log("Created draft: %s", result.ID)
	return out.WriteSuccess("gmail.drafts.create", result)
}

func (cmd *GmailDraftsUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.UpdateDraft(ctx, reqCtx, cmd.DraftID, cmd.To, cmd.Subject, cmd.Body)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.update", err)
	}

	out.Log("Updated draft: %s", result.ID)
	return out.WriteSuccess("gmail.drafts.update", result)
}

func (cmd *GmailDraftsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	err = mgr.DeleteDraft(ctx, reqCtx, cmd.DraftID)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.delete", err)
	}

	out.Log("Deleted draft: %s", cmd.DraftID)
	return out.WriteSuccess("gmail.drafts.delete", map[string]string{"id": cmd.DraftID, "status": "deleted"})
}

func (cmd *GmailDraftsSendCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.drafts.send", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.SendDraft(ctx, reqCtx, cmd.DraftID)
	if err != nil {
		return handleCLIError(out, "gmail.drafts.send", err)
	}

	out.Log("Sent draft: %s", result.ID)
	return out.WriteSuccess("gmail.drafts.send", result)
}

// --- Labels ---

func (cmd *GmailLabelsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.labels.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListLabels(ctx, reqCtx)
	if err != nil {
		return handleCLIError(out, "gmail.labels.list", err)
	}

	return out.WriteSuccess("gmail.labels.list", result)
}

func (cmd *GmailLabelsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.labels.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateLabel(ctx, reqCtx, cmd.Name)
	if err != nil {
		return handleCLIError(out, "gmail.labels.create", err)
	}

	out.Log("Created label: %s", result.Name)
	return out.WriteSuccess("gmail.labels.create", result)
}

func (cmd *GmailLabelsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.labels.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	err = mgr.DeleteLabel(ctx, reqCtx, cmd.LabelID)
	if err != nil {
		return handleCLIError(out, "gmail.labels.delete", err)
	}

	out.Log("Deleted label: %s", cmd.LabelID)
	return out.WriteSuccess("gmail.labels.delete", map[string]string{"id": cmd.LabelID, "status": "deleted"})
}

// --- Filters ---

func (cmd *GmailFiltersListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.filters.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListFilters(ctx, reqCtx)
	if err != nil {
		return handleCLIError(out, "gmail.filters.list", err)
	}

	return out.WriteSuccess("gmail.filters.list", result)
}

func (cmd *GmailFiltersCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.filters.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	var addLabelIDs []string
	if cmd.AddLabels != "" {
		addLabelIDs = splitCSV(cmd.AddLabels)
	}
	var removeLabelIDs []string
	if cmd.RemoveLabels != "" {
		removeLabelIDs = splitCSV(cmd.RemoveLabels)
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.CreateFilter(ctx, reqCtx, cmd.From, cmd.To, cmd.Subject, cmd.Query, cmd.HasAttachment, addLabelIDs, removeLabelIDs, cmd.Forward)
	if err != nil {
		return handleCLIError(out, "gmail.filters.create", err)
	}

	out.Log("Created filter: %s", result.ID)
	return out.WriteSuccess("gmail.filters.create", result)
}

func (cmd *GmailFiltersDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.filters.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	err = mgr.DeleteFilter(ctx, reqCtx, cmd.FilterID)
	if err != nil {
		return handleCLIError(out, "gmail.filters.delete", err)
	}

	out.Log("Deleted filter: %s", cmd.FilterID)
	return out.WriteSuccess("gmail.filters.delete", map[string]string{"id": cmd.FilterID, "status": "deleted"})
}

// --- Vacation ---

func (cmd *GmailVacationGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.vacation.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeGetByID
	result, err := mgr.GetVacation(ctx, reqCtx)
	if err != nil {
		return handleCLIError(out, "gmail.vacation.get", err)
	}

	return out.WriteSuccess("gmail.vacation.get", result)
}

func (cmd *GmailVacationSetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.vacation.set", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	settings := &types.GmailVacationSettings{
		EnableAutoReply:       cmd.Enable,
		ResponseSubject:       cmd.Subject,
		ResponseBodyPlainText: cmd.Body,
		ResponseBodyHtml:      cmd.HTMLBody,
		StartTime:             cmd.Start,
		EndTime:               cmd.End,
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.SetVacation(ctx, reqCtx, settings)
	if err != nil {
		return handleCLIError(out, "gmail.vacation.set", err)
	}

	out.Log("Updated vacation responder settings")
	return out.WriteSuccess("gmail.vacation.set", result)
}

// --- SendAs ---

func (cmd *GmailSendAsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.send-as.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeListOrSearch
	result, err := mgr.ListSendAs(ctx, reqCtx)
	if err != nil {
		return handleCLIError(out, "gmail.send-as.list", err)
	}

	return out.WriteSuccess("gmail.send-as.list", result)
}

// --- Batch operations ---

func (cmd *GmailBatchDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.batch-delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	ids := splitCSV(cmd.IDs)
	if len(ids) == 0 {
		return out.WriteError("gmail.batch-delete", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one message ID is required").Build())
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.BatchDelete(ctx, reqCtx, ids)
	if err != nil {
		return handleCLIError(out, "gmail.batch-delete", err)
	}

	out.Log("Batch deleted %d messages", result.Count)
	return out.WriteSuccess("gmail.batch-delete", result)
}

func (cmd *GmailBatchModifyCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.batch-modify", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	ids := splitCSV(cmd.IDs)
	if len(ids) == 0 {
		return out.WriteError("gmail.batch-modify", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one message ID is required").Build())
	}

	var addLabelIDs []string
	if cmd.AddLabels != "" {
		addLabelIDs = splitCSV(cmd.AddLabels)
	}
	var removeLabelIDs []string
	if cmd.RemoveLabels != "" {
		removeLabelIDs = splitCSV(cmd.RemoveLabels)
	}

	reqCtx.RequestType = types.RequestTypeMutation
	result, err := mgr.BatchModify(ctx, reqCtx, ids, addLabelIDs, removeLabelIDs)
	if err != nil {
		return handleCLIError(out, "gmail.batch-modify", err)
	}

	out.Log("Batch modified %d messages", result.Count)
	return out.WriteSuccess("gmail.batch-modify", result)
}

// --- Attachment ---

func (cmd *GmailAttachmentGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getGmailManager(ctx, flags)
	if err != nil {
		return out.WriteError("gmail.attachment.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}

	reqCtx.RequestType = types.RequestTypeDownloadOrExport
	data, meta, err := mgr.GetAttachment(ctx, reqCtx, cmd.MessageID, cmd.AttachmentID)
	if err != nil {
		return handleCLIError(out, "gmail.attachment.get", err)
	}

	outputPath := cmd.OutputPath
	if outputPath == "" && meta.Filename != "" {
		outputPath = meta.Filename
	}
	if outputPath == "" {
		outputPath = cmd.AttachmentID
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return out.WriteError("gmail.attachment.get", utils.NewCLIError(utils.ErrCodeUnknown, err.Error()).Build())
	}

	out.Log("Downloaded attachment to: %s", outputPath)
	return out.WriteSuccess("gmail.attachment.get", meta)
}
