// Package cli provides parse-level tests for all kong command structs.
// These tests verify that command paths resolve correctly, flags bind properly,
// positional args are populated, and required flags are enforced — without
// executing any Run() methods or making API calls.
package cli

import (
	"testing"

	"github.com/alecthomas/kong"
)

// newParser creates a fresh kong parser for the CLI struct.
// We suppress the default exit behavior so parse errors return errors instead of os.Exit.
func newParser(t *testing.T) *kong.Kong {
	t.Helper()
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit called during parse") }),
	)
	return p
}

// mustParse asserts that parsing args succeeds.
func mustParse(t *testing.T, args ...string) *kong.Kong {
	t.Helper()
	p := newParser(t)
	_, err := p.Parse(args)
	if err != nil {
		t.Fatalf("unexpected parse error for %v: %v", args, err)
	}
	return p
}

// mustFail asserts that parsing args returns an error (required flag missing, etc.).
func mustFail(t *testing.T, args ...string) {
	t.Helper()
	var c CLI
	var gotExit bool
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { gotExit = true }),
	)
	_, err := p.Parse(args)
	if err == nil && !gotExit {
		t.Fatalf("expected parse error for %v, but got none", args)
	}
}

// ============================================================
// Version
// ============================================================

func TestParseVersion(t *testing.T) {
	mustParse(t, "version")
}

// ============================================================
// About
// ============================================================

func TestParseAbout(t *testing.T) {
	mustParse(t, "about")
}

func TestParseAboutWithFields(t *testing.T) {
	mustParse(t, "about", "--fields", "version,api")
}

// ============================================================
// Files
// ============================================================

func TestParseFilesList(t *testing.T) {
	mustParse(t, "files", "list")
}

func TestParseFilesListWithFlags(t *testing.T) {
	mustParse(t, "files", "list", "--query", "name contains 'report'", "--limit", "50", "--paginate")
}

func TestParseFilesGet(t *testing.T) {
	mustParse(t, "files", "get", "FILEID")
}

func TestParseFilesGetWithFields(t *testing.T) {
	mustParse(t, "files", "get", "FILEID", "--fields", "id,name,mimeType")
}

func TestParseFilesUpload(t *testing.T) {
	mustParse(t, "files", "upload", "/path/to/file.txt")
}

func TestParseFilesUploadWithFlags(t *testing.T) {
	mustParse(t, "files", "upload", "/path/to/file.txt", "--parent", "PARENTID", "--name", "custom.txt")
}

func TestParseFilesDownload(t *testing.T) {
	mustParse(t, "files", "download", "FILEID")
}

func TestParseFilesDownloadWithFileOutput(t *testing.T) {
	// --file-output (renamed from --output to avoid global flag collision)
	mustParse(t, "files", "download", "FILEID", "--file-output", "/tmp/out.pdf")
}

func TestParseFilesDownloadWithDoc(t *testing.T) {
	mustParse(t, "files", "download", "FILEID", "--doc")
}

func TestParseFilesDelete(t *testing.T) {
	mustParse(t, "files", "delete", "FILEID")
}

func TestParseFilesDeletePermanent(t *testing.T) {
	mustParse(t, "files", "delete", "FILEID", "--permanent")
}

func TestParseFilesDeleteSkipConfirmation(t *testing.T) {
	// --skip-confirmation (renamed from --force to avoid global flag collision)
	mustParse(t, "files", "delete", "FILEID", "--skip-confirmation")
}

func TestParseFilesCopy(t *testing.T) {
	mustParse(t, "files", "copy", "FILEID")
}

func TestParseFilesCopyWithFlags(t *testing.T) {
	mustParse(t, "files", "copy", "FILEID", "--name", "Copy of File", "--parent", "PARENTID")
}

func TestParseFilesMove(t *testing.T) {
	// --parent is required for files move
	mustParse(t, "files", "move", "FILEID", "--parent", "NEWPARENTID")
}

func TestParseFilesMoveRequiresParent(t *testing.T) {
	mustFail(t, "files", "move", "FILEID")
}

func TestParseFilesTrash(t *testing.T) {
	mustParse(t, "files", "trash", "FILEID")
}

func TestParseFilesRestore(t *testing.T) {
	mustParse(t, "files", "restore", "FILEID")
}

func TestParseFilesRevisionsList(t *testing.T) {
	mustParse(t, "files", "revisions", "list", "FILEID")
}

// files revisions FILEID routes to list via default:"withargs"
func TestParseFilesRevisionsDefaultWithArgs(t *testing.T) {
	mustParse(t, "files", "revisions", "FILEID")
}

func TestParseFilesRevisionsDownload(t *testing.T) {
	mustParse(t, "files", "revisions", "download", "FILEID", "REVISIONID", "--revision-output", "/tmp/rev.pdf")
}

func TestParseFilesRevisionsDownloadRequiresOutput(t *testing.T) {
	mustFail(t, "files", "revisions", "download", "FILEID", "REVISIONID")
}

func TestParseFilesRevisionsRestore(t *testing.T) {
	mustParse(t, "files", "revisions", "restore", "FILEID", "REVISIONID")
}

func TestParseFilesListTrashed(t *testing.T) {
	mustParse(t, "files", "list-trashed")
}

func TestParseFilesListTrashedWithFlags(t *testing.T) {
	mustParse(t, "files", "list-trashed", "--limit", "25", "--paginate")
}

func TestParseFilesExportFormats(t *testing.T) {
	mustParse(t, "files", "export-formats", "FILEID")
}

// ============================================================
// Folders
// ============================================================

func TestParseFoldersCreate(t *testing.T) {
	mustParse(t, "folders", "create", "My Folder")
}

func TestParseFoldersCreateWithParent(t *testing.T) {
	mustParse(t, "folders", "create", "My Folder", "--parent", "PARENTID")
}

func TestParseFoldersList(t *testing.T) {
	mustParse(t, "folders", "list", "FOLDERID")
}

func TestParseFoldersListWithPaginate(t *testing.T) {
	mustParse(t, "folders", "list", "FOLDERID", "--paginate")
}

