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
	ScopeFull             = "https://www.googleapis.com/auth/drive"
	ScopeFile             = "https://www.googleapis.com/auth/drive.file"
	ScopeReadonly         = "https://www.googleapis.com/auth/drive.readonly"
	ScopeMetadataReadonly = "https://www.googleapis.com/auth/drive.metadata.readonly"
	ScopeAppdata          = "https://www.googleapis.com/auth/drive.appdata"
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
