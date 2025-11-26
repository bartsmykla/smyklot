// Package permissions provides CODEOWNERS-based authorization for Smyklot.
//
// It validates user permissions by parsing .github/CODEOWNERS files and
// checking if users have approval rights for repository changes.
package permissions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// GitHubClient defines the interface for GitHub API operations needed by Checker
type GitHubClient interface {
	IsTeamMember(ctx context.Context, org, teamSlug, username string) (bool, error)
}

// Checker validates user permissions based on CODEOWNERS files
//
// Phase 1: Global CODEOWNERS support only
//   - Uses .github/CODEOWNERS file (GitHub standard)
//   - Users listed as global owners (*) can approve any changes
//   - No path-specific permissions (future enhancement)
//
// Phase 2: Team membership support
//   - Supports team ownership (e.g., @org/team-name)
//   - Checks team membership via GitHub API
//
// Phase 3 (future): Path-specific permissions
//   - Support path-specific ownership patterns
//   - Users can approve changes in their scope
type Checker struct {
	rootApprovers []string
	githubClient  GitHubClient
}

// NewCheckerFromContent creates a new permission checker from CODEOWNERS content
//
// This is useful when the CODEOWNERS content is fetched from an API
// rather than read from the filesystem.
//
// The githubClient parameter is optional. If provided, the checker will support
// team membership validation. If nil, team ownership will be treated as individual
// usernames (backward compatible behavior).
//
// Returns an error if the content cannot be parsed.
func NewCheckerFromContent(content string, githubClient GitHubClient) (*Checker, error) {
	checker := &Checker{
		rootApprovers: []string{},
		githubClient:  githubClient,
	}

	if content == "" {
		return checker, nil
	}

	codeowners, err := ParseCodeownersContent(content)
	if err != nil {
		// Fail-closed: return error if CODEOWNERS cannot be parsed
		// This prevents privilege escalation via corrupted CODEOWNERS files
		return nil, NewCheckerError(ErrParseFailed, "content", err)
	}

	checker.rootApprovers = codeowners.GetGlobalOwners()

	return checker, nil
}

// NewChecker creates a new permission checker for the given repository path
//
// The checker loads the .github/CODEOWNERS file if it exists. If no
// CODEOWNERS file exists, all permission checks will return false.
//
// The githubClient parameter is optional. If provided, the checker will support
// team membership validation. If nil, team ownership will be treated as individual
// usernames (backward compatible behavior).
//
// Returns an error if:
//   - The repository path is empty
//   - The repository path does not exist
func NewChecker(repoPath string, githubClient GitHubClient) (*Checker, error) {
	if repoPath == "" {
		return nil, ErrEmptyRepoPath
	}

	// Check if repository path exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, NewCheckerError(ErrRepoPathNotExist, repoPath, err)
	}

	checker := &Checker{
		rootApprovers: []string{},
		githubClient:  githubClient,
	}

	// Try to load .github/CODEOWNERS file
	codeownersPath := filepath.Join(repoPath, ".github", "CODEOWNERS")
	if _, err := os.Stat(codeownersPath); err == nil {
		codeowners, err := ParseCodeownersFile(codeownersPath)
		if err != nil {
			// Fail-closed: return error if CODEOWNERS cannot be parsed
			// This prevents privilege escalation via corrupted CODEOWNERS files
			return nil, err
		}
		checker.rootApprovers = codeowners.GetGlobalOwners()
	}

	return checker, nil
}

// isTeamMember checks if a user is a member of a team via GitHub API
func (c *Checker) isTeamMember(approver, username string) (bool, error) {
	if c.githubClient == nil {
		return false, nil
	}

	parts := strings.SplitN(approver, "/", 2)
	if len(parts) != 2 {
		return false, nil
	}

	org, teamSlug := parts[0], parts[1]
	return c.githubClient.IsTeamMember(context.Background(), org, teamSlug, username)
}

// CanApprove checks if the given user can approve changes at the specified path.
//
// Phase 1: Root OWNERS only
//   - Returns true if the user is in the root OWNERS file
//   - The path parameter is ignored (reserved for future scoped permissions)
//   - Returns false if the username is empty or not in the approvers list
//
// Phase 2: Team membership support
//   - Supports team ownership (e.g., @org/team-name)
//   - Checks team membership via GitHub API if client is provided
//   - Falls back to string matching if no GitHub client
//
// Phase 3 (future): Scoped permissions
//   - Check path-specific OWNERS files
//   - Support hierarchical permission resolution
func (c *Checker) CanApprove(username, _ string) (bool, error) {
	if username == "" {
		return false, nil
	}

	// Check root approvers (Phase 1 & 2)
	for _, approver := range c.rootApprovers {
		// Check if approver is a team (contains '/')
		if strings.Contains(approver, "/") {
			isMember, err := c.isTeamMember(approver, username)
			if err != nil {
				return false, err
			}
			if isMember {
				return true, nil
			}
			continue
		}

		// Individual user ownership: exact match
		if approver == username {
			return true, nil
		}
	}

	return false, nil
}

// GetApprovers returns the list of users who can approve changes.
//
// Phase 1: Returns root approvers only
// Phase 2 (future): Could accept a path parameter to get scoped approvers
func (c *Checker) GetApprovers() []string {
	return c.rootApprovers
}
