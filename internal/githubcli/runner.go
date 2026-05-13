package githubcli

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const ghBinaryName = "gh"

var (
	ErrUnavailable                  = githubdomain.ErrUnavailable
	ErrUnauthenticated              = githubdomain.ErrUnauthenticated
	ErrInvalidConnectedUserResponse = errors.New("invalid connected user response")
	ErrEmptyConnectedUser           = githubdomain.ErrEmptyConnectedUser
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
}

type Runner interface {
	Run(name string, args ...string) (CommandResult, error)
	RunWithInput(name string, input []byte, args ...string) (CommandResult, error)
}

type execRunner struct{}

func (runner execRunner) Run(name string, args ...string) (CommandResult, error) {
	return runner.run(name, nil, args...)
}

func (runner execRunner) RunWithInput(name string, input []byte, args ...string) (CommandResult, error) {
	return runner.run(name, input, args...)
}

func (runner execRunner) run(name string, input []byte, args ...string) (CommandResult, error) {
	command := exec.Command(name, args...)
	if input != nil {
		command.Stdin = bytes.NewReader(input)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

func classifyCommandError(commandName string, err error, stderr []byte) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("%w: install `gh` and ensure it is in PATH", ErrUnavailable)
	}

	stderrText := strings.ToLower(strings.TrimSpace(string(stderr)))
	if isUnauthenticatedMessage(stderrText) {
		return fmt.Errorf("%w: run `gh auth login`", ErrUnauthenticated)
	}

	if stderrText == "" {
		return fmt.Errorf("run `%s`: %w", commandName, err)
	}

	return fmt.Errorf("run `%s`: %w: %s", commandName, err, strings.TrimSpace(string(stderr)))
}

func isUnauthenticatedMessage(stderr string) bool {
	patterns := []string{
		"gh auth login",
		"authentication required",
		"authentication failed",
		"not logged in",
		"not logged into",
	}

	for _, pattern := range patterns {
		if strings.Contains(stderr, pattern) {
			return true
		}
	}

	return false
}
