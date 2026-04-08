package platform

import (
	"os"
	"path/filepath"
)

// LocalFileSystem implements FileService using the host OS filesystem.
type LocalFileSystem struct {
	RootDir string
}

// NewLocalFileSystem creates a new LocalFileSystem rooted at the given path.
// If the directory does not exist, it is created automatically.
func NewLocalFileSystem(root string) (*LocalFileSystem, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	return &LocalFileSystem{RootDir: root}, nil
}

func (fs *LocalFileSystem) SaveFile(path, content string) error {
	full := fs.resolve(path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0644)
}

func (fs *LocalFileSystem) LoadFile(path string) (string, error) {
	data, err := os.ReadFile(fs.resolve(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (fs *LocalFileSystem) ListFiles(directory string) ([]FileEntry, error) {
	full := fs.resolve(directory)
	return walkDir(full, full)
}

func (fs *LocalFileSystem) CreateDirectory(path string) error {
	return os.MkdirAll(fs.resolve(path), 0755)
}

func (fs *LocalFileSystem) DeleteFile(path string) error {
	return os.RemoveAll(fs.resolve(path))
}

func (fs *LocalFileSystem) RenameFile(oldPath, newPath string) error {
	return os.Rename(fs.resolve(oldPath), fs.resolve(newPath))
}

func (fs *LocalFileSystem) GetRootDir() string {
	return fs.RootDir
}

func (fs *LocalFileSystem) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(fs.RootDir, filepath.Clean(path))
}

func walkDir(root, dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []FileEntry
	for _, e := range entries {
		relPath, _ := filepath.Rel(root, filepath.Join(dir, e.Name()))
		fe := FileEntry{
			Name:  e.Name(),
			Path:  relPath,
			IsDir: e.IsDir(),
		}
		if e.IsDir() {
			children, err := walkDir(root, filepath.Join(dir, e.Name()))
			if err == nil {
				fe.Items = children
			}
		}
		result = append(result, fe)
	}
	return result, nil
}
