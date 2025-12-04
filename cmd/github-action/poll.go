package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/smykla-labs/smyklot/pkg/config"
	"github.com/smykla-labs/smyklot/pkg/feedback"
	"github.com/smykla-labs/smyklot/pkg/github"
	"github.com/smykla-labs/smyklot/pkg/permissions"
)

const (
	flagPollRepo  = "repo"
	flagPollToken = "token"
	descPollRepo  = "Repository in format owner/name"
	descPollToken = "GitHub API token" //nolint:gosec // Flag description, not a credential
)

var pollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Poll reactions on open PRs",
	Long: `Poll reactions on all open pull requests and process them.

This command fetches all open PRs, checks reactions on all comments,
and processes reaction-based commands (approve, merge, cleanup).

Designed to be run on a schedule (cron) to enable reaction support
in GitHub Actions without webhooks.`,
	RunE: runPoll,
}

func init() {
	pollCmd.Flags().StringP(flagPollRepo, "r", "", descPollRepo)
	pollCmd.Flags().StringP(flagPollToken, "t", "", descPollToken)

	rootCmd.AddCommand(pollCmd)
}

func runPoll(cmd *cobra.Command, _ []string) error {
	// Create context from command
	ctx := cmd.Context()

	// Create Viper instance
	v := viper.New()
	config.SetupViper(v)

	// Create runtime config for GitHub App auth
	rc := &RuntimeConfig{}
	loadEnvIfEmpty(&rc.Token, envGitHubToken)
	loadEnvIfEmpty(&rc.GitHubAppPrivateKey, envGitHubAppPrivateKey)
	loadEnvIfEmpty(&rc.GitHubAppClientID, envGitHubAppClientID)
	loadEnvIfEmpty(&rc.GitHubAppID, envGitHubAppID)
	loadEnvIfEmpty(&rc.InstallationID, envInstallationID)

	// Load bot configuration
	bc, err := loadPollBotConfig(v)
	if err != nil {
		return err
	}

	// Get configuration from flags and environment
	repo, token, err := getPollConfig(cmd, rc)
	if err != nil {
		return err
	}

	// Parse repository owner and name
	repoOwner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	// Setup GitHub client and permission checker
	client, checker, err := setupPollClients(ctx, token, repoOwner, repoName)
	if err != nil {
		return err
	}

	// Poll and process all open PRs
	return pollAllPRs(ctx, client, checker, bc, repoOwner, repoName)
}

// loadPollBotConfig loads bot configuration from JSON config and Viper
func loadPollBotConfig(v *viper.Viper) (*config.Config, error) {
	// Load JSON configuration from SMYKLOT_CONFIG if present
	if err := config.LoadJSONConfig(v); err != nil {
		return nil, NewConfigError(ErrConfigLoad, err)
	}

	// Load bot configuration from Viper
	return config.LoadFromViper(v), nil
}

// getPollConfig retrieves repo and token from flags or environment
func getPollConfig(cmd *cobra.Command, rc *RuntimeConfig) (string, string, error) {
	repo, err := cmd.Flags().GetString(flagPollRepo)
	if err != nil {
		return "", "", err
	}

	token, err := cmd.Flags().GetString(flagPollToken)
	if err != nil {
		return "", "", err
	}

	// Get repo from environment if not provided via flag
	if repo == "" {
		owner := os.Getenv(envRepoOwner)
		name := os.Getenv(envRepoName)
		if owner == "" || name == "" {
			return "", "", fmt.Errorf("repository not specified (use --repo or REPO_OWNER/REPO_NAME env vars)")
		}
		repo = fmt.Sprintf("%s/%s", owner, name)
	}

	// Get token from environment if not provided via flag
	if token == "" {
		token = os.Getenv(envGitHubToken)
		if token == "" {
			// Try GitHub App auth
			installationToken, err := getInstallationToken(rc)
			if err != nil {
				return "", "", err
			}
			if installationToken != "" {
				token = installationToken
			}
		}
	}

	if token == "" {
		return "", "", fmt.Errorf("GitHub token not specified (use --token or GITHUB_TOKEN env var)")
	}

	return repo, token, nil
}

