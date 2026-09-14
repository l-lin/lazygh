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
	successMessage string
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

func (store statusStore) statusLineOperationFailureLabel(operationID uint64) string {
	if operationID == 0 || store.statusLineOperation.id != operationID {
		return ""
	}
	return store.statusLineOperation.failureLabel
}

func (store statusStore) withStatusLineOperationFinished(operationID uint64, operationErr error) statusStore {
	if operationID == 0 || store.statusLineOperation.id != operationID {
		return store
	}

	operation := store.statusLineOperation
	loadingLabel := operation.loadingLabel
	operation.loadingLabel = ""
	store.statusLineOperationStarted = false
	if operationErr == nil {
		operation.successMessage = iconStatusSuccess + " " + loadingLabel
		operation.failureMessage = ""
	} else {
		operation.successMessage = ""
		failureLabel := strings.TrimSpace(operation.failureLabel)
		if failureLabel != "" {
			operation.failureMessage = formatStatusLineFailure(failureLabel, operationErr)
		}
	}
	store.statusLineOperation = operation
	return store
}

func formatStatusLineFailureContext(action string, operationErr error) string {
	trimmedAction := strings.TrimSpace(action)
	if operationErr == nil {
		return ""
	}
	trimmedError := strings.TrimSpace(operationErr.Error())
	if trimmedAction == "" || trimmedError == "" {
		return ""
	}
	return trimmedAction + ": " + trimmedError
}

func formatStatusLineFailure(action string, operationErr error) string {
	failureContext := formatStatusLineFailureContext(action, operationErr)
	if failureContext == "" {
		return ""
	}
	return iconStatusFailure + " " + failureContext
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
	program.cancelScheduledPullRequestRefreshBatches()
	if program.manualRefreshStatusOperationID() != 0 {
		program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
			state.statusOperationID = 0
			return state
		})
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
	failureLabel := ""
	shouldRecord := operationID == 0
	if program != nil && program.statusStore != nil {
		if operationID != 0 && program.statusStore.statusLineOperation.id != operationID {
			shouldRecord = false
		} else if operationID != 0 {
			shouldRecord = true
			failureLabel = program.statusStore.statusLineOperationFailureLabel(operationID)
		}
	}
	if shouldRecord {
		program.recordStatusLineFailure(failureLabel, operationErr)
	}
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

func (program *Program) statusLineOperationSuccessStatus() string {
	if program == nil || program.statusStore == nil {
		return ""
	}
	return strings.TrimSpace(program.statusStore.statusLineOperation.successMessage)
}

func (program *Program) cancelStatusLineOperation(operationID uint64) {
	if program == nil || program.statusStore == nil || operationID == 0 {
		return
	}
	program.updateStatusStore(func(store statusStore) statusStore {
		if store.statusLineOperation.id != operationID {
			return store
		}
		store.statusLineOperation = statusLineOperationState{}
		store.statusLineOperationStarted = false
		return store
	})
}

func (program *Program) statusLineOperationStarted() bool {
	return program != nil && program.statusStore != nil && program.statusStore.statusLineOperationStarted
}
