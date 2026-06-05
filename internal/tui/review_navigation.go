package tui

import "github.com/jesseduffield/gocui"

type reviewNavigationDirection int

type reviewCommentNavigationFilter int

const (
	reviewNavigationBackward reviewNavigationDirection = -1
	reviewNavigationForward  reviewNavigationDirection = 1
)

const (
	reviewCommentNavigationFilterAll reviewCommentNavigationFilter = iota
	reviewCommentNavigationFilterUnresolvedOnly
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

func (program *Program) previousReviewUnresolvedComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationBackward, Filter: reviewCommentNavigationFilterUnresolvedOnly})
}

func (program *Program) nextReviewUnresolvedComment(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveReviewComment{Direction: reviewNavigationForward, Filter: reviewCommentNavigationFilterUnresolvedOnly})
}

type reviewCommentLocation struct {
	fileTreeRow  int
	renderedLine int
}

func reviewDiffRenderedRowIsThreadStatus(row reviewDiffRenderedRow) bool {
	return row.Thread != nil && row.Kind == reviewDiffRenderedRowKindInlineCommentHeader
}

func reviewDiffRenderedRowMatchesCommentNavigationFilter(row reviewDiffRenderedRow, filter reviewCommentNavigationFilter) bool {
	if !reviewDiffRenderedRowIsThreadStatus(row) {
		return false
	}
	if filter != reviewCommentNavigationFilterUnresolvedOnly {
		return true
	}
	return row.Thread != nil && !row.Thread.IsResolved
}
