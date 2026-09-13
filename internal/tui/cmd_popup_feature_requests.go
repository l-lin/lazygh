package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type notificationMutationRequest interface {
	run(notificationMutationCommandDeps) error
}

type storyReviewPreparationRequest interface {
	run(storyReviewPrepareCommandDeps) (preparedStoryReview, error)
}

type notificationMutationCmd struct {
	Snapshot               notificationMutationSnapshot
	SuccessFeedbackMessage string
	StatusLineOperationID  uint64
	request                notificationMutationRequest
}

type storyReviewPrepareCmd struct {
	request               storyReviewPreparationRequest
	statusLineOperationID uint64
}

func newNotificationMutationCommandDeps(program *Program) notificationMutationCommandDeps {
	if program == nil {
		return notificationMutationCommandDeps{}
	}
	return notificationMutationCommandDeps{
		notificationMutations:           program.notificationMutations,
		hideDoneNotificationsBestEffort: program.hideDoneNotificationsBestEffort,
	}
}

func newStoryReviewPrepareCommandDeps(program *Program) storyReviewPrepareCommandDeps {
	if program == nil {
		return storyReviewPrepareCommandDeps{}
	}
	deps := storyReviewPrepareCommandDeps{
		detailQueries:     program.detailQueries,
		reviewMutations:   program.reviewMutations,
		storyGenerator:    program.storyGenerator,
		storyReviewConfig: program.runtimeConfig.storyReviewConfig,
	}
	deps.pullRequestDetailForSummary = func(summary githubdomain.PullRequest) (pullRequestDetailResult, bool) {
		return program.pullRequestDetailForSummary(summary)
	}
	return deps
}

func (command notificationMutationCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.request == nil {
		return
	}

	capturedGUI := program.captureGUI(gui)
	deps := newNotificationMutationCommandDeps(program)
	run := func() {
		err := command.request.run(deps)
		message := MsgNotificationMutationFinished{
			Snapshot:               command.Snapshot,
			SuccessFeedbackMessage: command.SuccessFeedbackMessage,
			Err:                    err,
			StatusLineOperationID:  command.StatusLineOperationID,
		}
		if capturedGUI == nil {
			_ = program.executeRuntimeMessage(nil, message)
			return
		}
		program.dispatchAsyncMessage(message)
	}
	if capturedGUI == nil {
		run()
		return
	}
	program.runAsync(run)
}

func (command storyReviewPrepareCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.request == nil {
		return
	}

	capturedGUI := program.captureGUI(gui)
	deps := newStoryReviewPrepareCommandDeps(program)
	run := func() {
		prepared, err := command.request.run(deps)
		message := MsgStoryReviewPrepared{Prepared: prepared, Err: err, StatusLineOperationID: command.statusLineOperationID}
		if capturedGUI == nil {
			_ = program.executeRuntimeMessage(nil, message)
			return
		}
		program.dispatchAsyncMessage(message)
	}
	if capturedGUI == nil {
		run()
		return
	}
	program.runAsync(run)
}
