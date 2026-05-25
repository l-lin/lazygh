package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewUserFooterName          = "user-footer"
	viewPullRequestsFooterName  = "pull-requests-footer"
	viewNotificationsFooterName = "notifications-footer"
	viewDetailFooterName        = "detail-footer"

	pullRequestDetailLoadingTitle = "Loading pull request detail..."
)

type paneFooterState struct {
	searchSummary string
}

func (state paneFooterState) Text() string {
	return strings.TrimSpace(state.searchSummary)
}

func (state paneFooterState) Visible() bool {
	return state.Text() != ""
}

type statusLineHintSpec struct {
	label     string
	fallback  string
	actionIDs []keybindingActionID
}

func searchSummaryText(query string, count int) string {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return ""
	}

	return fmt.Sprintf("/%s (%d %s)", trimmedQuery, count, pluralize(count, "match", "matches"))
}

func paneFooterActionsActionID(focus Focus) (keybindingActionID, bool) {
	switch focus {
	case FocusUserView:
		return keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}, true
	case FocusPullRequestsView:
		return keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}, true
	case FocusNotificationsView:
		return keybindingActionID{scope: keymapScopeNotifications, action: "open_actions_popup"}, true
	case FocusDetailView:
		return keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}, true
	default:
		return keybindingActionID{}, false
	}
}

func (program *Program) configurePaneFooterView(view *gocui.View) {
	program.configureBottomPromptView(view, nil, false)
	view.Editable = false
	view.Editor = nil
}

func (program *Program) renderPaneFooterView(view *gocui.View, text string) {
	if view == nil {
		return
	}

	view.Clear()
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
	fmt.Fprint(view, strings.TrimSpace(text))
}

func paneViewName(focus Focus) string {
	switch focus {
	case FocusPullRequestsView:
		return viewPullRequestsName
	case FocusNotificationsView:
		return viewNotificationsName
	case FocusDetailView:
		return viewDetailName
	default:
		return viewUserName
	}
}
