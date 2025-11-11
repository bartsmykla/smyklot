// Package main provides the GitHub Actions entrypoint for Smyklot bot.
//
// Smyklot automates PR approvals and merges based on CODEOWNERS permissions.
// It reads environment variables from GitHub Actions and executes commands
// like /approve and /merge based on user permissions defined in the
// .github/CODEOWNERS file.
package main

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bartsmykla/smyklot/pkg/commands"
	"github.com/bartsmykla/smyklot/pkg/feedback"
	"github.com/bartsmykla/smyklot/pkg/github"
	"github.com/bartsmykla/smyklot/pkg/permissions"
)

const (
	envGitHubToken    = "GITHUB_TOKEN" //nolint:gosec // Environment variable name, not a credential
	envCommentBody    = "COMMENT_BODY"
	envCommentID      = "COMMENT_ID"
	envPRNumber       = "PR_NUMBER"
	envRepoOwner      = "REPO_OWNER"
	envRepoName       = "REPO_NAME"
	envCommentAuthor  = "COMMENT_AUTHOR"
	rootPath          = "/"
	emptyBaseURL      = ""
	flagToken         = "token"
	flagCommentBody   = "comment-body"
	flagCommentID     = "comment-id"
	flagPRNumber      = "pr-number"
	flagRepoOwner     = "repo-owner"
	flagRepoName      = "repo-name"
	flagCommentAuthor = "comment-author"
	descToken         = "GitHub API token" //nolint:gosec // Flag description, not a credential
	descCommentBody   = "PR comment body"
	descCommentID     = "PR comment ID"
	descPRNumber      = "Pull request number"
	descRepoOwner     = "Repository owner"
	descRepoName      = "Repository name"
	descCommentAuthor = "Comment author username"
	errInvalidPRNum   = "invalid PR number"
	errInvalidComment = "invalid comment ID"
)

// Config holds the runtime configuration for the action.
type Config struct {
	Token         string
	CommentBody   string
	CommentID     string
	PRNumber      string
	RepoOwner     string
	RepoName      string
	CommentAuthor string
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "smyklot-github-action",
	Short: "GitHub Actions bot for automated PR approvals and merges",
	Long: `Smyklot is a GitHub Actions bot that enables automated PR approvals
and merges based on CODEOWNERS permissions.

It reads environment variables from GitHub Actions and executes
commands like /approve and /merge based on user permissions.`,
	RunE: run,
}