func TestParseFoldersDelete(t *testing.T) {
	mustParse(t, "folders", "delete", "FOLDERID")
}

func TestParseFoldersDeleteRecursive(t *testing.T) {
	mustParse(t, "folders", "delete", "FOLDERID", "--recursive")
}

func TestParseFoldersMove(t *testing.T) {
	mustParse(t, "folders", "move", "FOLDERID", "NEWPARENTID")
}

func TestParseFoldersGet(t *testing.T) {
	mustParse(t, "folders", "get", "FOLDERID")
}

func TestParseFoldersGetWithFields(t *testing.T) {
	mustParse(t, "folders", "get", "FOLDERID", "--fields", "id,name")
}

// ============================================================
// Auth
// ============================================================

func TestParseAuthLogin(t *testing.T) {
	mustParse(t, "auth", "login")
}

func TestParseAuthLoginWithPreset(t *testing.T) {
	mustParse(t, "auth", "login", "--preset", "workspace-basic")
}

func TestParseAuthLoginWithNoBrowser(t *testing.T) {
	mustParse(t, "auth", "login", "--no-browser")
}

func TestParseAuthLoginWithWide(t *testing.T) {
	mustParse(t, "auth", "login", "--wide")
}

func TestParseAuthLogout(t *testing.T) {
	mustParse(t, "auth", "logout")
}

func TestParseAuthServiceAccount(t *testing.T) {
	mustParse(t, "auth", "service-account", "--key-file", "/path/to/key.json")
}

func TestParseAuthServiceAccountRequiresKeyFile(t *testing.T) {
	mustFail(t, "auth", "service-account")
}

func TestParseAuthServiceAccountWithImpersonate(t *testing.T) {
	mustParse(t, "auth", "service-account", "--key-file", "/path/to/key.json", "--impersonate-user", "admin@example.com")
}

func TestParseAuthStatus(t *testing.T) {
	mustParse(t, "auth", "status")
}

func TestParseAuthDevice(t *testing.T) {
	mustParse(t, "auth", "device")
}

func TestParseAuthDeviceWithPreset(t *testing.T) {
	mustParse(t, "auth", "device", "--preset", "workspace-full")
}

func TestParseAuthProfiles(t *testing.T) {
	mustParse(t, "auth", "profiles")
}

func TestParseAuthDiagnose(t *testing.T) {
	mustParse(t, "auth", "diagnose")
}

func TestParseAuthDiagnoseWithRefreshCheck(t *testing.T) {
	mustParse(t, "auth", "diagnose", "--refresh-check")
}

// ============================================================
// Permissions
// ============================================================

func TestParsePermissionsList(t *testing.T) {
	mustParse(t, "permissions", "list", "FILEID")
}

func TestParsePermissionsCreate(t *testing.T) {
	mustParse(t, "permissions", "create", "FILEID", "--type", "user", "--role", "reader", "--email", "user@example.com")
}

func TestParsePermissionsCreateRequiresTypeAndRole(t *testing.T) {
	mustFail(t, "permissions", "create", "FILEID")
}

func TestParsePermissionsUpdate(t *testing.T) {
	mustParse(t, "permissions", "update", "FILEID", "PERMID", "--role", "writer")
}

func TestParsePermissionsUpdateRequiresRole(t *testing.T) {
	mustFail(t, "permissions", "update", "FILEID", "PERMID")
}

func TestParsePermissionsRemove(t *testing.T) {
	mustParse(t, "permissions", "remove", "FILEID", "PERMID")
}

func TestParsePermissionsCreateLink(t *testing.T) {
	mustParse(t, "permissions", "create-link", "FILEID")
}

func TestParsePermissionsAuditPublic(t *testing.T) {
	mustParse(t, "permissions", "audit", "public")
}

func TestParsePermissionsAuditExternal(t *testing.T) {
	mustParse(t, "permissions", "audit", "external", "--internal-domain", "example.com")
}

func TestParsePermissionsAuditExternalRequiresDomain(t *testing.T) {
	mustFail(t, "permissions", "audit", "external")
}

func TestParsePermissionsAuditAnyoneWithLink(t *testing.T) {
	mustParse(t, "permissions", "audit", "anyone-with-link")
}

func TestParsePermissionsAuditUser(t *testing.T) {
	mustParse(t, "permissions", "audit", "user", "user@example.com")
}

func TestParsePermissionsAnalyze(t *testing.T) {
	mustParse(t, "permissions", "analyze", "FOLDERID")
}

func TestParsePermissionsAnalyzeWithFlags(t *testing.T) {
	mustParse(t, "permissions", "analyze", "FOLDERID", "--recursive", "--max-depth", "3")
}

func TestParsePermissionsReport(t *testing.T) {
	mustParse(t, "permissions", "report", "FILEID")
}

func TestParsePermissionsBulkRemovePublic(t *testing.T) {
	mustParse(t, "permissions", "bulk", "remove-public", "--folder-id", "FOLDERID")
}

func TestParsePermissionsBulkRemovePublicRequiresFolderID(t *testing.T) {
	mustFail(t, "permissions", "bulk", "remove-public")
}

func TestParsePermissionsBulkUpdateRole(t *testing.T) {
	mustParse(t, "permissions", "bulk", "update-role", "--folder-id", "FOLDERID", "--from-role", "writer", "--to-role", "reader")
}

func TestParsePermissionsBulkUpdateRoleRequiresFlags(t *testing.T) {
	mustFail(t, "permissions", "bulk", "update-role", "--folder-id", "FOLDERID")
}

func TestParsePermissionsSearch(t *testing.T) {
	mustParse(t, "permissions", "search", "--email", "user@example.com")
}

func TestParsePermissionsSearchByRole(t *testing.T) {
	mustParse(t, "permissions", "search", "--role", "commenter")
}

// ============================================================
// Drives
// ============================================================

func TestParseDrivesList(t *testing.T) {
	mustParse(t, "drives", "list")
}

func TestParseDrivesListWithPaginate(t *testing.T) {
	mustParse(t, "drives", "list", "--paginate")
}

