package prompter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter provides methods for interactive user input
type Prompter interface {
	// Input prompts for a text input value
	Input(prompt string) (string, error)

	// Password prompts for a password/token (hidden input)
	Password(prompt string) (string, error)

	// Confirm prompts for a yes/no confirmation
	Confirm(prompt string, defaultYes bool) (bool, error)
}

// New creates a new prompter using the given input and output streams
func New(in io.Reader, out io.Writer, errOut io.Writer) Prompter {
	return &stdPrompter{
		in:     in,
		out:    out,
		errOut: errOut,
	}
}

type stdPrompter struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

// Input prompts for text input
func (p *stdPrompter) Input(prompt string) (string, error) {
	_, _ = fmt.Fprint(p.errOut, prompt)

	reader := bufio.NewReader(p.in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// Password prompts for hidden password/token input
func (p *stdPrompter) Password(prompt string) (string, error) {
	_, _ = fmt.Fprint(p.errOut, prompt)

	// Check if stdin is a terminal
	if file, ok := p.in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		password, err := term.ReadPassword(int(file.Fd()))
		_, _ = fmt.Fprintln(p.errOut) // Print newline after hidden input
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(password)), nil
	}

	// Fallback to regular input for non-TTY (pipes, files, etc.)
	reader := bufio.NewReader(p.in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// Confirm prompts for yes/no confirmation
func (p *stdPrompter) Confirm(prompt string, defaultYes bool) (bool, error) {
	defaultLabel := "y/N"
	if defaultYes {
		defaultLabel = "Y/n"
	}

	fullPrompt := fmt.Sprintf("%s [%s]: ", prompt, defaultLabel)
	response, err := p.Input(fullPrompt)
	if err != nil {
		return false, err
	}

	if response == "" {
		return defaultYes, nil
	}

	response = strings.ToLower(response)
	if response == "y" || response == "yes" {
		return true, nil
	}
	if response == "n" || response == "no" {
		return false, nil
	}

	return defaultYes, nil
}