func init() {
	// Define CLI flags that can override environment variables
	rootCmd.Flags().StringVar(&config.Token, flagToken, "", descToken)
	rootCmd.Flags().StringVar(
		&config.CommentBody, flagCommentBody, "", descCommentBody,
	)
	rootCmd.Flags().StringVar(&config.CommentID, flagCommentID, "", descCommentID)
	rootCmd.Flags().StringVar(&config.PRNumber, flagPRNumber, "", descPRNumber)
	rootCmd.Flags().StringVar(&config.RepoOwner, flagRepoOwner, "", descRepoOwner)
	rootCmd.Flags().StringVar(&config.RepoName, flagRepoName, "", descRepoName)
	rootCmd.Flags().StringVar(
		&config.CommentAuthor, flagCommentAuthor, "", descCommentAuthor,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(_ *cobra.Command, _ []string) error {
	// Load configuration from environment variables if not provided via flags
	if err := loadConfig(); err != nil {
		return err
	}

	// Validate required configuration
	if err := validateConfig(); err != nil {
		return err
	}

	// Parse the command from the comment
	parsedCmd, err := commands.ParseCommand(config.CommentBody)
	if err != nil {
		// Not a valid command, ignore silently
		return nil
	}

	// Create a GitHub client
	client, err := github.NewClient(config.Token, emptyBaseURL)
	if err != nil {
		return NewGitHubError(ErrGitHubClient, err)
	}

	// Get the current working directory (repository root)
	repoPath, err := os.Getwd()
	if err != nil {
		return NewGitHubError(ErrGetWorkingDirectory, err)
	}

	// Initialize permission checker
	checker, err := permissions.NewChecker(repoPath)
	if err != nil {
		return NewGitHubError(ErrInitPermissions, err)
	}

	// Convert string IDs to integers
	prNum, err := strconv.Atoi(config.PRNumber)
	if err != nil {
		return NewInputError(ErrInvalidInput, config.PRNumber, errInvalidPRNum)
	}

	commentIDNum, err := strconv.Atoi(config.CommentID)
	if err != nil {
		return NewInputError(ErrInvalidInput, config.CommentID, errInvalidComment)
	}

	// Check if the user has permission to execute this command
	canApprove, err := checker.CanApprove(config.CommentAuthor, rootPath)
	if err != nil {
		return NewGitHubError(ErrPermissionCheck, err)
	}

	// Handle unauthorized users
	if !canApprove {
		return handleUnauthorized(client, checker, prNum, commentIDNum)
	}

	// Execute the command based on type
	switch parsedCmd.Type {
	case commands.CommandApprove:
		return handleApprove(client, prNum, commentIDNum)
	case commands.CommandMerge:
		return handleMerge(client, prNum, commentIDNum)
	default:
		// Unknown command type, ignore
		return nil
	}
}

// loadConfig loads configuration from environment variables if not set via flags.
func loadConfig() error {
	if config.Token == "" {
		config.Token = os.Getenv(envGitHubToken)
	}

	if config.CommentBody == "" {
		config.CommentBody = os.Getenv(envCommentBody)
	}

	if config.CommentID == "" {
		config.CommentID = os.Getenv(envCommentID)
	}

	if config.PRNumber == "" {
		config.PRNumber = os.Getenv(envPRNumber)
	}

	if config.RepoOwner == "" {
		config.RepoOwner = os.Getenv(envRepoOwner)
	}

	if config.RepoName == "" {
		config.RepoName = os.Getenv(envRepoName)
	}

	if config.CommentAuthor == "" {
		config.CommentAuthor = os.Getenv(envCommentAuthor)
	}

	return nil
}

// validateConfig validates that all required configuration is present.
func validateConfig() error {
	if config.Token == "" {
		return NewEnvVarError(ErrMissingEnvVar, envGitHubToken)
	}

	if config.CommentBody == "" {
		return NewEnvVarError(ErrMissingEnvVar, envCommentBody)
	}

	if config.CommentID == "" {
		return NewEnvVarError(ErrMissingEnvVar, envCommentID)
	}

	if config.PRNumber == "" {
		return NewEnvVarError(ErrMissingEnvVar, envPRNumber)
	}

	if config.RepoOwner == "" {
		return NewEnvVarError(ErrMissingEnvVar, envRepoOwner)
	}

	if config.RepoName == "" {
		return NewEnvVarError(ErrMissingEnvVar, envRepoName)
	}

	if config.CommentAuthor == "" {
		return NewEnvVarError(ErrMissingEnvVar, envCommentAuthor)
	}

	return nil
}

// postFeedback posts a comment and adds a reaction to a PR.
func postFeedback(
	client *github.Client,
	prNum, commentID int,
	message string,
	reaction github.ReactionType,
) error {
	if err := client.PostComment(
		config.RepoOwner,
		config.RepoName,
		prNum,
		message,
	); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(
		config.RepoOwner,
		config.RepoName,
		commentID,
		reaction,
	); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}

// addReaction adds a reaction to a comment.
func addReaction(
	client *github.Client,
	commentID int,
	reaction github.ReactionType,
) error {
	if err := client.AddReaction(
		config.RepoOwner,
		config.RepoName,
		commentID,
		reaction,
	); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}

// postOperationFailure posts failure feedback for a failed operation.
func postOperationFailure(
	client *github.Client,
	prNum, commentID int,
	operationErr error,
	feedbackFunc func(string) *feedback.Feedback,
	sentinelErr error,
) error {
	fb := feedbackFunc(operationErr.Error())

	if err := postFeedback(
		client,
		prNum,
		commentID,
		fb.Message,
		github.ReactionError,
	); err != nil {
		return err
	}

	return NewGitHubError(sentinelErr, operationErr)
}

// handleUnauthorized posts feedback for unauthorized users.
func handleUnauthorized(
	client *github.Client,
	checker *permissions.Checker,
	prNum, commentID int,
) error {
	fb := feedback.NewUnauthorized(config.CommentAuthor, checker.GetApprovers())

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionError)
}

// handleApprove handles the /approve command.
func handleApprove(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := addReaction(client, commentID, github.ReactionEyes); err != nil {
		return err
	}

	// Approve the PR
	if err := client.ApprovePR(config.RepoOwner, config.RepoName, prNum); err != nil {
		return postOperationFailure(
			client,
			prNum,
			commentID,
			err,
			feedback.NewApprovalFailed,
			ErrApprovePR,
		)
	}

	// Post-success feedback
	fb := feedback.NewApprovalSuccess(config.CommentAuthor)

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionSuccess)
}

// handleMerge handles the /merge command.
func handleMerge(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := addReaction(client, commentID, github.ReactionEyes); err != nil {
		return err
	}

	// Get PR info to check if it's mergeable
	info, err := client.GetPRInfo(config.RepoOwner, config.RepoName, prNum)
	if err != nil {
		return postOperationFailure(
			client,
			prNum,
			commentID,
			err,
			feedback.NewMergeFailed,
			ErrMergePR,
		)
	}

	// Check if PR is mergeable
	if !info.Mergeable {
		return postNotMergeable(client, prNum, commentID)
	}

	// Merge the PR
	if err := client.MergePR(config.RepoOwner, config.RepoName, prNum); err != nil {
		return postOperationFailure(
			client,
			prNum,
			commentID,
			err,
			feedback.NewMergeFailed,
			ErrMergePR,
		)
	}

	// Post-success feedback
	fb := feedback.NewMergeSuccess(config.CommentAuthor)

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionSuccess)
}

// postNotMergeable posts feedback when PR is not mergeable.
func postNotMergeable(client *github.Client, prNum, commentID int) error {
	fb := feedback.NewNotMergeable()

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionWarning)
}