func TestParseDrivesGet(t *testing.T) {
	mustParse(t, "drives", "get", "DRIVEID")
}

// ============================================================
// Sheets
// ============================================================

func TestParseSheetsList(t *testing.T) {
	mustParse(t, "sheets", "list")
}

func TestParseSheetsListWithFlags(t *testing.T) {
	mustParse(t, "sheets", "list", "--limit", "50", "--paginate")
}

func TestParseSheetsGet(t *testing.T) {
	mustParse(t, "sheets", "get", "SPREADSHEETID")
}

func TestParseSheetsCreate(t *testing.T) {
	mustParse(t, "sheets", "create", "My Spreadsheet")
}

func TestParseSheetsCreateWithParent(t *testing.T) {
	mustParse(t, "sheets", "create", "My Spreadsheet", "--parent", "PARENTID")
}

func TestParseSheetsBatchUpdate(t *testing.T) {
	mustParse(t, "sheets", "batch-update", "SPREADSHEETID")
}

func TestParseSheetsBatchUpdateWithRequestsFile(t *testing.T) {
	mustParse(t, "sheets", "batch-update", "SPREADSHEETID", "--requests-file", "/path/to/requests.json")
}

func TestParseSheetsValuesGet(t *testing.T) {
	mustParse(t, "sheets", "values", "get", "SPREADSHEETID", "Sheet1!A1:B10")
}

func TestParseSheetsValuesUpdate(t *testing.T) {
	mustParse(t, "sheets", "values", "update", "SPREADSHEETID", "Sheet1!A1:B2")
}

func TestParseSheetsValuesUpdateWithValues(t *testing.T) {
	mustParse(t, "sheets", "values", "update", "SPREADSHEETID", "Sheet1!A1:B2", "--values", "[[1,2],[3,4]]")
}

func TestParseSheetsValuesAppend(t *testing.T) {
	mustParse(t, "sheets", "values", "append", "SPREADSHEETID", "Sheet1!A1")
}

func TestParseSheetsValuesClear(t *testing.T) {
	mustParse(t, "sheets", "values", "clear", "SPREADSHEETID", "Sheet1!A1:B10")
}

// ============================================================
// Docs
// ============================================================

func TestParseDocsList(t *testing.T) {
	mustParse(t, "docs", "list")
}

func TestParseDocsGet(t *testing.T) {
	mustParse(t, "docs", "get", "DOCUMENTID")
}

func TestParseDocsRead(t *testing.T) {
	mustParse(t, "docs", "read", "DOCUMENTID")
}

func TestParseDocsCreate(t *testing.T) {
	mustParse(t, "docs", "create", "My Document")
}

func TestParseDocsCreateWithParent(t *testing.T) {
	mustParse(t, "docs", "create", "My Document", "--parent", "PARENTID")
}

func TestParseDocsUpdate(t *testing.T) {
	mustParse(t, "docs", "update", "DOCUMENTID")
}

func TestParseDocsUpdateWithRequestsFile(t *testing.T) {
	mustParse(t, "docs", "update", "DOCUMENTID", "--requests-file", "/path/to/updates.json")
}

// ============================================================
// Slides
// ============================================================

func TestParseSlidesList(t *testing.T) {
	mustParse(t, "slides", "list")
}

func TestParseSlidesGet(t *testing.T) {
	mustParse(t, "slides", "get", "PRESENTATIONID")
}

func TestParseSlidesRead(t *testing.T) {
	mustParse(t, "slides", "read", "PRESENTATIONID")
}

func TestParseSlidesCreate(t *testing.T) {
	mustParse(t, "slides", "create", "My Presentation")
}

func TestParseSlidesCreateWithParent(t *testing.T) {
	mustParse(t, "slides", "create", "My Presentation", "--parent", "PARENTID")
}

func TestParseSlidesUpdate(t *testing.T) {
	mustParse(t, "slides", "update", "PRESENTATIONID")
}

func TestParseSlidesUpdateWithRequestsFile(t *testing.T) {
	mustParse(t, "slides", "update", "PRESENTATIONID", "--requests-file", "/path/to/updates.json")
}

func TestParseSlidesReplace(t *testing.T) {
	mustParse(t, "slides", "replace", "PRESENTATIONID")
}

func TestParseSlidesReplaceWithData(t *testing.T) {
	mustParse(t, "slides", "replace", "PRESENTATIONID", "--data", `{"{{NAME}}":"Alice"}`)
}

func TestParseSlidesReplaceWithFile(t *testing.T) {
	mustParse(t, "slides", "replace", "PRESENTATIONID", "--file", "/path/to/replacements.json")
}

// ============================================================
// Admin
// ============================================================

func TestParseAdminUsersList(t *testing.T) {
	mustParse(t, "admin", "users", "list")
}

func TestParseAdminUsersListWithDomain(t *testing.T) {
	mustParse(t, "admin", "users", "list", "--domain", "example.com")
}

func TestParseAdminUsersGet(t *testing.T) {
	mustParse(t, "admin", "users", "get", "user@example.com")
}

func TestParseAdminUsersCreate(t *testing.T) {
	mustParse(t, "admin", "users", "create", "new@example.com",
		"--given-name", "John",
		"--family-name", "Doe",
		"--password", "TempPass123!")
}

func TestParseAdminUsersCreateRequiresGivenName(t *testing.T) {
	mustFail(t, "admin", "users", "create", "new@example.com",
		"--family-name", "Doe",
		"--password", "TempPass123!")
}

func TestParseAdminUsersCreateRequiresFamilyName(t *testing.T) {
	mustFail(t, "admin", "users", "create", "new@example.com",
		"--given-name", "John",
		"--password", "TempPass123!")
}

func TestParseAdminUsersCreateRequiresPassword(t *testing.T) {
	mustFail(t, "admin", "users", "create", "new@example.com",
		"--given-name", "John",
		"--family-name", "Doe")
}

func TestParseAdminUsersDelete(t *testing.T) {
	mustParse(t, "admin", "users", "delete", "user@example.com")
}

