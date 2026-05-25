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
	request                notificationMutationRequest
}

type storyReviewPrepareCmd struct {
	request storyReviewPreparationRequest
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

	deps := newNotificationMutationCommandDeps(program)
	run := func() {
		err := command.request.run(deps)
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

	deps := newStoryReviewPrepareCommandDeps(program)
	run := func() {
		prepared, err := command.request.run(deps)
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
