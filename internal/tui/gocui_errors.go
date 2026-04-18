package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

func isUnknownViewError(err error) bool {
	return isGocuiError(err, gocui.ErrUnknownView)
}

func isQuitError(err error) bool {
	return isGocuiError(err, gocui.ErrQuit)
}

func isGocuiError(err error, target error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, target) {
		return true
	}

	return strings.Contains(err.Error(), target.Error())
}
