// Package github provides a GitHub API client for Smyklot operations.
//
// It supports PR operations (approve, merge, info), comment posting, and
// emoji reactions through the GitHub REST API v3.
package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL        = "https://api.github.com"
	userAgent             = "smyklot-github-app"
	defaultTimeout        = 30 * time.Second
	maxIdleConns          = 100
	maxIdleConnsPerHost   = 10
	idleConnTimeout       = 90 * time.Second
	maxRetries            = 3
	maxCodeownersSize     = 1024 * 1024 // 1MB
	maxCommentBodyLength  = 10000       // 10KB
)

// Client is a GitHub API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewClient creates a new GitHub API client
//
// The token parameter is required and must not be empty. The baseURL parameter
// is optional; if empty, the default GitHub API URL will be used.
func NewClient(token, baseURL string) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        maxIdleConns,
				MaxIdleConnsPerHost: maxIdleConnsPerHost,
				IdleConnTimeout:     idleConnTimeout,
			},
		},
		token:   token,
		baseURL: baseURL,
	}, nil
}

// AddReaction adds an emoji reaction to a comment
//
// The reaction parameter should be one of the ReactionType constants
// (ReactionSuccess, ReactionError, ReactionWarning, ReactionEyes).
func (c *Client) AddReaction(owner, repo string, commentID int, reaction ReactionType) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID)

	body := map[string]string{
		"content": string(reaction),
	}

	_, err := c.makeRequest("POST", path, body)
	return err
}

// RemoveReaction removes an emoji reaction from a comment
//
// The reaction parameter should be one of the ReactionType constants.
// This retrieves all reactions on the comment and deletes matching ones.
func (c *Client) RemoveReaction(owner, repo string, commentID int, reaction ReactionType) error {
	// First, get all reactions on the comment
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return err
	}

	var reactions []map[string]interface{}
	if err := json.Unmarshal(data, &reactions); err != nil {
		return NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	// Find and delete matching reactions
	for _, r := range reactions {
		if content, ok := r["content"].(string); ok && content == string(reaction) {
			if id, ok := r["id"].(float64); ok {
				deletePath := fmt.Sprintf(
					"/repos/%s/%s/issues/comments/%d/reactions/%d",
					owner,
					repo,
					commentID,
					int(id),
				)
				if _, err := c.makeRequest("DELETE", deletePath, nil); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// PostComment posts a comment on a pull request
//
// The body parameter must not be empty.
func (c *Client) PostComment(owner, repo string, prNumber int, body string) error {
	if body == "" {
		return ErrEmptyComment
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)

	payload := map[string]string{
		"body": body,
	}

	_, err := c.makeRequest("POST", path, payload)
	return err
}

// ApprovePR approves a pull request
//
// This creates a review with the APPROVE event.
func (c *Client) ApprovePR(owner, repo string, prNumber int) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, prNumber)

	payload := map[string]string{
		"event": "APPROVE",
	}

	_, err := c.makeRequestWithRetry("POST", path, payload)
	return err
}

// DismissReview dismisses all approved reviews by the authenticated user
//
// WARNING: This method calls GET /user which requires special permissions
// not granted to GitHub App installation tokens. Use DismissReviewByUsername instead.
func (c *Client) DismissReview(owner, repo string, prNumber int) error {
	username, err := c.GetAuthenticatedUser()
	if err != nil {
		return err
	}

	return c.DismissReviewByUsername(owner, repo, prNumber, username)
}

// DismissReviewByUsername dismisses all approved reviews by the specified username
//
// This finds all APPROVED reviews by the specified user and dismisses them.
// Recommended for GitHub App installations to avoid GET /user permission issues.
func (c *Client) DismissReviewByUsername(owner, repo string, prNumber int, username string) error {
	reviews, err := c.getPullRequestReviews(owner, repo, prNumber)
	if err != nil {
		return err
	}

	return c.dismissApprovedReviews(owner, repo, prNumber, username, reviews)
}

// GetAuthenticatedUser retrieves the authenticated user's username
func (c *Client) GetAuthenticatedUser() (string, error) {
	userPath := "/user"
	userData, err := c.makeRequest("GET", userPath, nil)
	if err != nil {
		return "", err
	}

	var user map[string]interface{}
	if err := json.Unmarshal(userData, &user); err != nil {
		return "", NewAPIError(ErrResponseParse, 0, "GET", userPath, err)
	}

	username, ok := user["login"].(string)
	if !ok {
		return "", NewAPIError(ErrResponseParse, 0, "GET", userPath, fmt.Errorf("unable to get username"))
	}

	return username, nil
}

// getPullRequestReviews retrieves all reviews for a pull request
func (c *Client) getPullRequestReviews(owner, repo string, prNumber int) ([]map[string]interface{}, error) {
	reviewsPath := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, prNumber)
	reviewsData, err := c.makeRequest("GET", reviewsPath, nil)
	if err != nil {
		return nil, err
	}

	var reviews []map[string]interface{}
	if err := json.Unmarshal(reviewsData, &reviews); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", reviewsPath, err)
	}

	return reviews, nil
}