// parseRepo parses owner and name from repo string
func parseRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format (expected owner/name, got %q)", repo)
	}
	return parts[0], parts[1], nil
}

// setupPollClients creates GitHub client and permission checker
func setupPollClients(
	ctx context.Context,
	token, repoOwner, repoName string,
) (*github.Client, *permissions.Checker, error) {
	// Create GitHub client
	client, err := github.NewClient(token, emptyBaseURL)
	if err != nil {
		return nil, nil, NewGitHubError(ErrGitHubClient, err)
	}

	// Fetch CODEOWNERS (returns empty string if not found)
	codeownersContent, err := client.GetCodeowners(ctx, repoOwner, repoName)
	if err != nil {
		return nil, nil, NewGitHubError(ErrGetCodeowners, err)
	}

	// Log if CODEOWNERS is missing
	if codeownersContent == "" {
		fmt.Println("CODEOWNERS file not found, defaulting to repository admin permissions")
	}

	// Initialize permission checker
	checker, err := permissions.NewCheckerFromContent(codeownersContent, client)
	if err != nil {
		return nil, nil, NewGitHubError(ErrInitPermissions, err)
	}

	return client, checker, nil
}

// pollAllPRs polls and processes reactions on all open PRs
func pollAllPRs(
	ctx context.Context,
	client *github.Client,
	checker *permissions.Checker,
	bc *config.Config,
	repoOwner, repoName string,
) error {
	fmt.Printf("Polling PR reactions in %s/%s\n", repoOwner, repoName)

	// Get all open PRs
	prs, err := client.GetOpenPRs(ctx, repoOwner, repoName)
	if err != nil {
		return NewGitHubError(ErrGetCodeowners, err)
	}

	if len(prs) == 0 {
		fmt.Println("No open PRs found")
		return nil
	}

	fmt.Printf("Found %d open PR(s)\n", len(prs))

	// Process reactions on each PR
	for _, pr := range prs {
		if err := processPR(ctx, client, checker, bc, repoOwner, repoName, pr); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
		}
	}

	// Process pending-ci PRs (waiting for CI to pass before merge)
	if err := processPendingCIPRs(ctx, client, bc, repoOwner, repoName, prs); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: failed to process pending-ci PRs: %v\n", err)
	}

	fmt.Println("\nPolling complete")

	return nil
}

// processPR processes reactions on a single PR
func processPR(
	ctx context.Context,
	client *github.Client,
	checker *permissions.Checker,
	bc *config.Config,
	repoOwner, repoName string,
	pr map[string]interface{},
) error {
	prNumberFloat, ok := pr["number"].(float64)
	if !ok {
		return fmt.Errorf("invalid PR number")
	}
	prNumber := int(prNumberFloat)

	fmt.Printf("\nProcessing PR #%d\n", prNumber)

	// Get PR author and title for RuntimeConfig
	var author, title string
	if user, ok := pr["user"].(map[string]interface{}); ok {
		if login, ok := user["login"].(string); ok {
			author = login
		}
	}
	if t, ok := pr["title"].(string); ok {
		title = t
	}

	// Create request-scoped runtime config for this PR
	rc := &RuntimeConfig{
		CommentBody:   title, // Use PR title as body
		CommentID:     strconv.Itoa(prNumber),
		CommentAction: "created",
		PRNumber:      strconv.Itoa(prNumber),
		RepoOwner:     repoOwner,
		RepoName:      repoName,
		CommentAuthor: author,
		BotUsername:   defaultBotUsername, // Use default bot username
	}

	// Process reactions if not disabled
	if !bc.DisableReactions {
		if err := handleReactions(ctx, client, rc, bc, checker, prNumber, prNumber); err != nil {
			return fmt.Errorf("failed to process reactions on PR #%d: %w", prNumber, err)
		}
	}

	return nil
}

