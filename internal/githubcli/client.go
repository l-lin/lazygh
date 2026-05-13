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

type ConnectedUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	URL         string `json:"html_url"`
}

type execRunner struct{}

func (service *SessionService) GetConnectedUser() (ConnectedUser, error) {
	result, err := service.doREST(RESTRequest{Path: "user"})
	if err != nil {
		return ConnectedUser{}, err
	}

	var user ConnectedUser
	if err := service.transport.decoder.DecodeJSON(result.Stdout, &user); err != nil {
		return ConnectedUser{}, fmt.Errorf("%w: %v", ErrInvalidConnectedUserResponse, err)
	}

	user = user.normalized()
	if user.Login == "" {
		return ConnectedUser{}, ErrEmptyConnectedUser
	}

	return user, nil
}

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

func (user ConnectedUser) normalized() ConnectedUser {
	user.Login = strings.TrimSpace(user.Login)
	user.Name = strings.TrimSpace(user.Name)
	user.Bio = strings.TrimSpace(user.Bio)
	user.Company = strings.TrimSpace(user.Company)
	user.Location = strings.TrimSpace(user.Location)
	user.URL = strings.TrimSpace(user.URL)
	return user
}