// dismissApprovedReviews dismisses all approved reviews by the specified user
func (c *Client) dismissApprovedReviews(
	owner, repo string,
	prNumber int,
	username string,
	reviews []map[string]interface{},
) error {
	for _, review := range reviews {
		if err := c.dismissReviewIfApprovedByUser(owner, repo, prNumber, username, review); err != nil {
			return err
		}
	}

	return nil
}

// dismissReviewIfApprovedByUser dismisses a review if it's approved by the specified user
func (c *Client) dismissReviewIfApprovedByUser(
	owner, repo string,
	prNumber int,
	username string,
	review map[string]interface{},
) error {
	state, ok := review["state"].(string)
	if !ok || state != "APPROVED" {
		return nil
	}

	reviewUser, ok := review["user"].(map[string]interface{})
	if !ok {
		return nil
	}

	reviewUsername, ok := reviewUser["login"].(string)
	if !ok || reviewUsername != username {
		return nil
	}

	reviewID, ok := review["id"].(float64)
	if !ok {
		return nil
	}

	dismissPath := fmt.Sprintf(
		"/repos/%s/%s/pulls/%d/reviews/%d/dismissals",
		owner,
		repo,
		prNumber,
		int(reviewID),
	)
	payload := map[string]string{
		"message": "Review dismissed",
	}

	_, err := c.makeRequest("PUT", dismissPath, payload)
	return err
}

// MergePR merges a pull request using the specified merge method
//
// Supported merge methods: merge, squash, rebase
func (c *Client) MergePR(owner, repo string, prNumber int, method MergeMethod) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, prNumber)

	body := map[string]interface{}{
		"merge_method": string(method),
	}

	_, err := c.makeRequestWithRetry("PUT", path, body)
	return err
}

// EnableAutoMerge enables auto-merge for a pull request
//
// This will automatically merge the PR when all required checks pass.
// Uses GraphQL API as auto-merge is not available in REST API.
func (c *Client) EnableAutoMerge(owner, repo string, prNumber int, method MergeMethod) error {
	// Get PR node ID first (required for GraphQL)
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, prNumber)
	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return err
	}

	var prData map[string]interface{}
	if err := json.Unmarshal(data, &prData); err != nil {
		return NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	nodeID, ok := prData["node_id"].(string)
	if !ok {
		return NewAPIError(
			ErrResponseParse,
			0,
			"GET",
			path,
			fmt.Errorf("no node_id in response"),
		)
	}

	// Map merge method to GraphQL enum
	var gqlMethod string
	switch method {
	case MergeMethodMerge:
		gqlMethod = "MERGE"
	case MergeMethodSquash:
		gqlMethod = "SQUASH"
	case MergeMethodRebase:
		gqlMethod = "REBASE"
	default:
		gqlMethod = "MERGE"
	}

	// Enable auto-merge via GraphQL (using parameterized query to prevent injection)
	graphqlPath := "/graphql"
	query := map[string]interface{}{
		"query": `mutation($pullRequestId: ID!, $mergeMethod: PullRequestMergeMethod!) {
			enablePullRequestAutoMerge(input: {pullRequestId: $pullRequestId, mergeMethod: $mergeMethod}) {
				clientMutationId
			}
		}`,
		"variables": map[string]interface{}{
			"pullRequestId": nodeID,
			"mergeMethod":   gqlMethod,
		},
	}

	_, err = c.makeRequest("POST", graphqlPath, query)
	return err
}

// GetPRReactions retrieves all reactions for a pull request (issue)
//
// Returns a slice of Reaction structs containing user and reaction type information.
// This gets reactions on the PR description/body, not on comments.
func (c *Client) GetPRReactions(owner, repo string, prNumber int) ([]Reaction, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/reactions", owner, repo, prNumber)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var rawReactions []map[string]interface{}
	if err := json.Unmarshal(data, &rawReactions); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	reactions := make([]Reaction, 0, len(rawReactions))
	for _, r := range rawReactions {
		reaction := Reaction{}

		if content, ok := r["content"].(string); ok {
			reaction.Type = ReactionType(content)
		}

		if user, ok := r["user"].(map[string]interface{}); ok {
			if login, ok := user["login"].(string); ok {
				reaction.User = login
			}
		}

		if reaction.Type != "" && reaction.User != "" {
			reactions = append(reactions, reaction)
		}
	}

	return reactions, nil
}

