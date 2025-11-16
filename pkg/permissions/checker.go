// Package permissions provides CODEOWNERS-based authorization for Smyklot.
//
// It validates user permissions by parsing .github/CODEOWNERS files and
// checking if users have approval rights for repository changes.
package permissions

import (
	"os"
	"path/filepath"
)

// Checker validates user permissions based on CODEOWNERS files
//
// Phase 1: Global CODEOWNERS support only
//   - Uses .github/CODEOWNERS file (GitHub standard)
//   - Users listed as global owners (*) can approve any changes
//   - No path-specific permissions (future enhancement)
//
// Phase 2 (future): Path-specific permissions
//   - Support path-specific ownership patterns
//   - Users can approve changes in their scope
type Checker struct {
	rootApprovers []string
}

// NewCheckerFromContent creates a new permission checker from CODEOWNERS content
//
// This is useful when the CODEOWNERS content is fetched from an API
// rather than read from the filesystem.
//
// Returns an error if the content cannot be parsed.
func NewCheckerFromContent(content string) (*Checker, error) {
	checker := &Checker{
		rootApprovers: []string{},
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
// Returns an error if:
//   - The repository path is empty
//   - The repository path does not exist
func NewChecker(repoPath string) (*Checker, error) {
	if repoPath == "" {
		return nil, ErrEmptyRepoPath
	}

	// Check if repository path exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, NewCheckerError(ErrRepoPathNotExist, repoPath, err)
	}

	checker := &Checker{
		rootApprovers: []string{},
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

// CanApprove checks if the given user can approve changes at the specified path.
//
// Phase 1: Root OWNERS only
//   - Returns true if the user is in the root OWNERS file
//   - The path parameter is ignored (reserved for future scoped permissions)
//   - Returns false if the username is empty or not in the approvers list
//
// Phase 2 (future): Scoped permissions
//   - Check path-specific OWNERS files
//   - Support hierarchical permission resolution
func (c *Checker) CanApprove(username, _ string) (bool, error) {
	if username == "" {
		return false, nil
	}

	// Check root approvers (Phase 1)
	for _, approver := range c.rootApprovers {
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
