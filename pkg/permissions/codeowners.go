// Package permissions provides CODEOWNERS file parsing.
//
// This file provides parsing functionality for GitHub CODEOWNERS format,
// extracting ownership patterns and associated users/teams.
package permissions

import (
	"bufio"
	"os"
	"strings"
)

// parseEntries parses CODEOWNERS entries from a scanner
func parseEntries(scanner *bufio.Scanner) []CodeownersEntry {
	entries := make([]CodeownersEntry, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line: <pattern> <owners...>
		parts := strings.Fields(line)
		if len(parts) < 2 {
			// Invalid format: need at least pattern and one owner
			continue
		}

		pattern := parts[0]
		owners := make([]string, 0)

		// Extract owners (strings starting with @)
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "@") {
				// Remove the @ prefix for consistency
				owner := strings.TrimPrefix(part, "@")
				owners = append(owners, owner)
			}
		}

		if len(owners) > 0 {
			entries = append(entries, CodeownersEntry{
				Pattern: pattern,
				Owners:  owners,
			})
		}
	}

	return entries
}

// ParseCodeownersContent parses CODEOWNERS content from a string
//
// CODEOWNERS format:
//   - Lines starting with # are comments
//   - Empty lines are ignored
//   - Format: <pattern> <owner1> <owner2> ...
//   - Owners start with @ (e.g., @username, @org/team)
//
// Example:
//   # Global owners
//   * @global-owner1 @global-owner2
//   /docs/ @doc-team
//   *.js @js-team
//
// For Phase 1, this parser focuses on global owners (pattern: *)
// Path-specific ownership will be implemented in later phases
func ParseCodeownersContent(content string) (*CodeownersFile, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	entries := parseEntries(scanner)

	if err := scanner.Err(); err != nil {
		return nil, NewParseError(ErrReadFailed, "", err)
	}

	return &CodeownersFile{
		Path:    "",
		Entries: entries,
	}, nil
}

// ParseCodeownersFile reads and parses a CODEOWNERS file from the given path
//
// CODEOWNERS format:
//   - Lines starting with # are comments
//   - Empty lines are ignored
//   - Format: <pattern> <owner1> <owner2> ...
//   - Owners start with @ (e.g., @username, @org/team)
//
// Example:
//   # Global owners
//   * @global-owner1 @global-owner2
//   /docs/ @doc-team
//   *.js @js-team
//
// For Phase 1, this parser focuses on global owners (pattern: *)
// Path-specific ownership will be implemented in later phases
func ParseCodeownersFile(filePath string) (*CodeownersFile, error) {
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}

	// Read the file
	file, err := os.Open(filePath) //nolint:gosec // File path is validated by caller
	if err != nil {
		return nil, NewParseError(ErrReadFailed, filePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	entries := parseEntries(scanner)

	if err := scanner.Err(); err != nil {
		return nil, NewParseError(ErrReadFailed, filePath, err)
	}

	return &CodeownersFile{
		Path:    filePath,
		Entries: entries,
	}, nil
}

// GetGlobalOwners returns the list of global owners (pattern: *)
//
// Global owners have approval rights for any path in the repository
func (c *CodeownersFile) GetGlobalOwners() []string {
	for _, entry := range c.Entries {
		if entry.Pattern == "*" {
			return entry.Owners
		}
	}

	return []string{}
}

// GetAllOwners returns a deduplicated list of all owners from all patterns
func (c *CodeownersFile) GetAllOwners() []string {
	ownerSet := make(map[string]bool)
	for _, entry := range c.Entries {
		for _, owner := range entry.Owners {
			ownerSet[owner] = true
		}
	}

	owners := make([]string, 0, len(ownerSet))
	for owner := range ownerSet {
		owners = append(owners, owner)
	}

	return owners
}