// GetCommentReactions retrieves all reactions for a comment
//
// Returns a slice of Reaction structs containing user and reaction type information.
func (c *Client) GetCommentReactions(owner, repo string, commentID int) ([]Reaction, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var rawReactions []map[string]interface{}
	if err := json.Unmarshal(data, &rawReactions); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	reactions := make([]Reaction, 0, len(rawReactions))
	for _, r := range rawReactions {
		reaction := Reaction{}

		if content, ok := r["content"].(string); ok {
			reaction.Type = ReactionType(content)
		}

		if user, ok := r["user"].(map[string]interface{}); ok {
			if login, ok := user["login"].(string); ok {
				reaction.User = login
			}
		}

		if reaction.Type != "" && reaction.User != "" {
			reactions = append(reactions, reaction)
		}
	}

	return reactions, nil
}

// AddLabel adds a label to a pull request
func (c *Client) AddLabel(owner, repo string, prNumber int, label string) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, prNumber)

	payload := map[string][]string{
		"labels": {label},
	}

	_, err := c.makeRequest("POST", path, payload)
	return err
}

// RemoveLabel removes a label from a pull request
func (c *Client) RemoveLabel(owner, repo string, prNumber int, label string) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels/%s", owner, repo, prNumber, label)

	_, err := c.makeRequest("DELETE", path, nil)
	return err
}

// GetLabels retrieves all labels from a pull request
func (c *Client) GetLabels(owner, repo string, prNumber int) ([]string, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, prNumber)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var rawLabels []map[string]interface{}
	if err := json.Unmarshal(data, &rawLabels); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	labels := make([]string, 0, len(rawLabels))
	for _, l := range rawLabels {
		if name, ok := l["name"].(string); ok {
			labels = append(labels, name)
		}
	}

	return labels, nil
}

// GetCodeowners fetches the CODEOWNERS file content from the repository
//
// Returns the decoded content of .github/CODEOWNERS file.
// Returns empty string (not error) if file doesn't exist (404).
func (c *Client) GetCodeowners(owner, repo string) (string, error) {
	path := fmt.Sprintf("/repos/%s/%s/contents/.github/CODEOWNERS", owner, repo)

	data, err := c.makeRequestWithRetry("GET", path, nil)
	if err != nil {
		// Return empty string if CODEOWNERS doesn't exist (404)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return "", nil
		}
		return "", err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return "", NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	content, ok := response["content"].(string)
	if !ok {
		return "", NewAPIError(
			ErrResponseParse,
			0,
			"GET",
			path,
			fmt.Errorf("no content field in response"),
		)
	}

	// GitHub API returns base64-encoded content, decode it
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content, "\n", ""))
	if err != nil {
		return "", NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	// Validate decoded content size to prevent memory exhaustion
	if len(decoded) > maxCodeownersSize {
		return "", NewAPIError(
			ErrResponseParse,
			0,
			"GET",
			path,
			fmt.Errorf("CODEOWNERS file too large: %d bytes (max: %d)", len(decoded), maxCodeownersSize),
		)
	}

	return string(decoded), nil
}

// GetPRInfo retrieves information about a pull request
//
// Returns a PRInfo struct with details about the PR including number, state,
// mergeable status, author, and approvers.
func (c *Client) GetPRInfo(owner, repo string, prNumber int) (*PRInfo, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	data, err := c.makeRequestWithRetry("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	info := &PRInfo{
		Number: prNumber,
	}

	if state, ok := response["state"].(string); ok {
		info.State = state
	}

	if mergeable, ok := response["mergeable"].(bool); ok {
		info.Mergeable = mergeable
	}

	if title, ok := response["title"].(string); ok {
		info.Title = title
	}

	if body, ok := response["body"].(string); ok {
		info.Body = body
	}

	if user, ok := response["user"].(map[string]interface{}); ok {
		if login, ok := user["login"].(string); ok {
			info.Author = login
		}
	}

	if base, ok := response["base"].(map[string]interface{}); ok {
		if ref, ok := base["ref"].(string); ok {
			info.BaseBranch = ref
		}
	}

	// Populate ApprovedBy field
	info.ApprovedBy = c.getApprovers(owner, repo, prNumber)

	return info, nil
}

// getApprovers retrieves the list of users who have approved a PR
func (c *Client) getApprovers(owner, repo string, prNumber int) []string {
	reviews, err := c.getPullRequestReviews(owner, repo, prNumber)
	if err != nil {
		// Return empty slice if we can't get reviews
		return []string{}
	}

	approvers := make([]string, 0)
	approverSet := make(map[string]bool)

	for _, review := range reviews {
		login := c.extractApproverFromReview(review)
		if login == "" {
			continue
		}

		// Use a set to deduplicate approvers
		if !approverSet[login] {
			approverSet[login] = true
			approvers = append(approvers, login)
		}
	}

	return approvers
}

