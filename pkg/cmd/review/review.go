package review

import (
	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/pkg/cmdutil"
)

// NewCmdReview creates the review command group
func NewCmdReview(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <command>",
		Short: "Agent-optimized PR review commands",
		Long: `Review pull requests with token-efficient output optimized for AI agents.

Commands aggregate multiple API calls and provide flat JSON structures
designed for efficient LLM consumption during code review workflows.`,
	}

	// Add subcommands
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdView(f))
	cmd.AddCommand(NewCmdComment(f))
	cmd.AddCommand(NewCmdReply(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdApprove(f))
	cmd.AddCommand(NewCmdRequestChange(f))

	return cmd
}
