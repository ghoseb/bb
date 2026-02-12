package review

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type viewOptions struct {
	repo     string
	prNumber int
	file     string
	json     bool

	factory *cmdutil.Factory
	client  *bbcloud.Client
}

// NewCmdView creates the review view command
func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	opts := &viewOptions{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "view <pr-number> [file-path]",
		Short: "View PR details or specific file diff",
		Long: `View pull request with complete context for review.

Requires --repo flag to specify the repository.

Without file argument: Shows PR metadata, files, build status, and review status.
With file argument: Shows file diff and inline comments.

For actions, use dedicated commands:
  bbc review comment <pr> --repo <repo> "message"
  bbc review approve <pr> --repo <repo>
  bbc review request-change <pr> --repo <repo>

Examples:
  # View complete PR context
  bbc review view 450 --repo test_repo

  # View specific file diff with comments
  bbc review view 450 src/auth.ts --repo test_repo`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize client
			client, err := opts.factory.NewBBCloudClient("")
			if err != nil {
				return err
			}
			opts.client = client

			// Parse PR number
			prNum, err := parsePRNumber(args[0])
			if err != nil {
				return err
			}
			opts.prNumber = prNum

			// Check for file argument
			if len(args) > 1 {
				opts.file = args[1]
				return runViewFile(cmd.Context(), opts)
			}

			// Default: full PR view
			return runViewPR(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output JSON instead of markdown")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

type reviewerInfo struct {
	Username string `json:"username"`
	State    string `json:"state"` // "approved" or "changes_requested"
}

type fileInfo struct {
	Path      string `json:"path"`
	OldPath   string `json:"old_path,omitempty"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Comments  int    `json:"comments"`
}

type prViewOutput struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	State       string         `json:"state"`
	Source      string         `json:"source"`
	Target      string         `json:"target"`
	Created     string         `json:"created"`
	Updated     string         `json:"updated"`
	Reviewers   []reviewerInfo `json:"reviewers"`
	BuildStatus string         `json:"build_status"`
	Files       []fileInfo     `json:"files"`
	TotalFiles  int            `json:"total_files"`
	TotalAdds   int            `json:"total_additions"`
	TotalDels   int            `json:"total_deletions"`
	TotalComments int          `json:"total_comments"`
}

func runViewPR(ctx context.Context, opts *viewOptions) error {
	ios, _ := opts.factory.Streams()

	// Fetch PR metadata first (needed for output structure)
	pr, err := opts.client.GetPullRequest(ctx, opts.repo, opts.prNumber)
	if err != nil {
		return fmt.Errorf("get pull request: %w", err)
	}

	// Parallelize remaining API calls
	var (
		diffstat    []bbcloud.FileStats
		pipelines   []bbcloud.CommitStatus
		comments    []bbcloud.Comment
		buildStatus = "unknown"
	)

	g, gctx := errgroup.WithContext(ctx)

	// Fetch diffstat (critical - return error on failure)
	g.Go(func() error {
		var err error
		diffstat, err = opts.client.GetPRDiffStats(gctx, opts.repo, opts.prNumber)
		if err != nil {
			return fmt.Errorf("get diffstat: %w", err)
		}
		return nil
	})

	// Fetch build status (non-critical - log warning on failure, return nil)
	g.Go(func() error {
		var err error
		pipelines, err = opts.client.GetPRPipelines(gctx, opts.repo, opts.prNumber)
		if err != nil {
			_, _ = fmt.Fprintf(ios.ErrOut, "warning: failed to fetch pipeline status: %v\n", err)
		}
		return nil
	})

	// Fetch comments (non-critical - log warning on failure, return nil)
	g.Go(func() error {
		var err error
		comments, err = opts.client.ListPRComments(gctx, opts.repo, opts.prNumber)
		if err != nil {
			_, _ = fmt.Fprintf(ios.ErrOut, "warning: failed to fetch comments: %v\n", err)
		}
		return nil
	})

	// Wait for all goroutines
	if err := g.Wait(); err != nil {
		return err
	}

	// Process pipeline status
	if len(pipelines) > 0 && pipelines[0].State != "" {
		buildStatus = pipelines[0].State
	}

	// Count comments per file
	commentCounts := make(map[string]int)
	totalComments := len(comments)
	for _, comment := range comments {
		if comment.Inline != nil && comment.Inline.Path != "" {
			commentCounts[comment.Inline.Path]++
		}
	}

	// Build file list
	files := make([]fileInfo, 0, len(diffstat))
	totalAdds := 0
	totalDels := 0
	for _, stat := range diffstat {
		path := stat.GetPath()
		fi := fileInfo{
			Path:      path,
			Status:    stat.Status,
			Additions: stat.LinesAdded,
			Deletions: stat.LinesRemoved,
			Comments:  commentCounts[path],
		}
		if stat.Status == "renamed" && stat.Old != nil {
			fi.OldPath = stat.Old.Path
		}
		files = append(files, fi)
		totalAdds += stat.LinesAdded
		totalDels += stat.LinesRemoved
	}

	// Build reviewers list - only include those who have taken action (approved or requested changes)
	reviewers := make([]reviewerInfo, 0)
	for _, participant := range pr.Participants {
		if participant.Role == "REVIEWER" && participant.State != "" {
			reviewers = append(reviewers, reviewerInfo{
				Username: participant.User.DisplayName,
				State:    participant.State,
			})
		}
	}

	output := prViewOutput{
		ID:          pr.ID,
		Title:       pr.Title,
		Description: pr.Description,
		Author:      pr.Author.DisplayName,
		State:       pr.State,
		Source:      pr.Source.Branch.Name,
		Target:      pr.Destination.Branch.Name,
		Created:     pr.CreatedOn.Format("2006-01-02T15:04:05Z07:00"),
		Updated:     pr.UpdatedOn.Format("2006-01-02T15:04:05Z07:00"),
		Reviewers:   reviewers,
		BuildStatus: buildStatus,
		Files:       files,
		TotalFiles:  len(files),
		TotalAdds:   totalAdds,
		TotalDels:   totalDels,
		TotalComments: totalComments,
	}

	// Output format based on flag
	if opts.json {
		// Output JSON
		if err := cmdutil.WriteJSON(ios.Out, output); err != nil {
			return fmt.Errorf("encode output: %w", err)
		}
		return nil
	}

	// Output markdown (default)
	return renderMarkdownPRView(ios.Out, output, comments)
}

type fileViewOutput struct {
	PR        int            `json:"pr"`
	File      string         `json:"file"`
	Status    string         `json:"status"`
	Additions int            `json:"additions"`
	Deletions int            `json:"deletions"`
	Diff      string         `json:"diff"`      // Raw unified diff
	Comments  []commentInfo  `json:"comments"`
}

type commentInfo struct {
	ID        int             `json:"id"`
	Line      int             `json:"line"`
	Author    string          `json:"author"`
	AuthorID  string          `json:"author_id"`  // UUID for @mentions
	Text      string          `json:"text"`
	Created   string          `json:"created"`
	Inline    bool            `json:"inline"`
	Replies   []replyInfo     `json:"replies"`
}

type replyInfo struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	AuthorID  string `json:"author_id"`  // UUID for @mentions
	Text      string `json:"text"`
	Created   string `json:"created"`
}

// extractFileDiff extracts the diff section for a renamed file from the full PR diff.
// It looks for the "rename from/rename to" pattern and returns the hunks.
func extractFileDiff(fullDiff, oldPath, newPath string) string {
	lines := strings.Split(fullDiff, "\n")
	var capturing bool
	var result []string

	for i, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if capturing {
				break // hit next file, stop
			}
			// Check if this is our renamed file
			if strings.Contains(line, "a/"+oldPath) && strings.Contains(line, "b/"+newPath) {
				capturing = true
				continue
			}
		}
		if capturing {
			// Skip the rename metadata lines, capture from first @@ hunk
			if strings.HasPrefix(line, "@@") || (len(result) > 0) {
				result = append(result, line)
			}
			// Also stop if we somehow reach the end
			if i == len(lines)-1 {
				break
			}
		}
	}

	return strings.TrimRight(strings.Join(result, "\n"), "\n")
}

func runViewFile(ctx context.Context, opts *viewOptions) error {
	// Fetch diff for this file
	diff, err := opts.client.GetPRFileDiff(ctx, opts.repo, opts.prNumber, opts.file)
	if err != nil {
		return fmt.Errorf("get file diff: %w", err)
	}

	// Fetch diffstat for file stats
	diffstat, err := opts.client.GetPRDiffStats(ctx, opts.repo, opts.prNumber)
	if err != nil {
		return fmt.Errorf("get diffstat: %w", err)
	}

	// Find stats for this file
	var fileStatus string
	var oldPath string
	var additions, deletions int
	for _, stat := range diffstat {
		if stat.GetPath() == opts.file {
			fileStatus = stat.Status
			additions = stat.LinesAdded
			deletions = stat.LinesRemoved
			if stat.Old != nil {
				oldPath = stat.Old.Path
			}
			break
		}
	}

	// For renames, BB's per-file diff endpoint shows the entire file as "new file"
	// additions. Use diffstat as source of truth for whether there are real changes.
	if fileStatus == "renamed" && oldPath != "" {
		header := fmt.Sprintf("renamed: %s → %s\n", oldPath, opts.file)
		if additions == 0 && deletions == 0 {
			diff = header
		} else {
			// Real changes alongside rename — extract from full PR diff which has proper hunks
			fullDiff, err := opts.client.GetPRDiff(ctx, opts.repo, opts.prNumber)
			if err == nil {
				if section := extractFileDiff(fullDiff, oldPath, opts.file); section != "" {
					diff = header + "\n" + section
				} else {
					diff = fmt.Sprintf("%s(+%d/-%d lines changed)\n", header, additions, deletions)
				}
			} else {
				diff = fmt.Sprintf("%s(+%d/-%d lines changed)\n", header, additions, deletions)
			}
		}
	}

	// Fetch comments for this file
	allComments, err := opts.client.ListPRComments(ctx, opts.repo, opts.prNumber)
	if err != nil {
		return fmt.Errorf("get comments: %w", err)
	}

	// Filter comments for this file
	comments := make([]commentInfo, 0)
	for _, comment := range allComments {
		if comment.Inline != nil && comment.Inline.Path == opts.file {
			replies := make([]replyInfo, 0)
			// Note: Bitbucket API doesn't support nested replies in the same call
			// Would need separate API call for each comment's replies

			line := 0
			if comment.Inline.To != nil {
				line = *comment.Inline.To
			}

			comments = append(comments, commentInfo{
				ID:       comment.ID,
				Line:     line,
				Author:   comment.User.DisplayName,
				AuthorID: comment.User.UUID,
				Text:     comment.Content.Raw,
				Created:  comment.CreatedOn.Format("2006-01-02T15:04:05Z07:00"),
				Inline:   true,
				Replies:  replies,
			})
		}
	}

	output := fileViewOutput{
		PR:        opts.prNumber,
		File:      opts.file,
		Status:    fileStatus,
		Additions: additions,
		Deletions: deletions,
		Diff:      diff,
		Comments:  comments,
	}

	// Output format based on flag
	ios, _ := opts.factory.Streams()
	if opts.json {
		// Output JSON
		if err := cmdutil.WriteJSON(ios.Out, output); err != nil {
			return fmt.Errorf("encode output: %w", err)
		}
		return nil
	}

	// Output markdown (default)
	return renderMarkdownFileView(ios.Out, output)
}

func renderMarkdownPRView(w io.Writer, output prViewOutput, comments []bbcloud.Comment) error {
	_, _ = fmt.Fprintf(w, "# PR %d: %s\n", output.ID, output.Title)
	_, _ = fmt.Fprintf(w, "Author: %s | State: %s | Build: %s\n", output.Author, output.State, output.BuildStatus)
	_, _ = fmt.Fprintf(w, "Source: %s → %s\n", output.Source, output.Target)
	
	if len(output.Reviewers) > 0 {
		_, _ = fmt.Fprintf(w, "Reviewers: ")
		for i, r := range output.Reviewers {
			if i > 0 {
				_, _ = fmt.Fprintf(w, ", ")
			}
			_, _ = fmt.Fprintf(w, "%s (%s)", r.Username, r.State)
		}
		_, _ = fmt.Fprintf(w, "\n")
	}
	
	_, _ = fmt.Fprintf(w, "\n## Files (%d files, +%d, -%d)\n", output.TotalFiles, output.TotalAdds, output.TotalDels)
	for _, f := range output.Files {
		commentStr := ""
		if f.Comments > 0 {
			commentStr = fmt.Sprintf(", %d comments", f.Comments)
		}
		switch {
		case f.Status == "renamed" && f.OldPath != "" && f.Additions == 0 && f.Deletions == 0:
			_, _ = fmt.Fprintf(w, "- %s ← %s (renamed%s)\n", f.Path, f.OldPath, commentStr)
		case f.Status == "renamed" && f.OldPath != "":
			_, _ = fmt.Fprintf(w, "- %s ← %s (renamed, +%d/-%d%s)\n", f.Path, f.OldPath, f.Additions, f.Deletions, commentStr)
		default:
			_, _ = fmt.Fprintf(w, "- %s (+%d/-%d%s)\n", f.Path, f.Additions, f.Deletions, commentStr)
		}
	}
	
	if output.TotalComments > 0 {
		_, _ = fmt.Fprintf(w, "\n## Comments (%d)\n", output.TotalComments)
		for _, comment := range comments {
			if comment.Inline != nil {
				line := 0
				if comment.Inline.To != nil {
					line = *comment.Inline.To
				}
				_, _ = fmt.Fprintf(w, "**%s** (id:%s) on %s:%d (comment:%d): %s\n",
					comment.User.DisplayName,
					comment.User.UUID,
					comment.Inline.Path,
					line,
					comment.ID,
					comment.Content.Raw)
			} else {
				_, _ = fmt.Fprintf(w, "**%s** (id:%s, general) (comment:%d): %s\n",
					comment.User.DisplayName,
					comment.User.UUID,
					comment.ID,
					comment.Content.Raw)
			}
			
			// Render replies
			if comment.Parent == nil {
				for _, reply := range comments {
					if reply.Parent != nil && reply.Parent.ID == comment.ID {
						_, _ = fmt.Fprintf(w, "  > **%s** (id:%s, reply to comment:%d): %s\n",
							reply.User.DisplayName,
							reply.User.UUID,
							comment.ID,
							reply.Content.Raw)
					}
				}
			}
		}
	}
	
	return nil
}

func renderMarkdownFileView(w io.Writer, output fileViewOutput) error {
	_, _ = fmt.Fprintf(w, "# PR %d — %s\n", output.PR, output.File)
	_, _ = fmt.Fprintf(w, "Status: %s | +%d -%d\n\n", output.Status, output.Additions, output.Deletions)
	
	_, _ = fmt.Fprintf(w, "```diff\n%s```\n", output.Diff)
	
	if len(output.Comments) > 0 {
		_, _ = fmt.Fprintf(w, "\n## Comments (%d)\n", len(output.Comments))
		for _, comment := range output.Comments {
			lineStr := ""
			if comment.Line > 0 {
				lineStr = fmt.Sprintf(", line %d", comment.Line)
			}
			_, _ = fmt.Fprintf(w, "**%s** (id:%s%s) (comment:%d): %s\n",
				comment.Author,
				comment.AuthorID,
				lineStr,
				comment.ID,
				comment.Text)
			
			// Render replies
			for _, reply := range comment.Replies {
				_, _ = fmt.Fprintf(w, "  > **%s** (id:%s, reply to comment:%d): %s\n",
					reply.Author,
					reply.AuthorID,
					comment.ID,
					reply.Text)
			}
		}
	}
	
	return nil
}