func TestParseAdminUsersUpdate(t *testing.T) {
	mustParse(t, "admin", "users", "update", "user@example.com", "--given-name", "Jane")
}

func TestParseAdminUsersSuspend(t *testing.T) {
	mustParse(t, "admin", "users", "suspend", "user@example.com")
}

func TestParseAdminUsersUnsuspend(t *testing.T) {
	mustParse(t, "admin", "users", "unsuspend", "user@example.com")
}

func TestParseAdminGroupsList(t *testing.T) {
	mustParse(t, "admin", "groups", "list")
}

func TestParseAdminGroupsGet(t *testing.T) {
	mustParse(t, "admin", "groups", "get", "group@example.com")
}

func TestParseAdminGroupsCreate(t *testing.T) {
	mustParse(t, "admin", "groups", "create", "group@example.com", "Engineering Team")
}

func TestParseAdminGroupsCreateWithDescription(t *testing.T) {
	mustParse(t, "admin", "groups", "create", "group@example.com", "Engineering Team",
		"--description", "Engineering team members")
}

func TestParseAdminGroupsDelete(t *testing.T) {
	mustParse(t, "admin", "groups", "delete", "group@example.com")
}

func TestParseAdminGroupsUpdate(t *testing.T) {
	mustParse(t, "admin", "groups", "update", "group@example.com", "--name", "New Name")
}

func TestParseAdminGroupsMembersList(t *testing.T) {
	mustParse(t, "admin", "groups", "members", "list", "group@example.com")
}

func TestParseAdminGroupsMembersAdd(t *testing.T) {
	mustParse(t, "admin", "groups", "members", "add", "group@example.com", "user@example.com")
}

func TestParseAdminGroupsMembersAddWithRole(t *testing.T) {
	mustParse(t, "admin", "groups", "members", "add", "group@example.com", "user@example.com", "--role", "MANAGER")
}

func TestParseAdminGroupsMembersRemove(t *testing.T) {
	mustParse(t, "admin", "groups", "members", "remove", "group@example.com", "user@example.com")
}

// ============================================================
// Changes
// ============================================================

func TestParseChangesStartPageToken(t *testing.T) {
	mustParse(t, "changes", "start-page-token")
}

func TestParseChangesList(t *testing.T) {
	mustParse(t, "changes", "list", "--page-token", "12345")
}

func TestParseChangesListRequiresPageToken(t *testing.T) {
	mustFail(t, "changes", "list")
}

func TestParseChangesListWithFlags(t *testing.T) {
	mustParse(t, "changes", "list", "--page-token", "12345", "--limit", "50", "--paginate")
}

func TestParseChangesWatch(t *testing.T) {
	mustParse(t, "changes", "watch", "--page-token", "12345", "--webhook-url", "https://example.com/webhook")
}

func TestParseChangesWatchRequiresPageToken(t *testing.T) {
	mustFail(t, "changes", "watch", "--webhook-url", "https://example.com/webhook")
}

func TestParseChangesWatchRequiresWebhookURL(t *testing.T) {
	mustFail(t, "changes", "watch", "--page-token", "12345")
}

func TestParseChangesStop(t *testing.T) {
	mustParse(t, "changes", "stop", "CHANNELID", "RESOURCEID")
}

// ============================================================
// Labels
// ============================================================

func TestParseLabelsList(t *testing.T) {
	mustParse(t, "labels", "list")
}

func TestParseLabelsListWithFlags(t *testing.T) {
	mustParse(t, "labels", "list", "--published-only", "--limit", "50")
}

func TestParseLabelsGet(t *testing.T) {
	mustParse(t, "labels", "get", "LABELID")
}

func TestParseLabelsCreate(t *testing.T) {
	mustParse(t, "labels", "create", "Document Type")
}

func TestParseLabelsPublish(t *testing.T) {
	mustParse(t, "labels", "publish", "LABELID")
}

func TestParseLabelsDisable(t *testing.T) {
	mustParse(t, "labels", "disable", "LABELID")
}

func TestParseLabelsFileList(t *testing.T) {
	mustParse(t, "labels", "file", "list", "FILEID")
}

func TestParseLabelsFileApply(t *testing.T) {
	mustParse(t, "labels", "file", "apply", "FILEID", "LABELID")
}

func TestParseLabelsFileApplyWithFields(t *testing.T) {
	mustParse(t, "labels", "file", "apply", "FILEID", "LABELID", "--fields", `{"field1":"value1"}`)
}

func TestParseLabelsFileUpdate(t *testing.T) {
	mustParse(t, "labels", "file", "update", "FILEID", "LABELID")
}

func TestParseLabelsFileRemove(t *testing.T) {
	mustParse(t, "labels", "file", "remove", "FILEID", "LABELID")
}

// ============================================================
// Activity
// ============================================================

func TestParseActivityQuery(t *testing.T) {
	mustParse(t, "activity", "query")
}

func TestParseActivityQueryWithFileID(t *testing.T) {
	mustParse(t, "activity", "query", "--file-id", "FILEID")
}

func TestParseActivityQueryWithTimeRange(t *testing.T) {
	mustParse(t, "activity", "query",
		"--start-time", "2026-01-01T00:00:00Z",
		"--end-time", "2026-01-31T23:59:59Z")
}

func TestParseActivityQueryWithActionTypes(t *testing.T) {
	mustParse(t, "activity", "query", "--action-types", "edit,share,permission_change")
}

// ============================================================
// Chat
// ============================================================

func TestParseChatSpacesList(t *testing.T) {
	mustParse(t, "chat", "spaces", "list")
}

func TestParseChatSpacesGet(t *testing.T) {
	mustParse(t, "chat", "spaces", "get", "SPACEID")
}

func TestParseChatSpacesCreate(t *testing.T) {
	mustParse(t, "chat", "spaces", "create", "--display-name", "Team Chat")
}

func TestParseChatSpacesCreateWithType(t *testing.T) {
	mustParse(t, "chat", "spaces", "create", "--display-name", "Team Chat", "--type", "SPACE")
}

