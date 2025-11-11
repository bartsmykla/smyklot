package permissions

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseOwnersFile reads and parses an OWNERS file from the given path.
//
// The function performs the following operations:
//   - Reads the OWNERS file from disk
//   - Parses the YAML content
//   - Extracts the approvers list
//   - Filters out empty or whitespace-only usernames
//   - Sets the directory path where the OWNERS file is located
//
// Returns an error if:
//   - The file path is empty
//   - The file cannot be read
//   - The YAML syntax is invalid
func ParseOwnersFile(filePath string) (*OwnersFile, error) {
	if filePath == "" {
		return nil, ErrEmptyFilePath
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewParseError(ErrReadFailed, filePath, err)
	}

	// Parse YAML
	var owners OwnersFile
	if err := yaml.Unmarshal(data, &owners); err != nil {
		return nil, NewParseError(ErrInvalidYAML, filePath, err)
	}

	// Set the path to the directory containing the OWNERS file
	owners.Path = filepath.Dir(filePath)

	// Filter out empty or whitespace-only approvers
	owners.Approvers = filterEmptyApprovers(owners.Approvers)

	return &owners, nil
}

// filterEmptyApprovers removes empty strings and whitespace-only strings
// from the approvers list.
func filterEmptyApprovers(approvers []string) []string {
	if len(approvers) == 0 {
		return approvers
	}

	filtered := make([]string, 0, len(approvers))
	for _, approver := range approvers {
		trimmed := strings.TrimSpace(approver)
		if trimmed != "" {
			filtered = append(filtered, approver)
		}
	}

	return filtered
}
