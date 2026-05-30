package tui

import (
	"fmt"
	"strings"
)

func (program *Program) currentConnectedUserLogin() string {
	return strings.TrimSpace(program.connectedUserLogin)
}

func (program *Program) currentConnectedUserName() string {
	return strings.TrimSpace(program.connectedUserName)
}

func (program *Program) shouldHighlightSelection(focus Focus, selectable bool) bool {
	if !selectable {
		return false
	}

	if program.model.Focus() == focus {
		return true
	}

	return program.model.Focus() == FocusDetailView && program.model.currentSideFocus() == focus
}

func searchNoMatchesMessage(query string) string {
	return fmt.Sprintf("No matches for %q.", strings.TrimSpace(query))
}

func (program *Program) layoutContentHeight(maxY int) int {
	if maxY < 1 {
		return 1
	}
	if maxY > 1 {
		return maxY - 1
	}
	return maxY
}
