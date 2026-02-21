package utils

// Upload thresholds (binary units)
const (
	UploadSimpleMaxBytes = 5 * 1024 * 1024  // 5 MiB
	UploadChunkSize      = 8 * 1024 * 1024  // 8 MiB
	ExportMaxBytes       = 10 * 1024 * 1024 // 10 MiB
)

// Revision limits
const RevisionKeepForeverLimit = 200

// OAuth scopes
const (
	ScopeFull                        = "https://www.googleapis.com/auth/drive"
	ScopeFile                        = "https://www.googleapis.com/auth/drive.file"
	ScopeReadonly                    = "https://www.googleapis.com/auth/drive.readonly"
	ScopeMetadataReadonly            = "https://www.googleapis.com/auth/drive.metadata.readonly"
	ScopeAppdata                     = "https://www.googleapis.com/auth/drive.appdata"
	ScopeSheets                      = "https://www.googleapis.com/auth/spreadsheets"
	ScopeSheetsReadonly              = "https://www.googleapis.com/auth/spreadsheets.readonly"
	ScopeDocs                        = "https://www.googleapis.com/auth/documents"
	ScopeDocsReadonly                = "https://www.googleapis.com/auth/documents.readonly"
	ScopeSlides                      = "https://www.googleapis.com/auth/presentations"
	ScopeSlidesReadonly              = "https://www.googleapis.com/auth/presentations.readonly"
	ScopeAdminDirectoryUser          = "https://www.googleapis.com/auth/admin.directory.user"
	ScopeAdminDirectoryUserReadonly  = "https://www.googleapis.com/auth/admin.directory.user.readonly"
	ScopeAdminDirectoryGroup         = "https://www.googleapis.com/auth/admin.directory.group"
	ScopeAdminDirectoryGroupReadonly = "https://www.googleapis.com/auth/admin.directory.group.readonly"
	ScopeLabels                      = "https://www.googleapis.com/auth/drive.labels"
	ScopeLabelsReadonly              = "https://www.googleapis.com/auth/drive.labels.readonly"
	ScopeAdminLabels                 = "https://www.googleapis.com/auth/drive.admin.labels"
	ScopeAdminLabelsReadonly         = "https://www.googleapis.com/auth/drive.admin.labels.readonly"
	ScopeActivity                    = "https://www.googleapis.com/auth/drive.activity"
	ScopeActivityReadonly            = "https://www.googleapis.com/auth/drive.activity.readonly"
	ScopeChat                        = "https://www.googleapis.com/auth/chat"
	ScopeChatReadonly                = "https://www.googleapis.com/auth/chat.readonly"

	// Gmail scopes
	ScopeGmailReadonly      = "https://www.googleapis.com/auth/gmail.readonly"
	ScopeGmailSend          = "https://www.googleapis.com/auth/gmail.send"
	ScopeGmailCompose       = "https://www.googleapis.com/auth/gmail.compose"
	ScopeGmailModify        = "https://www.googleapis.com/auth/gmail.modify"
	ScopeGmailLabels        = "https://www.googleapis.com/auth/gmail.labels"
	ScopeGmailSettingsBasic = "https://www.googleapis.com/auth/gmail.settings.basic"
	ScopeGmailFull          = "https://mail.google.com/" // restricted — batch-delete only

	// Calendar scopes
	ScopeCalendar         = "https://www.googleapis.com/auth/calendar"
	ScopeCalendarReadonly = "https://www.googleapis.com/auth/calendar.readonly"

	// People / Contacts scopes
	ScopeContacts              = "https://www.googleapis.com/auth/contacts"
	ScopeContactsReadonly      = "https://www.googleapis.com/auth/contacts.readonly"
	ScopeContactsOtherReadonly = "https://www.googleapis.com/auth/contacts.other.readonly"
	ScopeDirectoryReadonly     = "https://www.googleapis.com/auth/directory.readonly"

	// Tasks scope
	ScopeTasks = "https://www.googleapis.com/auth/tasks"

	// Forms scopes
	ScopeFormsBody              = "https://www.googleapis.com/auth/forms.body"
	ScopeFormsBodyReadonly      = "https://www.googleapis.com/auth/forms.body.readonly"
	ScopeFormsResponsesReadonly = "https://www.googleapis.com/auth/forms.responses.readonly"

	// Apps Script scopes
	ScopeScriptProjects         = "https://www.googleapis.com/auth/script.projects"
	ScopeScriptProjectsReadonly = "https://www.googleapis.com/auth/script.projects.readonly"

	// Cloud Identity Groups scopes
	ScopeCloudIdentityGroups         = "https://www.googleapis.com/auth/cloud-identity.groups"
	ScopeCloudIdentityGroupsReadonly = "https://www.googleapis.com/auth/cloud-identity.groups.readonly"
)

