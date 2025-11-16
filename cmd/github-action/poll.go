package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bartsmykla/smyklot/pkg/github"
	"github.com/bartsmykla/smyklot/pkg/permissions"
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
	// Get configuration from flags and environment
	repo, token, err := getPollConfig(cmd)
	if err != nil {
		return err
	}

	// Parse repository owner and name
	repoOwner, repoName, err := parseRepo(repo)
	if err != nil {
		return err
	}

	// Setup GitHub client and permission checker
	client, checker, err := setupPollClients(token, repoOwner, repoName)
	if err != nil {
		return err
	}

	// Poll and process all open PRs
	return pollAllPRs(client, checker, repoOwner, repoName)
}

// getPollConfig retrieves repo and token from flags or environment
func getPollConfig(cmd *cobra.Command) (string, string, error) {
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
			installationToken, err := getInstallationToken()
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
	var repoOwner, repoName string
	if _, err := fmt.Sscanf(repo, "%[^/]/%s", &repoOwner, &repoName); err != nil {
		return "", "", fmt.Errorf("invalid repository format (expected owner/name): %w", err)
	}
	return repoOwner, repoName, nil
}

// setupPollClients creates GitHub client and permission checker
func setupPollClients(
	token, repoOwner, repoName string,
) (*github.Client, *permissions.Checker, error) {
	// Create GitHub client
	client, err := github.NewClient(token, emptyBaseURL)
	if err != nil {
		return nil, nil, NewGitHubError(ErrGitHubClient, err)
	}

	// Fetch CODEOWNERS
	codeownersContent, err := client.GetCodeowners(repoOwner, repoName)
	if err != nil {
		return nil, nil, NewGitHubError(ErrGetCodeowners, err)
	}

	// Initialize permission checker
	checker, err := permissions.NewCheckerFromContent(codeownersContent)
	if err != nil {
		return nil, nil, NewGitHubError(ErrInitPermissions, err)
	}

	return client, checker, nil
}

// pollAllPRs polls and processes all open PRs
func pollAllPRs(
	client *github.Client,
	checker *permissions.Checker,
	repoOwner, repoName string,
) error {
	fmt.Printf("Polling reactions on open PRs in %s/%s\n", repoOwner, repoName)

	// Get all open PRs
	prs, err := client.GetOpenPRs(repoOwner, repoName)
	if err != nil {
		return NewGitHubError(ErrGetCodeowners, err)
	}

	if len(prs) == 0 {
		fmt.Println("No open PRs found")
		return nil
	}

	fmt.Printf("Found %d open PR(s)\n", len(prs))

	// Process each PR
	for _, pr := range prs {
		if err := processPR(client, checker, repoOwner, repoName, pr); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
		}
	}

	fmt.Println("\nPolling complete")

	return nil
}

// processPR processes reactions on all comments in a single PR
func processPR(
	client *github.Client,
	checker *permissions.Checker,
	repoOwner, repoName string,
	pr map[string]interface{},
) error {
	prNumberFloat, ok := pr["number"].(float64)
	if !ok {
		return fmt.Errorf("invalid PR number")
	}
	prNumber := int(prNumberFloat)

	fmt.Printf("\nProcessing PR #%d\n", prNumber)

	// Get all comments on the PR
	comments, err := client.GetPRComments(repoOwner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get comments for PR #%d: %w", prNumber, err)
	}

	if len(comments) == 0 {
		fmt.Printf("  No comments on PR #%d\n", prNumber)
		return nil
	}

	fmt.Printf("  Found %d comment(s)\n", len(comments))

	// Process each comment
	for _, comment := range comments {
		if err := processComment(client, checker, repoOwner, repoName, prNumber, comment); err != nil {
			fmt.Fprintf(os.Stderr, "    Warning: %v\n", err)
		}
	}

	return nil
}

// processComment processes reactions on a single comment
func processComment(
	client *github.Client,
	checker *permissions.Checker,
	repoOwner, repoName string,
	prNumber int,
	comment map[string]interface{},
) error {
	commentIDFloat, ok := comment["id"].(float64)
	if !ok {
		return fmt.Errorf("invalid comment ID")
	}
	commentID := int(commentIDFloat)

	user, ok := comment["user"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid user data")
	}

	author, ok := user["login"].(string)
	if !ok {
		return fmt.Errorf("invalid author login")
	}

	body, ok := comment["body"].(string)
	if !ok {
		body = ""
	}

	fmt.Printf("  Processing comment %d by %s\n", commentID, author)

	// Process reactions on this comment
	if err := pollCommentReactions(
		client,
		checker,
		repoOwner,
		repoName,
		prNumber,
		commentID,
		author,
		body,
	); err != nil {
		return fmt.Errorf("failed to process reactions on comment %d: %w", commentID, err)
	}

	return nil
}

// pollCommentReactions checks and processes reactions on a specific comment
func pollCommentReactions(
	client *github.Client,
	checker *permissions.Checker,
	repoOwner, repoName string,
	prNumber, commentID int,
	author, body string,
) error {
	// Set runtime config for this comment
	runtimeConfig = RuntimeConfig{
		CommentBody:   body,
		CommentID:     strconv.Itoa(commentID),
		CommentAction: "created",
		PRNumber:      strconv.Itoa(prNumber),
		RepoOwner:     repoOwner,
		RepoName:      repoName,
		CommentAuthor: author,
	}

	// Process reactions if not disabled
	if !botConfig.DisableReactions {
		if err := handleReactions(client, checker, prNumber, commentID); err != nil {
			return err
		}
	}

	return nil
}
