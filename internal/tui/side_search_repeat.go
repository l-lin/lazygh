package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextUserSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatSideSearch{Focus: FocusUserView, Direction: searchRepeatForward})
}

func (program *Program) previousUserSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatSideSearch{Focus: FocusUserView, Direction: searchRepeatBackward})
}

func (program *Program) nextNotificationsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatSideSearch{Focus: FocusNotificationsView, Direction: searchRepeatForward})
}

func (program *Program) previousNotificationsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatSideSearch{Focus: FocusNotificationsView, Direction: searchRepeatBackward})
}

func (program *Program) sideSearchState(focus Focus) (string, []int, int) {
	switch focus {
	case FocusNotificationsView:
		return program.model.NotificationSearchQuery(), program.model.visibleNotificationIndexes(), program.model.SelectedNotificationIndex()
	default:
		return program.model.UserSearchQuery(), program.model.visibleUserIndexes(), program.model.SelectedUserIndex()
	}
}
