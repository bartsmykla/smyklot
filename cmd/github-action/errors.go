package main

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure cases.
var (
	ErrMissingEnvVar       = errors.New("required environment variable is missing")
	ErrInvalidInput        = errors.New("invalid input provided")
	ErrGitHubClient        = errors.New("GitHub client error")
	ErrPermissionCheck     = errors.New("permission check failed")
	ErrCommandExecution    = errors.New("command execution failed")
	ErrPostComment         = errors.New("failed to post comment")
	ErrAddReaction         = errors.New("failed to add reaction")
	ErrApprovePR           = errors.New("failed to approve PR")
	ErrMergePR             = errors.New("failed to merge PR")
	ErrGetPRInfo           = errors.New("failed to get PR info")
	ErrGetWorkingDirectory = errors.New("failed to get the working directory")
	ErrInitPermissions     = errors.New("failed to initialize permissions")
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
