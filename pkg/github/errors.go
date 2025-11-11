package github

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors for GitHub API operations.
var (
	// ErrEmptyToken is returned when an empty token is provided
	ErrEmptyToken = errors.New("empty GitHub token provided")

	// ErrEmptyComment is returned when an empty comment body is provided
	ErrEmptyComment = errors.New("empty comment body provided")

	// ErrAPIRequest is returned when an API request fails
	ErrAPIRequest = errors.New("GitHub API request failed")

	// ErrResponseParse is returned when parsing a response fails
	ErrResponseParse = errors.New("failed to parse GitHub API response")
)

// APIError represents an error from the GitHub API
type APIError struct {
	Op         error
	StatusCode int
	Method     string
	Path       string
	Detail     string
}

// NewAPIError creates a new API error
func NewAPIError(op error, statusCode int, method, path string, err error) error {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	return &APIError{
		Op:         op,
		StatusCode: statusCode,
		Method:     method,
		Path:       path,
		Detail:     detail,
	}
}

// Error returns the error message
func (e *APIError) Error() string {
	var builder strings.Builder

	if e.Op != nil {
		builder.WriteString(e.Op.Error())
	} else {
		builder.WriteString("API error")
	}

	if e.StatusCode != 0 {
		builder.WriteString(fmt.Sprintf(" (status: %d)", e.StatusCode))
	}

	if e.Method != "" || e.Path != "" {
		builder.WriteString(fmt.Sprintf(" [%s %s]", e.Method, e.Path))
	}

	if e.Detail != "" {
		builder.WriteString(fmt.Sprintf(": %s", e.Detail))
	}

	return builder.String()
}

// Unwrap returns the wrapped error
func (e *APIError) Unwrap() error {
	return e.Op
}

// Is checks if the target error matches this error type
func (e *APIError) Is(target error) bool {
	var apiErr *APIError
	if errors.As(target, &apiErr) {
		return true
	}

	return errors.Is(e.Op, target)
}