func TestParseChatSpacesDelete(t *testing.T) {
	mustParse(t, "chat", "spaces", "delete", "SPACEID")
}

func TestParseChatMessagesList(t *testing.T) {
	mustParse(t, "chat", "messages", "list", "SPACEID")
}

func TestParseChatMessagesGet(t *testing.T) {
	mustParse(t, "chat", "messages", "get", "SPACEID", "MESSAGEID")
}

func TestParseChatMessagesCreate(t *testing.T) {
	mustParse(t, "chat", "messages", "create", "SPACEID", "--text", "Hello everyone!")
}

func TestParseChatMessagesCreateWithThread(t *testing.T) {
	mustParse(t, "chat", "messages", "create", "SPACEID", "--text", "Reply", "--thread", "THREADID")
}

func TestParseChatMessagesUpdate(t *testing.T) {
	mustParse(t, "chat", "messages", "update", "SPACEID", "MESSAGEID", "--text", "Updated text")
}

func TestParseChatMessagesDelete(t *testing.T) {
	mustParse(t, "chat", "messages", "delete", "SPACEID", "MESSAGEID")
}

func TestParseChatMembersList(t *testing.T) {
	mustParse(t, "chat", "members", "list", "SPACEID")
}

func TestParseChatMembersGet(t *testing.T) {
	mustParse(t, "chat", "members", "get", "SPACEID", "MEMBERID")
}

func TestParseChatMembersCreate(t *testing.T) {
	mustParse(t, "chat", "members", "create", "SPACEID", "--email", "user@example.com")
}

func TestParseChatMembersCreateWithRole(t *testing.T) {
	mustParse(t, "chat", "members", "create", "SPACEID", "--email", "user@example.com", "--role", "MANAGER")
}

func TestParseChatMembersDelete(t *testing.T) {
	mustParse(t, "chat", "members", "delete", "SPACEID", "MEMBERID")
}

// ============================================================
// Sync
// ============================================================

func TestParseSyncInit(t *testing.T) {
	mustParse(t, "sync", "init", "/local/path", "REMOTEFOLDEID")
}

func TestParseSyncInitWithFlags(t *testing.T) {
	mustParse(t, "sync", "init", "/local/path", "REMOTEFOLDEID",
		"--conflict", "local-wins",
		"--direction", "push",
		"--exclude", "*.tmp,*.log")
}

func TestParseSyncPush(t *testing.T) {
	mustParse(t, "sync", "push", "CONFIGID")
}

func TestParseSyncPushWithFlags(t *testing.T) {
	mustParse(t, "sync", "push", "CONFIGID", "--delete", "--concurrency", "10")
}

func TestParseSyncPull(t *testing.T) {
	mustParse(t, "sync", "pull", "CONFIGID")
}

func TestParseSyncStatus(t *testing.T) {
	mustParse(t, "sync", "status", "CONFIGID")
}

func TestParseSyncList(t *testing.T) {
	mustParse(t, "sync", "list")
}

func TestParseSyncRemove(t *testing.T) {
	mustParse(t, "sync", "remove", "CONFIGID")
}

// ============================================================
// Config
// ============================================================

func TestParseConfigShow(t *testing.T) {
	mustParse(t, "config", "show")
}

func TestParseConfigSet(t *testing.T) {
	mustParse(t, "config", "set", "cachettl", "600")
}

func TestParseConfigReset(t *testing.T) {
	mustParse(t, "config", "reset")
}

// ============================================================
// Completion
// ============================================================

func TestParseCompletionBash(t *testing.T) {
	mustParse(t, "completion", "bash")
}

func TestParseCompletionZsh(t *testing.T) {
	mustParse(t, "completion", "zsh")
}

func TestParseCompletionFish(t *testing.T) {
	mustParse(t, "completion", "fish")
}

// ============================================================
// TestPermAliasWorks — "perm" is an alias for "permissions"
// ============================================================

func TestPermAliasWorks(t *testing.T) {
	var c CLI
	var gotExit bool
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { gotExit = true }),
	)
	_, err := p.Parse([]string{"perm", "list", "FILEID"})
	if err != nil || gotExit {
		t.Fatalf("expected 'perm list FILEID' to parse successfully, got err=%v, exit=%v", err, gotExit)
	}
}

// ============================================================
// TestGlobalFlagInheritance — global flags work on subcommands
// ============================================================

func TestGlobalFlagInheritance(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"json on files list", []string{"--json", "files", "list"}},
		{"quiet on files list", []string{"--quiet", "files", "list"}},
		{"profile on files list", []string{"--profile", "work", "files", "list"}},
		{"json on auth status", []string{"--json", "auth", "status"}},
		{"quiet on auth status", []string{"--quiet", "auth", "status"}},
		{"profile on auth login", []string{"--profile", "personal", "auth", "login"}},
		{"json on admin users list", []string{"--json", "admin", "users", "list"}},
		{"quiet on sheets list", []string{"--quiet", "sheets", "list"}},
		{"profile on drives list", []string{"--profile", "default", "drives", "list"}},
		{"json on activity query", []string{"--json", "activity", "query"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustParse(t, tc.args...)
		})
	}
}

// ============================================================
// TestShortFlags — short flags work on subcommands
// ============================================================

func TestShortFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"-q (quiet) on files list", []string{"-q", "files", "list"}},
		{"-v (verbose) on files list", []string{"-v", "files", "list"}},
		{"-f (force) on files delete", []string{"-f", "files", "delete", "FILEID"}},
		{"-y (yes) on files delete", []string{"-y", "files", "delete", "FILEID"}},
		{"-q combined with -v", []string{"-q", "-v", "files", "list"}},
		{"-f on permissions bulk remove-public", []string{"-f", "permissions", "bulk", "remove-public", "--folder-id", "FID"}},
		{"-y on config reset", []string{"-y", "config", "reset"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustParse(t, tc.args...)
		})
	}
}

// ============================================================
// TestJSONAliasWorks — --json sets output to json
// ============================================================

func TestJSONAliasWorks(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"--json", "files", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.JSON {
		t.Error("expected c.JSON to be true after --json flag")
	}
}

