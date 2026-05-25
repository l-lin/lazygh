package tui

import "github.com/jesseduffield/gocui"

type notificationMutationRequest interface {
	run(*Program) error
}

type storyReviewPreparationRequest interface {
	run(*Program) (preparedStoryReview, error)
}

type notificationMutationCmd struct {
	Snapshot               notificationMutationSnapshot
	SuccessFeedbackMessage string
	request                notificationMutationRequest
}

type storyReviewPrepareCmd struct {
	request storyReviewPreparationRequest
}

func (command notificationMutationCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.request == nil {
		return
	}

	run := func() {
		err := command.request.run(program)
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
	if program == nil || command.request == nil {
		return
	}

	run := func() {
		prepared, err := command.request.run(program)
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
