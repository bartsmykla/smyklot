// Package main provides the GitHub Actions entrypoint for Smyklot bot.
//
// Smyklot automates PR approvals and merges based on CODEOWNERS permissions.
// It reads environment variables from GitHub Actions and executes commands
// like /approve and /merge based on user permissions defined in the
// .github/CODEOWNERS file.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bartsmykla/smyklot/pkg/commands"
	"github.com/bartsmykla/smyklot/pkg/feedback"
	"github.com/bartsmykla/smyklot/pkg/github"
	"github.com/bartsmykla/smyklot/pkg/permissions"
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
	rootCmd.Flags().StringVar(&config.Token, "token", "", "GitHub API token")
	rootCmd.Flags().StringVar(&config.CommentBody, "comment-body", "", "PR comment body")
	rootCmd.Flags().StringVar(&config.CommentID, "comment-id", "", "PR comment ID")
	rootCmd.Flags().StringVar(&config.PRNumber, "pr-number", "", "Pull request number")
	rootCmd.Flags().StringVar(&config.RepoOwner, "repo-owner", "", "Repository owner")
	rootCmd.Flags().StringVar(&config.RepoName, "repo-name", "", "Repository name")
	rootCmd.Flags().StringVar(&config.CommentAuthor, "comment-author", "", "Comment author username")
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

	// Create GitHub client
	client, err := github.NewClient(config.Token, "")
	if err != nil {
		return NewGitHubError(ErrGitHubClient, err)
	}

	// Get current working directory (repository root)
	repoPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Initialize permission checker
	checker, err := permissions.NewChecker(repoPath)
	if err != nil {
		return fmt.Errorf("failed to initialize permission checker: %w", err)
	}

	// Convert string IDs to integers
	prNum, err := strconv.Atoi(config.PRNumber)
	if err != nil {
		return NewInputError(ErrInvalidInput, config.PRNumber, "invalid PR number")
	}
	commentIDNum, err := strconv.Atoi(config.CommentID)
	if err != nil {
		return NewInputError(ErrInvalidInput, config.CommentID, "invalid comment ID")
	}

	// Check if the user has permission to execute this command
	canApprove, err := checker.CanApprove(config.CommentAuthor, "/")
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
		config.Token = os.Getenv("GITHUB_TOKEN")
	}

	if config.CommentBody == "" {
		config.CommentBody = os.Getenv("COMMENT_BODY")
	}

	if config.CommentID == "" {
		config.CommentID = os.Getenv("COMMENT_ID")
	}

	if config.PRNumber == "" {
		config.PRNumber = os.Getenv("PR_NUMBER")
	}

	if config.RepoOwner == "" {
		config.RepoOwner = os.Getenv("REPO_OWNER")
	}

	if config.RepoName == "" {
		config.RepoName = os.Getenv("REPO_NAME")
	}

	if config.CommentAuthor == "" {
		config.CommentAuthor = os.Getenv("COMMENT_AUTHOR")
	}

	return nil
}

// validateConfig validates that all required configuration is present.
func validateConfig() error {
	if config.Token == "" {
		return NewEnvVarError(ErrMissingEnvVar, "GITHUB_TOKEN")
	}

	if config.CommentBody == "" {
		return NewEnvVarError(ErrMissingEnvVar, "COMMENT_BODY")
	}

	if config.CommentID == "" {
		return NewEnvVarError(ErrMissingEnvVar, "COMMENT_ID")
	}

	if config.PRNumber == "" {
		return NewEnvVarError(ErrMissingEnvVar, "PR_NUMBER")
	}

	if config.RepoOwner == "" {
		return NewEnvVarError(ErrMissingEnvVar, "REPO_OWNER")
	}

	if config.RepoName == "" {
		return NewEnvVarError(ErrMissingEnvVar, "REPO_NAME")
	}

	if config.CommentAuthor == "" {
		return NewEnvVarError(ErrMissingEnvVar, "COMMENT_AUTHOR")
	}

	return nil
}

// handleUnauthorized posts feedback for unauthorized users.
func handleUnauthorized(
	client *github.Client,
	checker *permissions.Checker,
	prNum, commentID int,
) error {
	fb := feedback.NewUnauthorized(config.CommentAuthor, checker.GetApprovers())

	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionError); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}

// handleApprove handles the /approve command.
func handleApprove(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionEyes); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	// Approve the PR
	if err := client.ApprovePR(config.RepoOwner, config.RepoName, prNum); err != nil {
		return postApprovalFailure(client, prNum, commentID, err)
	}

	// Post success feedback
	fb := feedback.NewApprovalSuccess(config.CommentAuthor)
	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	// Add success reaction
	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionSuccess); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}

// postApprovalFailure posts failure feedback for failed approval.
func postApprovalFailure(client *github.Client, prNum, commentID int, approvalErr error) error {
	fb := feedback.NewApprovalFailed(approvalErr.Error())

	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionError); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return NewGitHubError(ErrApprovePR, approvalErr)
}

// handleMerge handles the /merge command.
func handleMerge(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionEyes); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	// Get PR info to check if it's mergeable
	info, err := client.GetPRInfo(config.RepoOwner, config.RepoName, prNum)
	if err != nil {
		return postMergeFailure(client, prNum, commentID, err)
	}

	// Check if PR is mergeable
	if !info.Mergeable {
		return postNotMergeable(client, prNum, commentID)
	}

	// Merge the PR
	if err := client.MergePR(config.RepoOwner, config.RepoName, prNum); err != nil {
		return postMergeFailure(client, prNum, commentID, err)
	}

	// Post success feedback
	fb := feedback.NewMergeSuccess(config.CommentAuthor)
	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	// Add success reaction
	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionSuccess); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}

// postMergeFailure posts failure feedback for a failed merge.
func postMergeFailure(client *github.Client, prNum, commentID int, mergeErr error) error {
	fb := feedback.NewMergeFailed(mergeErr.Error())

	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionError); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return NewGitHubError(ErrMergePR, mergeErr)
}

// postNotMergeable posts feedback when PR is not mergeable.
func postNotMergeable(client *github.Client, prNum, commentID int) error {
	fb := feedback.NewNotMergeable()

	if err := client.PostComment(config.RepoOwner, config.RepoName, prNum, fb.Message); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(config.RepoOwner, config.RepoName, commentID, github.ReactionWarning); err != nil {
		return NewGitHubError(ErrAddReaction, err)
	}

	return nil
}
