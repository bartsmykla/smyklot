package github

// MergeableState represents the merge state of a PR from GitHub REST API
type MergeableState string

const (
	// MergeableStateClean indicates PR can be merged
	MergeableStateClean MergeableState = "clean"

	// MergeableStateDirty indicates PR has conflicts
	MergeableStateDirty MergeableState = "dirty"

	// MergeableStateBlocked indicates PR is blocked by branch protection
	MergeableStateBlocked MergeableState = "blocked"

	// MergeableStateUnstable indicates PR has failing status checks
	MergeableStateUnstable MergeableState = "unstable"

	// MergeableStateUnknown indicates mergeability not yet computed
	MergeableStateUnknown MergeableState = "unknown"
)

// PRInfo contains information about a pull request
type PRInfo struct {
	// Number is the PR number
	Number int

	// State is the current state (open, closed, merged)
	State string

	// Mergeable indicates whether the PR can be merged (no conflicts)
	Mergeable bool

	// MergeableState provides detailed merge state (clean, dirty, blocked, unstable, unknown)
	MergeableState MergeableState

	// Author is the username of the PR author
	Author string

	// ApprovedBy contains usernames of approvers
	ApprovedBy []string

	// Title is the PR title
	Title string

	// Body is the PR description
	Body string

	// BaseBranch is the base branch (e.g. "main", "master")
	BaseBranch string
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

	// ReactionCleanup represents cleanup command (❤️)
	ReactionCleanup ReactionType = "heart"
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

	// LabelReactionCleanup indicates cleanup was triggered via ❤️ reaction
	LabelReactionCleanup = "smyklot:reaction-cleanup"
)

// MergeMethod represents the type of merge method to use
type MergeMethod string

const (
	// MergeMethodMerge creates a merge commit
	MergeMethodMerge MergeMethod = "merge"

	// MergeMethodSquash squashes all commits into one
	MergeMethodSquash MergeMethod = "squash"

	// MergeMethodRebase rebases and merges
	MergeMethodRebase MergeMethod = "rebase"
)
