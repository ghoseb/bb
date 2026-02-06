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

type commentOptions struct {
	repo      string
	prNumber  int
	file      string
	lineStart int
	lineEnd   int // 0 means single line
	message   string
	edit      int // comment ID to edit
	delete    int // comment ID to delete
	resolve   int // comment ID to resolve
	reopen    int // comment ID to reopen

	factory *cmdutil.Factory
}

// NewCmdComment creates the review comment command
func NewCmdComment(f *cmdutil.Factory) *cobra.Command {
	opts := &commentOptions{factory: f}

	cmd := &cobra.Command{
		Use:   "comment <pr-number> [file-path line-start [line-end]] <message>",
		Short: "Manage comments on pull requests",
		Long: `Add, edit, delete, resolve, or reopen comments on pull requests.

Requires --repo flag to specify the repository.

General comment:
  bb review comment <pr> --repo <repo> "message"

Inline comment (single line):
  bb review comment <pr> <file> <line> --repo <repo> "message"

Inline comment (line range):
  bb review comment <pr> <file> <start> <end> --repo <repo> "message"

Edit comment:
  bb review comment <pr> --repo <repo> --edit <comment-id> "updated message"

Delete comment:
  bb review comment <pr> --repo <repo> --delete <comment-id>

Resolve comment:
  bb review comment <pr> --repo <repo> --resolve <comment-id>

Reopen comment:
  bb review comment <pr> --repo <repo> --reopen <comment-id>

Examples:
  # General comment
  bb review comment 450 --repo test_repo "Looks good overall"

  # Inline comment on single line
  bb review comment 450 src/auth.ts 23 --repo test_repo "Fix this typo"

  # Inline comment on line range
  bb review comment 450 src/auth.ts 23 27 --repo test_repo "Refactor this block"

  # Edit existing comment
  bb review comment 450 --repo test_repo --edit 753222173 "Updated text"

  # Delete comment
  bb review comment 450 --repo test_repo --delete 753222173

  # Resolve comment
  bb review comment 450 --repo test_repo --resolve 753222173

  # Reopen comment
  bb review comment 450 --repo test_repo --reopen 753222173`,
		Args: cobra.MinimumNArgs(1),
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

			// Handle --edit flag
			if opts.edit > 0 {
				if len(args) < 2 {
					return fmt.Errorf("message is required when editing a comment")
				}
				opts.message = args[1]
				if strings.TrimSpace(opts.message) == "" {
					return fmt.Errorf("message cannot be empty")
				}
				return runUpdateComment(cmd.Context(), opts, client)
			}

			// Handle --delete flag
			if opts.delete > 0 {
				return runDeleteComment(cmd.Context(), opts, client)
			}

			// Handle --resolve flag
			if opts.resolve > 0 {
				return runResolveComment(cmd.Context(), opts, client)
			}

			// Handle --reopen flag
			if opts.reopen > 0 {
				return runReopenComment(cmd.Context(), opts, client)
			}

			// Validate minimum args for create operations
			if len(args) < 2 {
				return fmt.Errorf("message is required")
			}

			// Determine comment type based on args count
			switch len(args) {
			case 2:
				// General comment: pr + message
				opts.message = args[1]
				if strings.TrimSpace(opts.message) == "" {
					return fmt.Errorf("message cannot be empty")
				}
				return runGeneralComment(cmd.Context(), opts, client)

			case 4:
				// Inline comment: pr + file + line + message
				opts.file = args[1]
				line, err := strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("invalid line number: %s", args[2])
				}
				if line <= 0 {
					return fmt.Errorf("line number must be positive, got %d", line)
				}
				opts.lineStart = line
				opts.lineEnd = 0 // Single line
				opts.message = args[3]
				if strings.TrimSpace(opts.message) == "" {
					return fmt.Errorf("message cannot be empty")
				}
				return runInlineComment(cmd.Context(), opts, client)

			case 5:
				// Line range comment: pr + file + start + end + message
				opts.file = args[1]
				start, err := strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("invalid line start: %s", args[2])
				}
				end, err := strconv.Atoi(args[3])
				if err != nil {
					return fmt.Errorf("invalid line end: %s", args[3])
				}
				if start <= 0 || end <= 0 {
					return fmt.Errorf("line numbers must be positive")
				}

				// Simple logic: if end <= start, treat as single line
				if end <= start {
					opts.lineStart = start
					opts.lineEnd = 0
				} else {
					opts.lineStart = start
					opts.lineEnd = end
				}

				opts.message = args[4]
				if strings.TrimSpace(opts.message) == "" {
					return fmt.Errorf("message cannot be empty")
				}
				return runInlineComment(cmd.Context(), opts, client)

			default:
				return fmt.Errorf("invalid number of arguments (expected 2, 4, or 5)")
			}
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().IntVar(&opts.edit, "edit", 0, "Edit existing comment by ID")
	cmd.Flags().IntVar(&opts.delete, "delete", 0, "Delete existing comment by ID")
	cmd.Flags().IntVar(&opts.resolve, "resolve", 0, "Resolve comment by ID")
	cmd.Flags().IntVar(&opts.reopen, "reopen", 0, "Reopen comment by ID")

	return cmd
}

func runGeneralComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	comment, err := client.CreateComment(ctx, opts.repo, opts.prNumber, opts.message)
	if err != nil {
		return fmt.Errorf("create comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": comment.ID,
		"type":       "general",
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}

func runInlineComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	// Determine line range for API call
	var lineStart, lineEnd int
	if opts.lineEnd == 0 {
		// Single-line comment
		lineStart = 0
		lineEnd = opts.lineStart
	} else {
		// Range comment
		lineStart = opts.lineStart
		lineEnd = opts.lineEnd
	}
	
	comment, err := client.CreateInlineComment(ctx, opts.repo, opts.prNumber,
		opts.message, opts.file, lineStart, lineEnd)
	if err != nil {
		return fmt.Errorf("create inline comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": comment.ID,
		"type":       "inline",
		"file":       opts.file,
		"line_start": opts.lineStart,
	}

	if opts.lineEnd > 0 {
		output["line_end"] = opts.lineEnd
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}

func runUpdateComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	comment, err := client.UpdateComment(ctx, opts.repo, opts.prNumber, opts.edit, opts.message)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": comment.ID,
		"action":     "updated",
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}

func runDeleteComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	err := client.DeleteComment(ctx, opts.repo, opts.prNumber, opts.delete)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": opts.delete,
		"action":     "deleted",
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}

func runResolveComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	err := client.ResolveComment(ctx, opts.repo, opts.prNumber, opts.resolve)
	if err != nil {
		return fmt.Errorf("resolve comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": opts.resolve,
		"action":     "resolved",
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}

func runReopenComment(ctx context.Context, opts *commentOptions, client *bbcloud.Client) error {
	err := client.ReopenComment(ctx, opts.repo, opts.prNumber, opts.reopen)
	if err != nil {
		return fmt.Errorf("reopen comment: %w", err)
	}

	output := map[string]interface{}{
		"pr":         opts.prNumber,
		"repo":       opts.repo,
		"comment_id": opts.reopen,
		"action":     "reopened",
	}

	return cmdutil.WriteJSON(opts.factory.IOStreams.Out, output)
}
