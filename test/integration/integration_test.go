package integration

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ghoseb/bb/pkg/bbcloud"
)

// loadEnv loads environment variables from .env file
func loadEnv(t *testing.T) {
	file, err := os.Open("../../.env")
	if err != nil {
		t.Skip("No .env file found, skipping integration tests")
		return
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"")
			_ = os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading .env file: %v", err)
	}
}

// getClient creates a Bitbucket Cloud client from environment variables
func getClient(t *testing.T) *bbcloud.Client {
	workspace := os.Getenv("BB_WORKSPACE")
	username := os.Getenv("BB_USERNAME")
	token := os.Getenv("BB_TOKEN")

	if workspace == "" || username == "" || token == "" {
		t.Skip("Missing BB_WORKSPACE, BB_USERNAME, or BB_TOKEN environment variables")
	}

	client, err := bbcloud.New(bbcloud.Options{
		Workspace: workspace,
		Username:  username,
		Token:     token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func TestIntegration_CurrentUser(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := client.CurrentUser(ctx)
	if err != nil {
		t.Fatalf("CurrentUser failed: %v", err)
	}

	if user.Username == "" {
		t.Error("Expected username to be non-empty")
	}

	t.Logf("✓ Authenticated as: %s (%s)", user.Username, user.DisplayName)
}

func TestIntegration_ListRepositories(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 10)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}

	if len(repos) == 0 {
		t.Skip("No repositories found in workspace")
	}

	t.Logf("✓ Found %d repositories", len(repos))
	for i, repo := range repos {
		t.Logf("  %d. %s (%s) - private: %v", i+1, repo.Name, repo.Slug, repo.IsPrivate)
	}

	// Store first repo for other tests
	if len(repos) > 0 {
		_ = os.Setenv("TEST_REPO_SLUG", repos[0].Slug)
	}
}

func TestIntegration_GetRepository(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// First get a repo to test with
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 1)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found in workspace")
	}

	repoSlug := repos[0].Slug

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	repo, err := client.GetRepository(ctx2, repoSlug)
	if err != nil {
		t.Fatalf("GetRepository failed: %v", err)
	}

	if repo.Slug != repoSlug {
		t.Errorf("Expected slug %s, got %s", repoSlug, repo.Slug)
	}

	t.Logf("✓ Repository: %s", repo.FullName)
	t.Logf("  Description: %s", repo.Description)
	t.Logf("  Private: %v", repo.IsPrivate)
	t.Logf("  Created: %s", repo.CreatedOn.Format(time.RFC3339))
}

func TestIntegration_ListPullRequests(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo first
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	// Try to find a repo with PRs
	var testRepo string
	var prs []bbcloud.PullRequest

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err = client.ListPullRequests(ctx2, repo.Slug, "", 5)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found in any repository")
	}

	t.Logf("✓ Found %d pull requests in %s", len(prs), testRepo)
	for i, pr := range prs {
		t.Logf("  %d. PR #%d: %s (%s)", i+1, pr.ID, pr.Title, pr.State)
	}

	// Store for other tests
	if len(prs) > 0 {
		_ = os.Setenv("TEST_REPO_SLUG", testRepo)
		_ = os.Setenv("TEST_PR_ID", string(rune(prs[0].ID)))
	}
}

func TestIntegration_GetPullRequest(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo with PRs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	var testRepo string
	var testPRID int

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err := client.ListPullRequests(ctx2, repo.Slug, "", 1)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			testPRID = prs[0].ID
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found")
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()

	pr, err := client.GetPullRequest(ctx3, testRepo, testPRID)
	if err != nil {
		t.Fatalf("GetPullRequest failed: %v", err)
	}

	t.Logf("✓ Pull Request #%d: %s", pr.ID, pr.Title)
	t.Logf("  State: %s", pr.State)
	if pr.Author != nil {
		t.Logf("  Author: %s", pr.Author.Username)
	}
	if pr.Source != nil && pr.Source.Branch != nil {
		t.Logf("  Source: %s", pr.Source.Branch.Name)
	}
	if pr.Destination != nil && pr.Destination.Branch != nil {
		t.Logf("  Target: %s", pr.Destination.Branch.Name)
	}
	t.Logf("  Created: %s", pr.CreatedOn.Format(time.RFC3339))
}

