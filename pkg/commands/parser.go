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
		Raw:      commentBody,
		Type:     CommandUnknown,
		Commands: []CommandType{},
		IsValid:  false,
	}

	if commentBody == "" || strings.TrimSpace(commentBody) == "" {
		return cmd, nil
	}

	// Collect all commands found in the text
	commandsFound := make(map[CommandType]bool)

	// Find all slash commands
	findSlashCommands(commentBody, cfg, commandsFound)

	// Find all mention commands if not disabled
	if !cfg.DisableMentions {
		findMentionCommands(commentBody, cfg, commandsFound)
	}

	// Find bare commands if not disabled (check each word)
	if !cfg.DisableBareCommands {
		findBareCommands(commentBody, cfg, commandsFound)
	}

	// Convert map to slice (deduplicated and ordered)
	commands := buildCommandList(commandsFound)

	// Populate the command struct
	if len(commands) > 0 {
		cmd.Commands = commands
		cmd.Type = commands[0] // For backward compatibility
		cmd.IsValid = true
	}

	return cmd, nil
}

// findSlashCommands finds all slash commands in the comment body
func findSlashCommands(commentBody string, cfg *config.Config, commandsFound map[CommandType]bool) {
	// Build slash command regex dynamically based on the prefix
	slashPattern := fmt.Sprintf(`(?i)%s(\w+)`, regexp.QuoteMeta(cfg.CommandPrefix))
	slashCommandRegex := regexp.MustCompile(slashPattern)

	slashMatches := slashCommandRegex.FindAllStringSubmatch(commentBody, -1)
	for _, matches := range slashMatches {
		if len(matches) > 1 {
			commandName := strings.ToLower(matches[1])
			if cmdType := resolveCommand(commandName, cfg); cmdType != CommandUnknown {
				commandsFound[cmdType] = true
			}
		}
	}
}

// findMentionCommands finds all mention commands in the comment body
func findMentionCommands(commentBody string, cfg *config.Config, commandsFound map[CommandType]bool) {
	mentionMatches := mentionCommandRegex.FindAllStringSubmatch(commentBody, -1)
	for _, matches := range mentionMatches {
		if len(matches) > 1 {
			commandName := strings.ToLower(matches[1])
			if cmdType := resolveCommand(commandName, cfg); cmdType != CommandUnknown {
				commandsFound[cmdType] = true
			}
		}
	}
}

// findBareCommands finds all bare commands in the comment body
func findBareCommands(commentBody string, cfg *config.Config, commandsFound map[CommandType]bool) {
	// Split by whitespace and newlines
	words := strings.Fields(commentBody)
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if cmdType := resolveCommand(word, cfg); cmdType != CommandUnknown {
			commandsFound[cmdType] = true
		}
	}
}

// buildCommandList converts the commands map to an ordered slice
func buildCommandList(commandsFound map[CommandType]bool) []CommandType {
	var commands []CommandType
	// Maintain consistent order: approve then merge
	if commandsFound[CommandApprove] {
		commands = append(commands, CommandApprove)
	}
	if commandsFound[CommandMerge] {
		commands = append(commands, CommandMerge)
	}

	return commands
}

// resolveCommand resolves aliases, validates the command, and returns the CommandType
// Returns CommandUnknown if the command is invalid or not allowed
func resolveCommand(commandName string, cfg *config.Config) CommandType {
	// Resolve alias
	commandName = cfg.ResolveAlias(commandName)

	cmdType, ok := validCommands[commandName]
	if !ok {
		return CommandUnknown
	}

	// Check if command is allowed
	if !cfg.IsCommandAllowed(commandName) {
		return CommandUnknown
	}

	return cmdType
}
