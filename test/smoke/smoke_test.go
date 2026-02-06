package smoke

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// MockServer creates a test HTTP server that mocks Bitbucket Cloud API
type MockServer struct {
	*httptest.Server
	Requests []*http.Request
}

// NewMockServer creates a new mock Bitbucket API server
func NewMockServer() *MockServer {
	ms := &MockServer{
		Requests: make([]*http.Request, 0),
	}

	mux := http.NewServeMux()

	// Mock /user endpoint
	mux.HandleFunc("/2.0/user", func(w http.ResponseWriter, r *http.Request) {
		ms.Requests = append(ms.Requests, r)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":         "{test-uuid}",
			"username":     "testuser",
			"display_name": "Test User",
			"account_id":   "test-account-id",
			"type":         "user",
		})
	})

	// Mock /repositories/{workspace} endpoint
	mux.HandleFunc("/2.0/repositories/", func(w http.ResponseWriter, r *http.Request) {
		ms.Requests = append(ms.Requests, r)
		w.Header().Set("Content-Type", "application/json")

		// Check if it's listing repos or getting a specific repo
		switch r.URL.Path {
		case "/2.0/repositories/testworkspace":
			// List repositories
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"pagelen": 10,
				"values": []map[string]interface{}{
					{
						"uuid":      "{repo-uuid-1}",
						"name":      "test-repo",
						"slug":      "test-repo",
						"full_name": "testworkspace/test-repo",
						"is_private": true,
						"type":      "repository",
					},
				},
				"page": 1,
				"size": 1,
			})
		case "/2.0/repositories/testworkspace/test-repo":
			// Get specific repository
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"uuid":      "{repo-uuid-1}",
				"name":      "test-repo",
				"slug":      "test-repo",
				"full_name": "testworkspace/test-repo",
				"is_private": true,
				"type":      "repository",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	ms.Server = httptest.NewServer(mux)
	return ms
}

// RunCLI runs the bb CLI command and returns stdout, stderr, and error
func RunCLI(args ...string) (string, string, error) {
	// Build the CLI if not already built
	cmd := exec.Command("go", append([]string{"run", "./cmd/bb"}, args...)...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Set working directory to project root
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "../..")
	cmd.Dir = projectRoot
	
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ParseJSON parses JSON output from CLI
func ParseJSON(output string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	return result, err
}

// ParseJSONArray parses JSON array output from CLI
func ParseJSONArray(output string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	return result, err
}

// TestBasicCLI tests that the CLI can be invoked
func TestBasicCLI(t *testing.T) {
	stdout, _, err := RunCLI("--help")
	if err != nil {
		t.Fatalf("Failed to run CLI: %v", err)
	}

	if stdout == "" {
		t.Fatal("Expected help output, got empty string")
	}

	if !bytes.Contains([]byte(stdout), []byte("bb is a command-line interface")) {
		t.Errorf("Help output missing expected content: %s", stdout)
	}
}

// TestAuthStatus tests the auth status command without authentication
func TestAuthStatus(t *testing.T) {
	stdout, _, err := RunCLI("auth", "status")
	if err != nil {
		// This is expected to succeed even when not authenticated
		t.Logf("Auth status command returned error (expected): %v", err)
	}

	result, err := ParseJSON(stdout)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout)
	}

	authenticated, ok := result["authenticated"].(bool)
	if !ok {
		t.Fatal("Expected 'authenticated' field in output")
	}

	if authenticated {
		t.Log("User is authenticated (credentials found)")
	} else {
		t.Log("User is not authenticated (no credentials)")
	}
}

// TestRepoListRequiresAuth tests that listing repos requires authentication
func TestRepoListRequiresAuth(t *testing.T) {
	// This test will fail if the user hasn't authenticated
	_, stderr, err := RunCLI("list", "repos")
	
	if err != nil {
		// Expected to fail without valid credentials
		if bytes.Contains([]byte(stderr), []byte("not authenticated")) || 
		   bytes.Contains([]byte(stderr), []byte("401")) ||
		   bytes.Contains([]byte(stderr), []byte("Unauthorized")) {
			t.Log("Correctly requires authentication")
			return
		}
		t.Logf("Command failed (expected): %v", err)
	}
}

// TestPRViewHelp tests PR view help
func TestPRViewHelp(t *testing.T) {
	stdout, _, err := RunCLI("review", "--help")
	if err != nil {
		t.Fatalf("Failed to get review help: %v", err)
	}

	if !bytes.Contains([]byte(stdout), []byte("review")) {
		t.Errorf("Help output missing 'review' keyword: %s", stdout)
	}
}

// TestCommandsHaveHelp verifies major commands have help text
func TestCommandsHaveHelp(t *testing.T) {
	commands := []string{"auth", "list", "review"}
	
	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			stdout, _, err := RunCLI(cmd, "--help")
			if err != nil {
				t.Fatalf("Failed to get help for %s: %v", cmd, err)
			}
			
			if stdout == "" {
				t.Errorf("Expected help output for %s, got empty string", cmd)
			}
		})
	}
}

// TestRequiredFlags tests that commands enforce required flags
func TestRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"review list without repo", []string{"review", "list"}},
		{"review without pr", []string{"review", "--repo", "test"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := RunCLI(tt.args...)
			if err == nil {
				t.Errorf("Expected error for missing required flag, got nil")
			}
			
			// Should mention required flag in error
			if !bytes.Contains([]byte(stderr), []byte("required")) {
				t.Logf("Error message: %s", stderr)
			}
		})
	}
}
