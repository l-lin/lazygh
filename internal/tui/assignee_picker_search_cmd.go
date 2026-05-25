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

type assigneePickerSearchCommandDeps struct {
	visible              bool
	repository           string
	pullRequestMutations PullRequestMutations
	requestCurrent       func(int, string) bool
	runAsync             func(func())
	dispatchAsync        func(*gocui.Gui, Msg)
}

func newAssigneePickerSearchCommandDeps(program *Program) assigneePickerSearchCommandDeps {
	if program == nil {
		return assigneePickerSearchCommandDeps{}
	}

	repository := ""
	if program.actionsPopupWidget.assigneePicker != nil {
		repository = program.actionsPopupWidget.assigneePicker.target.repository
	}
	return assigneePickerSearchCommandDeps{
		visible:              program.assigneePickerVisible(),
		repository:           repository,
		pullRequestMutations: program.pullRequestMutations,
		requestCurrent:       program.assigneePickerSearchRequestCurrent,
		runAsync:             program.runAsync,
		dispatchAsync:        program.dispatchAsync,
	}
}

func (command assigneePickerSearchCmd) execute(program *Program, gui *gocui.Gui) {
	executeAssigneePickerSearchCommand(newAssigneePickerSearchCommandDeps(program), gui, command)
}

func executeAssigneePickerSearchCommand(deps assigneePickerSearchCommandDeps, gui *gocui.Gui, command assigneePickerSearchCmd) {
	if command.RequestID <= 0 || !deps.visible || deps.pullRequestMutations == nil {
		return
	}

	trimmedQuery := strings.TrimSpace(command.Query)
	run := func() {
		if command.Delay > 0 {
			timer := time.NewTimer(command.Delay)
			defer timer.Stop()
			<-timer.C
		}

		if command.DispatchLoading && deps.dispatchAsync != nil {
			deps.dispatchAsync(gui, MsgAssigneePickerSearchLoadingStarted{RequestID: command.RequestID, Query: trimmedQuery})
		}
		if deps.requestCurrent == nil || !deps.requestCurrent(command.RequestID, trimmedQuery) {
			return
		}
		results, err := deps.pullRequestMutations.SearchAssignableUsers(deps.repository, trimmedQuery)
		if deps.dispatchAsync != nil {
			deps.dispatchAsync(gui, MsgAssigneePickerSearchLoaded{RequestID: command.RequestID, Query: trimmedQuery, Results: results, Err: err})
		}
	}
	if deps.runAsync != nil {
		deps.runAsync(run)
		return
	}
	run()
}
