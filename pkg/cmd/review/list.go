package review

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/ghoseb/bb/pkg/bbcloud"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

type listOptions struct {
	repo  string
	state string
	limit int
	json  bool

	factory *cmdutil.Factory
	client  *bbcloud.Client
}

// NewCmdList creates the review list command
func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests with review stats",
		Long: `List pull requests with token-efficient output for agent review.

Requires --repo flag to specify the repository.

Includes file counts, line changes, and reviewer approval status.

Examples:
  # List open PRs in a repository
  bb review list --repo test_repo

  # List merged PRs
  bb review list --repo test_repo --state MERGED

  # List more PRs
  bb review list --repo test_repo --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := opts.factory.NewBBCloudClient("")
			if err != nil {
				return err
			}
			opts.client = client
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repo, "repo", "r", "", "Repository slug (required)")
	cmd.Flags().StringVar(&opts.state, "state", "OPEN", "PR state (OPEN, MERGED, DECLINED)")
	cmd.Flags().IntVar(&opts.limit, "limit", 20, "Maximum number of PRs to list")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Output JSON instead of markdown")
	_ = cmd.MarkFlagRequired("repo")

	return cmd
}

type prListItem struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	State     string `json:"state"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Created   string `json:"created"`
	Updated   string `json:"updated"`
	Files     int    `json:"files"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Approved  int    `json:"approved"`
	Declined  int    `json:"declined"`
}

type listOutput struct {
	PRs []prListItem `json:"prs"`
}

func runList(ctx context.Context, opts *listOptions) error {
	// Fetch PRs from Bitbucket
	prs, err := opts.client.ListPullRequests(ctx, opts.repo, opts.state, opts.limit)
	if err != nil {
		return fmt.Errorf("list pull requests: %w", err)
	}

	// Transform to agent-optimized format
	items := make([]prListItem, len(prs))
	
	for i, pr := range prs {
		// Count approvals and declines
		approved := 0
		declined := 0
		for _, participant := range pr.Participants {
			if participant.Approved {
				approved++
			}
			if participant.State == "changes_requested" {
				declined++
			}
		}

		items[i] = prListItem{
			ID:        pr.ID,
			Title:     pr.Title,
			Author:    pr.Author.DisplayName,
			State:     pr.State,
			Source:    pr.Source.Branch.Name,
			Target:    pr.Destination.Branch.Name,
			Created:   pr.CreatedOn.Format("2006-01-02T15:04:05Z07:00"),
			Updated:   pr.UpdatedOn.Format("2006-01-02T15:04:05Z07:00"),
			Files:     0, // Will be populated below
			Additions: 0, // Will be populated below
			Deletions: 0, // Will be populated below
			Approved:  approved,
			Declined:  declined,
		}
	}

	// Fetch diffstats concurrently with rate limiting (max 5 concurrent)
	sem := make(chan struct{}, 5)
	g, gctx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	ios, _ := opts.factory.Streams()

	for i := range items {
		i := i // capture loop variable
		sem <- struct{}{} // acquire semaphore
		g.Go(func() error {
			defer func() { <-sem }() // release semaphore

			diffstats, err := opts.client.GetPRDiffStats(gctx, opts.repo, items[i].ID)
			if err != nil {
				// Non-critical: log warning and continue
				_, _ = fmt.Fprintf(ios.ErrOut, "warning: failed to fetch stats for PR %d: %v\n", items[i].ID, err)
				return nil
			}

			// Calculate totals
			totalFiles := len(diffstats)
			totalAdds := 0
			totalDels := 0
			for _, stat := range diffstats {
				totalAdds += stat.LinesAdded
				totalDels += stat.LinesRemoved
			}

			// Update item (thread-safe)
			mu.Lock()
			items[i].Files = totalFiles
			items[i].Additions = totalAdds
			items[i].Deletions = totalDels
			mu.Unlock()

			return nil
		})
	}

	// Wait for all diffstat fetches
	if err := g.Wait(); err != nil {
		return err
	}

	// Output format based on flag
	if opts.json {
		output := listOutput{
			PRs: items,
		}

		// Output JSON
		if err := cmdutil.WriteJSON(ios.Out, output); err != nil {
			return fmt.Errorf("encode output: %w", err)
		}
		return nil
	}

	// Output markdown (default)
	return renderMarkdownList(ios.Out, opts.repo, items)
}

func renderMarkdownList(w io.Writer, repo string, items []prListItem) error {
	if len(items) == 0 {
		_, _ = fmt.Fprintf(w, "# No PRs found — %s\n", repo)
		return nil
	}

	state := items[0].State
	_, _ = fmt.Fprintf(w, "# %s PRs — %s\n\n", state, repo)
	_, _ = fmt.Fprintf(w, "| PR | Title | Author | Build | Files | +/- |\n")
	_, _ = fmt.Fprintf(w, "|----|-------|--------|-------|-------|-----|\n")

	for _, item := range items {
		buildStatus := "—"
		// We don't have build status in the current data structure
		// This will be added when available
		
		_, _ = fmt.Fprintf(w, "| %d | %s | %s | %s | %d | +%d/-%d |\n",
			item.ID,
			item.Title,
			item.Author,
			buildStatus,
			item.Files,
			item.Additions,
			item.Deletions,
		)
	}

	return nil
}
