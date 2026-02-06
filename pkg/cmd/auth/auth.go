package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/internal/secret"
	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type loginOptions struct {
	workspace string
	username  string
	token     string
	
	factory *cmdutil.Factory
}

// NewCmdAuth creates the auth command group
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	opts := &loginOptions{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Bitbucket Cloud",
		Long: `Authenticate with Bitbucket Cloud (default action is login).

Credentials are validated by fetching the authenticated user info,
then stored securely in your system keyring.

The token should be a Bitbucket App Password with appropriate permissions.
You can create one at: https://bitbucket.org/account/settings/app-passwords/

To check authentication status:
  bb auth status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default action: run login
			return runLogin(cmd.Context(), opts)
		},
	}

	// Move login flags to parent command
	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "",
		"Bitbucket workspace")
	cmd.Flags().StringVarP(&opts.username, "username", "u", "",
		"Bitbucket username")
	cmd.Flags().StringVarP(&opts.token, "token", "t", "",
		"Bitbucket App Password")

	// Add subcommands
	cmd.AddCommand(NewCmdStatus(f))

	return cmd
}

func runLogin(ctx context.Context, opts *loginOptions) error {
	ios, _ := opts.factory.Streams()
	prompter := opts.factory.Prompter

	// Prompt for missing fields interactively
	if opts.workspace == "" {
		// Try environment variable fallback
		if envWorkspace := os.Getenv("BB_WORKSPACE"); envWorkspace != "" {
			opts.workspace = envWorkspace
		} else {
			_, _ = fmt.Fprintln(ios.ErrOut, "Log in to Bitbucket Cloud")
			_, _ = fmt.Fprintln(ios.ErrOut)
			workspace, err := prompter.Input("Bitbucket workspace: ")
			if err != nil {
				return fmt.Errorf("read workspace: %w", err)
			}
			if workspace == "" {
				return fmt.Errorf("workspace is required")
			}
			opts.workspace = workspace
		}
	}

	if opts.username == "" {
		// Try environment variable fallback
		if envUsername := os.Getenv("BB_USERNAME"); envUsername != "" {
			opts.username = envUsername
		} else {
			username, err := prompter.Input("Bitbucket username: ")
			if err != nil {
				return fmt.Errorf("read username: %w", err)
			}
			if username == "" {
				return fmt.Errorf("username is required")
			}
			opts.username = username
		}
	}

	if opts.token == "" {
		// Try environment variable fallback
		if envToken := os.Getenv("BB_TOKEN"); envToken != "" {
			opts.token = envToken
		} else {
			_, _ = fmt.Fprintln(ios.ErrOut, "Tip: Create an App Password at https://bitbucket.org/account/settings/app-passwords/")
			token, err := prompter.Password("App Password (input hidden): ")
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			if token == "" {
				return fmt.Errorf("token is required")
			}
			opts.token = token
		}
	}

	// Test credentials by creating a client and fetching user info
	client, err := bbcloud.New(bbcloud.Options{
		Workspace: opts.workspace,
		Username:  opts.username,
		Token:     opts.token,
	})
	if err != nil {
		return fmt.Errorf("create API client: %w", err)
	}

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Credentials are valid, store them in keyring
	store, err := secret.Open(secret.WithAllowFileFallback(true))
	if err != nil {
		return fmt.Errorf("open secret store: %w", err)
	}

	// Store credentials as a single JSON blob to avoid multiple keyring unlock prompts
	creds := &cmdutil.Credentials{
		Workspace: opts.workspace,
		Username:  opts.username,
		Token:     opts.token,
	}
	if err := cmdutil.SaveCredentialsToStore(store, creds); err != nil {
		return err
	}

	// Output JSON result
	result := map[string]interface{}{
		"status":    "success",
		"username":  user.Username,
		"workspace": opts.workspace,
	}

	if err := cmdutil.WriteJSON(ios.Out, result); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}
