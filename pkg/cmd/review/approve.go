package review

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type approveOptions struct {
	repo     string
	prNumber int
	undo     bool

	factory *cmdutil.Factory
}

// NewCmdApprove creates the review approve command
func NewCmdApprove(f *cmdutil.Factory) *cobra.Command {
	opts := &approveOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "approve <pr-number>",
		Short: "Approve a pull request",
		Long: `Approve a pull request.

Requires --repo flag to specify the repository.

To add a comment with your approval, use bb review comment separately.

Use --undo to remove your approval.

Examples:
  # Approve PR
  bbc review approve 450 --repo test_repo

  # Remove approval
  bbc review approve 450 --repo test_repo --undo

  # Approve and comment (two commands)
  bbc review approve 450 --repo test_repo
  bbc review comment 450 --repo test_repo "LGTM! Ship it."`,
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

			return runApprove(cmd.Context(), opts, client)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().BoolVar(&opts.undo, "undo", false, "Remove approval instead of approving")

	return cmd
}

func friendlyError(errMsg string) string {
	switch {
	case strings.Contains(errMsg, "already been merged"):
		return "PR is already merged"
	case strings.Contains(errMsg, "already been declined"):
		return "PR is already declined"
	case strings.Contains(errMsg, "haven't approved"):
		return "no approval to remove"
	case strings.Contains(errMsg, "Request changes"):
		return "no request-change to remove"
	case strings.Contains(errMsg, "not found"):
		return "PR not found"
	default:
		return errMsg
	}
}

func runApprove(ctx context.Context, opts *approveOptions, client *bbcloud.Client) error {
	if opts.undo {
		// Remove approval
		err := client.UnapprovePR(ctx, opts.repo, opts.prNumber)
		if err != nil {
			output := map[string]interface{}{
				"pr":     opts.prNumber,
				"repo":   opts.repo,
				"action": "unapprove",
				"error":  friendlyError(err.Error()),
			}
			
			return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
		}

		output := map[string]interface{}{
			"pr":     opts.prNumber,
			"repo":   opts.repo,
			"action": "unapproved",
		}

		return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
	}

	// Approve PR
	participant, err := client.ApprovePR(ctx, opts.repo, opts.prNumber)
	if err != nil {
		output := map[string]interface{}{
			"pr":     opts.prNumber,
			"repo":   opts.repo,
			"action": "approve",
			"error":  friendlyError(err.Error()),
		}
		
		return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
	}

	output := map[string]interface{}{
		"pr":       opts.prNumber,
		"repo":     opts.repo,
		"action":   "approved",
		"approved": participant.Approved,
		"state":    participant.State,
	}

	if participant.User != nil {
		output["user"] = participant.User.GetName()
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
