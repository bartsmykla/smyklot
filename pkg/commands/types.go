package commands

// CommandType represents the type of command
type CommandType string

const (
	// CommandApprove represents the approve command
	CommandApprove CommandType = "approve"

	// CommandMerge represents the merge command (uses default or first allowed method)
	CommandMerge CommandType = "merge"

	// CommandSquash represents the squash merge command
	CommandSquash CommandType = "squash"

	// CommandRebase represents the rebase merge command
	CommandRebase CommandType = "rebase"

	// CommandUnapprove represents the unapprove command
	CommandUnapprove CommandType = "unapprove"

	// CommandCleanup represents the cleanup command
	CommandCleanup CommandType = "cleanup"

	// CommandHelp represents the help command
	CommandHelp CommandType = "help"

	// CommandUnknown represents an unknown or invalid command
	CommandUnknown CommandType = "unknown"
)

// Command represents a parsed command from a comment
type Command struct {
	// Type is the command type (approve, merge, or unknown)
	//
	// Deprecated: Use Commands field for multiple command support
	Type CommandType

	// Commands is the list of parsed command types
	Commands []CommandType

	// Raw is the original comment text
	Raw string

	// IsValid indicates whether the command was successfully parsed
	IsValid bool

	// Error contains any parsing error message
	Error string

	// WaitForCI indicates merge should be deferred until CI passes
	// Set when "after CI", "when green", "when checks pass" modifier is detected
	WaitForCI bool

	// RequiredChecksOnly indicates only required checks should be considered
	// Set when "required" modifier is detected (e.g., "after required CI")
	RequiredChecksOnly bool
}
