// Package feedback provides user-facing messages for Smyklot operations.
//
// It creates formatted feedback messages with emoji and Markdown text for
// success, error, and warning scenarios during PR approval and merge operations.
package feedback

import (
	"fmt"
	"strings"
)

// NewSuccess creates success feedback
//
//	uses only an emoji reaction (✅); no comment is posted
func NewSuccess() *Feedback {
	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: "",
	}
}

// NewUnauthorized creates error feedback for an unauthorized user
//
// The message includes:
//   - The username that attempted the action
//   - A list of authorized approvers (if any)
//   - A suggestion to check the CODEOWNERS file
func NewUnauthorized(username string, approvers []string) *Feedback {
	var message string

	if len(approvers) == 0 {
		message = fmt.Sprintf(
			"❌ **Not Authorized**\n\n"+
				"User `%s` is not authorized to perform this action.\n\n"+
				"No approvers are configured in the CODEOWNERS file. "+
				"Please add approvers to `.github/CODEOWNERS` in the repository root.",
			username,
		)
	} else {
		approverList := formatApproverList(approvers)
		message = fmt.Sprintf(
			"❌ **Not Authorized**\n\n"+
				"User `%s` is not authorized to perform this action.\n\n"+
				"**Authorized approvers:**\n%s\n\n"+
				"Please contact one of the approvers listed above.",
			username,
			approverList,
		)
	}

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewInvalidCommand creates error feedback for an invalid command
//
// The message includes:
//   - The invalid command that was used
//   - A list of valid commands
func NewInvalidCommand(command string) *Feedback {
	message := fmt.Sprintf(
		"❌ **Invalid Command**\n\n"+
			"Command `%s` is not recognized.\n\n"+
			"**Valid commands:**\n"+
			"- `/approve` or `@smyklot approve` - Approve the PR\n"+
			"- `/merge` or `@smyklot merge` - Merge the PR\n\n"+
			"Please use one of the valid commands listed above.",
		command,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewAlreadyApproved creates warning feedback for an already-approved PR
//
// The message indicates who already approved the PR
func NewAlreadyApproved(approver string) *Feedback {
	message := fmt.Sprintf(
		"⚠️ **Already Approved**\n\n"+
			"This pull request has already been approved by `%s`.\n\n"+
			"No action has been taken.",
		approver,
	)

	return &Feedback{
		Type:    Warning,
		Emoji:   "⚠️",
		Message: message,
	}
}

// NewAlreadyMerged creates warning feedback for an already-merged PR
func NewAlreadyMerged() *Feedback {
	message := "⚠️ **Already Merged**\n\n" +
		"This pull request has already been merged.\n\n" +
		"No action has been taken."

	return &Feedback{
		Type:    Warning,
		Emoji:   "⚠️",
		Message: message,
	}
}

// NewPRNotReady creates error feedback for a PR that is not ready to merge
//
// The reason parameter should describe why the PR is not ready
// (e.g., "CI checks failing", "required reviews not met")
func NewPRNotReady(reason string) *Feedback {
	message := fmt.Sprintf(
		"❌ **PR Not Ready**\n\n"+
			"This pull request is not ready to be merged.\n\n"+
			"**Reason:** %s\n\n"+
			"Please resolve the issues before attempting to merge.",
		reason,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewMergeConflict creates error feedback for a merge conflict
func NewMergeConflict() *Feedback {
	message := "❌ **Merge Conflict**\n\n" +
		"This pull request has conflicts with the base branch.\n\n" +
		"Please resolve the conflicts before attempting to merge."

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewNoCodeownersFile creates error feedback for a missing CODEOWNERS file
func NewNoCodeownersFile() *Feedback {
	message := "❌ **No CODEOWNERS File**\n\n" +
		"The CODEOWNERS file was not found at `.github/CODEOWNERS`.\n\n" +
		"Please create a CODEOWNERS file with global owners:\n" +
		"```\n" +
		"* @username1 @username2\n" +
		"```"

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewApprovalSuccess creates success feedback for a successful PR approval
//
// The message acknowledges the approver and indicates the approval was successful
// If quietSuccess is true, only an emoji reaction is used (no comment)
func NewApprovalSuccess(approver string, quietSuccess bool) *Feedback {
	message := ""
	if !quietSuccess {
		message = fmt.Sprintf(
			"✅ **PR Approved**\n\n"+
				"This pull request has been approved by `%s`.",
			approver,
		)
	}

	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: message,
	}
}

// NewApprovalFailed creates error feedback for a failed PR approval
//
// The reason parameter should describe why the approval failed
func NewApprovalFailed(reason string) *Feedback {
	message := fmt.Sprintf(
		"❌ **Approval Failed**\n\n"+
			"Failed to approve this pull request.\n\n"+
			"**Reason:** %s",
		reason,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewMergeSuccess creates success feedback for a successful PR merge
//
// The message acknowledges who merged the PR
// If quietSuccess is true, only an emoji reaction is used (no comment)
func NewMergeSuccess(author string, quietSuccess bool) *Feedback {
	message := ""
	if !quietSuccess {
		message = fmt.Sprintf(
			"✅ **PR Merged**\n\n"+
				"This pull request has been successfully merged by `%s`.",
			author,
		)
	}

	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: message,
	}
}

// NewMergeFailed creates error feedback for a failed PR merge
//
// The reason parameter should describe why the merge failed
func NewMergeFailed(reason string) *Feedback {
	message := fmt.Sprintf(
		"❌ **Merge Failed**\n\n"+
			"Failed to merge this pull request.\n\n"+
			"**Reason:** %s",
		reason,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewNotMergeable creates error feedback for a PR that cannot be merged
func NewNotMergeable() *Feedback {
	message := "❌ **PR Not Mergeable**\n\n" +
		"This pull request cannot be merged at this time.\n\n" +
		"Possible reasons:\n" +
		"- Merge conflicts exist\n" +
		"- Required checks have not passed\n" +
		"- Branch protection rules are not satisfied\n\n" +
		"Please resolve the issues before attempting to merge."

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewUnapproveSuccess creates success feedback for dismissing a review
//
// The message acknowledges who dismissed the review
// If quietSuccess is true, only an emoji reaction is used (no comment)
func NewUnapproveSuccess(author string, quietSuccess bool) *Feedback {
	message := ""
	if !quietSuccess {
		message = fmt.Sprintf(
			"✅ **Review Dismissed**\n\n"+
				"The approval has been dismissed by `%s`.",
			author,
		)
	}

	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: message,
	}
}

// NewUnapproveFailed creates error feedback for a failed review dismissal
//
// The reason parameter should describe why the dismissal failed
func NewUnapproveFailed(reason string) *Feedback {
	message := fmt.Sprintf(
		"❌ **Dismiss Failed**\n\n"+
			"Failed to dismiss the approval.\n\n"+
			"**Reason:** %s",
		reason,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewReactionApprovalSuccess creates success feedback for reaction-based approval
//
// The message acknowledges the approver and indicates the approval was triggered by reaction
// If quietReactions is true, only an emoji reaction is used (no comment)
func NewReactionApprovalSuccess(approver string, quietReactions bool) *Feedback {
	message := ""
	if !quietReactions {
		message = fmt.Sprintf(
			"✅ **PR Approved (via 👍 reaction)**\n\n"+
				"This pull request has been approved by `%s` using a 👍 reaction.",
			approver,
		)
	}

	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: message,
	}
}

// NewReactionMergeSuccess creates success feedback for reaction-based merge
//
// The message acknowledges who merged the PR via reaction
// If quietReactions is true, only an emoji reaction is used (no comment)
func NewReactionMergeSuccess(author string, quietReactions bool) *Feedback {
	message := ""
	if !quietReactions {
		message = fmt.Sprintf(
			"✅ **PR Merged (via 🚀 reaction)**\n\n"+
				"This pull request has been successfully merged by `%s` using a 🚀 reaction.",
			author,
		)
	}

	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: message,
	}
}

// NewHelp creates help feedback with usage instructions
func NewHelp() *Feedback {
	message := "ℹ️ **Smyklot Bot - Help**\n\n" +
		"I can help you manage pull requests through simple commands.\n\n" +
		"**Available Commands:**\n\n" +
		"**Approval Commands:**\n" +
		"- `/approve` or `@smyklot approve` or `approve` - Approve the PR\n" +
		"- `accept` or `lgtm` - Alternative ways to approve\n\n" +
		"**Merge Commands:**\n" +
		"- `/merge` or `@smyklot merge` or `merge` - Merge the PR\n\n" +
		"**Review Management:**\n" +
		"- `/unapprove` or `@smyklot unapprove` or `unapprove` - Dismiss your approval\n" +
		"- `disapprove` - Alternative way to unapprove\n\n" +
		"**Help:**\n" +
		"- `/help` or `@smyklot help` or `help` - Show this help message\n\n" +
		"**Command Formats:**\n" +
		"- **Slash commands**: `/approve`, `/merge`, `/help`\n" +
		"- **Mention commands**: `@smyklot approve`, `@smyklot merge`\n" +
		"- **Bare commands**: `approve`, `lgtm`, `merge`, `help`\n\n" +
		"**Multiple Commands:**\n" +
		"You can use multiple commands in one comment:\n" +
		"```\napprove\nmerge\n```\n" +
		"or `approve merge`\n\n" +
		"**Permissions:**\n" +
		"Only users listed in `.github/CODEOWNERS` can execute commands.\n\n" +
		"**Note:** All commands are case-insensitive."

	return &Feedback{
		Type:    Success,
		Emoji:   "ℹ️",
		Message: message,
	}
}

// String returns a string representation of the feedback
//
// For success: Returns emoji only
// For error/warning: Returns emoji + message
func (f *Feedback) String() string {
	if f.Message == "" {
		return f.Emoji
	}

	return fmt.Sprintf("%s\n\n%s", f.Emoji, f.Message)
}

// RequiresComment returns true if the feedback requires a comment
//
// Feedback with an empty message only uses emoji reactions
// Feedback with a message requires a comment to be posted
func (f *Feedback) RequiresComment() bool {
	return f.Message != ""
}

// formatApproverList formats a list of approvers as a Markdown bulleted list
func formatApproverList(approvers []string) string {
	if len(approvers) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, approver := range approvers {
		builder.WriteString(fmt.Sprintf("- `%s`\n", approver))
	}

	return strings.TrimSuffix(builder.String(), "\n")
}
