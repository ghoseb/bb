package bbcloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ghoseb/bb/pkg/httpx"
)

const (
	// DefaultBaseURL is the default Bitbucket Cloud API base URL
	DefaultBaseURL = "https://api.bitbucket.org/2.0"
	
	// DefaultUserAgent is the default User-Agent header value
	DefaultUserAgent = "bb-cli"
)

// Client provides access to the Bitbucket Cloud API
type Client struct {
	client    *httpx.Client
	workspace string
}

// Options configures a Bitbucket Cloud client
type Options struct {
	// BaseURL is the Bitbucket Cloud API base URL (defaults to DefaultBaseURL)
	BaseURL string
	
	// Username is the Bitbucket username for authentication
	Username string
	
	// Token is the Bitbucket App Password or API token
	Token string
	
	// Workspace is the Bitbucket workspace slug
	Workspace string
	
	// UserAgent is the User-Agent header value (defaults to DefaultUserAgent)
	UserAgent string
	
	// Timeout is the HTTP request timeout (defaults to 30 seconds)
	Timeout time.Duration
	
	// Debug enables debug logging
	Debug bool
}

// New creates a new Bitbucket Cloud API client
func New(opts Options) (*Client, error) {
	if opts.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if opts.Token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if opts.Workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	// Configure retry policy with exponential backoff
	retryPolicy := httpx.RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
	
	httpClient, err := httpx.New(httpx.Options{
		BaseURL:   baseURL,
		Username:  opts.Username,
		Password:  opts.Token, // Bitbucket uses Basic Auth with username:app_password
		UserAgent: userAgent,
		Timeout:   timeout,
		Retry:     retryPolicy,
		Debug:     opts.Debug,
	})
	if err != nil {
		return nil, fmt.Errorf("create HTTP client: %w", err)
	}
	
	return &Client{
		client:    httpClient,
		workspace: opts.Workspace,
	}, nil
}

// HTTP returns the underlying HTTP client for advanced usage
func (c *Client) HTTP() *httpx.Client {
	return c.client
}

// Workspace returns the configured workspace slug
func (c *Client) Workspace() string {
	return c.workspace
}

// NewRequest creates a new HTTP request for the Bitbucket Cloud API
func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	return c.client.NewRequest(ctx, method, path, body)
}

// Do executes an HTTP request and decodes the response into v
func (c *Client) Do(req *http.Request, v any) error {
	return c.client.Do(req, v)
}

// Get is a convenience method for GET requests
func (c *Client) Get(ctx context.Context, path string, v any) error {
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	return c.client.Do(req, v)
}

// GetWithHeaders is a convenience method for GET requests that need response headers
func (c *Client) GetWithHeaders(ctx context.Context, path string, v any) (http.Header, error) {
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return c.client.DoWithHeaders(req, v)
}

// Post is a convenience method for POST requests
func (c *Client) Post(ctx context.Context, path string, body any, v any) error {
	req, err := c.client.NewRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	return c.client.Do(req, v)
}

// Put is a convenience method for PUT requests
func (c *Client) Put(ctx context.Context, path string, body any, v any) error {
	req, err := c.client.NewRequest(ctx, "PUT", path, body)
	if err != nil {
		return err
	}
	return c.client.Do(req, v)
}

// Delete is a convenience method for DELETE requests
func (c *Client) Delete(ctx context.Context, path string) error {
	req, err := c.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	return c.client.Do(req, nil)
}
