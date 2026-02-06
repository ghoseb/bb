package bbcloud

import (
	"context"
	"fmt"
	"net/url"
)

// ListRepositories lists repositories in the configured workspace
// If limit is 0, all repositories are returned (with pagination)
// If limit > 0, at most limit repositories are returned
func (c *Client) ListRepositories(ctx context.Context, limit int) ([]Repository, error) {
	var allRepos []Repository
	page := 1
	pageLen := 100 // Bitbucket Cloud max page size
	
	// If limit is set and less than pageLen, use it
	if limit > 0 && limit < pageLen {
		pageLen = limit
	}
	
	for {
		path := fmt.Sprintf("/repositories/%s?pagelen=%d&page=%d", 
			url.PathEscape(c.workspace), pageLen, page)
		
		var result RepositoryList
		err := c.Get(ctx, path, &result)
		if err != nil {
			return nil, fmt.Errorf("list repositories (page %d): %w", page, err)
		}
		
		allRepos = append(allRepos, result.Values...)
		
		// Check if we've hit the limit or there's no more data
		if limit > 0 && len(allRepos) >= limit {
			// Trim to exact limit if we exceeded it
			if len(allRepos) > limit {
				allRepos = allRepos[:limit]
			}
			break
		}
		
		// Check if there's a next page
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return allRepos, nil
}

// GetRepository retrieves a single repository by slug
func (c *Client) GetRepository(ctx context.Context, slug string) (*Repository, error) {
	if slug == "" {
		return nil, fmt.Errorf("repository slug is required")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s", 
		url.PathEscape(c.workspace), 
		url.PathEscape(slug))
	
	var repo Repository
	err := c.Get(ctx, path, &repo)
	if err != nil {
		return nil, fmt.Errorf("get repository %q: %w", slug, err)
	}
	
	return &repo, nil
}
