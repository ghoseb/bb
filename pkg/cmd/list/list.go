package list

import (
	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/cmdutil"
)

// NewCmdList creates the list command
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <command>",
		Short: "List repositories",
		Long: `List repositories in your Bitbucket workspace.

For pull requests, use:
  bb review list --repo <repo>`,
	}

	cmd.AddCommand(NewCmdRepos(f))

	return cmd
}
