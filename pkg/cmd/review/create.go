package review

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type createOptions struct {
	repo              string
	sourceBranch      string
	targetBranch      string
	title             string
	closeSourceBranch bool
	draft             bool

	factory *cmdutil.Factory
}

// NewCmdCreate creates the review create command
func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &createOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "create <source-branch> <title>",
		Short: "Create a new pull request",
		Long: `Create a new pull request.

Requires --repo flag to specify the repository.
If --target is not specified, the repository's main branch is used.

Examples:
  # Create PR to main branch
  bb review create feat/auth --repo test_repo "Add JWT authentication"

  # Create PR to specific branch
  bb review create feat/auth --target develop --repo test_repo "Add JWT authentication"

  # Create draft PR, close source branch after merge
  bb review create feat/auth --draft --close-source --repo test_repo "Add JWT authentication"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.factory.NewBBCloudClient("")
			if err != nil {
				return err
			}

			opts.sourceBranch = args[0]
			opts.title = args[1]

			if strings.TrimSpace(opts.title) == "" {
				return fmt.Errorf("title cannot be empty")
			}

			return runCreate(cmd.Context(), opts, client)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	cmd.Flags().StringVarP(&opts.targetBranch, "target", "t", "", "Target branch (default: repo main branch)")
	cmd.Flags().BoolVar(&opts.closeSourceBranch, "close-source", false, "Close source branch after merge")
	cmd.Flags().BoolVar(&opts.draft, "draft", false, "Create as draft pull request")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

func runCreate(ctx context.Context, opts *createOptions, client *bbcloud.Client) error {
	pr, err := client.CreatePR(ctx, opts.repo, bbcloud.CreatePROptions{
		Title:             opts.title,
		SourceBranch:      opts.sourceBranch,
		DestinationBranch: opts.targetBranch,
		CloseSourceBranch: opts.closeSourceBranch,
		Draft:             opts.draft,
	})
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}

	// Extract branch names from response
	source := opts.sourceBranch
	target := ""
	if pr.Destination != nil && pr.Destination.Branch != nil {
		target = pr.Destination.Branch.Name
	}

	// Extract URL
	url := ""
	if pr.Links.HTML != nil {
		url = pr.Links.HTML.Href
	}

	output := map[string]interface{}{
		"pr":     pr.ID,
		"repo":   opts.repo,
		"url":    url,
		"source": source,
		"target": target,
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
