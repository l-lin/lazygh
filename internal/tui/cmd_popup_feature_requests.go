package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type notificationMutationCmd struct {
	Snapshot               notificationMutationSnapshot
	SuccessFeedbackMessage string
	Work                   func(*Program) error
}

type storyReviewPrepareCmd struct {
	Summary githubdomain.PullRequest
}

func (command notificationMutationCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.Work == nil {
		return
	}

	run := func() {
		err := command.Work(program)
		if gui == nil {
			program.executeCmds(gui, Update(program, MsgNotificationMutationFinished{Snapshot: command.Snapshot, SuccessFeedbackMessage: command.SuccessFeedbackMessage, Err: err}))
			return
		}
		program.dispatchAsync(gui, MsgNotificationMutationFinished{Snapshot: command.Snapshot, SuccessFeedbackMessage: command.SuccessFeedbackMessage, Err: err})
	}
	if gui == nil {
		run()
		return
	}
	program.runAsync(run)
}

func (command storyReviewPrepareCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	run := func() {
		prepared, err := program.prepareStoryReview(command.Summary)
		if gui == nil {
			program.executeCmds(gui, Update(program, MsgStoryReviewPrepared{Prepared: prepared, Err: err}))
			return
		}
		program.dispatchAsync(gui, MsgStoryReviewPrepared{Prepared: prepared, Err: err})
	}
	if gui == nil {
		run()
		return
	}
	program.runAsync(run)
}
