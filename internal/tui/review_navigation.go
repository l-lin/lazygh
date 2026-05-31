package tui

import "github.com/jesseduffield/gocui"

type reviewNavigationDirection int

const (
	reviewNavigationBackward reviewNavigationDirection = -1
	reviewNavigationForward  reviewNavigationDirection = 1
)

func (program *Program) previousReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewFile{Delta: int(reviewNavigationBackward)})
}

func (program *Program) nextReviewFile(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewFile{Delta: int(reviewNavigationForward)})
}

func (program *Program) previousReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationBackward})
}

func (program *Program) nextReviewComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationForward})
}

type reviewCommentLocation struct {
	fileTreeRow  int
	renderedLine int
}

func reviewDiffRenderedRowIsThreadStatus(row reviewDiffRenderedRow) bool {
	return row.Thread != nil && row.Kind == reviewDiffRenderedRowKindInlineCommentHeader
}
