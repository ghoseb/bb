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

type replyOptions struct {
	repo      string
	prNumber  int
	commentID int
	message   string

	factory *cmdutil.Factory
}

// NewCmdReply creates the review reply command
func NewCmdReply(f *cmdutil.Factory) *cobra.Command {
	opts := &replyOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "reply <pr-number> <comment-id> <message>",
		Short: "Reply to a comment on a pull request",
		Long: `Reply to an existing comment on a pull request.

Requires --repo flag to specify the repository.

The comment ID can be found in the output of bb review view commands.

Examples:
  bbc review reply 450 123456 --repo test_repo "Fixed in latest commit"
  bbc review reply 450 789012 --repo test_repo "Good catch, updated"`,
		Args: cobra.ExactArgs(3),
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

			// Parse comment ID
			commentID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid comment ID: %s", args[1])
			}
			if commentID <= 0 {
				return fmt.Errorf("comment ID must be positive")
			}
			opts.commentID = commentID

			// Get message
			opts.message = args[2]
			if strings.TrimSpace(opts.message) == "" {
				return fmt.Errorf("message cannot be empty")
			}

			return runReply(cmd.Context(), opts, client)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

func runReply(ctx context.Context, opts *replyOptions, client *bbcloud.Client) error {
	reply, err := client.ReplyToComment(ctx, opts.repo, opts.prNumber,
		opts.commentID, opts.message)
	if err != nil {
		return fmt.Errorf("reply to comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":        opts.prNumber,
		"repo":      opts.repo,
		"reply_id":  reply.ID,
		"parent_id": opts.commentID,
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
