package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

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

func (program *Program) repeatSideSearch(gui *gocui.Gui, focus Focus, choose searchMatchIndexChooser) error {
	if program.reviewModeActive() || program.model.Focus() != focus {
		return nil
	}

	query, matchIndexes, selectedIndex := program.sideSearchState(focus)
	if strings.TrimSpace(query) == "" {
		return nil
	}

	matchIndex := choose(matchIndexes, selectedIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return nil
	}

	program.setSideSearchSelection(focus, matchIndexes[matchIndex])
	return nil
}

func (program *Program) sideSearchState(focus Focus) (string, []int, int) {
	switch focus {
	case FocusNotificationsView:
		return program.model.NotificationSearchQuery(), program.model.visibleNotificationIndexes(), program.model.SelectedNotificationIndex()
	default:
		return program.model.UserSearchQuery(), program.model.visibleUserIndexes(), program.model.SelectedUserIndex()
	}
}

func (program *Program) setSideSearchSelection(focus Focus, index int) {
	switch focus {
	case FocusNotificationsView:
		program.model.SelectNotificationIndex(index)
	default:
		program.model.SelectUserIndex(index)
	}
}
