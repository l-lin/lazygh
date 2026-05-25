package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextPullRequestsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewModeActive() {
		return program.nextReviewFileTreeSearchMatch(gui, view)
	}
	return program.dispatch(gui, MsgRepeatPullRequestSearch{Direction: searchRepeatForward})
}

func (program *Program) previousPullRequestsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewModeActive() {
		return program.previousReviewFileTreeSearchMatch(gui, view)
	}
	return program.dispatch(gui, MsgRepeatPullRequestSearch{Direction: searchRepeatBackward})
}
