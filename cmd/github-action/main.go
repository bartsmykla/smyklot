package main

import (
	"fmt"
	"os"

	"github.com/bartsmykla/smyklot/pkg/commands"
	"github.com/bartsmykla/smyklot/pkg/feedback"
	"github.com/bartsmykla/smyklot/pkg/github"
	"github.com/bartsmykla/smyklot/pkg/permissions"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Read environment variables from GitHub Actions
	token := os.Getenv("GITHUB_TOKEN")
	commentBody := os.Getenv("COMMENT_BODY")
	commentID := os.Getenv("COMMENT_ID")
	prNumber := os.Getenv("PR_NUMBER")
	repoOwner := os.Getenv("REPO_OWNER")
	repoName := os.Getenv("REPO_NAME")
	commentAuthor := os.Getenv("COMMENT_AUTHOR")

	// Validate required environment variables
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	if commentBody == "" {
		return fmt.Errorf("COMMENT_BODY environment variable is required")
	}
	if commentID == "" {
		return fmt.Errorf("COMMENT_ID environment variable is required")
	}
	if prNumber == "" {
		return fmt.Errorf("PR_NUMBER environment variable is required")
	}
	if repoOwner == "" {
		return fmt.Errorf("REPO_OWNER environment variable is required")
	}
	if repoName == "" {
		return fmt.Errorf("REPO_NAME environment variable is required")
	}
	if commentAuthor == "" {
		return fmt.Errorf("COMMENT_AUTHOR environment variable is required")
	}

	// Parse the command from the comment
	cmd, err := commands.ParseCommand(commentBody)
	if err != nil {
		// Not a valid command, ignore
		return nil
	}

	// Create GitHub client
	client, err := github.NewClient(token, "")
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
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

	// Check if user has permission to execute this command
	canApprove, err := checker.CanApprove(commentAuthor, "/")
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	// Convert string IDs to integers
	var prNum, commentIDNum int
	if _, err := fmt.Sscanf(prNumber, "%d", &prNum); err != nil {
		return fmt.Errorf("invalid PR number: %w", err)
	}
	if _, err := fmt.Sscanf(commentID, "%d", &commentIDNum); err != nil {
		return fmt.Errorf("invalid comment ID: %w", err)
	}

	// Handle unauthorized users
	if !canApprove {
		fb := feedback.NewUnauthorized(commentAuthor, checker.GetApprovers())
		if err := client.PostComment(repoOwner, repoName, prNum, fb.Message); err != nil {
			return fmt.Errorf("failed to post unauthorized feedback: %w", err)
		}
		if err := client.AddReaction(repoOwner, repoName, commentIDNum, github.ReactionError); err != nil {
			return fmt.Errorf("failed to add error reaction: %w", err)
		}
		return nil
	}

	// Execute the command based on type
	switch cmd.Type {
	case commands.CommandApprove:
		if err := handleApprove(client, repoOwner, repoName, prNum, commentIDNum, commentAuthor); err != nil {
			return err
		}
	case commands.CommandMerge:
		if err := handleMerge(client, repoOwner, repoName, prNum, commentIDNum, commentAuthor); err != nil {
			return err
		}
	default:
		// Unknown command type
		return nil
	}

	return nil
}

func handleApprove(client *github.Client, owner, repo string, prNum, commentID int, author string) error {
	// Add eyes reaction to acknowledge
	if err := client.AddReaction(owner, repo, commentID, github.ReactionEyes); err != nil {
		return fmt.Errorf("failed to add acknowledgment reaction: %w", err)
	}

	// Approve the PR
	if err := client.ApprovePR(owner, repo, prNum); err != nil {
		// Post error feedback
		fb := feedback.NewApprovalFailed(err.Error())
		if postErr := client.PostComment(owner, repo, prNum, fb.Message); postErr != nil {
			return fmt.Errorf("approval failed and unable to post feedback: %w (original error: %v)", postErr, err)
		}
		if reactionErr := client.AddReaction(owner, repo, commentID, github.ReactionError); reactionErr != nil {
			return fmt.Errorf("approval failed and unable to add error reaction: %w (original error: %v)", reactionErr, err)
		}
		return fmt.Errorf("failed to approve PR: %w", err)
	}

	// Post success feedback
	fb := feedback.NewApprovalSuccess(author)
	if err := client.PostComment(owner, repo, prNum, fb.Message); err != nil {
		return fmt.Errorf("approval succeeded but failed to post feedback: %w", err)
	}

	// Add success reaction
	if err := client.AddReaction(owner, repo, commentID, github.ReactionSuccess); err != nil {
		return fmt.Errorf("approval succeeded but failed to add success reaction: %w", err)
	}

	return nil
}

func handleMerge(client *github.Client, owner, repo string, prNum, commentID int, author string) error {
	// Add eyes reaction to acknowledge
	if err := client.AddReaction(owner, repo, commentID, github.ReactionEyes); err != nil {
		return fmt.Errorf("failed to add acknowledgment reaction: %w", err)
	}

	// Get PR info to check if it's mergeable
	info, err := client.GetPRInfo(owner, repo, prNum)
	if err != nil {
		fb := feedback.NewMergeFailed(fmt.Sprintf("failed to get PR info: %v", err))
		if postErr := client.PostComment(owner, repo, prNum, fb.Message); postErr != nil {
			return fmt.Errorf("failed to get PR info and unable to post feedback: %w (original error: %v)", postErr, err)
		}
		if reactionErr := client.AddReaction(owner, repo, commentID, github.ReactionError); reactionErr != nil {
			return fmt.Errorf("failed to get PR info and unable to add error reaction: %w (original error: %v)", reactionErr, err)
		}
		return fmt.Errorf("failed to get PR info: %w", err)
	}

	// Check if PR is mergeable
	if !info.Mergeable {
		fb := feedback.NewNotMergeable()
		if err := client.PostComment(owner, repo, prNum, fb.Message); err != nil {
			return fmt.Errorf("PR not mergeable and failed to post feedback: %w", err)
		}
		if err := client.AddReaction(owner, repo, commentID, github.ReactionWarning); err != nil {
			return fmt.Errorf("PR not mergeable and failed to add warning reaction: %w", err)
		}
		return nil
	}

	// Merge the PR
	if err := client.MergePR(owner, repo, prNum); err != nil {
		fb := feedback.NewMergeFailed(err.Error())
		if postErr := client.PostComment(owner, repo, prNum, fb.Message); postErr != nil {
			return fmt.Errorf("merge failed and unable to post feedback: %w (original error: %v)", postErr, err)
		}
		if reactionErr := client.AddReaction(owner, repo, commentID, github.ReactionError); reactionErr != nil {
			return fmt.Errorf("merge failed and unable to add error reaction: %w (original error: %v)", reactionErr, err)
		}
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	// Post success feedback
	fb := feedback.NewMergeSuccess(author)
	if err := client.PostComment(owner, repo, prNum, fb.Message); err != nil {
		return fmt.Errorf("merge succeeded but failed to post feedback: %w", err)
	}

	// Add success reaction
	if err := client.AddReaction(owner, repo, commentID, github.ReactionSuccess); err != nil {
		return fmt.Errorf("merge succeeded but failed to add success reaction: %w", err)
	}

	return nil
}
