package engine

// SpecialtyPlugin defines the interface for all domain-specific plugins.
type SpecialtyPlugin interface {
	GetName() string
	GetCommands() map[string]ParserFunc
	InitData() interface{}
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
