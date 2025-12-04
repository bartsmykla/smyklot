// Package commands provides command parsing for Smyklot PR comments
//
// It supports parsing slash commands (/approve, /merge, /unapprove) and
// mention commands (@smyklot approve, @smyklot merge, @smyklot unapprove)
// from GitHub PR comment text
package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/smykla-labs/smyklot/pkg/config"
)

var (
	// mentionCommandRegex matches @smyklot command patterns
	mentionCommandRegex = regexp.MustCompile(`(?i)@smyklot\s+(\w+)`)

	// waitForCIRegex matches "after CI" modifier phrases
	// Patterns: "after CI", "after checks", "when green", "when CI passes",
	// "when checks pass", "once CI passes", "once checks pass"
	// Supports optional "required" modifier: "after required CI"
	waitForCIRegex = regexp.MustCompile(`(?i)\b(?:after|when|once)\s+(?:required\s+)?(?:CI|CD|GHA|checks?|green|workflows?|github\s+actions?)(?:\s+(?:pass(?:es)?|are?\s+green|finish(?:es)?|complete(?:s)?))?\b`)

	// requiredChecksRegex matches "required" modifier in CI phrases
	// Patterns: "required CI", "required checks", etc.
	requiredChecksRegex = regexp.MustCompile(`(?i)\b(?:after|when|once)\s+required\s+(?:CI|CD|GHA|checks?|workflows?|github\s+actions?)\b`)

	// validCommands maps command names to their corresponding types
	validCommands = map[string]CommandType{
		"approve":    CommandApprove,
		"accept":     CommandApprove,
		"lgtm":       CommandApprove,
		"merge":      CommandMerge,
		"squash":     CommandSquash,
		"rebase":     CommandRebase,
		"unapprove":  CommandUnapprove,
		"disapprove": CommandUnapprove,
		"cleanup":    CommandCleanup,
		"help":       CommandHelp,
	}

	// ErrContradictingCommands indicates contradicting commands were found
	ErrContradictingCommands = errors.New("contradicting commands found: cannot use approve/merge with unapprove")

	// ErrCleanupWithOtherCommands indicates cleanup was used with other commands
	ErrCleanupWithOtherCommands = errors.New("cleanup command cannot be combined with other commands")
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

	// Check for contradicting commands
	if hasContradictingCommands(commands) {
		cmd.Error = ErrContradictingCommands.Error()
		return cmd, ErrContradictingCommands
	}

	// Check if cleanup is combined with other commands
	if hasCleanupWithOtherCommands(commands) {
		cmd.Error = ErrCleanupWithOtherCommands.Error()
		return cmd, ErrCleanupWithOtherCommands
	}

	// Populate the command struct
	if len(commands) > 0 {
		cmd.Commands = commands
		cmd.Type = commands[0] // For backward compatibility
		cmd.IsValid = true

		// Check for "after CI" modifier - only applies to merge commands
		if hasMergeCommand(commands) && detectWaitForCIModifier(commentBody) {
			cmd.WaitForCI = true

			// Check for "required checks only" modifier
			if detectRequiredChecksModifier(commentBody) {
				cmd.RequiredChecksOnly = true
			}
		}
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
// Uses natural language detection to avoid matching commands in regular sentences
func findBareCommands(commentBody string, cfg *config.Config, commandsFound map[CommandType]bool) {
	// Process each line separately to avoid matching in regular sentences
	lines := strings.Split(commentBody, "\n")
	for _, line := range lines {
		// Skip lines that look like natural language sentences
		if looksLikeNaturalLanguage(line) {
			continue
		}

		// Only process command-heavy lines
		if isCommandHeavyLine(line, cfg) {
			extractCommandsFromLine(line, cfg, commandsFound)
		}
	}
}

// looksLikeNaturalLanguage detects if a line looks like natural language
// rather than a command. Returns true if the line contains patterns that
// indicate it's a sentence rather than a command list.
func looksLikeNaturalLanguage(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}

	// Indicators of natural language
	naturalLanguageWords := map[string]bool{
		// Articles
		"a": true, "an": true, "the": true,
		// Common verbs		"is": true, "are": true, "was": true, "were": true,
		"have": true, "has": true, "had": true,
		"think": true, "should": true, "would": true, "could": true,
		"will": true, "can": true, "do": true, "does": true,
		"looks": true, "seems": true, "appears": true,
		// Pronouns
		"i": true, "you": true, "we": true, "they": true,
		"this": true, "that": true, "these": true, "those": true,
		// Prepositions
		"in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "from": true, "of": true,
		"after": true, "before": true,
		// Question words
		"what": true, "when": true, "where": true, "why": true,
		"how": true, "who": true, "which": true,
		// Conjunctions
		"but": true, "or": true, "so": true, "if": true,
		// Other common words
		"let": true, "lets": true, "let's": true,
		"someone": true, "anyone": true,
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return false
	}

	// Single words are never natural language (they're commands)
	if len(words) == 1 {
		return false
	}

	// Check for sentence patterns
	firstWord := strings.ToLower(strings.Trim(words[0], ".,!?;:"))

	// Starts with capital and first word is natural language indicator
	if len(words[0]) > 0 && words[0][0] >= 'A' && words[0][0] <= 'Z' {
		if naturalLanguageWords[firstWord] || len(words) > 3 {
			return true
		}
	}

	// Contains question mark
	if strings.Contains(line, "?") {
		return true
	}

	// Count natural language indicators
	nlWordCount := 0
	for _, word := range words {
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:"))
		if naturalLanguageWords[cleanWord] {
			nlWordCount++
		}
	}

	// For short lines (2-3 words), require higher threshold to avoid false positives
	// For longer lines, use lower threshold
	threshold := 0.6
	if len(words) > 3 {
		threshold = 0.4
	}

	// If enough words are natural language indicators, it's likely a sentence
	if float64(nlWordCount)/float64(len(words)) > threshold {
		return true
	}

	return false
}

// isCommandHeavyLine checks if a line is "command-heavy" (>66% commands/fillers/CI-terms)
func isCommandHeavyLine(line string, cfg *config.Config) bool {
	fillerWords := map[string]bool{
		"and": true, "please": true, "this": true, "the": true, "it": true,
	}

	// CI-related technical terms that indicate command context
	ciRelatedWords := map[string]bool{
		"ci": true, "cd": true, "gha": true, "checks": true, "check": true,
		"workflows": true, "workflow": true, "green": true,
		"github": true, "actions": true, "action": true,
		"after": true, "when": true, "once": true,
		"pass": true, "passes": true, "passing": true,
		"finish": true, "finishes": true, "complete": true, "completes": true,
		"required": true,
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return false
	}

	commandWordCount := 0
	totalNonMentionWords := 0

	for _, word := range words {
		cleanWord := strings.ToLower(strings.TrimSpace(word))
		cleanWord = strings.Trim(cleanWord, ".,!?;:")

		// Skip @ mentions in the word count
		if strings.HasPrefix(word, "@") {
			continue
		}

		totalNonMentionWords++

		// Count as command-related if it's a command, filler, or CI-related term
		if resolveCommand(cleanWord, cfg) != CommandUnknown || fillerWords[cleanWord] || ciRelatedWords[cleanWord] {
			commandWordCount++
		}
	}

	if totalNonMentionWords == 0 {
		return false
	}

	// Only process this line if > 66% of words are commands/fillers/CI-terms
	return float64(commandWordCount)/float64(totalNonMentionWords) > 0.66
}

// extractCommandsFromLine extracts all commands from a line
// Single-letter commands are only matched if ALL words are commands/fillers
func extractCommandsFromLine(line string, cfg *config.Config, commandsFound map[CommandType]bool) {
	words := strings.Fields(line)
	onlyCommandsAndFillers := isLineOnlyCommandsAndFillers(line, cfg)

	for _, word := range words {
		cleanWord := strings.ToLower(strings.TrimSpace(word))
		cleanWord = strings.Trim(cleanWord, ".,!?;:")

		// Skip @ mentions
		if strings.HasPrefix(word, "@") {
			continue
		}

		cmdType := resolveCommand(cleanWord, cfg)
		if cmdType == CommandUnknown {
			continue
		}

		// Skip single-letter commands unless line is only commands/fillers
		if len(cleanWord) == 1 && !onlyCommandsAndFillers {
			continue
		}

		commandsFound[cmdType] = true
	}
}

// isLineOnlyCommandsAndFillers checks if line contains ONLY commands, fillers, and CI terms
func isLineOnlyCommandsAndFillers(line string, cfg *config.Config) bool {
	fillerWords := map[string]bool{
		"and": true, "please": true, "this": true, "the": true, "it": true,
	}

	// CI-related technical terms
	ciRelatedWords := map[string]bool{
		"ci": true, "cd": true, "gha": true, "checks": true, "check": true,
		"workflows": true, "workflow": true, "green": true,
		"github": true, "actions": true, "action": true,
		"after": true, "when": true, "once": true,
		"pass": true, "passes": true, "passing": true,
		"finish": true, "finishes": true, "complete": true, "completes": true,
		"required": true,
	}

	words := strings.Fields(line)
	for _, word := range words {
		cleanWord := strings.ToLower(strings.TrimSpace(word))
		cleanWord = strings.Trim(cleanWord, ".,!?;:")

		// Skip @ mentions
		if strings.HasPrefix(word, "@") {
			continue
		}

		// If it's not a command, filler, or CI-related word, return false
		if resolveCommand(cleanWord, cfg) == CommandUnknown && !fillerWords[cleanWord] && !ciRelatedWords[cleanWord] {
			return false
		}
	}

	return true
}

// buildCommandList converts the commands map to an ordered slice
func buildCommandList(commandsFound map[CommandType]bool) []CommandType {
	var commands []CommandType
	// Maintain consistent order: approve, merge, squash, rebase, unapprove, cleanup, help
	if commandsFound[CommandApprove] {
		commands = append(commands, CommandApprove)
	}
	if commandsFound[CommandMerge] {
		commands = append(commands, CommandMerge)
	}
	if commandsFound[CommandSquash] {
		commands = append(commands, CommandSquash)
	}
	if commandsFound[CommandRebase] {
		commands = append(commands, CommandRebase)
	}
	if commandsFound[CommandUnapprove] {
		commands = append(commands, CommandUnapprove)
	}
	if commandsFound[CommandCleanup] {
		commands = append(commands, CommandCleanup)
	}
	if commandsFound[CommandHelp] {
		commands = append(commands, CommandHelp)
	}

	return commands
}

// hasContradictingCommands checks if the command list contains contradicting commands
// Returns true if both unapprove and (approve or merge) are present
func hasContradictingCommands(commands []CommandType) bool {
	hasUnapprove := false
	hasApproveOrMerge := false

	for _, cmd := range commands {
		if cmd == CommandUnapprove {
			hasUnapprove = true
		}
		if cmd == CommandApprove || cmd == CommandMerge {
			hasApproveOrMerge = true
		}
	}

	return hasUnapprove && hasApproveOrMerge
}

// hasCleanupWithOtherCommands checks if cleanup is combined with other commands
// Returns true if cleanup is present with any other command
func hasCleanupWithOtherCommands(commands []CommandType) bool {
	hasCleanup := false
	hasOtherCommand := false

	for _, cmd := range commands {
		if cmd == CommandCleanup {
			hasCleanup = true
		} else {
			hasOtherCommand = true
		}
	}

	return hasCleanup && hasOtherCommand
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

	// Check if unapprove is disabled
	if cmdType == CommandUnapprove && cfg.DisableUnapprove {
		return CommandUnknown
	}

	// Check if command is allowed
	if !cfg.IsCommandAllowed(commandName) {
		return CommandUnknown
	}

	return cmdType
}

// detectWaitForCIModifier checks if the comment contains "after CI" modifier phrases
// Matches patterns like: "after CI", "when green", "when checks pass",
// "once CI passes", "after checks", etc.
func detectWaitForCIModifier(text string) bool {
	return waitForCIRegex.MatchString(text)
}

// detectRequiredChecksModifier checks if the comment specifies "required checks only"
// Matches patterns like: "after required CI", "when required checks pass", etc.
func detectRequiredChecksModifier(text string) bool {
	return requiredChecksRegex.MatchString(text)
}

// isMergeCommand checks if a command type is a merge-related command
func isMergeCommand(cmdType CommandType) bool {
	return cmdType == CommandMerge || cmdType == CommandSquash || cmdType == CommandRebase
}

// hasMergeCommand checks if any merge-related command is present
func hasMergeCommand(commands []CommandType) bool {
	for _, cmd := range commands {
		if isMergeCommand(cmd) {
			return true
		}
	}

	return false
}
