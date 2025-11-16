// Package commands provides command parsing for Smyklot PR comments
//
// It supports parsing slash commands (/approve, /merge) and mention commands
// (@smyklot approve, @smyklot merge) from GitHub PR comment text
package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bartsmykla/smyklot/pkg/config"
)

var (
	// mentionCommandRegex matches @smyklot command patterns
	mentionCommandRegex = regexp.MustCompile(`(?i)@smyklot\s+(\w+)`)

	// validCommands maps command names to their corresponding types
	validCommands = map[string]CommandType{
		"approve": CommandApprove,
		"accept":  CommandApprove,
		"lgtm":    CommandApprove,
		"merge":   CommandMerge,
	}
)

// ParseCommand parses comment text and extracts a command if present
//
// Supported formats:
//   - /approve or /merge (slash commands with custom prefix)
//   - @smyklot approve or @smyklot merge (mention commands)
//   - approve, accept, lgtm, or merge (bare commands - exact match only)
//
// Commands are case-insensitive and can appear anywhere in the text
// If multiple commands are present, the first one found is returned
// Priority: slash commands > mention commands > bare commands
//
// Configuration options:
//   - CommandPrefix: Custom prefix for slash commands (default: "/")
//   - CommandAliases: Map aliases to command names (e.g., "app" -> "approve")
//   - DisableMentions: Disable mention-style commands
//   - DisableBareCommands: Disable bare commands (approve, lgtm, merge)
//   - AllowedCommands: Only allow specified commands (empty = all allowed)
//
// If cfg is nil, default configuration is used
func ParseCommand(commentBody string, cfg *config.Config) (Command, error) {
	// Use default config if nil
	if cfg == nil {
		cfg = config.Default()
	}

	cmd := Command{
		Raw:     commentBody,
		Type:    CommandUnknown,
		IsValid: false,
	}

	if commentBody == "" || strings.TrimSpace(commentBody) == "" {
		return cmd, nil
	}

	// Build slash command regex dynamically based on the prefix
	slashPattern := fmt.Sprintf(`(?i)^[\s]*%s(\w+)`, regexp.QuoteMeta(cfg.CommandPrefix))
	slashCommandRegex := regexp.MustCompile(slashPattern)

	// Try to match a slash command first (higher priority)
	if matches := slashCommandRegex.FindStringSubmatch(commentBody); len(matches) > 1 {
		commandName := strings.ToLower(matches[1])

		if processCommand(commandName, cfg, &cmd) {
			return cmd, nil
		}

		// Invalid slash command found - return unknown
		return cmd, nil
	}

	// Try to match a mention command if not disabled
	if !cfg.DisableMentions {
		if matches := mentionCommandRegex.FindStringSubmatch(commentBody); len(matches) > 1 {
			commandName := strings.ToLower(matches[1])

			if processCommand(commandName, cfg, &cmd) {
				return cmd, nil
			}

			// Invalid mention command found - return unknown
			return cmd, nil
		}
	}

	// Try to match a bare command if not disabled (lowest priority)
	// Bare commands must be exact matches (no extra text before or after)
	if !cfg.DisableBareCommands {
		trimmedBody := strings.TrimSpace(commentBody)
		commandName := strings.ToLower(trimmedBody)

		if processCommand(commandName, cfg, &cmd) {
			return cmd, nil
		}
	}

	// No command found
	return cmd, nil
}

// processCommand resolves aliases, validates the command, and updates the cmd struct
// Returns true if the command was valid and processed successfully
func processCommand(commandName string, cfg *config.Config, cmd *Command) bool {
	// Resolve alias
	commandName = cfg.ResolveAlias(commandName)

	cmdType, ok := validCommands[commandName]
	if !ok {
		return false
	}

	// Check if command is allowed
	if !cfg.IsCommandAllowed(commandName) {
		return false
	}

	cmd.Type = cmdType
	cmd.IsValid = true

	return true
}
