package commands

import (
	"regexp"
	"strings"
)

var (
	// slashCommandRegex matches /command patterns
	slashCommandRegex = regexp.MustCompile(`(?i)^[\s]*/(\w+)`)

	// mentionCommandRegex matches @smyklot command patterns
	mentionCommandRegex = regexp.MustCompile(`(?i)@smyklot\s+(\w+)`)

	// validCommands maps command names to their types
	validCommands = map[string]CommandType{
		"approve": CommandApprove,
		"merge":   CommandMerge,
	}
)

// ParseCommand parses a comment text and extracts a command if present
//
// Supported formats:
//   - /approve or /merge (slash commands)
//   - @smyklot approve or @smyklot merge (mention commands)
//
// Commands are case-insensitive and can appear anywhere in the text.
// If multiple commands are present, the first one found is returned.
// Slash commands take priority over mention commands.
func ParseCommand(commentBody string) (Command, error) {
	cmd := Command{
		Raw:     commentBody,
		Type:    CommandUnknown,
		IsValid: false,
	}

	if commentBody == "" || strings.TrimSpace(commentBody) == "" {
		return cmd, nil
	}

	// Try to match slash command first (priority)
	if matches := slashCommandRegex.FindStringSubmatch(commentBody); len(matches) > 1 {
		commandName := strings.ToLower(matches[1])
		if cmdType, ok := validCommands[commandName]; ok {
			cmd.Type = cmdType
			cmd.IsValid = true
			return cmd, nil
		}
		// Invalid slash command - return unknown
		return cmd, nil
	}

	// Try to match mention command
	if matches := mentionCommandRegex.FindStringSubmatch(commentBody); len(matches) > 1 {
		commandName := strings.ToLower(matches[1])
		if cmdType, ok := validCommands[commandName]; ok {
			cmd.Type = cmdType
			cmd.IsValid = true
			return cmd, nil
		}
		// Invalid mention command - return unknown
		return cmd, nil
	}

	// No command found
	return cmd, nil
}
