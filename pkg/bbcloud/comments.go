package bbcloud

import (
	"context"
	"fmt"
	"net/url"
)

// ListPRComments retrieves all comments for a pull request
// Returns both general and inline comments
func (c *Client) ListPRComments(ctx context.Context, repoSlug string, prID int) ([]Comment, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	var allComments []Comment
	page := 1
	
	for {
		pagedPath := fmt.Sprintf("%s?pagelen=100&page=%d", path, page)
		
		var result CommentList
		err := c.Get(ctx, pagedPath, &result)
		if err != nil {
			return nil, fmt.Errorf("list PR comments (page %d): %w", page, err)
		}
		
		allComments = append(allComments, result.Values...)
		
		// Check if there's a next page
		if result.Next == "" {
			break
		}
		
		page++
	}
	
	return allComments, nil
}

// ListGeneralComments retrieves only general (non-inline) comments for a pull request
func (c *Client) ListGeneralComments(ctx context.Context, repoSlug string, prID int) ([]Comment, error) {
	allComments, err := c.ListPRComments(ctx, repoSlug, prID)
	if err != nil {
		return nil, err
	}
	
	var generalComments []Comment
	for _, comment := range allComments {
		if !comment.IsInline() {
			generalComments = append(generalComments, comment)
		}
	}
	
	return generalComments, nil
}

// ListInlineComments retrieves only inline comments for a pull request
// If filePath is non-empty, only returns comments for that specific file
func (c *Client) ListInlineComments(ctx context.Context, repoSlug string, prID int, filePath string) ([]Comment, error) {
	allComments, err := c.ListPRComments(ctx, repoSlug, prID)
	if err != nil {
		return nil, err
	}
	
	var inlineComments []Comment
	for _, comment := range allComments {
		if comment.IsInline() {
			// If filePath is specified, filter by file
			if filePath == "" || comment.Inline.Path == filePath {
				inlineComments = append(inlineComments, comment)
			}
		}
	}
	
	return inlineComments, nil
}

// GetComment retrieves a specific comment by ID
func (c *Client) GetComment(ctx context.Context, repoSlug string, prID int, commentID int) (*Comment, error) {
	if err := c.validateCommentArgs(repoSlug, prID, commentID); err != nil {
		return nil, err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		commentID)
	
	var comment Comment
	err := c.Get(ctx, path, &comment)
	if err != nil {
		return nil, fmt.Errorf("get comment %d: %w", commentID, err)
	}
	
	return &comment, nil
}

// CreateComment creates a new general (non-inline) comment on a pull request
func (c *Client) CreateComment(ctx context.Context, repoSlug string, prID int, message string) (*Comment, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	body := map[string]any{
		"content": map[string]string{
			"raw": message,
		},
	}
	
	var comment Comment
	err := c.Post(ctx, path, body, &comment)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	
	return &comment, nil
}

// CreateInlineComment creates a new inline comment on a specific line or range
// For single-line: pass lineStart = 0, lineEnd = the line number
// For range: pass lineStart = start line, lineEnd = end line
func (c *Client) CreateInlineComment(ctx context.Context, repoSlug string, prID int, message string, filePath string, lineStart int, lineEnd int) (*Comment, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	if filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if lineEnd <= 0 {
		return nil, fmt.Errorf("line number must be positive")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	inline := map[string]any{
		"path": filePath,
		"to":   lineEnd,
	}
	
	// For range comments, set start_to if lineStart is provided and different from lineEnd
	if lineStart > 0 && lineStart != lineEnd {
		inline["start_to"] = lineStart
	}
	
	body := map[string]any{
		"content": map[string]string{
			"raw": message,
		},
		"inline": inline,
	}
	
	var comment Comment
	err := c.Post(ctx, path, body, &comment)
	if err != nil {
		return nil, fmt.Errorf("create inline comment: %w", err)
	}
	
	return &comment, nil
}

// ReplyToComment creates a reply to an existing comment
func (c *Client) ReplyToComment(ctx context.Context, repoSlug string, prID int, parentID int, message string) (*Comment, error) {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return nil, err
	}
	if parentID <= 0 {
		return nil, fmt.Errorf("parent comment ID must be positive")
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID)
	
	body := map[string]any{
		"content": map[string]string{
			"raw": message,
		},
		"parent": map[string]int{
			"id": parentID,
		},
	}
	
	var comment Comment
	err := c.Post(ctx, path, body, &comment)
	if err != nil {
		return nil, fmt.Errorf("reply to comment: %w", err)
	}
	
	return &comment, nil
}

// UpdateComment updates an existing comment
func (c *Client) UpdateComment(ctx context.Context, repoSlug string, prID int, commentID int, message string) (*Comment, error) {
	if err := c.validateCommentArgs(repoSlug, prID, commentID); err != nil {
		return nil, err
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		commentID)
	
	body := map[string]any{
		"content": map[string]string{
			"raw": message,
		},
	}
	
	var comment Comment
	err := c.Put(ctx, path, body, &comment)
	if err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}
	
	return &comment, nil
}

// DeleteComment deletes a comment
func (c *Client) DeleteComment(ctx context.Context, repoSlug string, prID int, commentID int) error {
	if err := c.validateCommentArgs(repoSlug, prID, commentID); err != nil {
		return err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		commentID)
	
	err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	
	return nil
}

// ResolveComment marks a comment as resolved
func (c *Client) ResolveComment(ctx context.Context, repoSlug string, prID int, commentID int) error {
	if err := c.validateCommentArgs(repoSlug, prID, commentID); err != nil {
		return err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		commentID)
	
	// POST returns 204 No Content on success
	err := c.Post(ctx, path, nil, nil)
	if err != nil {
		return fmt.Errorf("resolve comment: %w", err)
	}
	
	return nil
}

// ReopenComment marks a resolved comment as unresolved (reopens it)
func (c *Client) ReopenComment(ctx context.Context, repoSlug string, prID int, commentID int) error {
	if err := c.validateCommentArgs(repoSlug, prID, commentID); err != nil {
		return err
	}
	
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve",
		url.PathEscape(c.workspace),
		url.PathEscape(repoSlug),
		prID,
		commentID)
	
	// DELETE returns 204 No Content on success
	err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("reopen comment: %w", err)
	}
	
	return nil
}
