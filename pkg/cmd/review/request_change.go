package review

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type requestChangeOptions struct {
	repo     string
	prNumber int
	undo     bool

	factory *cmdutil.Factory
}

// NewCmdRequestChange creates the review request-change command
func NewCmdRequestChange(f *cmdutil.Factory) *cobra.Command {
	opts := &requestChangeOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "request-change <pr-number>",
		Short: "Request changes on a pull request",
		Long: `Request changes on a pull request.

Requires --repo flag to specify the repository.

To explain what needs to change, use bb review comment separately.

Use --undo to remove your request-change status.

Examples:
  # Request changes
  bb review request-change 450 --repo test_repo

  # Remove request-change
  bb review request-change 450 --repo test_repo --undo

  # Request changes with explanation (two commands)
  bb review request-change 450 --repo test_repo
  bb review comment 450 --repo test_repo "Please add tests for the new feature"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize client
			client, err := opts.factory.NewBBCloudClient("")
			if err != nil {
				return err
			}

			// Parse PR number
			prNum, err := parsePRNumber(args[0])
			if err != nil {
				return err
			}
			opts.prNumber = prNum

			return runRequestChange(cmd.Context(), opts, client)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().BoolVar(&opts.undo, "undo", false, "Remove request-change instead of requesting changes")

	return cmd
}

func runRequestChange(ctx context.Context, opts *requestChangeOptions, client *bbcloud.Client) error {
	if opts.undo {
		// Remove request-change
		err := client.UnrequestChangesPR(ctx, opts.repo, opts.prNumber)
		if err != nil {
			output := map[string]interface{}{
				"pr":     opts.prNumber,
				"repo":   opts.repo,
				"action": "unrequest-change",
				"error":  friendlyError(err.Error()),
			}
			
			return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
		}

		output := map[string]interface{}{
			"pr":     opts.prNumber,
			"repo":   opts.repo,
			"action": "unrequested_change",
		}

		return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
	}

	// Request changes on PR
	participant, err := client.RequestChangesPR(ctx, opts.repo, opts.prNumber)
	if err != nil {
		output := map[string]interface{}{
			"pr":     opts.prNumber,
			"repo":   opts.repo,
			"action": "request-change",
			"error":  friendlyError(err.Error()),
		}
		
		return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
	}

	output := map[string]interface{}{
		"pr":       opts.prNumber,
		"repo":     opts.repo,
		"action":   "changes_requested",
		"approved": participant.Approved,
		"state":    participant.State,
	}

	if participant.User != nil {
		output["user"] = participant.User.GetName()
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
