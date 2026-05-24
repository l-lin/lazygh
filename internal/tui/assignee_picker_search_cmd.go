package tui

import (
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
)

type assigneePickerSearchCmd struct {
	RequestID       int
	Query           string
	Delay           time.Duration
	DispatchLoading bool
}

func (command assigneePickerSearchCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.RequestID <= 0 || !program.assigneePickerVisible() || !program.hasPullRequestMutations() {
		return
	}

	repository := program.actionsPopupWidget.assigneePicker.target.repository
	trimmedQuery := strings.TrimSpace(command.Query)
	program.runAsync(func() {
		if command.Delay > 0 {
			timer := time.NewTimer(command.Delay)
			defer timer.Stop()
			<-timer.C
		}

		if command.DispatchLoading {
			program.dispatchAsync(gui, MsgAssigneePickerSearchLoadingStarted{RequestID: command.RequestID, Query: trimmedQuery})
		}
		if !program.assigneePickerSearchRequestCurrent(command.RequestID, trimmedQuery) {
			return
		}
		results, err := program.pullRequestMutations.SearchAssignableUsers(repository, trimmedQuery)
		program.dispatchAsync(gui, MsgAssigneePickerSearchLoaded{RequestID: command.RequestID, Query: trimmedQuery, Results: results, Err: err})
	})
}
