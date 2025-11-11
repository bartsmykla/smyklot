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
	repoPath      string
	rootApprovers []string
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
		repoPath:      repoPath,
		rootApprovers: []string{},
	}

	// Try to load .github/CODEOWNERS file
	codeownersPath := filepath.Join(repoPath, ".github", "CODEOWNERS")
	if _, err := os.Stat(codeownersPath); err == nil {
		codeowners, err := ParseCodeownersFile(codeownersPath)
		if err != nil {
			// If the CODEOWNERS file exists but cannot be parsed, treat it as
			// having no approvers. This is a soft failure to avoid blocking
			// operations due to syntax errors.
			return checker, nil
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
