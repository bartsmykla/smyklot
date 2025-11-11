package feedback

import (
	"fmt"
	"strings"
)

// NewSuccess creates a success feedback
//
// Success feedback only uses emoji reaction (✅), no comment is posted.
func NewSuccess() *Feedback {
	return &Feedback{
		Type:    Success,
		Emoji:   "✅",
		Message: "",
	}
}

// NewUnauthorized creates an error feedback for unauthorized user
//
// The message includes:
//   - The username that attempted the action
//   - List of authorized approvers (if any)
//   - Suggestion to check OWNERS file
func NewUnauthorized(username string, approvers []string) *Feedback {
	var message string

	if len(approvers) == 0 {
		message = fmt.Sprintf(
			"❌ **Not Authorized**\n\n"+
				"User `%s` is not authorized to perform this action.\n\n"+
				"There are no approvers configured in the OWNERS file. "+
				"Please add approvers to the OWNERS file in the repository root.",
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

// NewInvalidCommand creates an error feedback for invalid command
//
// The message includes:
//   - The invalid command that was used
//   - List of valid commands
func NewInvalidCommand(command string) *Feedback {
	message := fmt.Sprintf(
		"❌ **Invalid Command**\n\n"+
			"Command `%s` is not recognized.\n\n"+
			"**Valid commands:**\n"+
			"- `/approve` or `@smyklot approve` - Approve the PR\n"+
			"- `/merge` or `@smyklot merge` - Merge the PR\n\n"+
			"Please use one of the valid commands above.",
		command,
	)

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
		Message: message,
	}
}

// NewAlreadyApproved creates a warning feedback for already approved PR
//
// The message indicates who already approved the PR
func NewAlreadyApproved(approver string) *Feedback {
	message := fmt.Sprintf(
		"⚠️ **Already Approved**\n\n"+
			"This PR has already been approved by `%s`.\n\n"+
			"No action taken.",
		approver,
	)

	return &Feedback{
		Type:    Warning,
		Emoji:   "⚠️",
		Message: message,
	}
}

// NewAlreadyMerged creates a warning feedback for already merged PR
func NewAlreadyMerged() *Feedback {
	message := "⚠️ **Already Merged**\n\n" +
		"This pull request has already been merged.\n\n" +
		"No action taken."

	return &Feedback{
		Type:    Warning,
		Emoji:   "⚠️",
		Message: message,
	}
}

// NewPRNotReady creates an error feedback for PR not ready to merge
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

// NewMergeConflict creates an error feedback for merge conflict
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

// NewNoOWNERSFile creates an error feedback for missing OWNERS file
func NewNoOWNERSFile() *Feedback {
	message := "❌ **No OWNERS File**\n\n" +
		"The OWNERS file was not found in the repository root.\n\n" +
		"Please create an OWNERS file with the list of approvers."

	return &Feedback{
		Type:    Error,
		Emoji:   "❌",
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
// Success feedback only uses emoji reactions
// Error and warning feedback require comments with details
func (f *Feedback) RequiresComment() bool {
	return f.Type != Success
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
