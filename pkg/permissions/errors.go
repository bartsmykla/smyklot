package permissions

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for permission operations.
var (
	// ErrEmptyFilePath is returned when an empty file path is provided
	ErrEmptyFilePath = errors.New("empty file path provided")

	// ErrReadFailed is returned when reading the CODEOWNERS file fails
	ErrReadFailed = errors.New("failed to read the CODEOWNERS file")

	// ErrEmptyRepoPath is returned when an empty repository path is provided
	ErrEmptyRepoPath = errors.New("empty repository path provided")

	// ErrRepoPathNotExist is returned when the repository path does not exist
	ErrRepoPathNotExist = errors.New("repository path does not exist")
)

// ParseError represents an error that occurred during CODEOWNERS file parsing
type ParseError struct {
	Op       error
	FilePath string
	Detail   string
}

// NewParseError creates a new parse error
func NewParseError(op error, filePath string, err error) error {
	return &ParseError{
		Op:       op,
		FilePath: filePath,
		Detail:   err.Error(),
	}
}

// Error implements the error interface
func (e *ParseError) Error() string {
	var builder strings.Builder

	if e.Op != nil {
		builder.WriteString(e.Op.Error())
	} else {
		builder.WriteString("parse error")
	}

	if e.FilePath != "" {
		builder.WriteString(fmt.Sprintf(" (file: %s)", e.FilePath))
	}

	if e.Detail != "" {
		builder.WriteString(fmt.Sprintf(": %s", e.Detail))
	}

	return builder.String()
}

// Unwrap returns the wrapped error
func (e *ParseError) Unwrap() error {
	return e.Op
}

// Is checks if the target error matches
func (e *ParseError) Is(target error) bool {
	var parseErr *ParseError
	if errors.As(target, &parseErr) {
		return true
	}

	return errors.Is(e.Op, target)
}

// CheckerError represents an error that occurred during permission checking
type CheckerError struct {
	Op       error
	RepoPath string
	Detail   string
}

// NewCheckerError creates a new checker error
func NewCheckerError(op error, repoPath string, err error) error {
	return &CheckerError{
		Op:       op,
		RepoPath: repoPath,
		Detail:   err.Error(),
	}
}

// Error implements the error interface
func (e *CheckerError) Error() string {
	var builder strings.Builder

	if e.Op != nil {
		builder.WriteString(e.Op.Error())
	} else {
		builder.WriteString("checker error")
	}

	if e.RepoPath != "" {
		builder.WriteString(fmt.Sprintf(" (repo: %s)", e.RepoPath))
	}

	if e.Detail != "" {
		builder.WriteString(fmt.Sprintf(": %s", e.Detail))
	}

	return builder.String()
}

// Unwrap returns the wrapped error
func (e *CheckerError) Unwrap() error {
	return e.Op
}

// Is checks if the target error matches
func (e *CheckerError) Is(target error) bool {
	var checkerErr *CheckerError
	if errors.As(target, &checkerErr) {
		return true
	}

	return errors.Is(e.Op, target)
}
