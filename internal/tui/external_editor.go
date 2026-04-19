package tui

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/jesseduffield/gocui"
)

type externalEditor interface {
	Edit(*gocui.Gui, string) (string, error)
}

type systemExternalEditor struct{}

func (systemExternalEditor) Edit(gui *gocui.Gui, text string) (string, error) {
	editorCommand := strings.TrimSpace(os.Getenv("EDITOR"))
	if editorCommand == "" {
		return "", errors.New("EDITOR is not set")
	}

	tempFile, err := os.CreateTemp("", "lazygh-pr-description-*.md")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()
	if _, err := tempFile.WriteString(text); err != nil {
		_ = tempFile.Close()
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		return "", err
	}

	suspended := false
	if gui != nil {
		if err := gui.Suspend(); err != nil {
			return "", err
		}
		suspended = true
	}

	command := exec.Command("sh", "-c", "$EDITOR \"$1\"", "sh", tempFile.Name())
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	runErr := command.Run()
	if suspended {
		if resumeErr := gui.Resume(); resumeErr != nil && runErr == nil {
			runErr = resumeErr
		}
	}
	if runErr != nil {
		return "", runErr
	}

	updatedText, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", err
	}

	return string(updatedText), nil
}
