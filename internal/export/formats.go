package export

import (
	"fmt"
	"strings"

	"github.com/dl-alexandre/gdrv/internal/utils"
)

// Google Workspace MIME types
const (
	MimeTypeGoogleDocs        = "application/vnd.google-apps.document"
	MimeTypeGoogleSheets      = "application/vnd.google-apps.spreadsheet"
	MimeTypeGoogleSlides      = "application/vnd.google-apps.presentation"
	MimeTypeGoogleDrawing     = "application/vnd.google-apps.drawing"
	MimeTypeGoogleForm        = "application/vnd.google-apps.form"
	MimeTypeGoogleSite        = "application/vnd.google-apps.site"
	MimeTypeGoogleScript      = "application/vnd.google-apps.script"
	MimeTypeGoogleJamboard    = "application/vnd.google-apps.jam"
	MimeTypeGoogleShortcut    = "application/vnd.google-apps.shortcut"
	MimeTypeGoogleFolder      = "application/vnd.google-apps.folder"
	MimeTypeGoogleMap         = "application/vnd.google-apps.map"
	MimeTypeGoogleFusiontable = "application/vnd.google-apps.fusiontable"
)

// Export MIME types as per Google's reference
// https://developers.google.com/drive/api/guides/ref-export-formats
var exportFormats = map[string][]string{
	MimeTypeGoogleDocs: {
		"application/rtf",
		"application/vnd.oasis.opendocument.text",
		"text/html",
		"application/pdf",
		"application/epub+zip",
		"application/zip",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
	},
	MimeTypeGoogleSheets: {
		"application/x-vnd.oasis.opendocument.spreadsheet",
		"text/tab-separated-values",
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/csv",
		"application/zip",
	},
	MimeTypeGoogleSlides: {
		"application/vnd.oasis.opendocument.presentation",
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
	},
	MimeTypeGoogleDrawing: {
		"image/svg+xml",
		"image/png",
		"application/pdf",
		"image/jpeg",
	},
	MimeTypeGoogleScript: {
		"application/vnd.google-apps.script+json",
	},
}

// Convenience format mapping
var convenienceFormats = map[string]string{
	"pdf":  "application/pdf",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"txt":  "text/plain",
	"html": "text/html",
	"rtf":  "application/rtf",
	"odt":  "application/vnd.oasis.opendocument.text",
	"ods":  "application/x-vnd.oasis.opendocument.spreadsheet",
	"odp":  "application/vnd.oasis.opendocument.presentation",
	"csv":  "text/csv",
	"tsv":  "text/tab-separated-values",
	"zip":  "application/zip",
	"epub": "application/epub+zip",
	"svg":  "image/svg+xml",
	"png":  "image/png",
	"jpeg": "image/jpeg",
	"jpg":  "image/jpeg",
	"json": "application/vnd.google-apps.script+json",
}

// IsGoogleWorkspaceFile checks if a MIME type is a Google Workspace file
func IsGoogleWorkspaceFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/vnd.google-apps.")
}

// GetConvenienceFormat maps a format shorthand to a MIME type
func GetConvenienceFormat(formatShorthand string) (string, error) {
	shorthand := strings.ToLower(formatShorthand)

	// Check if it's already a MIME type
	if strings.Contains(shorthand, "/") {
		return formatShorthand, nil
	}

	// Look up convenience mapping
	mimeType, ok := convenienceFormats[shorthand]
	if !ok {
		return "", utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			fmt.Sprintf("Unknown format shorthand: %s", formatShorthand)).
			WithContext("formatShorthand", formatShorthand).
			WithContext("supportedFormats", getSupportedShorthands()).
			Build())
	}

	return mimeType, nil
}

// GetAvailableFormats returns available export formats for a source MIME type
func GetAvailableFormats(sourceMimeType string) ([]string, error) {
	if !IsGoogleWorkspaceFile(sourceMimeType) {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Export is only supported for Google Workspace files").
			WithContext("sourceMimeType", sourceMimeType).
			Build())
	}

	formats, ok := exportFormats[sourceMimeType]
	if !ok {
		return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			fmt.Sprintf("No export formats available for MIME type: %s", sourceMimeType)).
			WithContext("sourceMimeType", sourceMimeType).
			Build())
	}

	return formats, nil
}

// ValidateExportFormat validates an export format against Google's reference mapping
func ValidateExportFormat(sourceMimeType, targetMimeType string) error {
	if !IsGoogleWorkspaceFile(sourceMimeType) {
		return utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			"Export is only supported for Google Workspace files").
			WithContext("sourceMimeType", sourceMimeType).
			Build())
	}

	availableFormats, ok := exportFormats[sourceMimeType]
	if !ok {
		return utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			fmt.Sprintf("No export formats available for file type: %s", sourceMimeType)).
			WithContext("sourceMimeType", sourceMimeType).
			Build())
	}

	// Check if target MIME type is supported
	for _, format := range availableFormats {
		if format == targetMimeType {
			return nil
		}
	}

	return utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
		fmt.Sprintf("Export format '%s' is not supported for source type '%s'", targetMimeType, sourceMimeType)).
		WithContext("sourceMimeType", sourceMimeType).
		WithContext("targetMimeType", targetMimeType).
		WithContext("availableFormats", availableFormats).
		Build())
}

// getSupportedShorthands returns a list of supported format shorthands
func getSupportedShorthands() []string {
	shorthands := make([]string, 0, len(convenienceFormats))
	for k := range convenienceFormats {
		shorthands = append(shorthands, k)
	}
	return shorthands
}

// GetMimeTypeForWorkspaceType returns the MIME type for a workspace type name
func GetMimeTypeForWorkspaceType(typeName string) (string, error) {
	switch strings.ToLower(typeName) {
	case "document", "doc", "docs":
		return MimeTypeGoogleDocs, nil
	case "spreadsheet", "sheet", "sheets":
		return MimeTypeGoogleSheets, nil
	case "presentation", "slide", "slides":
		return MimeTypeGoogleSlides, nil
	case "drawing":
		return MimeTypeGoogleDrawing, nil
	case "form":
		return MimeTypeGoogleForm, nil
	case "script":
		return MimeTypeGoogleScript, nil
	default:
		return "", utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
			fmt.Sprintf("Unknown workspace type: %s", typeName)).
			WithContext("typeName", typeName).
			Build())
	}
}
