// Package github provides a GitHub API client for Smyklot operations.
//
// It supports PR operations (approve, merge, info), comment posting, and
// emoji reactions through the GitHub REST API v3.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultBaseURL = "https://api.github.com"
	userAgent      = "smyklot-github-app"
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
		httpClient: &http.Client{},
		token:      token,
		baseURL:    baseURL,
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

	_, err := c.makeRequest("POST", path, payload)
	return err
}

// MergePR merges a pull request
//
// This attempts to merge the PR using the default merge method.
func (c *Client) MergePR(owner, repo string, prNumber int) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, prNumber)

	_, err := c.makeRequest("PUT", path, nil)
	return err
}

// GetPRInfo retrieves information about a pull request
//
// Returns a PRInfo struct with details about the PR including number, state,
// mergeable status, author, and approvers.
func (c *Client) GetPRInfo(owner, repo string, prNumber int) (*PRInfo, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	data, err := c.makeRequest("GET", path, nil)
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

	return info, nil
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
