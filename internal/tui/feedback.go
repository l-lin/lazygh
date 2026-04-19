package tui

import "strings"

func (program *Program) setFeedback(focus Focus, message string) {
	program.feedbackFocus = focus
	program.feedbackMessage = strings.TrimSpace(message)
}

func (program *Program) feedbackSuffix(focus Focus) string {
	message := strings.TrimSpace(program.feedbackMessage)
	if message == "" || program.feedbackFocus != focus {
		return ""
	}

	return " · " + message
}
