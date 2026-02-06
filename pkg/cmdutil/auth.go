package cmdutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ghoseb/bb/internal/secret"
	"github.com/ghoseb/bb/pkg/bbcloud"
)

// Credentials holds Bitbucket Cloud authentication credentials
type Credentials struct {
	Workspace string
	Username  string
	Token     string
}

// LoadCredentialsFromStore loads credentials from an existing secret store.
// Credentials are stored as a single JSON blob to avoid multiple keyring unlock prompts.
func LoadCredentialsFromStore(store *secret.Store) (*Credentials, error) {
	credsJSON, err := store.Get("bb/credentials")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not authenticated (run 'bb auth')")
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(credsJSON), &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return &creds, nil
}

// SaveCredentialsToStore saves credentials to the secret store as a single JSON blob
// to avoid multiple keyring unlock prompts on subsequent reads.
func SaveCredentialsToStore(store *secret.Store, creds *Credentials) error {
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := store.Set("bb/credentials", string(credsJSON)); err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	return nil
}

// NewBBCloudClient creates a new Bitbucket Cloud API client using cached credentials
// If workspace is provided, it overrides the stored workspace
func (f *Factory) NewBBCloudClient(workspaceOverride string) (*bbcloud.Client, error) {
	creds, err := f.GetCredentials()
	if err != nil {
		return nil, err
	}

	workspace := creds.Workspace
	if workspaceOverride != "" {
		workspace = workspaceOverride
	}

	client, err := bbcloud.New(bbcloud.Options{
		Workspace: workspace,
		Username:  creds.Username,
		Token:     creds.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}

	return client, nil
}