// extractApproverFromReview extracts the approver username from a review
func (c *Client) extractApproverFromReview(review map[string]interface{}) string {
	state, ok := review["state"].(string)
	if !ok || state != "APPROVED" {
		return ""
	}

	user, ok := review["user"].(map[string]interface{})
	if !ok {
		return ""
	}

	login, ok := user["login"].(string)
	if !ok {
		return ""
	}

	return login
}

// GetPRComments retrieves all comments on a pull request
//
// Returns a slice of comment data including ID, user, and body.
func (c *Client) GetPRComments(owner, repo string, prNumber int) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)

	data, err := c.makeRequestWithRetry("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var comments []map[string]interface{}
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	return comments, nil
}

// DeleteComment deletes a comment from a pull request
func (c *Client) DeleteComment(owner, repo string, commentID int) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/comments/%d", owner, repo, commentID)

	_, err := c.makeRequest("DELETE", path, nil)
	return err
}

// GetOpenPRs retrieves all open pull requests in a repository
//
// Returns a slice of PR data including number, title, and state.
func (c *Client) GetOpenPRs(owner, repo string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)

	data, err := c.makeRequestWithRetry("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var prs []map[string]interface{}
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	// Filter only open PRs
	var openPRs []map[string]interface{}
	for _, pr := range prs {
		if state, ok := pr["state"].(string); ok && state == "open" {
			openPRs = append(openPRs, pr)
		}
	}

	return openPRs, nil
}

// HasWritePermission checks if the user has write/admin permission to the repository
func (c *Client) HasWritePermission(owner, repo, username string) (bool, error) {
	path := fmt.Sprintf("/repos/%s/%s/collaborators/%s/permission", owner, repo, username)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		// If user is not a collaborator, return false (not an error)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return false, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	permission, ok := response["permission"].(string)
	if !ok {
		return false, NewAPIError(
			ErrResponseParse,
			0,
			"GET",
			path,
			fmt.Errorf("no permission field in response"),
		)
	}

	// admin and write permissions allow approving/merging
	return permission == "admin" || permission == "write", nil
}

// IsMergeQueueEnabled checks if merge queue is enabled for a branch
func (c *Client) IsMergeQueueEnabled(owner, repo, branch string) (bool, error) {
	path := fmt.Sprintf("/repos/%s/%s/branches/%s/protection", owner, repo, branch)

	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		// 404 means branch protection not enabled
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}

	var protection map[string]interface{}
	if err := json.Unmarshal(data, &protection); err != nil {
		return false, NewAPIError(ErrResponseParse, 0, "GET", path, err)
	}

	// Check if merge queue is enabled
	if mergeQueue, ok := protection["required_pull_request_reviews"].(map[string]interface{}); ok {
		if enabled, ok := mergeQueue["require_last_push_approval"].(bool); ok && enabled {
			return true, nil
		}
	}

	// Also check for the merge_queue field directly
	if mergeQueue, ok := protection["merge_queue"].(map[string]interface{}); ok {
		if enabled, ok := mergeQueue["enabled"].(bool); ok {
			return enabled, nil
		}
	}

	return false, nil
}

// makeRequestWithRetry makes an HTTP request with retry logic and exponential backoff
func (c *Client) makeRequestWithRetry(method, path string, payload interface{}) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		data, err := c.makeRequest(method, path, payload)
		if err == nil {
			return data, nil
		}

		lastErr = err

		// Check for rate limiting (429) or server errors (5xx)
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode == 429 || (apiErr.StatusCode >= 500 && apiErr.StatusCode < 600) {
				// Exponential backoff: 1s, 2s, 4s
				backoff := time.Duration(1<<uint(attempt)) * time.Second
				time.Sleep(backoff)
				continue
			}
		}

		// For other errors, don't retry
		return nil, err
	}

	return nil, lastErr
}

// makeRequest makes an HTTP request to the GitHub API
func (c *Client) makeRequest(method, path string, payload interface{}) ([]byte, error) {
	url := c.baseURL + path

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, NewAPIError(ErrAPIRequest, 0, method, path, err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, NewAPIError(ErrAPIRequest, 0, method, path, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewAPIError(ErrAPIRequest, 0, method, path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewAPIError(ErrAPIRequest, resp.StatusCode, method, path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("status code %d", resp.StatusCode)

		var errResp map[string]interface{}
		if json.Unmarshal(data, &errResp) == nil {
			if message, ok := errResp["message"].(string); ok {
				errMsg = message
			}
		}

		return nil, NewAPIError(
			ErrAPIRequest,
			resp.StatusCode,
			method,
			path,
			fmt.Errorf("%s", errMsg),
		)
	}

	return data, nil
}
