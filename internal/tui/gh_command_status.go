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
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withGHCommandLoadingStarted(command)
	})
}

func (program *Program) clearGHCommandLoading() {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withGHCommandLoadingCleared()
	})
}
