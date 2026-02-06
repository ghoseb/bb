package cmdutil

import (
	"os"
	"sync"

	"github.com/ghoseb/bb/internal/secret"
	"github.com/ghoseb/bb/pkg/iostreams"
	"github.com/ghoseb/bb/pkg/prompter"
)

// Factory holds shared dependencies for command execution.
type Factory struct {
	AppVersion string
	IOStreams  *iostreams.IOStreams
	Prompter   prompter.Prompter

	// secret store cache - keeps keyring unlocked for the session
	storeOnce sync.Once
	store     *secret.Store
	storeErr  error

	// credentials cache
	credsOnce sync.Once
	creds     *Credentials
	credsErr  error
}

// NewFactory constructs a new Factory instance.
func NewFactory(appVersion string, ios *iostreams.IOStreams) *Factory {
	return &Factory{
		AppVersion: appVersion,
		IOStreams:  ios,
		Prompter:   prompter.New(ios.In, ios.Out, ios.ErrOut),
	}
}

// Streams returns the configured IO streams.
func (f *Factory) Streams() (*iostreams.IOStreams, error) {
	return f.IOStreams, nil
}

// GetSecretStore opens the secret store once and caches it for the lifetime of the Factory.
// This keeps the keyring session open and prevents multiple unlock prompts.
func (f *Factory) GetSecretStore() (*secret.Store, error) {
	f.storeOnce.Do(func() {
		f.store, f.storeErr = secret.Open(secret.WithAllowFileFallback(true))
	})
	return f.store, f.storeErr
}

// GetCredentials loads credentials from the keyring once and caches them for the lifetime of the Factory.
// This prevents multiple keyring unlock prompts during a single CLI invocation.
func (f *Factory) GetCredentials() (*Credentials, error) {
	f.credsOnce.Do(func() {
		f.creds, f.credsErr = f.loadCredentials()
	})
	return f.creds, f.credsErr
}

// loadCredentialsFromEnv returns credentials from environment variables if all three are set.
// This bypasses the keyring entirely — useful for development and CI.
func loadCredentialsFromEnv() *Credentials {
	ws := os.Getenv("BB_WORKSPACE")
	user := os.Getenv("BB_USERNAME")
	token := os.Getenv("BB_TOKEN")
	if ws != "" && user != "" && token != "" {
		return &Credentials{Workspace: ws, Username: user, Token: token}
	}
	return nil
}

// loadCredentials loads credentials from env vars first, then falls back to keyring.
func (f *Factory) loadCredentials() (*Credentials, error) {
	if creds := loadCredentialsFromEnv(); creds != nil {
		return creds, nil
	}

	store, err := f.GetSecretStore()
	if err != nil {
		return nil, err
	}

	return LoadCredentialsFromStore(store)
}
