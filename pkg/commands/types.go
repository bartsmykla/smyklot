package commands

// CommandType represents the type of command
type CommandType string

const (
	// CommandApprove represents the approve command
	CommandApprove CommandType = "approve"

	// CommandMerge represents the merge command
	CommandMerge CommandType = "merge"

	// CommandUnapprove represents the unapprove command
	CommandUnapprove CommandType = "unapprove"

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
}
