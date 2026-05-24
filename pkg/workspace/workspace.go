// Package workspace provides filesystem-backed storage for ClinLang
// notes (.cln files) and associated workspace metadata. It is the same
// abstraction used by the HTTP API server in both local and hosted
// deployment modes; the deployment-mode caller is responsible for
// selecting the appropriate root (per-user in hosted mode, single
// shared root in local mode) before constructing a LocalFileSystem.
//
// All public methods reject paths that resolve outside the configured
// root. There is no escape hatch.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileEntry is a single node in a workspace listing.
type FileEntry struct {
	Name  string      `json:"name"`
	Path  string      `json:"path"`
	IsDir bool        `json:"is_dir"`
	Items []FileEntry `json:"items,omitempty"`
}

// LocalFileSystem is a filesystem-backed workspace rooted at RootDir.
// All caller-supplied paths are treated as relative; absolute paths are
// re-interpreted as relative, and "../" sequences cannot climb above
// RootDir.
type LocalFileSystem struct {
	RootDir string
	absRoot string
}

// NewLocalFileSystem creates a new LocalFileSystem rooted at the given
// path. If the directory does not exist, it is created.
func NewLocalFileSystem(root string) (*LocalFileSystem, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &LocalFileSystem{RootDir: root, absRoot: abs}, nil
}

func (fs *LocalFileSystem) SaveFile(path, content string) error {
	full, err := fs.resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0644)
}

func (fs *LocalFileSystem) LoadFile(path string) (string, error) {
	full, err := fs.resolve(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (fs *LocalFileSystem) ListFiles(directory string) ([]FileEntry, error) {
	full, err := fs.resolve(directory)
	if err != nil {
		return nil, err
	}
	return walkDir(fs.absRoot, full)
}

func (fs *LocalFileSystem) CreateDirectory(path string) error {
	full, err := fs.resolve(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(full, 0755)
}

func (fs *LocalFileSystem) DeleteFile(path string) error {
	full, err := fs.resolve(path)
	if err != nil {
		return err
	}
	// Refuse to delete the root itself.
	if full == fs.absRoot {
		return fmt.Errorf("cannot delete workspace root")
	}
	return os.RemoveAll(full)
}

func (fs *LocalFileSystem) RenameFile(oldPath, newPath string) error {
	oldFull, err := fs.resolve(oldPath)
	if err != nil {
		return err
	}
	newFull, err := fs.resolve(newPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(newFull), 0755); err != nil {
		return err
	}
	return os.Rename(oldFull, newFull)
}

// GetRootDir returns the workspace root as originally configured.
func (fs *LocalFileSystem) GetRootDir() string {
	return fs.RootDir
}

// resolve maps a caller-supplied path to an absolute on-disk path inside
// RootDir. It rejects any input that would land outside RootDir.
//
//   - Absolute paths supplied by the caller are NOT honored — they are
//     re-interpreted as relative to RootDir.
//   - "../" sequences are collapsed by filepath.Clean before joining,
//     so the final path can never climb above RootDir.
//   - After joining we re-canonicalise with filepath.Abs and verify
//     the prefix; any mismatch is rejected.
func (fs *LocalFileSystem) resolve(rel string) (string, error) {
	rooted := filepath.Clean("/" + filepath.ToSlash(rel))
	full := filepath.Join(fs.RootDir, rooted)

	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", rel, err)
	}
	if abs != fs.absRoot && !strings.HasPrefix(abs, fs.absRoot+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes workspace root", rel)
	}
	return abs, nil
}

func walkDir(root, dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []FileEntry
	for _, e := range entries {
		fullPath := filepath.Join(dir, e.Name())
		relPath, _ := filepath.Rel(root, fullPath)
		// Normalise to forward slashes so the API output is the same
		// shape on all platforms.
		relPath = filepath.ToSlash(relPath)
		fe := FileEntry{
			Name:  e.Name(),
			Path:  relPath,
			IsDir: e.IsDir(),
		}
		if e.IsDir() {
			children, err := walkDir(root, fullPath)
			if err == nil {
				fe.Items = children
			}
		}
		result = append(result, fe)
	}
	return result, nil
}
