package config

const (
	// DefaultCommandPrefix is the default prefix for slash-style commands
	DefaultCommandPrefix = "/"
)

// Config represents the configuration for Smyklot
type Config struct {
	// QuietSuccess disables success comments (keeps reactions only)
	QuietSuccess bool `json:"quiet_success"`

	// AllowedCommands is a list of allowed command names
	// an Empty list means all commands are allowed
	AllowedCommands []string `json:"allowed_commands"`

	// CommandAliases maps aliases to command names,
	// For example, {"app": "approve", "a": "approve"}
	CommandAliases map[string]string `json:"command_aliases"`

	// CommandPrefix is the prefix for slash-style commands
	// The default is "/" (e.g., /approve)
	CommandPrefix string `json:"command_prefix"`

	// DisableMentions disables mention-style commands (@smyklot approve)
	DisableMentions bool `json:"disable_mentions"`

	// DisableBareCommands disables bare commands (approve, lgtm, merge)
	DisableBareCommands bool `json:"disable_bare_commands"`
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		QuietSuccess:        false,
		AllowedCommands:     []string{},
		CommandAliases:      make(map[string]string),
		CommandPrefix:       DefaultCommandPrefix,
		DisableMentions:     false,
		DisableBareCommands: false,
	}
}

// IsCommandAllowed checks if a command is allowed
// If AllowedCommands is empty, all commands are allowed
func (c *Config) IsCommandAllowed(command string) bool {
	if len(c.AllowedCommands) == 0 {
		return true
	}

	for _, allowed := range c.AllowedCommands {
		if allowed == command {
			return true
		}
	}

	return false
}

// ResolveAlias resolves a command alias to the actual command name
// If no alias exists, returns the original command
func (c *Config) ResolveAlias(command string) string {
	if resolved, ok := c.CommandAliases[command]; ok {
		return resolved
	}

	return command
}
