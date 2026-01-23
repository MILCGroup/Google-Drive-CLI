package pkg

// DriveService defines the public interface for Drive operations
type DriveService interface {
	List(query string) ([]FileInfo, error)
	Upload(path string, parentID string) (*FileInfo, error)
	Download(fileID string, destPath string) error
	Delete(fileID string) error
}

// FileInfo represents public file information
type FileInfo struct {
	ID       string
	Name     string
	MimeType string
	Size     int64
}