var (
	ScopesWorkspaceBasic = []string{
		ScopeFile,
		ScopeReadonly,
		ScopeMetadataReadonly,
		ScopeSheetsReadonly,
		ScopeDocsReadonly,
		ScopeSlidesReadonly,
		ScopeLabelsReadonly,
	}
	ScopesWorkspaceFull = []string{
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeLabels,
	}
	ScopesAdmin = []string{
		ScopeAdminDirectoryUser,
		ScopeAdminDirectoryGroup,
		ScopeAdminLabels,
	}
	ScopesWorkspaceWithAdmin = []string{
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeAdminDirectoryUser,
		ScopeAdminDirectoryGroup,
		ScopeLabels,
		ScopeAdminLabels,
	}
	// New presets for additional APIs
	ScopesWorkspaceActivity = []string{
		ScopeFile,
		ScopeReadonly,
		ScopeMetadataReadonly,
		ScopeSheetsReadonly,
		ScopeDocsReadonly,
		ScopeSlidesReadonly,
		ScopeLabelsReadonly,
		ScopeActivityReadonly,
	}
	ScopesWorkspaceLabels = []string{
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeLabels,
	}
	ScopesWorkspaceSync = []string{
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeLabels,
		// Changes API uses standard Drive scopes
	}
	ScopesWorkspaceComplete = []string{
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeLabels,
		ScopeActivity,
		// Changes API uses standard Drive scopes
	}

	// Gmail presets
	ScopesGmail = []string{
		ScopeGmailSend,
		ScopeGmailCompose,
		ScopeGmailModify,
		ScopeGmailLabels,
		ScopeGmailSettingsBasic,
	}
	ScopesGmailReadonly = []string{
		ScopeGmailReadonly,
	}

	// Calendar presets
	ScopesCalendar = []string{
		ScopeCalendar,
	}
	ScopesCalendarReadonly = []string{
		ScopeCalendarReadonly,
	}

	// People preset
	ScopesPeople = []string{
		ScopeContacts,
		ScopeContactsOtherReadonly,
		ScopeDirectoryReadonly,
	}

	// Tasks preset
	ScopesTasks = []string{
		ScopeTasks,
	}

	// Forms preset
	ScopesForms = []string{
		ScopeFormsBody,
		ScopeFormsResponsesReadonly,
	}

	// Apps Script preset
	ScopesAppScript = []string{
		ScopeScriptProjects,
	}

	// Cloud Identity Groups preset
	ScopesGroups = []string{
		ScopeCloudIdentityGroups,
	}

	// Suite-complete mega-preset: ALL existing + ALL new API scopes
	ScopesSuiteComplete = []string{
		// Drive
		ScopeFull,
		ScopeSheets,
		ScopeDocs,
		ScopeSlides,
		ScopeLabels,
		ScopeActivity,
		ScopeAdminDirectoryUser,
		ScopeAdminDirectoryGroup,
		ScopeAdminLabels,
		ScopeChat,
		// Gmail (excludes mail.google.com — must be explicit)
		ScopeGmailSend,
		ScopeGmailCompose,
		ScopeGmailModify,
		ScopeGmailLabels,
		ScopeGmailSettingsBasic,
		// Calendar
		ScopeCalendar,
		// People
		ScopeContacts,
		ScopeContactsOtherReadonly,
		ScopeDirectoryReadonly,
		// Tasks
		ScopeTasks,
		// Forms
		ScopeFormsBody,
		ScopeFormsResponsesReadonly,
		// Apps Script
		ScopeScriptProjects,
		// Cloud Identity Groups
		ScopeCloudIdentityGroups,
	}
)

// Drive API base URLs
const (
	DriveAPIBase    = "https://www.googleapis.com/drive/v3"
	DriveUploadBase = "https://www.googleapis.com/upload/drive/v3"
)

// Retry configuration
const (
	DefaultMaxRetries   = 3
	DefaultRetryDelayMs = 1000
	MaxRetryDelayMs     = 32000
)

// Cache TTL
const DefaultCacheTTLSeconds = 300

// Schema version
const SchemaVersion = "1.0"

// Google Workspace MIME types
const (
	MimeTypeDocument     = "application/vnd.google-apps.document"
	MimeTypeSpreadsheet  = "application/vnd.google-apps.spreadsheet"
	MimeTypePresentation = "application/vnd.google-apps.presentation"
	MimeTypeDrawing      = "application/vnd.google-apps.drawing"
	MimeTypeForm         = "application/vnd.google-apps.form"
	MimeTypeScript       = "application/vnd.google-apps.script"
	MimeTypeFolder       = "application/vnd.google-apps.folder"
	MimeTypeShortcut     = "application/vnd.google-apps.shortcut"
)

// FormatMappings maps convenience format names to MIME types
var FormatMappings = map[string]string{
	"pdf":  "application/pdf",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"txt":  "text/plain",
	"html": "text/html",
	"csv":  "text/csv",
	"png":  "image/png",
	"jpg":  "image/jpeg",
	"svg":  "image/svg+xml",
}

// IsWorkspaceMimeType checks if a MIME type is a Google Workspace type
func IsWorkspaceMimeType(mimeType string) bool {
	switch mimeType {
	case MimeTypeDocument, MimeTypeSpreadsheet, MimeTypePresentation,
		MimeTypeDrawing, MimeTypeForm, MimeTypeScript:
		return true
	}
	return false
}
