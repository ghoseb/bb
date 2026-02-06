package root

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ghoseb/bb/internal/build"
	"github.com/ghoseb/bb/pkg/cmd/auth"
	"github.com/ghoseb/bb/pkg/cmd/list"
	"github.com/ghoseb/bb/pkg/cmd/review"
	"github.com/ghoseb/bb/pkg/cmdutil"
)

// NewCmdRoot creates the root command for the bb CLI.
func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bb <command> <subcommand> [flags]",
		Short: "Bitbucket Cloud CLI for PR workflows",
		Long: `bb is a command-line interface for Bitbucket Cloud.

It provides tools for managing pull requests, comments, and pipelines
directly from your terminal, making code review workflows faster and
more efficient.`,
		Version:       fmt.Sprintf("%s (%s, %s)", build.Version, build.Commit, build.Date),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help when no subcommand is provided
			return cmd.Help()
		},
	}

	// Set output streams for help and error messages
	ios, _ := f.Streams()
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetIn(ios.In)

	// Global flags
	cmd.PersistentFlags().StringP("workspace", "w", "", 
		"Override workspace (env: BB_WORKSPACE, or from stored credentials)")

	// Add command groups
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(review.NewCmdReview(f))
	cmd.AddCommand(list.NewCmdList(f))

	// Custom help that shows subcommand usage inline
	cmd.SetHelpFunc(expandedHelp)

	return cmd
}

// skipCommands are utility commands excluded from expanded help.
var skipCommands = map[string]bool{"completion": true, "help": true}

func expandedHelp(cmd *cobra.Command, _ []string) {
	var b strings.Builder

	if cmd.Long != "" {
		b.WriteString(cmd.Long)
	} else {
		b.WriteString(cmd.Short)
	}
	b.WriteString("\n\nUsage:\n")

	for _, child := range cmd.Commands() {
		if child.Hidden || !child.IsAvailableCommand() || skipCommands[child.Name()] {
			continue
		}
		subs := child.Commands()
		if len(subs) == 0 {
			fmt.Fprintf(&b, "  bb %-50s  %s\n", child.Use, child.Short)
		} else {
			// If the parent command itself is runnable, show it too
			if child.RunE != nil || child.Run != nil {
				fmt.Fprintf(&b, "  bb %-50s  %s\n", child.Name(), child.Short)
			}
			for _, sub := range subs {
				if sub.Hidden || !sub.IsAvailableCommand() {
					continue
				}
				flags := strings.TrimRight(sub.NonInheritedFlags().FlagUsages(), "\n")
				if flags != "" {
					fmt.Fprintf(&b, "  bb %s %-40s  %s\n", child.Name(), sub.Use, sub.Short)
					for _, line := range strings.Split(flags, "\n") {
						fmt.Fprintf(&b, "      %s\n", strings.TrimSpace(line))
					}
				} else {
					fmt.Fprintf(&b, "  bb %s %-40s  %s\n", child.Name(), sub.Use, sub.Short)
				}
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("Global Flags:\n")
	b.WriteString(cmd.PersistentFlags().FlagUsages())
	fmt.Fprintf(&b, "\nVersion: %s\n", cmd.Version)

	_, _ = fmt.Fprint(cmd.OutOrStdout(), b.String())
}
