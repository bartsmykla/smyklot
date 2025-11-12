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

	"github.com/jferrl/go-githubauth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bartsmykla/smyklot/pkg/commands"
	"github.com/bartsmykla/smyklot/pkg/config"
	"github.com/bartsmykla/smyklot/pkg/feedback"
	"github.com/bartsmykla/smyklot/pkg/github"
	"github.com/bartsmykla/smyklot/pkg/permissions"
)

const (
	envGitHubToken         = "GITHUB_TOKEN" //nolint:gosec // Environment variable name, not a credential
	envCommentBody         = "COMMENT_BODY"
	envCommentID           = "COMMENT_ID"
	envPRNumber            = "PR_NUMBER"
	envRepoOwner           = "REPO_OWNER"
	envRepoName            = "REPO_NAME"
	envCommentAuthor       = "COMMENT_AUTHOR"
	envGitHubAppPrivateKey = "GITHUB_APP_PRIVATE_KEY" //nolint:gosec // Environment variable name, not a credential
	envGitHubAppClientID   = "GITHUB_APP_CLIENT_ID"   //nolint:gosec // Environment variable name, not a credential
	envGitHubAppID         = "GITHUB_APP_ID"          //nolint:gosec // Environment variable name, not a credential
	envInstallationID      = "GITHUB_INSTALLATION_ID"
	rootPath               = "/"
	emptyBaseURL           = ""
	flagToken              = "token"
	flagCommentBody        = "comment-body"
	flagCommentID          = "comment-id"
	flagPRNumber           = "pr-number"
	flagRepoOwner          = "repo-owner"
	flagRepoName           = "repo-name"
	flagCommentAuthor      = "comment-author"
	descToken              = "GitHub API token" //nolint:gosec // Flag description, not a credential
	descCommentBody        = "PR comment body"
	descCommentID          = "PR comment ID"
	descPRNumber           = "Pull request number"
	descRepoOwner          = "Repository owner"
	descRepoName           = "Repository name"
	descCommentAuthor      = "Comment author username"
	errInvalidPRNum        = "invalid PR number"
	errInvalidComment      = "invalid comment ID"
	errInvalidInstallID    = "invalid installation ID"
)

// RuntimeConfig holds the runtime configuration for the action
type RuntimeConfig struct {
	Token               string
	CommentBody         string
	CommentID           string
	PRNumber            string
	RepoOwner           string
	RepoName            string
	CommentAuthor       string
	GitHubAppPrivateKey string
	GitHubAppClientID   string
	GitHubAppID         string
	InstallationID      string
}

