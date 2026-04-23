package tui

import "strings"

func (program *Program) setFeedback(_ Focus, message string) {
	program.feedbackMessage = strings.TrimSpace(message)
}
