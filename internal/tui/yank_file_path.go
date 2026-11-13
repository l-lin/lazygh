package tui

import "github.com/jesseduffield/gocui"

const (
	yankPathSuccessMessage     = iconCopy + " Path copied"
	yankPathUnavailableMessage = iconUnavailable + " No path selected"
)

func (program *Program) copyReviewFilePath(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCopyReviewFilePathRequested{})
}
