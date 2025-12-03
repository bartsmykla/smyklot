package feedback

// Type represents the type of feedback
type Type string

const (
	// Success represents a successful operation
	// Response: Emoji reaction only (✅)
	Success Type = "success"

	// Error represents a failed operation
	// Response: Emoji reaction (❌) + detailed comment explaining the error
	Error Type = "error"

	// Warning represents a non-critical issue
	// Response: Emoji reaction (⚠️) + informative comment
	Warning Type = "warning"

	// Pending represents an operation waiting for external condition
	// Response: Emoji reaction (⏳) + informative comment
	Pending Type = "pending"
)

// Feedback represents a response to a command
//
// Success feedback: Emoji reaction only
// Error/Warning feedback: Emoji reaction + comment with details
type Feedback struct {
	// Type is the feedback type (success, error, or warning)
	Type Type

	// Emoji is the emoji used for the reaction
	Emoji string

	// Message is the optional comment message
	// Empty for success, populated for errors and warnings
	Message string
}