// processPendingCIPRs processes PRs that are waiting for CI to pass before merge
//
// It queries PRs with pending-ci labels, checks their CI status, and:
// - Merges if CI passes
// - Removes label and posts failure feedback if CI fails
// - Skips if CI is still pending
//
//nolint:unparam // error return kept for consistent function signature and future error handling
func processPendingCIPRs(
	ctx context.Context,
	client *github.Client,
	bc *config.Config,
	repoOwner, repoName string,
	prs []map[string]interface{},
) error {
	// Filter PRs with pending-ci labels
	pendingPRs := filterPendingCIPRs(prs)

	if len(pendingPRs) == 0 {
		return nil
	}

	fmt.Printf("\nProcessing %d PR(s) waiting for CI\n", len(pendingPRs))

	for _, pr := range pendingPRs {
		if err := processPendingCIPR(ctx, client, bc, repoOwner, repoName, pr); err != nil {
			prNum := extractPRNumber(pr.prData)
			fmt.Fprintf(os.Stderr, "  Warning: failed to process pending-ci PR #%d: %v\n", prNum, err)
		}
	}

	return nil
}

// pendingCIPR holds data about a PR waiting for CI
type pendingCIPR struct {
	prData       map[string]interface{}
	method       github.MergeMethod
	label        string
	requiredOnly bool // true if only required checks should be considered
}

// filterPendingCIPRs filters PRs that have pending-ci labels
func filterPendingCIPRs(prs []map[string]interface{}) []pendingCIPR {
	var result []pendingCIPR

	for _, pr := range prs {
		labels, ok := pr["labels"].([]interface{})
		if !ok {
			continue
		}

		for _, l := range labels {
			labelMap, ok := l.(map[string]interface{})
			if !ok {
				continue
			}

			labelName, ok := labelMap["name"].(string)
			if !ok {
				continue
			}

			method, requiredOnly, label := parsePendingCILabel(labelName)
			if label != "" {
				result = append(result, pendingCIPR{
					prData:       pr,
					method:       method,
					label:        label,
					requiredOnly: requiredOnly,
				})

				break // Only one pending-ci label per PR
			}
		}
	}

	return result
}

// parsePendingCILabel parses a pending-ci label and returns the merge method and required flag
//
// Returns:
// - MergeMethod, requiredOnly bool, and label name if valid pending-ci label
// - Empty string if not a pending-ci label
func parsePendingCILabel(label string) (github.MergeMethod, bool, string) {
	switch label {
	case github.LabelPendingCIMerge:
		return github.MergeMethodMerge, false, label
	case github.LabelPendingCISquash:
		return github.MergeMethodSquash, false, label
	case github.LabelPendingCIRebase:
		return github.MergeMethodRebase, false, label
	case github.LabelPendingCIMergeRequired:
		return github.MergeMethodMerge, true, label
	case github.LabelPendingCISquashRequired:
		return github.MergeMethodSquash, true, label
	case github.LabelPendingCIRebaseRequired:
		return github.MergeMethodRebase, true, label
	default:
		return "", false, ""
	}
}

// extractPRNumber extracts PR number from PR data
func extractPRNumber(pr map[string]interface{}) int {
	if num, ok := pr["number"].(float64); ok {
		return int(num)
	}

	return 0
}

