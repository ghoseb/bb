package review

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type updateOptions struct {
	repo        string
	prID        int
	title       string
	description string

	factory *cmdutil.Factory
}

// NewCmdUpdate creates the review update command
func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	opts := &updateOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "update <pr-id>",
		Short: "Update a pull request",
		Long: `Update an existing pull request's title or description.

Requires --repo flag to specify the repository.

Examples:
  # Update PR title
  bbc review update 123 --repo test_repo --title "New title"

  # Update PR description
  bbc review update 123 --repo test_repo --description "New description"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.factory.NewBBCloudClient("")
			if err != nil {
				return err
			}

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid PR ID: %w", err)
			}
			opts.prID = id

			if strings.TrimSpace(opts.title) == "" && strings.TrimSpace(opts.description) == "" {
				return fmt.Errorf("at least one of --title or --description must be provided")
			}

			return runUpdate(cmd.Context(), opts, client)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Pull request title")
	cmd.Flags().StringVarP(&opts.description, "description", "d", "", "Pull request description")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

func runUpdate(ctx context.Context, opts *updateOptions, client *bbcloud.Client) error {
	pr, err := client.UpdatePR(ctx, opts.repo, opts.prID, bbcloud.UpdatePROptions{
		Title:       opts.title,
		Description: opts.description,
	})
	if err != nil {
		return fmt.Errorf("update PR: %w", err)
	}

	output := map[string]interface{}{
		"pr":          pr.ID,
		"repo":        opts.repo,
		"title":       pr.Title,
		"description": pr.Description,
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
