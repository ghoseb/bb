package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/cmdutil"
)

type reposOptions struct {
	workspace string

	factory *cmdutil.Factory
}

// NewCmdRepos creates the list repos command
func NewCmdRepos(f *cmdutil.Factory) *cobra.Command {
	opts := &reposOptions{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List repositories in a workspace",
		Long: `List all repositories in a Bitbucket workspace.

Example:
  bb list repos
  bb list repos --workspace other-workspace`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListRepos(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", 
		"Workspace to list repos from (uses authenticated workspace if not specified)")

	return cmd
}

type repoInfo struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
	Language    string `json:"language,omitempty"`
}

func runListRepos(ctx context.Context, opts *reposOptions) error {
	// Create client with specified workspace (or default)
	client, err := opts.factory.NewBBCloudClient(opts.workspace)
	if err != nil {
		return err
	}

	repos, err := client.ListRepositories(ctx, 0)
	if err != nil {
		return fmt.Errorf("list repositories: %w", err)
	}

	// Convert to output format
	output := make([]repoInfo, len(repos))
	for i, repo := range repos {
		output[i] = repoInfo{
			Name:        repo.Name,
			Slug:        repo.Slug,
			Description: repo.Description,
			IsPrivate:   repo.IsPrivate,
			Language:    repo.Language,
		}
	}

	if err := cmdutil.WriteJSON(opts.factory.IOStreams.Out, output); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}
