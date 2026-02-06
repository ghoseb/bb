package prompter

import (
	"bytes"
	"strings"
	"testing"
)

func TestInput(t *testing.T) {
	in := strings.NewReader("test-input\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := New(in, out, errOut)
	result, err := p.Input("Enter value: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "test-input" {
		t.Errorf("got %q, want %q", result, "test-input")
	}

	prompt := errOut.String()
	if prompt != "Enter value: " {
		t.Errorf("got prompt %q, want %q", prompt, "Enter value: ")
	}
}

func TestInputWithWhitespace(t *testing.T) {
	in := strings.NewReader("  test-input  \n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := New(in, out, errOut)
	result, err := p.Input("Enter value: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "test-input" {
		t.Errorf("got %q, want %q (whitespace should be trimmed)", result, "test-input")
	}
}

func TestPasswordFallback(t *testing.T) {
	// When input is not a TTY (e.g., pipe), Password falls back to regular input
	in := strings.NewReader("secret-token\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := New(in, out, errOut)
	result, err := p.Password("Enter password: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "secret-token" {
		t.Errorf("got %q, want %q", result, "secret-token")
	}

	prompt := errOut.String()
	if prompt != "Enter password: " {
		t.Errorf("got prompt %q, want %q", prompt, "Enter password: ")
	}
}

func TestConfirmDefaultYes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     bool
		wantErr  bool
	}{
		{"empty_default_yes", "\n", true, false},
		{"explicit_yes", "y\n", true, false},
		{"explicit_yes_full", "yes\n", true, false},
		{"explicit_no", "n\n", false, false},
		{"explicit_no_full", "no\n", false, false},
		{"uppercase_yes", "Y\n", true, false},
		{"uppercase_no", "N\n", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}

			p := New(in, out, errOut)
			result, err := p.Confirm("Continue?", true)

			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}

			if result != tt.want {
				t.Errorf("got %v, want %v (input: %q)", result, tt.want, tt.input)
			}
		})
	}
}

func TestConfirmDefaultNo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     bool
		wantErr  bool
	}{
		{"empty_default_no", "\n", false, false},
		{"explicit_yes", "y\n", true, false},
		{"explicit_no", "n\n", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			errOut := &bytes.Buffer{}

			p := New(in, out, errOut)
			result, err := p.Confirm("Continue?", false)

			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}

			if result != tt.want {
				t.Errorf("got %v, want %v (input: %q)", result, tt.want, tt.input)
			}
		})
	}
}

func TestConfirmInvalidInput(t *testing.T) {
	in := strings.NewReader("invalid\n")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	p := New(in, out, errOut)
	result, err := p.Confirm("Continue?", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Invalid input should use default
	if result != true {
		t.Errorf("got %v, want %v (invalid input should use default)", result, true)
	}
}
