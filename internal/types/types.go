package types

// File represents a Google Drive file
type File struct {
	ID       string
	Name     string
	MimeType string
	Size     int64
	Parents  []string
}

// Folder represents a Google Drive folder
type Folder struct {
	ID      string
	Name    string
	Parents []string
}
