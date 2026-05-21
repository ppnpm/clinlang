package main

import (
	"context"
	// "fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn" // registers via init()
	"clinlang_desktop/backend/autocomplete"
	"clinlang_desktop/backend/platform"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DocEntry represents a documentation file.
type DocEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// App is the main Wails application struct.
// It exposes methods to the frontend via Wails bindings.
type App struct {
	ctx context.Context
	fs  *platform.LocalFileSystem
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load user config
	cfg := platform.LoadConfig()

	root := cfg.WorkspaceDir
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		root = filepath.Join(home, "ClinCases")
	}

	fs, err := platform.NewLocalFileSystem(root)
	if err != nil {
		panic("Failed to initialize file storage: " + err.Error())
	}
	a.fs = fs
}

// ---- File Management Methods (exposed to frontend) ----

// SaveFile saves a .cln file at the given relative path.
func (a *App) SaveFile(path, content string) error {
	return a.fs.SaveFile(path, content)
}

// LoadFile reads a file and returns its contents.
func (a *App) LoadFile(path string) (string, error) {
	return a.fs.LoadFile(path)
}

// ListFiles returns the folder tree for the given directory.
func (a *App) ListFiles(directory string) ([]platform.FileEntry, error) {
	return a.fs.ListFiles(directory)
}

// CreateDirectory creates a new folder.
func (a *App) CreateDirectory(path string) error {
	return a.fs.CreateDirectory(path)
}

// DeleteFile removes a file or directory.
func (a *App) DeleteFile(path string) error {
	return a.fs.DeleteFile(path)
}

// RenameFile renames a file.
func (a *App) RenameFile(oldPath, newPath string) error {
	return a.fs.RenameFile(oldPath, newPath)
}

// GetRootDir returns the absolute path to the ClinCases directory.
func (a *App) GetRootDir() string {
	return a.fs.GetRootDir()
}

// DuplicateFile creates a copy of a file with a suffix.
func (a *App) DuplicateFile(path string) error {
	content, err := a.fs.LoadFile(path)
	if err != nil {
		return err
	}
	
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	newPath := base + "_copy" + ext
	
	// Ensure unique name if _copy already exists
	counter := 1
	for {
		if _, err := os.Stat(filepath.Join(a.fs.GetRootDir(), newPath)); os.IsNotExist(err) {
			break
		}
		newPath = base + "_copy_" + string(rune('0'+counter)) + ext
		counter++
		if counter > 9 { break } // safety break
	}

	return a.fs.SaveFile(newPath, content)
}

// ToggleFullscreen toggles the window fullscreen state.
func (a *App) ToggleFullscreen() {
	if runtime.WindowIsFullscreen(a.ctx) {
		runtime.WindowUnfullscreen(a.ctx)
	} else {
		runtime.WindowFullscreen(a.ctx)
	}
}

// IsFullscreen returns the current fullscreen state.
func (a *App) IsFullscreen() bool {
	return runtime.WindowIsFullscreen(a.ctx)
}

// AddImage opens a file dialog, copies the selected image to the workspace's images/ folder,
// and returns the relative path for the frontend.
func (a *App) AddImage() (string, error) {
	opts := runtime.OpenDialogOptions{
		Title: "Add Image",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images (*.png;*.jpg;*.jpeg;*.gif;*.webp)", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.webp"},
		},
	}
	
	sourcePath, err := runtime.OpenFileDialog(a.ctx, opts)
	if err != nil || sourcePath == "" {
		return "", err
	}

	// Ensure images directory exists
	imgDir := filepath.Join(a.fs.GetRootDir(), "images")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return "", err
	}

	// Create unique destination filename
	ext := filepath.Ext(sourcePath)
	destName := time.Now().Format("20060102_150405") + ext
	destPath := filepath.Join(imgDir, destName)

	// Copy the file
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return "images/" + destName, nil
}

// SelectFolder opens a dialog to select a workspace directory and switches to it.
func (a *App) SelectFolder() string {
	opts := runtime.OpenDialogOptions{
		Title: "Select Workspace Directory",
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, opts)
	if err == nil && path != "" {
		cfg := platform.LoadConfig()
		cfg.WorkspaceDir = path
		platform.SaveConfig(cfg)

		if newFs, err := platform.NewLocalFileSystem(path); err == nil {
			a.fs = newFs
		}
		return path
	}
	return ""
}

// ---- Clinical Engine Methods (backed by clinlang/pkg/engine) ----

// ParseClinLang takes raw .cln text and returns a fully parsed engine.ClinicalCase.
func (a *App) ParseClinLang(raw string) engine.ClinicalCase {
	return engine.ParseString(raw)
}

// FormatSOAP takes raw .cln text and returns a formatted SOAP note string.
func (a *App) FormatSOAP(raw string) string {
	c := engine.ParseString(raw)
	return engine.FormatSOAP(c)
}

// FormatJSON takes raw .cln text and returns a formatted JSON string.
func (a *App) FormatJSON(raw string) string {
	c := engine.ParseString(raw)
	return engine.FormatJSON(c)
}

// FormatMarkdown takes raw .cln text and returns a "beautifully written" markdown string.
func (a *App) FormatMarkdown(raw string) string {
	c := engine.ParseString(raw)
	return engine.FormatMarkdown(c)
}


// GetSuggestions returns autocomplete suggestions for a given command and query.
func (a *App) GetSuggestions(command, query string) []autocomplete.Suggestion {
	return autocomplete.GetSuggestions(command, query)
}

// SearchDrugs returns drug names matching the given prefix via the engine's drug list.
func (a *App) SearchDrugs(prefix string) []string {
	return engine.SearchDrugs(prefix)
}

// GetPlugins returns metadata about all registered specialty plugins.
func (a *App) GetPlugins() []engine.PluginInfo {
	return engine.ListPlugins()
}

// GetDocsList returns a list of files in the docs/ directory.
func (a *App) GetDocsList() ([]DocEntry, error) {
	// Original docs directory is in the root of the workspace
	docsDir := "../docs"

	files, err := os.ReadDir(docsDir)
	if err != nil {
		return nil, err
	}

	var docs []DocEntry
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			docs = append(docs, DocEntry{
				Name: f.Name(),
				Path: filepath.Join(docsDir, f.Name()),
			})
		}
	}
	return docs, nil
}

// GetDocContent reads a documentation file by name.
func (a *App) GetDocContent(name string) (string, error) {
	docsDir := "../docs"

	content, err := os.ReadFile(filepath.Join(docsDir, name))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

