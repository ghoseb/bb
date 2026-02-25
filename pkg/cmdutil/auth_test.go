package cmdutil

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ghoseb/bb/internal/secret"
)

func TestCredentialsSerialization(t *testing.T) {
	creds := &Credentials{
		Workspace: "test-workspace",
		Email:     "test@example.com",
		Token:     "test-token",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Credentials
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Workspace != creds.Workspace {
		t.Errorf("workspace mismatch: got %q, want %q", decoded.Workspace, creds.Workspace)
	}
	if decoded.Email != creds.Email {
		t.Errorf("email mismatch: got %q, want %q", decoded.Email, creds.Email)
	}
	if decoded.Token != creds.Token {
		t.Errorf("token mismatch: got %q, want %q", decoded.Token, creds.Token)
	}
}

func TestSaveAndLoadCredentials(t *testing.T) {
	// Create temporary directory for file backend
	tmpDir := t.TempDir()
	keyringDir := filepath.Join(tmpDir, "test-keyring")

	// Set environment for file backend
	t.Setenv("BB_ALLOW_INSECURE_STORE", "1")
	t.Setenv("KEYRING_FILE_DIR", keyringDir)
	t.Setenv("BB_KEYRING_PASSPHRASE", "test-passphrase")

	// Open store
	store, err := secret.Open(
		secret.WithAllowFileFallback(true),
		secret.WithFileDir(keyringDir),
		secret.WithPassphrase("test-passphrase"),
	)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	// Save credentials
	original := &Credentials{
		Workspace: "my-workspace",
		Email:     "me@example.com",
		Token:     "my-secret-token",
	}

	if err := SaveCredentialsToStore(store, original); err != nil {
		t.Fatalf("save credentials: %v", err)
	}

	// Load credentials back
	loaded, err := LoadCredentialsFromStore(store)
	if err != nil {
		t.Fatalf("load credentials: %v", err)
	}

	// Verify
	if loaded.Workspace != original.Workspace {
		t.Errorf("workspace mismatch: got %q, want %q", loaded.Workspace, original.Workspace)
	}
	if loaded.Email != original.Email {
		t.Errorf("email mismatch: got %q, want %q", loaded.Email, original.Email)
	}
	if loaded.Token != original.Token {
		t.Errorf("token mismatch: got %q, want %q", loaded.Token, original.Token)
	}
}
