package tui

import (
	"errors"
	"os/exec"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

var ErrLinkOpenerUnavailable = errors.New("link opener is unavailable")

type linkOpener interface {
	Open(string) error
}

type systemLinkOpener struct {
	command []string
	start   func(string, ...string) error
}

func newSystemLinkOpener(command []string) *systemLinkOpener {
	return &systemLinkOpener{command: append([]string(nil), command...), start: startDetachedCommand}
}

func (program *Program) ApplyLinksConfig(config appconfig.LinksConfig) {
	if program == nil {
		return
	}

	resolved := appconfig.ResolveLinksConfig(config)
	if actual, ok := program.linkOpener.(*systemLinkOpener); ok {
		actual.command = append([]string(nil), resolved.OpenCommand...)
		return
	}

	program.linkOpener = newSystemLinkOpener(resolved.OpenCommand)
}

func (opener *systemLinkOpener) Open(url string) error {
	trimmedURL := strings.TrimSpace(url)
	if len(opener.command) == 0 || strings.TrimSpace(opener.command[0]) == "" {
		return ErrLinkOpenerUnavailable
	}
	if trimmedURL == "" {
		return ErrNoLinkUnderCursor
	}

	return opener.start(opener.command[0], append(opener.command[1:], trimmedURL)...)
}

func startDetachedCommand(name string, args ...string) error {
	command := exec.Command(name, args...)
	if err := command.Start(); err != nil {
		return err
	}
	if command.Process == nil {
		return nil
	}

	return command.Process.Release()
}
