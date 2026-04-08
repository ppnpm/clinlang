package engine

// SpecialtyPlugin defines the interface for all domain-specific plugins.
type SpecialtyPlugin interface {
	GetName() string
	GetDescription() string                    // short human summary, e.g. "OB/GYN Specialty"
	GetCommandSummary() map[string]string      // command → one-line description
	GetCommands() map[string]ParserFunc
	InitData() any
}

var pluginsRegistry = make(map[string]SpecialtyPlugin)

// RegisterPlugin is called by plugins in their init() function to register themselves.
func RegisterPlugin(p SpecialtyPlugin) {
	pluginsRegistry[p.GetName()] = p
}

// GetPlugin returns a registered plugin by name, or nil if not found.
func GetPlugin(name string) SpecialtyPlugin {
	return pluginsRegistry[name]
}

// PluginInfo is a serialisable summary of a plugin for the frontend.
type PluginInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Commands    map[string]string `json:"commands"`
}

// ListPlugins returns metadata for all registered plugins.
func ListPlugins() []PluginInfo {
	result := make([]PluginInfo, 0, len(pluginsRegistry))
	for _, p := range pluginsRegistry {
		result = append(result, PluginInfo{
			Name:        p.GetName(),
			Description: p.GetDescription(),
			Commands:    p.GetCommandSummary(),
		})
	}
	return result
}