// ============================================================
// TestDefaultValues — default values are applied
// ============================================================

func TestDefaultValues(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"files", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Global defaults
	if c.Output != "json" {
		t.Errorf("expected default Output=json, got %q", c.Output)
	}
	if c.Profile != "default" {
		t.Errorf("expected default Profile=default, got %q", c.Profile)
	}
	if c.CacheTTL != 300 {
		t.Errorf("expected default CacheTTL=300, got %d", c.CacheTTL)
	}

	// Files list defaults
	if c.Files.List.Limit != 100 {
		t.Errorf("expected default files list Limit=100, got %d", c.Files.List.Limit)
	}
}

func TestDefaultValuesSheetsValuesUpdate(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"sheets", "values", "update", "SPID", "Sheet1!A1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Sheets.Values.Update.ValueInputOption != "USER_ENTERED" {
		t.Errorf("expected default ValueInputOption=USER_ENTERED, got %q", c.Sheets.Values.Update.ValueInputOption)
	}
}

func TestDefaultValuesChangesListSupportsAllDrives(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"changes", "list", "--page-token", "TOKEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Changes.List.SupportsAllDrives {
		t.Error("expected default SupportsAllDrives=true for changes list")
	}
}

func TestDefaultValuesSyncPushConcurrency(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"sync", "push", "CONFIGID"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Sync.Push.Concurrency != 5 {
		t.Errorf("expected default sync push Concurrency=5, got %d", c.Sync.Push.Concurrency)
	}
}

func TestDefaultValuesAdminGroupsMembersAdd(t *testing.T) {
	var c CLI
	p := kong.Must(&c,
		kong.Name("gdrv"),
		kong.Exit(func(int) { t.Fatal("unexpected exit") }),
	)
	_, err := p.Parse([]string{"admin", "groups", "members", "add", "group@example.com", "user@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Admin.Groups.Members.Add.Role != "MEMBER" {
		t.Errorf("expected default Role=MEMBER, got %q", c.Admin.Groups.Members.Add.Role)
	}
}

// ============================================================
// TestPositionalArgs — positional args bind correctly
// ============================================================

func TestPositionalArgs(t *testing.T) {
	t.Run("files get FILEID", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"files", "get", "my-file-id-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Files.Get.FileID != "my-file-id-123" {
			t.Errorf("expected FileID=my-file-id-123, got %q", c.Files.Get.FileID)
		}
	})

	t.Run("files upload LOCAL-PATH", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"files", "upload", "/home/user/report.pdf"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Files.Upload.LocalPath != "/home/user/report.pdf" {
			t.Errorf("expected LocalPath=/home/user/report.pdf, got %q", c.Files.Upload.LocalPath)
		}
	})

	t.Run("admin users get USER-KEY", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"admin", "users", "get", "admin@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Admin.Users.Get.UserKey != "admin@example.com" {
			t.Errorf("expected UserKey=admin@example.com, got %q", c.Admin.Users.Get.UserKey)
		}
	})

	t.Run("changes stop CHANNELID RESOURCEID", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"changes", "stop", "chan-id-xyz", "res-id-abc"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Changes.Stop.ChannelID != "chan-id-xyz" {
			t.Errorf("expected ChannelID=chan-id-xyz, got %q", c.Changes.Stop.ChannelID)
		}
		if c.Changes.Stop.ResourceID != "res-id-abc" {
			t.Errorf("expected ResourceID=res-id-abc, got %q", c.Changes.Stop.ResourceID)
		}
	})

	t.Run("permissions update FILEID PERMID --role", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"permissions", "update", "file-abc", "perm-xyz", "--role", "writer"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Permissions.Update.FileID != "file-abc" {
			t.Errorf("expected FileID=file-abc, got %q", c.Permissions.Update.FileID)
		}
		if c.Permissions.Update.PermissionID != "perm-xyz" {
			t.Errorf("expected PermissionID=perm-xyz, got %q", c.Permissions.Update.PermissionID)
		}
		if c.Permissions.Update.Role != "writer" {
			t.Errorf("expected Role=writer, got %q", c.Permissions.Update.Role)
		}
	})

	t.Run("admin users create EMAIL positional", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"admin", "users", "create", "new@example.com",
			"--given-name", "Bob",
			"--family-name", "Builder",
			"--password", "Pa$$word1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Admin.Users.Create.Email != "new@example.com" {
			t.Errorf("expected Email=new@example.com, got %q", c.Admin.Users.Create.Email)
		}
		if c.Admin.Users.Create.GivenName != "Bob" {
			t.Errorf("expected GivenName=Bob, got %q", c.Admin.Users.Create.GivenName)
		}
	})

	t.Run("labels file apply FILEID LABELID", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"labels", "file", "apply", "file-abc", "label-xyz"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Labels.File.Apply.FileID != "file-abc" {
			t.Errorf("expected FileID=file-abc, got %q", c.Labels.File.Apply.FileID)
		}
		if c.Labels.File.Apply.LabelID != "label-xyz" {
			t.Errorf("expected LabelID=label-xyz, got %q", c.Labels.File.Apply.LabelID)
		}
	})

	t.Run("sync init LOCAL REMOTE", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"sync", "init", "/local/path", "drive-folder-id"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Sync.Init.LocalPath != "/local/path" {
			t.Errorf("expected LocalPath=/local/path, got %q", c.Sync.Init.LocalPath)
		}
		if c.Sync.Init.RemoteFolder != "drive-folder-id" {
			t.Errorf("expected RemoteFolder=drive-folder-id, got %q", c.Sync.Init.RemoteFolder)
		}
	})

	t.Run("config set KEY VALUE", func(t *testing.T) {
		var c CLI
		p := kong.Must(&c,
			kong.Name("gdrv"),
			kong.Exit(func(int) { t.Fatal("unexpected exit") }),
		)
		_, err := p.Parse([]string{"config", "set", "cachettl", "600"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Config.Set.Key != "cachettl" {
			t.Errorf("expected Key=cachettl, got %q", c.Config.Set.Key)
		}
		if c.Config.Set.Value != "600" {
			t.Errorf("expected Value=600, got %q", c.Config.Set.Value)
		}
	})
}

// ============================================================
// TestRequiredFlagsEnforced — required flags cause parse failures
// ============================================================

func TestRequiredFlagsEnforced(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"admin users create missing --given-name",
			[]string{"admin", "users", "create", "new@example.com", "--family-name", "Doe", "--password", "Pass123!"}},
		{"admin users create missing --family-name",
			[]string{"admin", "users", "create", "new@example.com", "--given-name", "John", "--password", "Pass123!"}},
		{"admin users create missing --password",
			[]string{"admin", "users", "create", "new@example.com", "--given-name", "John", "--family-name", "Doe"}},
		{"auth service-account missing --key-file",
			[]string{"auth", "service-account"}},
		{"changes list missing --page-token",
			[]string{"changes", "list"}},
		{"changes watch missing --page-token",
			[]string{"changes", "watch", "--webhook-url", "https://example.com"}},
		{"changes watch missing --webhook-url",
			[]string{"changes", "watch", "--page-token", "TOKEN"}},
		{"permissions create missing --type --role",
			[]string{"permissions", "create", "FILEID"}},
		{"permissions update missing --role",
			[]string{"permissions", "update", "FILEID", "PERMID"}},
		{"permissions audit external missing --internal-domain",
			[]string{"permissions", "audit", "external"}},
		{"permissions bulk remove-public missing --folder-id",
			[]string{"permissions", "bulk", "remove-public"}},
		{"permissions bulk update-role missing required flags",
			[]string{"permissions", "bulk", "update-role", "--folder-id", "FID"}},
		{"files move missing --parent",
			[]string{"files", "move", "FILEID"}},
		{"files revisions download missing --revision-output",
			[]string{"files", "revisions", "download", "FILEID", "REVID"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustFail(t, tc.args...)
		})
	}
}

