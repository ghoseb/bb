package bbcloud

import (
	"context"
	"fmt"
	"strings"
)

// CurrentUser returns information about the currently authenticated user
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var user User
	err := c.Get(ctx, "/user", &user)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	return &user, nil
}

// CurrentUserWithScopes fetches the current user and returns granted OAuth scopes
func (c *Client) CurrentUserWithScopes(ctx context.Context) (*User, []string, error) {
	var user User
	headers, err := c.GetWithHeaders(ctx, "/user", &user)
	if err != nil {
		return nil, nil, fmt.Errorf("get current user: %w", err)
	}
	
	// Parse x-oauth-scopes header (comma-separated list)
	scopesHeader := headers.Get("x-oauth-scopes")
	var scopes []string
	if scopesHeader != "" {
		parts := strings.Split(scopesHeader, ",")
		for _, part := range parts {
			scope := strings.TrimSpace(part)
			if scope != "" {
				scopes = append(scopes, scope)
			}
		}
	}
	
	return &user, scopes, nil
}
