package github

// PRInfo contains information about a pull request
type PRInfo struct {
	// Number is the PR number
	Number int

	// State is the current state (open, closed, merged)
	State string

	// Mergeable indicates whether the PR can be merged
	Mergeable bool

	// Author is the username of the PR author
	Author string

	// ApprovedBy contains usernames of approvers
	ApprovedBy []string

	// Title is the PR title
	Title string

	// Body is the PR description
	Body string
}

// ReactionType represents the type of emoji reaction
type ReactionType string

const (
	// ReactionSuccess represents success (✅)
	ReactionSuccess ReactionType = "+1"

	// ReactionError represents error (❌)
	ReactionError ReactionType = "-1"

	// ReactionWarning represents warning (⚠️)
	ReactionWarning ReactionType = "confused"

	// ReactionEyes represents acknowledgment (👀)
	ReactionEyes ReactionType = "eyes"

	// ReactionApprove represents approve command (👍)
	ReactionApprove ReactionType = "+1"

	// ReactionMerge represents merge command (🚀)
	ReactionMerge ReactionType = "rocket"
)

// Reaction represents a reaction on a comment
type Reaction struct {
	// Type is the reaction type
	Type ReactionType

	// User is the username of the user who reacted
	User string
}

const (
	// LabelReactionApprove indicates PR was approved via 👍 reaction
	LabelReactionApprove = "smyklot:reaction-approve"

	// LabelReactionMerge indicates PR was merged via 🚀 reaction
	LabelReactionMerge = "smyklot:reaction-merge"
)
