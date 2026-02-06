package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	envWorkspace = "BB_WORKSPACE"
)

// Workspace resolves the workspace from flags or environment.
func Workspace(cmd *cobra.Command) (string, error) {
	value := strings.TrimSpace(flagValue(cmd, "workspace"))
	if value == "" {
		value = strings.TrimSpace(os.Getenv(envWorkspace))
	}
	if value == "" {
		return "", fmt.Errorf("workspace is required (set --workspace or %s)", envWorkspace)
	}
	return value, nil
}

// Repo resolves the repository slug from flags.
func Repo(cmd *cobra.Command) (string, error) {
	value := strings.TrimSpace(flagValue(cmd, "repo"))
	if value == "" {
		return "", fmt.Errorf("repo is required (set --repo)")
	}
	return value, nil
}

// RequireFlags ensures a set of flags are provided.
func RequireFlags(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		if strings.TrimSpace(flagValue(cmd, name)) == "" {
			return &ValidationError{Field: name, Msg: "is required"}
		}
	}
	return nil
}

func flagValue(cmd *cobra.Command, name string) string {
	if cmd == nil {
		return ""
	}
	value, _ := cmd.Flags().GetString(name)
	return value
}