// processPendingCIPR processes a single PR waiting for CI
func processPendingCIPR(
	ctx context.Context,
	client *github.Client,
	bc *config.Config,
	repoOwner, repoName string,
	pr pendingCIPR,
) error {
	prNumber := extractPRNumber(pr.prData)
	if prNumber == 0 {
		return fmt.Errorf("invalid PR number")
	}

	fmt.Printf("  Checking CI status for PR #%d (method: %s)\n", prNumber, pr.method)

	// Get PR head SHA for CI status check
	headRef, err := client.GetPRHeadRef(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR head ref: %w", err)
	}

	// Get required checks list if filtering by required checks only
	var requiredChecks []string
	if pr.requiredOnly {
		// Get base branch from PR info
		info, err := client.GetPRInfo(ctx, repoOwner, repoName, prNumber)
		if err != nil {
			return fmt.Errorf("failed to get PR info: %w", err)
		}

		if info.BaseBranch != "" {
			requiredChecks, err = client.GetRequiredStatusChecks(ctx, repoOwner, repoName, info.BaseBranch)
			if err != nil {
				return fmt.Errorf("failed to get required checks: %w", err)
			}
		}
	}

	// Check current CI status
	checkStatus, err := client.GetCheckStatus(ctx, repoOwner, repoName, headRef, requiredChecks)
	if err != nil {
		return fmt.Errorf("failed to get CI status: %w", err)
	}

	// Handle based on CI status
	switch {
	case checkStatus.AllPassing:
		return handlePendingCIPassed(ctx, client, bc, repoOwner, repoName, prNumber, pr)

	case checkStatus.Failing:
		return handlePendingCIFailed(ctx, client, bc, repoOwner, repoName, prNumber, pr, checkStatus.Summary)

	default:
		// CI still pending, skip
		fmt.Printf("    CI still pending: %s\n", checkStatus.Summary)

		return nil
	}
}

// handlePendingCIPassed handles a PR where CI has passed
func handlePendingCIPassed(
	ctx context.Context,
	client *github.Client,
	bc *config.Config,
	repoOwner, repoName string,
	prNumber int,
	pr pendingCIPR,
) error {
	fmt.Printf("    CI passed! Merging PR #%d\n", prNumber)

	// Merge the PR
	if err := client.MergePR(ctx, repoOwner, repoName, prNumber, pr.method); err != nil {
		// Try fallback methods if merge commits not allowed
		if pr.method == github.MergeMethodMerge && strings.Contains(err.Error(), "Merge commits are not allowed") {
			if err := client.MergePR(ctx, repoOwner, repoName, prNumber, github.MergeMethodSquash); err != nil {
				if err := client.MergePR(ctx, repoOwner, repoName, prNumber, github.MergeMethodRebase); err != nil {
					return postPendingCIError(ctx, client, repoOwner, repoName, prNumber, pr.label, err.Error())
				}
			}
		} else {
			return postPendingCIError(ctx, client, repoOwner, repoName, prNumber, pr.label, err.Error())
		}
	}

	// Remove pending-ci label
	_ = client.RemoveLabel(ctx, repoOwner, repoName, prNumber, pr.label)

	// Post success feedback
	// We don't know who requested the merge, so use a generic message
	fb := feedback.NewPendingCIMerged("automation", bc.QuietSuccess)
	if fb.RequiresComment() {
		_ = client.PostComment(ctx, repoOwner, repoName, prNumber, fb.Message)
	}

	return nil
}

// handlePendingCIFailed handles a PR where CI has failed
//
//nolint:unparam // bc kept for API consistency with handlePendingCIPassed
func handlePendingCIFailed(
	ctx context.Context,
	client *github.Client,
	_ *config.Config, // kept for API consistency with handlePendingCIPassed
	repoOwner, repoName string,
	prNumber int,
	pr pendingCIPR,
	summary string,
) error {
	fmt.Printf("    CI failed: %s\n", summary)

	return postPendingCIError(ctx, client, repoOwner, repoName, prNumber, pr.label, summary)
}

// postPendingCIError posts error feedback and removes label for failed pending-ci
func postPendingCIError(
	ctx context.Context,
	client *github.Client,
	repoOwner, repoName string,
	prNumber int,
	label, reason string,
) error {
	// Remove pending-ci label
	_ = client.RemoveLabel(ctx, repoOwner, repoName, prNumber, label)

	// Post failure feedback
	fb := feedback.NewPendingCIFailed(reason)
	_ = client.PostComment(ctx, repoOwner, repoName, prNumber, fb.Message)

	return nil
}

