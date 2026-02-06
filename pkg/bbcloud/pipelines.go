package bbcloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// GetPRPipelines retrieves all pipelines (build statuses) for a pull request
// This returns the commit statuses which include pipeline runs
func (c *Client) GetPRPipelines(ctx context.Context, repoSlug string, prID int) ([]CommitStatus, error) {
	if repoSlug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	if prID <= 0 {
		return nil, fmt.Errorf("pull request ID must be positive")
	}
	
	// First, get the PR to find the source commit
	pr, err := c.GetPullRequest(ctx, repoSlug, prID)
	if err != nil {
		return nil, fmt.Errorf("get pull request: %w", err)
	}
	
	if pr.Source == nil || pr.Source.Commit == nil {
		return nil, fmt.Errorf("pull request has no source commit")
	}
	
	commitHash := pr.Source.Commit.Hash
	
	// Get statuses for the commit
	path := fmt.Sprintf("/repositories/%s/%s/commit/%s/statuses",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		commitHash)
	
	var statuses []CommitStatus
	page := 1
	
	for {
		pagedPath := fmt.Sprintf("%s?pagelen=100&page=%d", path, page)
		
		var result CommitStatusList
		err := c.Get(ctx, pagedPath, &result)
		if err != nil {
			return nil, fmt.Errorf("get commit statuses (page %d): %w", page, err)
		}
		
		statuses = append(statuses, result.Values...)
		
		// Check if there's a next page
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return statuses, nil
}

// GetPipelineStatus retrieves the status of a specific pipeline by UUID
func (c *Client) GetPipelineStatus(ctx context.Context, repoSlug string, pipelineUUID string) (*Pipeline, error) {
	if repoSlug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	if pipelineUUID == "" {
		return nil, fmt.Errorf("pipeline UUID is required")
	}
	
	// Bitbucket UUIDs include braces like {uuid}, which must be properly encoded
	// If the UUID doesn't already have braces, add them
	if !strings.HasPrefix(pipelineUUID, "{") {
		pipelineUUID = "{" + pipelineUUID + "}"
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		url.PathEscape(pipelineUUID))
	
	var pipeline Pipeline
	err := c.Get(ctx, path, &pipeline)
	if err != nil {
		return nil, fmt.Errorf("get pipeline status: %w", err)
	}
	
	return &pipeline, nil
}

// ListPipelines lists all pipelines for a repository
func (c *Client) ListPipelines(ctx context.Context, repoSlug string, limit int) ([]Pipeline, error) {
	if repoSlug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	
	var allPipelines []Pipeline
	page := 1
	pageLen := 50 // Reasonable default for pipelines
	
	if limit > 0 && limit < pageLen {
		pageLen = limit
	}
	
	for {
		path := fmt.Sprintf("/repositories/%s/%s/pipelines/?pagelen=%d&page=%d",
			url.PathEscape(c.workspace),
			url.PathEscape(repoSlug),
			pageLen,
			page)
		
		var result PipelineList
		err := c.Get(ctx, path, &result)
		if err != nil {
			return nil, fmt.Errorf("list pipelines (page %d): %w", page, err)
		}
		
		allPipelines = append(allPipelines, result.Values...)
		
		// Check if we've hit the limit or there's no more data
		if limit > 0 && len(allPipelines) >= limit {
			if len(allPipelines) > limit {
				allPipelines = allPipelines[:limit]
			}
			break
		}
		
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return allPipelines, nil
}
