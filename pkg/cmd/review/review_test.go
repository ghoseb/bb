package review

import (
	"testing"

	"github.com/ghoseb/bb/pkg/cmdutil"
	"github.com/ghoseb/bb/pkg/iostreams"
)

func TestCommandStructure(t *testing.T) {
	// Create a real factory for command structure testing
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdReview(factory)
	
	if cmd.Use != "review <command>" {
		t.Errorf("expected Use to be 'review <command>', got %q", cmd.Use)
	}
	
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	
	// Check subcommands are registered
	subcommands := cmd.Commands()
	if len(subcommands) != 7 {
		t.Errorf("expected 7 subcommands, got %d", len(subcommands))
	}
	
	// Verify subcommand names
	names := make(map[string]bool)
	for _, sub := range subcommands {
		names[sub.Name()] = true
	}
	
	if !names["list"] {
		t.Error("expected 'list' subcommand")
	}
	if !names["view"] {
		t.Error("expected 'view' subcommand")
	}
	if !names["comment"] {
		t.Error("expected 'comment' subcommand")
	}
	if !names["reply"] {
		t.Error("expected 'reply' subcommand")
	}
	if !names["create"] {
		t.Error("expected 'create' subcommand")
	}
	if !names["approve"] {
		t.Error("expected 'approve' subcommand")
	}
	if !names["request-change"] {
		t.Error("expected 'request-change' subcommand")
	}
}

func TestListCommand(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdList(factory)
	
	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}
	
	// Check required flags
	repoFlag := cmd.Flags().Lookup("repo")
	if repoFlag == nil {
		t.Error("expected --repo flag")
	}
	
	stateFlag := cmd.Flags().Lookup("state")
	if stateFlag == nil {
		t.Error("expected --state flag")
	}
	
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected --limit flag")
	}
}

func TestViewCommand(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdView(factory)
	
	if cmd.Use != "view <pr-number> [file-path]" {
		t.Errorf("expected Use to be 'view <pr-number> [file-path]', got %q", cmd.Use)
	}
	
	// Check flags
	repoFlag := cmd.Flags().Lookup("repo")
	if repoFlag == nil {
		t.Error("expected --repo flag")
	}
	
	// Verify --files and --comments flags no longer exist
	if cmd.Flags().Lookup("files") != nil {
		t.Error("--files flag should not exist")
	}
	
	if cmd.Flags().Lookup("comments") != nil {
		t.Error("--comments flag should not exist")
	}
	
	// Verify at least 1 arg required
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

func TestViewCommandNoApprovalFlags(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdView(factory)
	
	// Verify approval flags don't exist on view command
	if cmd.Flags().Lookup("approve") != nil {
		t.Error("--approve flag should not exist on view command")
	}
	
	if cmd.Flags().Lookup("request-changes") != nil {
		t.Error("--request-changes flag should not exist on view command")
	}
	
	if cmd.Flags().Lookup("comment") != nil {
		t.Error("--comment flag should not exist on view command")
	}
}
