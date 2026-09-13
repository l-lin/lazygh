package tui

import (
	"strings"
	"unicode/utf8"
)

type statusLineOperationDescriptor struct {
	loadingLabel string
	failureLabel string
}

type statusLineOperationState struct {
	id             uint64
	loadingLabel   string
	failureLabel   string
	failureMessage string
}

func (store statusStore) withStatusLineOperationStarted(descriptor statusLineOperationDescriptor) (statusStore, uint64) {
	loadingLabel := strings.TrimSpace(descriptor.loadingLabel)
	if loadingLabel == "" {
		return store, 0
	}

	store.nextStatusLineOperationID++
	store.statusLineOperationStarted = true
	failureLabel := strings.TrimSpace(descriptor.failureLabel)
	if failureLabel == "" {
		failureLabel = loadingLabel
	}
	store.statusLineOperation = statusLineOperationState{
		id:           store.nextStatusLineOperationID,
		loadingLabel: loadingLabel,
		failureLabel: failureLabel,
	}
	return store, store.nextStatusLineOperationID
}

func (store statusStore) withStatusLineOperationFinished(operationID uint64, operationErr error) statusStore {
	if operationID == 0 || store.statusLineOperation.id != operationID {
		return store
	}

	operation := store.statusLineOperation
	store.statusLineOperation = statusLineOperationState{}
	if operationErr == nil {
		return store
	}

	failureLabel := strings.TrimSpace(operation.failureLabel)
	if failureLabel == "" {
		return store
	}

	store.statusLineOperation.failureMessage = formatStatusLineFailure(failureLabel, operationErr)
	return store
}

func formatStatusLineFailure(action string, operationErr error) string {
	trimmedAction := strings.TrimSpace(action)
	if operationErr == nil {
		return ""
	}
	trimmedError := strings.TrimSpace(operationErr.Error())
	if trimmedAction == "" || trimmedError == "" {
		return ""
	}
	return iconStatusFailure + " " + trimmedAction + ": " + trimmedError
}

func truncateStatusLineText(text string, availableWidth int) string {
	trimmedText := strings.TrimSpace(text)
	if availableWidth <= 0 {
		return ""
	}
	textWidth := utf8.RuneCountInString(trimmedText)
	if textWidth <= availableWidth {
		return trimmedText
	}
	if availableWidth == 1 {
		return "…"
	}

	runes := []rune(trimmedText)
	return string(runes[:availableWidth-1]) + "…"
}

func (program *Program) startStatusLineOperation(descriptor statusLineOperationDescriptor) uint64 {
	if program == nil || program.statusStore == nil {
		return 0
	}

	var operationID uint64
	program.updateStatusStore(func(store statusStore) statusStore {
		updatedStore, actualOperationID := store.withStatusLineOperationStarted(descriptor)
		operationID = actualOperationID
		return updatedStore
	})
	return operationID
}

func (program *Program) finishStatusLineOperation(operationID uint64, operationErr error) {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withStatusLineOperationFinished(operationID, operationErr)
	})
}

func (program *Program) statusLineOperationLoadingStatus() string {
	if program == nil || program.statusStore == nil {
		return ""
	}
	return strings.TrimSpace(program.statusStore.statusLineOperation.loadingLabel)
}

func (program *Program) statusLineOperationFailureStatus() string {
	if program == nil || program.statusStore == nil {
		return ""
	}
	return strings.TrimSpace(program.statusStore.statusLineOperation.failureMessage)
}

func (program *Program) statusLineOperationStarted() bool {
	return program != nil && program.statusStore != nil && program.statusStore.statusLineOperationStarted
}
