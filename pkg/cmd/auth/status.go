package auth

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
	"github.com/ghoseb/bb/pkg/iostreams"
)

type statusOptions struct {
	factory *cmdutil.Factory
}

// NewCmdStatus creates the auth status command
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	opts := &statusOptions{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		Long: `Check if you are authenticated with Bitbucket Cloud.

Reads credentials from the system keyring and verifies them by
fetching the authenticated user info.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), opts)
		},
	}

	return cmd
}

// requiredScopes lists the OAuth scopes needed for bb to function correctly
var requiredScopes = []string{
	"read:user:bitbucket",
	"read:workspace:bitbucket",
	"read:repository:bitbucket",
	"read:pullrequest:bitbucket",
	"write:pullrequest:bitbucket",
	"read:pipeline:bitbucket",
}

func runStatus(ctx context.Context, opts *statusOptions) error {
	ios, _ := opts.factory.Streams()

	// Load credentials using factory cache (avoids multiple keyring prompts)
	creds, err := opts.factory.GetCredentials()
	if err != nil {
		return outputNotAuthenticated(ios, fmt.Sprintf("failed to load credentials: %v", err))
	}

	// Verify credentials by calling API
	client, err := bbcloud.New(bbcloud.Options{
		Workspace: creds.Workspace,
		Username:  creds.Username,
		Token:     creds.Token,
	})
	if err != nil {
		return outputNotAuthenticated(ios, fmt.Sprintf("failed to create API client: %v", err))
	}

	user, grantedScopes, err := client.CurrentUserWithScopes(ctx)
	if err != nil {
		return outputNotAuthenticated(ios, fmt.Sprintf("authentication failed: %v", err))
	}

	// Check for missing scopes
	missing := checkMissingScopes(grantedScopes, requiredScopes)
	
	// Output authenticated status
	result := map[string]interface{}{
		"authenticated": true,
		"username":      user.Username,
		"workspace":     creds.Workspace,
	}
	
	if len(missing) == 0 {
		result["scopes"] = "ok"
	} else {
		result["scopes"] = "missing"
		result["missing_scopes"] = missing
	}

	if err := cmdutil.WriteJSON(ios.Out, result); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}

func checkMissingScopes(granted []string, required []string) []string {
	grantedSet := make(map[string]bool)
	for _, scope := range granted {
		grantedSet[scope] = true
	}
	
	var missing []string
	for _, scope := range required {
		if !grantedSet[scope] {
			missing = append(missing, scope)
		}
	}
	
	return missing
}

func outputNotAuthenticated(ios *iostreams.IOStreams, reason string) error {
	result := map[string]interface{}{
		"authenticated": false,
		"reason":        reason,
	}

	if err := cmdutil.WriteJSON(ios.Out, result); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}
