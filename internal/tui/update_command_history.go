package tui

import (
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) updateCommandHistoryStore(transition func(commandHistoryStore) commandHistoryStore) {
	if program == nil || program.commandHistoryStore == nil || transition == nil {
		return
	}
	updatedStore := transition(*program.commandHistoryStore)
	program.commandHistoryStore = &updatedStore
}

func (program *Program) applyGHCommandStarted(message MsgGHCommandStarted) {
	program.updateCommandHistoryStore(func(store commandHistoryStore) commandHistoryStore {
		return store.withCommandStarted(message.Command, message.StartedAt)
	})
}

func (program *Program) ReportGHCommandStarted(command string, startedAt time.Time) {
	if program == nil {
		return
	}
	program.dispatchAsyncMessage(MsgGHCommandStarted{
		Command:   command,
		StartedAt: startedAt,
	})
}

func (program *Program) ReportCommandActivity(activity githubdomain.CommandActivity) {
	if program != nil {
		program.dispatchAsyncMessage(MsgCommandActivity{Activity: activity})
	}
}

func (program *Program) routeCommandHistoryMessages(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgCommandActivity:
		program.updateCommandHistoryStore(func(store commandHistoryStore) commandHistoryStore {
			return store.withActivity(actual.Activity)
		})
		return handledUpdate(nil)
	case MsgGHCommandStarted:
		program.applyGHCommandStarted(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}
