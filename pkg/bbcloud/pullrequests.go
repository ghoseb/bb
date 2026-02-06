package bbcloud

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
)

// GetPullRequest retrieves a single pull request by ID
func (c *Client) GetPullRequest(ctx context.Context, repoSlug string, prID int) (*PullRequest, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	var pr PullRequest
	err := c.Get(ctx, path, &pr)
	if err != nil {
		return nil, fmt.Errorf("get pull request %d: %w", prID, err)
	}
	
	return &pr, nil
}

// GetPRDiffStats retrieves the diffstat for a pull request
// Returns file-level statistics (lines added/removed per file)
func (c *Client) GetPRDiffStats(ctx context.Context, repoSlug string, prID int) ([]FileStats, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diffstat",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	var result FileStatsList
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, fmt.Errorf("get PR diffstat: %w", err)
	}
	
	return result.Values, nil
}

// GetPRDiff retrieves the full unified diff for a pull request
// Returns the diff as a string in unified diff format
func (c *Client) GetPRDiff(ctx context.Context, repoSlug string, prID int) (string, error) {
	if repoSlug == "" {
		return "", fmt.Errorf("repository slug is required")
	}
	if prID <= 0 {
		return "", fmt.Errorf("pull request ID must be positive")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	
	// For diff, we want plain text, not JSON
	req.Header.Set("Accept", "text/plain")
	
	// Use a bytes.Buffer as an io.Writer to capture the response
	var buf bytes.Buffer
	err = c.client.Do(req, &buf)
	if err != nil {
		return "", fmt.Errorf("get PR diff: %w", err)
	}
	
	return buf.String(), nil
}

// GetPRFileDiff retrieves the diff for a specific file in a pull request
// filePath should be the path to the file relative to the repository root
func (c *Client) GetPRFileDiff(ctx context.Context, repoSlug string, prID int, filePath string) (string, error) {
	if repoSlug == "" {
		return "", fmt.Errorf("repository slug is required")
	}
	if prID <= 0 {
		return "", fmt.Errorf("pull request ID must be positive")
	}
	if filePath == "" {
		return "", fmt.Errorf("file path is required")
	}
	
	// Use the PR diff endpoint with path query parameter
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff?path=%s",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		url.QueryEscape(filePath))
	
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	
	// For diff, we want plain text, not JSON
	req.Header.Set("Accept", "text/plain")
	
	// Use a bytes.Buffer as an io.Writer to capture the response
	var buf bytes.Buffer
	err = c.client.Do(req, &buf)
	if err != nil {
		return "", fmt.Errorf("get file diff: %w", err)
	}
	
	return buf.String(), nil
}

// GetPRActivity retrieves the activity timeline for a pull request
// This includes comments, updates, approvals, and other events
func (c *Client) GetPRActivity(ctx context.Context, repoSlug string, prID int) ([]Activity, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/activity",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	var activities []Activity
	page := 1
	
	for {
		pagedPath := fmt.Sprintf("%s?page=%d", path, page)
		
		var result struct {
			PaginatedResponse
			Values []Activity `json:"values"`
		}
		
		err := c.Get(ctx, pagedPath, &result)
		if err != nil {
			return nil, fmt.Errorf("get PR activity (page %d): %w", page, err)
		}
		
		activities = append(activities, result.Values...)
		
		// Check if there's a next page
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return activities, nil
}

// GetPRStatuses retrieves the commit statuses (build checks) for a pull request
// This is typically used to show CI/CD pipeline results
func (c *Client) GetPRStatuses(ctx context.Context, repoSlug string, prID int) ([]Pipeline, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/statuses",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	var result PipelineList
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, fmt.Errorf("get PR statuses: %w", err)
	}
	
	return result.Values, nil
}

// ListPullRequests lists pull requests for a repository
// state can be "OPEN", "MERGED", "DECLINED", or "" for all states
func (c *Client) ListPullRequests(ctx context.Context, repoSlug string, state string, limit int) ([]PullRequest, error) {
	if repoSlug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	
	var allPRs []PullRequest
	page := 1
	pageLen := 50 // Reasonable default for PRs
	
	if limit > 0 && limit < pageLen {
		pageLen = limit
	}
	
	for {
		path := fmt.Sprintf("/repositories/%s/%s/pullrequests?pagelen=%d&page=%d",
			url.PathEscape(c.workspace),
			url.PathEscape(repoSlug),
			pageLen,
			page)
		
		if state != "" {
			path += "&state=" + url.QueryEscape(state)
		}
		
		var result PullRequestList
		err := c.Get(ctx, path, &result)
		if err != nil {
			return nil, fmt.Errorf("list pull requests (page %d): %w", page, err)
		}
		
		allPRs = append(allPRs, result.Values...)
		
		// Check if we've hit the limit or there's no more data
		if limit > 0 && len(allPRs) >= limit {
			if len(allPRs) > limit {
				allPRs = allPRs[:limit]
			}
			break
		}
		
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return allPRs, nil
}

// ApprovePR approves a pull request
// Returns the updated participant information showing the approval
func (c *Client) ApprovePR(ctx context.Context, repoSlug string, prID int) (*Participant, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)

	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var participant Participant
	err = c.client.Do(req, &participant)
	if err != nil {
		return nil, fmt.Errorf("approve PR %d: %w", prID, err)
	}

	return &participant, nil
}

// UnapprovePR removes approval from a pull request
func (c *Client) UnapprovePR(ctx context.Context, repoSlug string, prID int) error {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return err
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)

	req, err := c.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// DELETE returns 204 No Content, so we don't expect a response body
	err = c.client.Do(req, nil)
	if err != nil {
		return fmt.Errorf("unapprove PR %d: %w", prID, err)
	}

	return nil
}

// RequestChangesPR requests changes on a pull request
// This is done by posting to the request-changes endpoint
func (c *Client) RequestChangesPR(ctx context.Context, repoSlug string, prID int) (*Participant, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/request-changes",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)

	req, err := c.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var participant Participant
	err = c.client.Do(req, &participant)
	if err != nil {
		return nil, fmt.Errorf("request changes on PR %d: %w", prID, err)
	}

	return &participant, nil
}

// UnrequestChangesPR removes the request-changes state from a pull request
func (c *Client) UnrequestChangesPR(ctx context.Context, repoSlug string, prID int) error {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return err
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/request-changes",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)

	req, err := c.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// DELETE returns 204 No Content
	err = c.client.Do(req, nil)
	if err != nil {
		return fmt.Errorf("unrequest changes on PR %d: %w", prID, err)
	}

	return nil
}

// CreatePROptions holds options for creating a pull request
type CreatePROptions struct {
	Title             string
	SourceBranch      string
	DestinationBranch string // empty = repo mainbranch
	CloseSourceBranch bool
	Draft             bool
}

// CreatePR creates a new pull request
func (c *Client) CreatePR(ctx context.Context, repoSlug string, opts CreatePROptions) (*PullRequest, error) {
	if repoSlug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	if opts.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if opts.SourceBranch == "" {
		return nil, fmt.Errorf("source branch is required")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug))

	body := map[string]any{
		"title": opts.Title,
		"source": map[string]any{
			"branch": map[string]string{
				"name": opts.SourceBranch,
			},
		},
		"close_source_branch": opts.CloseSourceBranch,
		"draft":               opts.Draft,
	}

	if opts.DestinationBranch != "" {
		body["destination"] = map[string]any{
			"branch": map[string]string{
				"name": opts.DestinationBranch,
			},
		}
	}

	var pr PullRequest
	err := c.Post(ctx, path, body, &pr)
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}

	return &pr, nil
}