var (
	runtimeConfig RuntimeConfig
	botConfig     *config.Config
	v             *viper.Viper
)

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
	// Initialize Viper
	v = viper.New()
	config.SetupViper(v)

	// Define CLI flags that can override environment variables
	rootCmd.Flags().StringVar(&runtimeConfig.Token, flagToken, "", descToken)
	rootCmd.Flags().StringVar(
		&runtimeConfig.CommentBody, flagCommentBody, "", descCommentBody,
	)
	rootCmd.Flags().StringVar(&runtimeConfig.CommentID, flagCommentID, "", descCommentID)
	rootCmd.Flags().StringVar(&runtimeConfig.PRNumber, flagPRNumber, "", descPRNumber)
	rootCmd.Flags().StringVar(&runtimeConfig.RepoOwner, flagRepoOwner, "", descRepoOwner)
	rootCmd.Flags().StringVar(&runtimeConfig.RepoName, flagRepoName, "", descRepoName)
	rootCmd.Flags().StringVar(
		&runtimeConfig.CommentAuthor, flagCommentAuthor, "", descCommentAuthor,
	)

	// Bind Viper config flags
	rootCmd.Flags().Bool(config.KeyQuietSuccess, false, "Disable success comments (emoji only)")
	rootCmd.Flags().StringSlice(config.KeyAllowedCommands, []string{}, "Allowed commands (empty = all)")
	rootCmd.Flags().StringToString(config.KeyCommandAliases, map[string]string{}, "Command aliases (JSON)")
	rootCmd.Flags().String(config.KeyCommandPrefix, config.DefaultCommandPrefix, "Command prefix")
	rootCmd.Flags().Bool(config.KeyDisableMentions, false, "Disable mention-style commands")

	// Bind flags to Viper
	bindViperFlags([]string{
		config.KeyQuietSuccess,
		config.KeyAllowedCommands,
		config.KeyCommandAliases,
		config.KeyCommandPrefix,
		config.KeyDisableMentions,
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// bindViperFlags binds multiple flags to Viper
func bindViperFlags(keys []string) {
	for _, key := range keys {
		_ = v.BindPFlag(key, rootCmd.Flags().Lookup(key))
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
	parsedCmd, err := commands.ParseCommand(runtimeConfig.CommentBody, botConfig)
	if err != nil {
		// Not a valid command, ignore silently
		return nil
	}

	// Get GitHub App installation token if configured
	token := runtimeConfig.Token
	installationToken, err := getInstallationToken()
	if err != nil {
		return err
	}

	if installationToken != "" {
		token = installationToken
	}

	// Create a GitHub client
	client, err := github.NewClient(token, emptyBaseURL)
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
	prNum, err := strconv.Atoi(runtimeConfig.PRNumber)
	if err != nil {
		return NewInputError(ErrInvalidInput, runtimeConfig.PRNumber, errInvalidPRNum)
	}

	commentIDNum, err := strconv.Atoi(runtimeConfig.CommentID)
	if err != nil {
		return NewInputError(ErrInvalidInput, runtimeConfig.CommentID, errInvalidComment)
	}

	// Check if the user has permission to execute this command
	canApprove, err := checker.CanApprove(runtimeConfig.CommentAuthor, rootPath)
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

// loadConfig loads configuration from environment variables if not set via flags
func loadConfig() error {
	loadEnvIfEmpty(&runtimeConfig.Token, envGitHubToken)
	loadEnvIfEmpty(&runtimeConfig.CommentBody, envCommentBody)
	loadEnvIfEmpty(&runtimeConfig.CommentID, envCommentID)
	loadEnvIfEmpty(&runtimeConfig.PRNumber, envPRNumber)
	loadEnvIfEmpty(&runtimeConfig.RepoOwner, envRepoOwner)
	loadEnvIfEmpty(&runtimeConfig.RepoName, envRepoName)
	loadEnvIfEmpty(&runtimeConfig.CommentAuthor, envCommentAuthor)
	loadEnvIfEmpty(&runtimeConfig.GitHubAppPrivateKey, envGitHubAppPrivateKey)
	loadEnvIfEmpty(&runtimeConfig.GitHubAppClientID, envGitHubAppClientID)
	loadEnvIfEmpty(&runtimeConfig.GitHubAppID, envGitHubAppID)
	loadEnvIfEmpty(&runtimeConfig.InstallationID, envInstallationID)

	// Load bot configuration from Viper
	botConfig = config.LoadFromViper(v)

	return nil
}

// loadEnvIfEmpty loads environment variable into target if target is empty
func loadEnvIfEmpty(target *string, envVar string) {
	if *target == "" {
		*target = os.Getenv(envVar)
	}
}

// validateConfig validates that all required configuration is present
func validateConfig() error {
	requiredFields := []struct {
		value  string
		envVar string
	}{
		{runtimeConfig.Token, envGitHubToken},
		{runtimeConfig.CommentBody, envCommentBody},
		{runtimeConfig.CommentID, envCommentID},
		{runtimeConfig.PRNumber, envPRNumber},
		{runtimeConfig.RepoOwner, envRepoOwner},
		{runtimeConfig.RepoName, envRepoName},
		{runtimeConfig.CommentAuthor, envCommentAuthor},
	}

	for _, field := range requiredFields {
		if field.value == "" {
			return NewEnvVarError(ErrMissingEnvVar, field.envVar)
		}
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
		runtimeConfig.RepoOwner,
		runtimeConfig.RepoName,
		prNum,
		message,
	); err != nil {
		return NewGitHubError(ErrPostComment, err)
	}

	if err := client.AddReaction(
		runtimeConfig.RepoOwner,
		runtimeConfig.RepoName,
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
		runtimeConfig.RepoOwner,
		runtimeConfig.RepoName,
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
	fb := feedback.NewUnauthorized(runtimeConfig.CommentAuthor, checker.GetApprovers())

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionError)
}

// handleApprove handles the /approve command.
func handleApprove(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := addReaction(client, commentID, github.ReactionEyes); err != nil {
		return err
	}

	// Approve the PR
	if err := client.ApprovePR(runtimeConfig.RepoOwner, runtimeConfig.RepoName, prNum); err != nil {
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
	fb := feedback.NewApprovalSuccess(runtimeConfig.CommentAuthor, botConfig.QuietSuccess)

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionSuccess)
}

// handleMerge handles the /merge command.
func handleMerge(client *github.Client, prNum, commentID int) error {
	// Add eyes reaction to acknowledge
	if err := addReaction(client, commentID, github.ReactionEyes); err != nil {
		return err
	}

	// Get PR info to check if it's mergeable
	info, err := client.GetPRInfo(runtimeConfig.RepoOwner, runtimeConfig.RepoName, prNum)
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
	if err := client.MergePR(runtimeConfig.RepoOwner, runtimeConfig.RepoName, prNum); err != nil {
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
	fb := feedback.NewMergeSuccess(runtimeConfig.CommentAuthor, botConfig.QuietSuccess)

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionSuccess)
}

// postNotMergeable posts feedback when PR is not mergeable.
func postNotMergeable(client *github.Client, prNum, commentID int) error {
	fb := feedback.NewNotMergeable()

	return postFeedback(client, prNum, commentID, fb.Message, github.ReactionWarning)
}

// getInstallationToken generates a GitHub App installation token if credentials are provided.
//
// Returns an empty string if GitHub App credentials are not configured.
// Returns the token on success.
func getInstallationToken() (string, error) {
	// Check if GitHub App credentials are provided
	if runtimeConfig.GitHubAppPrivateKey == "" || runtimeConfig.InstallationID == "" {
		return "", nil
	}

	// Determine which ID to use (ClientID is preferred, fallback to AppID)
	clientID := runtimeConfig.GitHubAppClientID
	if clientID == "" {
		clientID = runtimeConfig.GitHubAppID
	}

	if clientID == "" {
		return "", nil
	}

	// Convert installation ID to int64
	installationID, err := strconv.ParseInt(runtimeConfig.InstallationID, 10, 64)
	if err != nil {
		return "", NewInputError(ErrInvalidInput, runtimeConfig.InstallationID, errInvalidInstallID)
	}

	// Create GitHub App JWT token source
	appTokenSource, err := githubauth.NewApplicationTokenSource(
		clientID,
		[]byte(runtimeConfig.GitHubAppPrivateKey),
	)
	if err != nil {
		return "", NewGitHubError(ErrGitHubAppAuth, err)
	}

	// Create the installation token source
	installationTokenSource := githubauth.NewInstallationTokenSource(
		installationID,
		appTokenSource,
	)

	// Get the installation token
	token, err := installationTokenSource.Token()
	if err != nil {
		return "", NewGitHubError(ErrGitHubAppAuth, err)
	}

	return token.AccessToken, nil
}
