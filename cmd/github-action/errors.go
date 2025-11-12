package main

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure cases.
var (
	// ErrMissingEnvVar is returned when a required environment variable is missing
	ErrMissingEnvVar = errors.New("required environment variable is missing")

	// ErrInvalidInput is returned when invalid input is provided
	ErrInvalidInput = errors.New("invalid input provided")

	// ErrGitHubClient is returned when GitHub client creation fails
	ErrGitHubClient = errors.New("GitHub client error")

	// ErrGitHubAppAuth is returned when GitHub App authentication fails
	ErrGitHubAppAuth = errors.New("GitHub App authentication failed")

	// ErrPermissionCheck is returned when the permission check fails
	ErrPermissionCheck = errors.New("permission check failed")

	// ErrPostComment is returned when posting a comment fails
	ErrPostComment = errors.New("failed to post comment")

	// ErrAddReaction is returned when adding a reaction fails
	ErrAddReaction = errors.New("failed to add reaction")

	// ErrApprovePR is returned when approving a PR fails
	ErrApprovePR = errors.New("failed to approve PR")

	// ErrMergePR is returned when merging a PR fails
	ErrMergePR = errors.New("failed to merge PR")

	// ErrGetWorkingDirectory is returned when getting working directory fails
	ErrGetWorkingDirectory = errors.New("failed to get the working directory")

	// ErrInitPermissions is returned when initializing permissions fails
	ErrInitPermissions = errors.New("failed to initialize permissions")

	// ErrStepSummary is returned when step summary operations fail
	ErrStepSummary = errors.New("failed to write step summary")
)

// EnvVarError represents an error related to environment variable validation.
type EnvVarError struct {
	Op      error
	VarName string
}

func NewEnvVarError(op error, varName string) error {
	return &EnvVarError{
		Op:      op,
		VarName: varName,
	}
}

func (e *EnvVarError) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.VarName)
}

func (e *EnvVarError) Unwrap() error {
	return e.Op
}

func (e *EnvVarError) Is(target error) bool {
	var envVarErr *EnvVarError
	if errors.As(target, &envVarErr) {
		return true
	}

	return errors.Is(e.Op, target)
}

// InputError represents an error with parsing or validating input.
type InputError struct {
	Op      error
	Input   string
	Details string
}

func NewInputError(op error, input, details string) error {
	return &InputError{
		Op:      op,
		Input:   input,
		Details: details,
	}
}

func (e *InputError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Op, e.Input, e.Details)
	}

	return fmt.Sprintf("%s: %s", e.Op, e.Input)
}

func (e *InputError) Unwrap() error {
	return e.Op
}

func (e *InputError) Is(target error) bool {
	var inputErr *InputError
	if errors.As(target, &inputErr) {
		return true
	}

	return errors.Is(e.Op, target)
}

// GitHubError represents an error from GitHub API operations.
type GitHubError struct {
	Op  error
	Err error
}

func NewGitHubError(op, err error) error {
	return &GitHubError{
		Op:  op,
		Err: err,
	}
}

func (e *GitHubError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}

	return e.Op.Error()
}

func (e *GitHubError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}

	return e.Op
}

func (e *GitHubError) Is(target error) bool {
	var ghErr *GitHubError
	if errors.As(target, &ghErr) {
		return true
	}

	return errors.Is(e.Op, target)
}
