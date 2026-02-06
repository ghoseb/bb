package bbcloud

import "fmt"

// validatePRArgs validates common PR-related arguments
func (c *Client) validatePRArgs(repoSlug string, prID int) error {
	if repoSlug == "" {
		return fmt.Errorf("repository slug is required")
	}
	if prID <= 0 {
		return fmt.Errorf("pull request ID must be positive")
	}
	return nil
}

// validateCommentArgs validates comment-related arguments
func (c *Client) validateCommentArgs(repoSlug string, prID int, commentID int) error {
	if err := c.validatePRArgs(repoSlug, prID); err != nil {
		return err
	}
	if commentID <= 0 {
		return fmt.Errorf("comment ID must be positive")
	}
	return nil
}
