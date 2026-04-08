package platform

// FileService defines the storage interface.
// The desktop implementation uses the local OS filesystem.
// A future web implementation could use an API or browser storage.
type FileService interface {
	SaveFile(path, content string) error
	LoadFile(path string) (string, error)
	ListFiles(directory string) ([]FileEntry, error)
	CreateDirectory(path string) error
	DeleteFile(path string) error
	RenameFile(oldPath, newPath string) error
}

// FileEntry represents a single file or folder in the file tree.
type FileEntry struct {
	Name  string      `json:"name"`
	Path  string      `json:"path"`
	IsDir bool        `json:"is_dir"`
	Items []FileEntry `json:"items,omitempty"`
}
