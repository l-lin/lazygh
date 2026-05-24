package tui

import "time"

type searchWidgetState struct {
	editor         *lineEditor
	detailReversed bool
}

type actionsPopupWidgetState struct {
	searchEditor                      *lineEditor
	errorMessage                      string
	pendingConfirmationActionID       string
	reactionPicker                    *reactionPickerState
	themePicker                       *themePickerState
	assigneePicker                    *assigneePickerState
	assigneePickerSearchDebounceDelay time.Duration
	assigneePickerLoad                *assigneePickerLoadState
}
