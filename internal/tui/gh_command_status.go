package tui

import "strings"

func formatStatusLineCommand(arguments ...string) string {
	normalizedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		trimmedArgument := strings.TrimSpace(argument)
		if trimmedArgument == "" {
			continue
		}
		normalizedArguments = append(normalizedArguments, trimmedArgument)
	}
	return strings.Join(normalizedArguments, " ")
}

func formatRunningCommandStatus(command string) string {
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return ""
	}
	return "Running `" + trimmedCommand + "`."
}

func (program *Program) startGHCommandLoading(command string) {
	if program == nil {
		return
	}

	program.feedbackMessage = ""
	program.ghCommandLoadingMessage = formatRunningCommandStatus(command)
}

func (program *Program) clearGHCommandLoading() {
	if program == nil {
		return
	}

	program.ghCommandLoadingMessage = ""
}
