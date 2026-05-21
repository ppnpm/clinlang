package engine

import "strings"

// ─────────────────────────────────────────────────────────────────────────────
// PLUGIN INTERFACE
// ─────────────────────────────────────────────────────────────────────────────

// SpecialtyPlugin defines the mandatory contract for all specialty plugins.
// Every plugin must implement this interface.
type SpecialtyPlugin interface {
	GetName() string
	GetDescription() string                 // short human summary, e.g. "OB/GYN Specialty"
	GetCommandSummary() map[string]string   // command → one-line description (for docs/autocomplete)
	GetCommands() map[string]ParserFunc     // new standalone commands added by this plugin
	InitData() any                          // allocates and returns the plugin's data struct
}

// ─────────────────────────────────────────────────────────────────────────────
// OPTIONAL INTERFACE: CommandExtendable
// ─────────────────────────────────────────────────────────────────────────────

// TokenExtFunc is the handler signature for an inline token extension.
//
// When a core command parser encounters an unrecognized token, it consults
// the CommandTokenRegistry. If a matching extension is found, this function
// is called with the raw token and the ClinicalCase.
//
// Rules for plugin authors:
//   - Return true  → you handled the token; parser continues silently.
//   - Return false → you did not handle it; parser falls through to warning.
//   - ONLY write into c.SpecialtyData — never into c.Patient or other core fields.
//   - Always type-assert c.SpecialtyData safely (check the bool).
type TokenExtFunc func(token string, c *ClinicalCase) bool

// CommandExtendable is an OPTIONAL interface plugins may implement.
//
// It allows a plugin to register inline token handlers for any existing core
// command without modifying the core engine source.
//
// Return value layout:
//
//	map[commandName]map[tokenPrefix]TokenExtFunc
//	     "pt"           "ga:"          handler
//	     "vitals"       "fhr:"         handler
//
// Token prefixes are matched case-insensitively with strings.HasPrefix.
// Use a colon suffix (e.g. "ga:") to clearly delimit prefix from value.
//
// Example usage in a .cln file (when @profile is declared first):
//
//	@profile obgyn
//	pt 28F wt65 ga:34w       ← "ga:" handled by obgyn's pt extension
//	vitals bp120/75 fhr:142  ← "fhr:" handled by obgyn's vitals extension
type CommandExtendable interface {
	GetCommandTokens() map[string]map[string]TokenExtFunc
}

// ─────────────────────────────────────────────────────────────────────────────
// COMMAND TOKEN REGISTRY (per-parse session)
// ─────────────────────────────────────────────────────────────────────────────

// CommandTokenRegistry holds all active token extensions for one parse session.
//
// It is created fresh inside ParseString() — there is NO global state.
// This guarantees that extensions registered for an OB/GYN case never bleed
// into a Cardiology case parsed in the same process.
type CommandTokenRegistry struct {
	exts map[string]map[string]TokenExtFunc // cmd → lowercase_prefix → handler
}

// NewCommandTokenRegistry creates an empty registry for a single parse session.
func NewCommandTokenRegistry() *CommandTokenRegistry {
	return &CommandTokenRegistry{exts: make(map[string]map[string]TokenExtFunc)}
}

// Register adds a token extension for the given command name and token prefix.
// Both cmd and prefix are normalised to lowercase on storage.
func (r *CommandTokenRegistry) Register(cmd, prefix string, fn TokenExtFunc) {
	cmd = strings.ToLower(cmd)
	prefix = strings.ToLower(prefix)
	if r.exts[cmd] == nil {
		r.exts[cmd] = make(map[string]TokenExtFunc)
	}
	r.exts[cmd][prefix] = fn
}

// RegisterUnique adds a token extension only if command+prefix is not taken.
// Returns true if inserted, false if it already exists.
func (r *CommandTokenRegistry) RegisterUnique(cmd, prefix string, fn TokenExtFunc) bool {
	cmd = strings.ToLower(cmd)
	prefix = strings.ToLower(prefix)
	if r.exts[cmd] == nil {
		r.exts[cmd] = make(map[string]TokenExtFunc)
	}
	if _, exists := r.exts[cmd][prefix]; exists {
		return false
	}
	r.exts[cmd][prefix] = fn
	return true
}

// Try looks up and calls a registered extension for the given command and token.
//
// Returns true if an extension matched and handled the token.
// Returns false (no-op) if:
//   - the receiver is nil
//   - no extensions are registered for this command
//   - no prefix matches the token
//
// Safe to call with a nil receiver — core parsers always call Try even when
// no plugin is active, so no nil-check is required at each call site.
func (r *CommandTokenRegistry) Try(cmd, token string, c *ClinicalCase) bool {
	if r == nil {
		return false
	}
	exts, ok := r.exts[strings.ToLower(cmd)]
	if !ok {
		return false
	}
	lower := strings.ToLower(token)
	for prefix, fn := range exts {
		if strings.HasPrefix(lower, prefix) {
			return fn(token, c)
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// PLUGIN STORE
// ─────────────────────────────────────────────────────────────────────────────

var pluginsRegistry = make(map[string]SpecialtyPlugin)

// RegisterPlugin is called by plugins in their init() function to self-register.
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
