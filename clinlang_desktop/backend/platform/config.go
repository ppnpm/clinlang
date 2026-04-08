package platform

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	WorkspaceDir string `json:"workspace_dir"`
	Theme        string `json:"theme"`
}

var (
	configPath string
	mu         sync.Mutex
)

func init() {
	home, err := os.UserHomeDir()
	if err == nil {
		configPath = filepath.Join(home, ".clinlang_config.json")
	} else {
		configPath = ".clinlang_config.json"
	}
}

// LoadConfig reads the configuration from disk.
func LoadConfig() Config {
	mu.Lock()
	defer mu.Unlock()

	cfg := Config{
		Theme: "system",
	}
	
	data, err := os.ReadFile(configPath)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	return cfg
}

func SaveConfig(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
