package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghoseb/bb/internal/build"
	"github.com/ghoseb/bb/pkg/cmd/root"
	"github.com/ghoseb/bb/pkg/cmdutil"
	"github.com/ghoseb/bb/pkg/iostreams"
)

// Main initialises CLI dependencies and executes the root command.
func Main() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ios := iostreams.System()
	f := cmdutil.NewFactory(build.Version, ios)

	rootCmd := root.NewCmdRoot(f)
	rootCmd.SetContext(ctx)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		var exitErr *cmdutil.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Msg != "" {
				_, _ = fmt.Fprintln(ios.ErrOut, exitErr.Msg)
			}
			return exitErr.Code
		}
		_, _ = fmt.Fprintf(ios.ErrOut, "Error: %v\n", err)
		return 1
	}

	return 0
}
