package commands

// CommandType represents the type of command
type CommandType string

const (
	// CommandApprove represents the approve command
	CommandApprove CommandType = "approve"

	// CommandMerge represents the merge command
	CommandMerge CommandType = "merge"

	// CommandUnknown represents an unknown/invalid command
	CommandUnknown CommandType = "unknown"
)

// Command represents a parsed command from a comment
type Command struct {
	// Type is the command type (approve, merge, unknown)
	Type CommandType

	// Raw is the original comment text
	Raw string

	// IsValid indicates if the command was successfully parsed
	IsValid bool

	// Error contains any parsing error message
	Error string
}
