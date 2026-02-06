package list

import (
	"testing"

	"github.com/ghoseb/bb/pkg/cmdutil"
	"github.com/ghoseb/bb/pkg/iostreams"
)

func TestCommandStructure(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdList(factory)
	
	if cmd.Use != "list <command>" {
		t.Errorf("expected Use to be 'list <command>', got %q", cmd.Use)
	}
	
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	
	// Check subcommands
	subcommands := cmd.Commands()
	if len(subcommands) != 1 {
		t.Errorf("expected 1 subcommand, got %d", len(subcommands))
	}
	
	names := make(map[string]bool)
	for _, sub := range subcommands {
		names[sub.Name()] = true
	}
	
	if !names["repos"] {
		t.Error("expected 'repos' subcommand")
	}
}

func TestReposCommandFlags(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdRepos(factory)
	
	// Check workspace flag exists
	workspaceFlag := cmd.Flags().Lookup("workspace")
	if workspaceFlag == nil {
		t.Fatal("expected --workspace flag")
	}
	
	if workspaceFlag.Value.Type() != "string" {
		t.Errorf("expected --workspace to be string, got %s", workspaceFlag.Value.Type())
	}
}

func TestReposCommandArgValidation(t *testing.T) {
	ios := iostreams.System()
	factory := cmdutil.NewFactory("test", ios)
	
	cmd := NewCmdRepos(factory)
	
	// Verify Args validator is set to NoArgs
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
	
	// Test: should reject args
	err := cmd.Args(cmd, []string{"extra"})
	if err == nil {
		t.Error("expected error with args, got nil")
	}
	
	// Test: should accept no args
	err = cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("expected no error with no args, got: %v", err)
	}
}