// ============================================================
// TestAllCommandsReachable — every command path resolves
// ============================================================

func TestAllCommandsReachable(t *testing.T) {
	// Every parseable command path in the CLI.
	// For commands with required flags/args, we supply them.
	commands := []struct {
		name string
		args []string
	}{
		// Top-level
		{"version", []string{"version"}},
		{"about", []string{"about"}},

		// files
		{"files list", []string{"files", "list"}},
		{"files get", []string{"files", "get", "FILEID"}},
		{"files upload", []string{"files", "upload", "/path/to/file"}},
		{"files download", []string{"files", "download", "FILEID"}},
		{"files delete", []string{"files", "delete", "FILEID"}},
		{"files copy", []string{"files", "copy", "FILEID"}},
		{"files move", []string{"files", "move", "FILEID", "--parent", "PARENTID"}},
		{"files trash", []string{"files", "trash", "FILEID"}},
		{"files restore", []string{"files", "restore", "FILEID"}},
		{"files revisions list", []string{"files", "revisions", "list", "FILEID"}},
		{"files revisions download", []string{"files", "revisions", "download", "FILEID", "REVID", "--revision-output", "/tmp/out"}},
		{"files revisions restore", []string{"files", "revisions", "restore", "FILEID", "REVID"}},
		{"files list-trashed", []string{"files", "list-trashed"}},
		{"files export-formats", []string{"files", "export-formats", "FILEID"}},

		// folders
		{"folders create", []string{"folders", "create", "My Folder"}},
		{"folders list", []string{"folders", "list", "FOLDERID"}},
		{"folders delete", []string{"folders", "delete", "FOLDERID"}},
		{"folders move", []string{"folders", "move", "FOLDERID", "NEWPARENTID"}},
		{"folders get", []string{"folders", "get", "FOLDERID"}},

		// auth
		{"auth login", []string{"auth", "login"}},
		{"auth logout", []string{"auth", "logout"}},
		{"auth service-account", []string{"auth", "service-account", "--key-file", "/path/key.json"}},
		{"auth status", []string{"auth", "status"}},
		{"auth device", []string{"auth", "device"}},
		{"auth profiles", []string{"auth", "profiles"}},
		{"auth diagnose", []string{"auth", "diagnose"}},

		// permissions
		{"permissions list", []string{"permissions", "list", "FILEID"}},
		{"permissions create", []string{"permissions", "create", "FILEID", "--type", "user", "--role", "reader", "--email", "user@example.com"}},
		{"permissions update", []string{"permissions", "update", "FILEID", "PERMID", "--role", "writer"}},
		{"permissions remove", []string{"permissions", "remove", "FILEID", "PERMID"}},
		{"permissions create-link", []string{"permissions", "create-link", "FILEID"}},
		{"permissions audit public", []string{"permissions", "audit", "public"}},
		{"permissions audit external", []string{"permissions", "audit", "external", "--internal-domain", "example.com"}},
		{"permissions audit anyone-with-link", []string{"permissions", "audit", "anyone-with-link"}},
		{"permissions audit user", []string{"permissions", "audit", "user", "user@example.com"}},
		{"permissions analyze", []string{"permissions", "analyze", "FOLDERID"}},
		{"permissions report", []string{"permissions", "report", "FILEID"}},
		{"permissions bulk remove-public", []string{"permissions", "bulk", "remove-public", "--folder-id", "FID"}},
		{"permissions bulk update-role", []string{"permissions", "bulk", "update-role", "--folder-id", "FID", "--from-role", "writer", "--to-role", "reader"}},
		{"permissions search", []string{"permissions", "search", "--email", "user@example.com"}},

		// drives
		{"drives list", []string{"drives", "list"}},
		{"drives get", []string{"drives", "get", "DRIVEID"}},

		// sheets
		{"sheets list", []string{"sheets", "list"}},
		{"sheets get", []string{"sheets", "get", "SPREADSHEETID"}},
		{"sheets create", []string{"sheets", "create", "My Sheet"}},
		{"sheets batch-update", []string{"sheets", "batch-update", "SPREADSHEETID"}},
		{"sheets values get", []string{"sheets", "values", "get", "SPREADSHEETID", "Sheet1!A1:B10"}},
		{"sheets values update", []string{"sheets", "values", "update", "SPREADSHEETID", "Sheet1!A1"}},
		{"sheets values append", []string{"sheets", "values", "append", "SPREADSHEETID", "Sheet1!A1"}},
		{"sheets values clear", []string{"sheets", "values", "clear", "SPREADSHEETID", "Sheet1!A1:B10"}},

		// docs
		{"docs list", []string{"docs", "list"}},
		{"docs get", []string{"docs", "get", "DOCID"}},
		{"docs read", []string{"docs", "read", "DOCID"}},
		{"docs create", []string{"docs", "create", "My Doc"}},
		{"docs update", []string{"docs", "update", "DOCID"}},

		// slides
		{"slides list", []string{"slides", "list"}},
		{"slides get", []string{"slides", "get", "PRESID"}},
		{"slides read", []string{"slides", "read", "PRESID"}},
		{"slides create", []string{"slides", "create", "My Presentation"}},
		{"slides update", []string{"slides", "update", "PRESID"}},
		{"slides replace", []string{"slides", "replace", "PRESID"}},

		// admin users
		{"admin users list", []string{"admin", "users", "list"}},
		{"admin users get", []string{"admin", "users", "get", "user@example.com"}},
		{"admin users create", []string{"admin", "users", "create", "new@example.com", "--given-name", "John", "--family-name", "Doe", "--password", "Pass123!"}},
		{"admin users delete", []string{"admin", "users", "delete", "user@example.com"}},
		{"admin users update", []string{"admin", "users", "update", "user@example.com", "--given-name", "Jane"}},
		{"admin users suspend", []string{"admin", "users", "suspend", "user@example.com"}},
		{"admin users unsuspend", []string{"admin", "users", "unsuspend", "user@example.com"}},

		// admin groups
		{"admin groups list", []string{"admin", "groups", "list"}},
		{"admin groups get", []string{"admin", "groups", "get", "group@example.com"}},
		{"admin groups create", []string{"admin", "groups", "create", "group@example.com", "Engineering Team"}},
		{"admin groups delete", []string{"admin", "groups", "delete", "group@example.com"}},
		{"admin groups update", []string{"admin", "groups", "update", "group@example.com", "--name", "New Name"}},
		{"admin groups members list", []string{"admin", "groups", "members", "list", "group@example.com"}},
		{"admin groups members add", []string{"admin", "groups", "members", "add", "group@example.com", "user@example.com"}},
		{"admin groups members remove", []string{"admin", "groups", "members", "remove", "group@example.com", "user@example.com"}},

		// changes
		{"changes start-page-token", []string{"changes", "start-page-token"}},
		{"changes list", []string{"changes", "list", "--page-token", "TOKEN"}},
		{"changes watch", []string{"changes", "watch", "--page-token", "TOKEN", "--webhook-url", "https://example.com"}},
		{"changes stop", []string{"changes", "stop", "CHANNELID", "RESOURCEID"}},

		// labels
		{"labels list", []string{"labels", "list"}},
		{"labels get", []string{"labels", "get", "LABELID"}},
		{"labels create", []string{"labels", "create", "My Label"}},
		{"labels publish", []string{"labels", "publish", "LABELID"}},
		{"labels disable", []string{"labels", "disable", "LABELID"}},
		{"labels file list", []string{"labels", "file", "list", "FILEID"}},
		{"labels file apply", []string{"labels", "file", "apply", "FILEID", "LABELID"}},
		{"labels file update", []string{"labels", "file", "update", "FILEID", "LABELID"}},
		{"labels file remove", []string{"labels", "file", "remove", "FILEID", "LABELID"}},

		// activity
		{"activity query", []string{"activity", "query"}},

		// chat spaces
		{"chat spaces list", []string{"chat", "spaces", "list"}},
		{"chat spaces get", []string{"chat", "spaces", "get", "SPACEID"}},
		{"chat spaces create", []string{"chat", "spaces", "create", "--display-name", "Team Chat"}},
		{"chat spaces delete", []string{"chat", "spaces", "delete", "SPACEID"}},

		// chat messages
		{"chat messages list", []string{"chat", "messages", "list", "SPACEID"}},
		{"chat messages get", []string{"chat", "messages", "get", "SPACEID", "MESSAGEID"}},
		{"chat messages create", []string{"chat", "messages", "create", "SPACEID", "--text", "Hello!"}},
		{"chat messages update", []string{"chat", "messages", "update", "SPACEID", "MESSAGEID", "--text", "Updated"}},
		{"chat messages delete", []string{"chat", "messages", "delete", "SPACEID", "MESSAGEID"}},

		// chat members
		{"chat members list", []string{"chat", "members", "list", "SPACEID"}},
		{"chat members get", []string{"chat", "members", "get", "SPACEID", "MEMBERID"}},
		{"chat members create", []string{"chat", "members", "create", "SPACEID", "--email", "user@example.com"}},
		{"chat members delete", []string{"chat", "members", "delete", "SPACEID", "MEMBERID"}},

		// sync
		{"sync init", []string{"sync", "init", "/local/path", "drive-folder-id"}},
		{"sync push", []string{"sync", "push", "CONFIGID"}},
		{"sync pull", []string{"sync", "pull", "CONFIGID"}},
		{"sync status", []string{"sync", "status", "CONFIGID"}},
		{"sync list", []string{"sync", "list"}},
		{"sync remove", []string{"sync", "remove", "CONFIGID"}},

		// config
		{"config show", []string{"config", "show"}},
		{"config set", []string{"config", "set", "cachettl", "600"}},
		{"config reset", []string{"config", "reset"}},

		// completion
		{"completion bash", []string{"completion", "bash"}},
		{"completion zsh", []string{"completion", "zsh"}},
		{"completion fish", []string{"completion", "fish"}},

		// perm alias
		{"perm list (alias)", []string{"perm", "list", "FILEID"}},
		{"perm create-link (alias)", []string{"perm", "create-link", "FILEID"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			mustParse(t, tc.args...)
		})
	}
}
