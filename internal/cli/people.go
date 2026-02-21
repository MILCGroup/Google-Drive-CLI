package cli

import (
	"context"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	peoplemgr "github.com/dl-alexandre/gdrv/internal/people"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
	peopleapi "google.golang.org/api/people/v1"
)

// ============================================================
// People / Contacts commands
// ============================================================

type PeopleCmd struct {
	Contacts      PeopleContactsCmd      `cmd:"" help:"Manage Google Contacts"`
	OtherContacts PeopleOtherContactsCmd `cmd:"" help:"Manage Other Contacts"`
	Directory     PeopleDirectoryCmd     `cmd:"" help:"Domain directory people"`
}

// --- Contacts sub-commands ---

type PeopleContactsCmd struct {
	List   PeopleContactsListCmd   `cmd:"" help:"List contacts"`
	Search PeopleContactsSearchCmd `cmd:"" help:"Search contacts"`
	Get    PeopleContactsGetCmd    `cmd:"" help:"Get a contact by resource name"`
	Create PeopleContactsCreateCmd `cmd:"" help:"Create a new contact"`
	Update PeopleContactsUpdateCmd `cmd:"" help:"Update an existing contact"`
	Delete PeopleContactsDeleteCmd `cmd:"" help:"Delete a contact"`
}

type PeopleContactsListCmd struct {
	Limit     int    `help:"Maximum results to return" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
	Sort      string `help:"Sort order (FIRST_NAME_ASCENDING, LAST_NAME_ASCENDING)" name:"sort"`
}

type PeopleContactsSearchCmd struct {
	Query string `help:"Search query (required)" name:"query" required:""`
	Limit int    `help:"Maximum results to return" default:"100" name:"limit"`
}

type PeopleContactsGetCmd struct {
	ResourceName string `arg:"" name:"resource-name" help:"Contact resource name (e.g. people/c123)"`
}

type PeopleContactsCreateCmd struct {
	GivenName  string `help:"Given (first) name (required)" name:"given-name" required:""`
	FamilyName string `help:"Family (last) name" name:"family-name"`
	Email      string `help:"Comma-separated email addresses" name:"email"`
	Phone      string `help:"Comma-separated phone numbers" name:"phone"`
}

type PeopleContactsUpdateCmd struct {
	ResourceName string `arg:"" name:"resource-name" help:"Contact resource name (e.g. people/c123)"`
	GivenName    string `help:"Update given (first) name" name:"given-name"`
	FamilyName   string `help:"Update family (last) name" name:"family-name"`
	Email        string `help:"Comma-separated email addresses (replaces existing)" name:"email"`
	Phone        string `help:"Comma-separated phone numbers (replaces existing)" name:"phone"`
}

type PeopleContactsDeleteCmd struct {
	ResourceName string `arg:"" name:"resource-name" help:"Contact resource name (e.g. people/c123)"`
}

// --- OtherContacts sub-commands ---

type PeopleOtherContactsCmd struct {
	List   PeopleOtherContactsListCmd   `cmd:"" help:"List other contacts"`
	Search PeopleOtherContactsSearchCmd `cmd:"" help:"Search other contacts"`
}

type PeopleOtherContactsListCmd struct {
	Limit     int    `help:"Maximum results to return" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
}

type PeopleOtherContactsSearchCmd struct {
	Query string `help:"Search query (required)" name:"query" required:""`
	Limit int    `help:"Maximum results to return" default:"100" name:"limit"`
}

// --- Directory sub-commands ---

type PeopleDirectoryCmd struct {
	List   PeopleDirectoryListCmd   `cmd:"" help:"List domain directory people"`
	Search PeopleDirectorySearchCmd `cmd:"" help:"Search domain directory"`
}

type PeopleDirectoryListCmd struct {
	Limit     int    `help:"Maximum results to return" default:"100" name:"limit"`
	PageToken string `help:"Page token for pagination" name:"page-token"`
	Paginate  bool   `help:"Automatically fetch all pages" name:"paginate"`
	Query     string `help:"Filter query" name:"query"`
}

type PeopleDirectorySearchCmd struct {
	Query string `help:"Search query (required)" name:"query" required:""`
	Limit int    `help:"Maximum results to return" default:"100" name:"limit"`
}

// ============================================================
// Run methods -- Contacts
// ============================================================

