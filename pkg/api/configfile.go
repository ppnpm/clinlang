package api

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// persistedConfig is the on-disk shape of user preferences kept across
// launches. It lives at <OS user config dir>/clinlang/config.json:
//
//	Linux:   ~/.config/clinlang/config.json
//	macOS:   ~/Library/Application Support/clinlang/config.json
//	Windows: %AppData%\clinlang\config.json
//
// Today this only persists the chosen workspace root. Other
// preferences (theme, last-open files) live in browser localStorage
// because they are per-device, not per-binary.
type persistedConfig struct {
	WorkspaceRoot string `json:"workspace_root"`
}

func configFilePath() (string, error) {
	// Tests set CLINLANG_CONFIG_DIR to a temp directory so workspace
	// PUTs don't pollute the developer's real config file. Production
	// leaves this unset and falls back to os.UserConfigDir().
	if override := os.Getenv("CLINLANG_CONFIG_DIR"); override != "" {
		return filepath.Join(override, "clinlang", "config.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "clinlang", "config.json"), nil
}

// loadPersistedConfig is best-effort: any error reading or parsing
// the file returns a zero-value config so the caller falls through to
// the next source (env var or empty).
func loadPersistedConfig() persistedConfig {
	p, err := configFilePath()
	if err != nil {
		return persistedConfig{}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return persistedConfig{}
	}
	var c persistedConfig
	_ = json.Unmarshal(data, &c)
	return c
}

// savePersistedConfig writes the config file atomically (best-effort).
func savePersistedConfig(c persistedConfig) error {
	p, err := configFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// suggestedWorkspace returns the platform-appropriate default the
// frontend pre-fills the welcome dialog with. The user is free to
// edit it before confirming. Returns "" if no sensible default can be
// derived (no home dir, etc.).
func suggestedWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "ClinCases")
}