func TestIntegration_GetPRDiffStats(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo with PRs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	var testRepo string
	var testPRID int

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err := client.ListPullRequests(ctx2, repo.Slug, "", 1)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			testPRID = prs[0].ID
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found")
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()

	stats, err := client.GetPRDiffStats(ctx3, testRepo, testPRID)
	if err != nil {
		t.Fatalf("GetPRDiffStats failed: %v", err)
	}

	t.Logf("✓ Diff stats for PR #%d: %d files changed", testPRID, len(stats))
	for i, stat := range stats {
		if i >= 5 {
			t.Logf("  ... and %d more files", len(stats)-5)
			break
		}
		t.Logf("  %s: +%d -%d (%s)", stat.Path, stat.LinesAdded, stat.LinesRemoved, stat.Status)
	}
}

func TestIntegration_ListPRComments(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo with PRs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	var testRepo string
	var testPRID int

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err := client.ListPullRequests(ctx2, repo.Slug, "", 5)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			testPRID = prs[0].ID
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found")
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()

	comments, err := client.ListPRComments(ctx3, testRepo, testPRID)
	if err != nil {
		t.Fatalf("ListPRComments failed: %v", err)
	}

	t.Logf("✓ Found %d comments on PR #%d", len(comments), testPRID)
	for i, comment := range comments {
		if i >= 3 {
			t.Logf("  ... and %d more comments", len(comments)-3)
			break
		}
		author := "unknown"
		if comment.User != nil {
			author = comment.User.Username
		}
		content := ""
		if comment.Content != nil {
			content = comment.Content.Raw
			if len(content) > 50 {
				content = content[:50] + "..."
			}
		}
		t.Logf("  Comment #%d by %s: %s", comment.ID, author, content)
	}
}

func TestIntegration_GetPRActivity(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo with PRs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	var testRepo string
	var testPRID int

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err := client.ListPullRequests(ctx2, repo.Slug, "", 1)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			testPRID = prs[0].ID
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found")
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel3()

	activities, err := client.GetPRActivity(ctx3, testRepo, testPRID)
	if err != nil {
		t.Fatalf("GetPRActivity failed: %v", err)
	}

	t.Logf("✓ Found %d activity items on PR #%d", len(activities), testPRID)
	
	updateCount := 0
	commentCount := 0
	approvalCount := 0
	
	for _, activity := range activities {
		if activity.Update != nil {
			updateCount++
		}
		if activity.Comment != nil {
			commentCount++
		}
		if activity.Approval != nil {
			approvalCount++
		}
	}
	
	t.Logf("  Updates: %d, Comments: %d, Approvals: %d", updateCount, commentCount, approvalCount)
}

func TestIntegration_GetPRPipelines(t *testing.T) {
	loadEnv(t)
	client := getClient(t)

	// Get a repo with PRs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, 5)
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Skip("No repositories found")
	}

	var testRepo string
	var testPRID int

	for _, repo := range repos {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		prs, err := client.ListPullRequests(ctx2, repo.Slug, "", 1)
		cancel2()
		
		if err == nil && len(prs) > 0 {
			testRepo = repo.Slug
			testPRID = prs[0].ID
			break
		}
	}

	if testRepo == "" {
		t.Skip("No pull requests found")
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel3()

	statuses, err := client.GetPRPipelines(ctx3, testRepo, testPRID)
	if err != nil {
		// Some repos might not have pipelines enabled
		t.Logf("Note: GetPRPipelines returned error (pipelines may not be enabled): %v", err)
		t.Skip("Skipping pipeline test")
	}

	t.Logf("✓ Found %d commit statuses for PR #%d", len(statuses), testPRID)
	for i, status := range statuses {
		if i >= 3 {
			t.Logf("  ... and %d more statuses", len(statuses)-3)
			break
		}
		t.Logf("  %s: %s - %s", status.Key, status.State, status.Name)
	}
}