func (cmd *PeopleContactsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allContacts []types.Contact
		pageToken := ""
		for {
			result, nextToken, err := mgr.ListContacts(ctx, reqCtx, int64(cmd.Limit), pageToken, cmd.Sort)
			if err != nil {
				return handleCLIError(out, "people.contacts.list", err)
			}
			allContacts = append(allContacts, result.Contacts...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("people.contacts.list", allContacts)
		}
		return out.WriteSuccess("people.contacts.list", map[string]interface{}{
			"contacts": allContacts,
		})
	}

	result, nextToken, err := mgr.ListContacts(ctx, reqCtx, int64(cmd.Limit), cmd.PageToken, cmd.Sort)
	if err != nil {
		return handleCLIError(out, "people.contacts.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.contacts.list", result.Contacts)
	}
	resp := map[string]interface{}{
		"contacts":    result.Contacts,
		"totalPeople": result.TotalPeople,
		"totalItems":  result.TotalItems,
	}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("people.contacts.list", resp)
}

func (cmd *PeopleContactsSearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.search", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.SearchContacts(ctx, reqCtx, cmd.Query, int64(cmd.Limit))
	if err != nil {
		return handleCLIError(out, "people.contacts.search", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.contacts.search", result.Contacts)
	}
	return out.WriteSuccess("people.contacts.search", result)
}

func (cmd *PeopleContactsGetCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.get", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeGetByID

	result, err := mgr.GetContact(ctx, reqCtx, cmd.ResourceName)
	if err != nil {
		return handleCLIError(out, "people.contacts.get", err)
	}

	return out.WriteSuccess("people.contacts.get", result)
}

func (cmd *PeopleContactsCreateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.create", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	emails := splitCSV(cmd.Email)
	phones := splitCSV(cmd.Phone)

	result, err := mgr.CreateContact(ctx, reqCtx, cmd.GivenName, cmd.FamilyName, emails, phones)
	if err != nil {
		return handleCLIError(out, "people.contacts.create", err)
	}

	return out.WriteSuccess("people.contacts.create", result)
}

func (cmd *PeopleContactsUpdateCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.update", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	// Fetch current contact to obtain the etag.
	current, err := mgr.GetContact(ctx, reqCtx, cmd.ResourceName)
	if err != nil {
		return handleCLIError(out, "people.contacts.update", err)
	}

	var givenName *string
	if cmd.GivenName != "" {
		givenName = &cmd.GivenName
	}
	var familyName *string
	if cmd.FamilyName != "" {
		familyName = &cmd.FamilyName
	}

	var emails []string
	if cmd.Email != "" {
		emails = splitCSV(cmd.Email)
	}
	var phones []string
	if cmd.Phone != "" {
		phones = splitCSV(cmd.Phone)
	}

	if givenName == nil && familyName == nil && emails == nil && phones == nil {
		return out.WriteError("people.contacts.update", utils.NewCLIError(utils.ErrCodeInvalidArgument, "at least one field must be provided").Build())
	}

	result, err := mgr.UpdateContact(ctx, reqCtx, cmd.ResourceName, current.Etag, givenName, familyName, emails, phones)
	if err != nil {
		return handleCLIError(out, "people.contacts.update", err)
	}

	return out.WriteSuccess("people.contacts.update", result)
}

func (cmd *PeopleContactsDeleteCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.contacts.delete", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeMutation

	if err := mgr.DeleteContact(ctx, reqCtx, cmd.ResourceName); err != nil {
		return handleCLIError(out, "people.contacts.delete", err)
	}

	return out.WriteSuccess("people.contacts.delete", map[string]string{
		"resourceName": cmd.ResourceName,
		"status":       "deleted",
	})
}

// ============================================================
// Run methods -- OtherContacts
// ============================================================

func (cmd *PeopleOtherContactsListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.othercontacts.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allContacts []types.OtherContact
		pageToken := ""
		for {
			result, nextToken, err := mgr.ListOtherContacts(ctx, reqCtx, int64(cmd.Limit), pageToken)
			if err != nil {
				return handleCLIError(out, "people.othercontacts.list", err)
			}
			allContacts = append(allContacts, result.Contacts...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("people.othercontacts.list", allContacts)
		}
		return out.WriteSuccess("people.othercontacts.list", map[string]interface{}{
			"contacts": allContacts,
		})
	}

	result, nextToken, err := mgr.ListOtherContacts(ctx, reqCtx, int64(cmd.Limit), cmd.PageToken)
	if err != nil {
		return handleCLIError(out, "people.othercontacts.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.othercontacts.list", result.Contacts)
	}
	resp := map[string]interface{}{
		"contacts": result.Contacts,
	}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("people.othercontacts.list", resp)
}

func (cmd *PeopleOtherContactsSearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.othercontacts.search", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.SearchOtherContacts(ctx, reqCtx, cmd.Query, int64(cmd.Limit))
	if err != nil {
		return handleCLIError(out, "people.othercontacts.search", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.othercontacts.search", result.Contacts)
	}
	return out.WriteSuccess("people.othercontacts.search", result)
}

// ============================================================
// Run methods -- Directory
// ============================================================

func (cmd *PeopleDirectoryListCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.directory.list", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	if cmd.Paginate {
		var allPeople []types.DirectoryPerson
		pageToken := ""
		for {
			result, nextToken, err := mgr.ListDirectory(ctx, reqCtx, int64(cmd.Limit), pageToken, cmd.Query)
			if err != nil {
				return handleCLIError(out, "people.directory.list", err)
			}
			allPeople = append(allPeople, result.People...)
			if nextToken == "" {
				break
			}
			pageToken = nextToken
		}
		if flags.OutputFormat == types.OutputFormatTable {
			return out.WriteSuccess("people.directory.list", allPeople)
		}
		return out.WriteSuccess("people.directory.list", map[string]interface{}{
			"people": allPeople,
		})
	}

	result, nextToken, err := mgr.ListDirectory(ctx, reqCtx, int64(cmd.Limit), cmd.PageToken, cmd.Query)
	if err != nil {
		return handleCLIError(out, "people.directory.list", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.directory.list", result.People)
	}
	resp := map[string]interface{}{
		"people": result.People,
	}
	if nextToken != "" {
		resp["nextPageToken"] = nextToken
	}
	return out.WriteSuccess("people.directory.list", resp)
}

func (cmd *PeopleDirectorySearchCmd) Run(globals *Globals) error {
	flags := globals.ToGlobalFlags()
	ctx := context.Background()

	mgr, reqCtx, out, err := getPeopleManager(ctx, flags)
	if err != nil {
		return out.WriteError("people.directory.search", utils.NewCLIError(utils.ErrCodeAuthRequired, err.Error()).Build())
	}
	reqCtx.RequestType = types.RequestTypeListOrSearch

	result, err := mgr.SearchDirectory(ctx, reqCtx, cmd.Query, int64(cmd.Limit))
	if err != nil {
		return handleCLIError(out, "people.directory.search", err)
	}

	if flags.OutputFormat == types.OutputFormatTable {
		return out.WriteSuccess("people.directory.search", result.People)
	}
	return out.WriteSuccess("people.directory.search", result)
}

// ============================================================
// Helpers
// ============================================================

// getPeopleManager creates a People API manager and supporting objects
// from the global flags. It mirrors the pattern used by getChatService
// and getAdminService.
func getPeopleManager(ctx context.Context, flags types.GlobalFlags) (*peoplemgr.Manager, *types.RequestContext, *OutputWriter, error) {
	out := NewOutputWriter(flags.OutputFormat, flags.Quiet, flags.Verbose)
	configDir := getConfigDir()
	authMgr := auth.NewManager(configDir)

	creds, err := authMgr.GetValidCredentials(ctx, flags.Profile)
	if err != nil {
		return nil, nil, out, err
	}

	if err := authMgr.ValidateServiceScopes(creds, auth.ServicePeople); err != nil {
		return nil, nil, out, err
	}

	svc, err := authMgr.GetPeopleService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	driveSvc, err := authMgr.GetDriveService(ctx, creds)
	if err != nil {
		return nil, nil, out, err
	}

	client := api.NewClient(driveSvc, utils.DefaultMaxRetries, utils.DefaultRetryDelayMs, GetLogger())
	mgr := peoplemgr.NewManager(client, svc)
	reqCtx := api.NewRequestContext(flags.Profile, flags.DriveID, types.RequestTypeListOrSearch)
	return mgr, reqCtx, out, nil
}

// splitCSV splits a comma-separated string into a trimmed slice.
// Returns nil for empty input.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// Ensure the peopleapi import is used (service type referenced by helper).
var _ *peopleapi.Service
