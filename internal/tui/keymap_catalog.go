package tui

import "github.com/jesseduffield/gocui"

const (
	keymapScopePrefix               = "_prefix"
	keymapScopeGlobal               = "global"
	keymapScopeMain                 = "main"
	keymapScopeSide                 = "side"
	keymapScopeUser                 = "user"
	keymapScopePullRequests         = "pull_requests"
	keymapScopeNotifications        = "notifications"
	keymapScopeDetail               = "detail"
	keymapScopeSearch               = "search"
	keymapScopeActionsPopup         = "actions_popup"
	keymapScopeActionsPopupSearch   = "actions_popup_search"
	keymapScopeModalEditor          = "modal_editor"
	keymapScopePullRequestBuildInfo = "pull_request_build_info"
	keymapScopeHelp                 = "help"
)

var mainPaneViewNames = []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewDetailName}
var sidePaneViewNames = []string{viewUserName, viewPullRequestsName, viewNotificationsName}

func keybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...string) keybindingAction {
	return keybindingAction{
		id:              keybindingActionID{scope: scope, action: action},
		viewNames:       append([]string(nil), viewNames...),
		defaultBindings: mustConfiguredKeySequences(bindings...),
		handler:         handler,
		allowSequences:  true,
	}
}

func keybindingActionForDirectOnly(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...string) keybindingAction {
	definition := keybindingActionFor(scope, action, viewNames, handler, bindings...)
	definition.allowSequences = false
	return definition
}

func (program *Program) keybindingActions() []keybindingAction {
	return []keybindingAction{
		keybindingActionFor(keymapScopeGlobal, "quit", []string{""}, program.quit, "ctrl+c"),
		keybindingActionFor(keymapScopeGlobal, "next_side_view", []string{""}, program.nextSideView, "tab"),
		keybindingActionFor(keymapScopeGlobal, "previous_side_view", []string{""}, program.previousSideView, "shift+tab"),

		keybindingActionFor(keymapScopeMain, "toggle_help", mainPaneViewNames, program.toggleHelp, "?"),
		keybindingActionFor(keymapScopeMain, "focus_user_view", mainPaneViewNames, program.focusUserView, "1"),
		keybindingActionFor(keymapScopeMain, "focus_pull_requests_view", mainPaneViewNames, program.focusPullRequestsView, "2"),
		keybindingActionFor(keymapScopeMain, "focus_notifications_view", mainPaneViewNames, program.focusNotificationsView, "3"),
		keybindingActionFor(keymapScopeMain, "open_search", mainPaneViewNames, program.openSearch, "/"),
		keybindingActionFor(keymapScopeMain, "move_selection_down", mainPaneViewNames, program.moveSelectionDown, "j", "down"),
		keybindingActionFor(keymapScopeMain, "move_selection_up", mainPaneViewNames, program.moveSelectionUp, "k", "up"),
		keybindingActionFor(keymapScopeMain, "move_detail_view_down", mainPaneViewNames, program.moveDetailViewDown, "J"),
		keybindingActionFor(keymapScopeMain, "move_detail_view_up", mainPaneViewNames, program.moveDetailViewUp, "K"),
		keybindingActionFor(keymapScopeMain, "page_down", mainPaneViewNames, program.pageDown, "ctrl+d"),
		keybindingActionFor(keymapScopeMain, "page_up", mainPaneViewNames, program.pageUp, "ctrl+u"),
		keybindingActionFor(keymapScopeMain, "full_page_down", mainPaneViewNames, program.fullPageDown, "ctrl+f", "pagedown"),
		keybindingActionFor(keymapScopeMain, "full_page_up", mainPaneViewNames, program.fullPageUp, "ctrl+b", "pageup"),
		keybindingActionFor(keymapScopeMain, "grow_focused_pane", mainPaneViewNames, program.growFocusedPane, "+"),
		keybindingActionFor(keymapScopeMain, "shrink_focused_pane", mainPaneViewNames, program.shrinkFocusedPane, "-"),

		keybindingActionFor(keymapScopeSide, "next_side_view", sidePaneViewNames, program.nextSideView, "l"),
		keybindingActionFor(keymapScopeSide, "previous_side_view", sidePaneViewNames, program.previousSideView, "h"),
		keybindingActionFor(keymapScopeSide, "focus_detail_view", sidePaneViewNames, program.focusDetailView, "0"),
		keybindingActionFor(keymapScopeSide, "move_selection_to_top", sidePaneViewNames, program.moveSideSelectionToTop, "gg"),
		keybindingActionFor(keymapScopeSide, "move_selection_to_bottom", sidePaneViewNames, program.moveSideSelectionToBottom, "G"),
		keybindingActionFor(keymapScopeSide, "recenter_selection", sidePaneViewNames, program.recenterSideSelection, "zz"),
		keybindingActionFor(keymapScopeSide, "place_selection_at_viewport_top", sidePaneViewNames, program.moveSideSelectionToViewportTop, "zt"),
		keybindingActionFor(keymapScopeSide, "place_selection_at_viewport_bottom", sidePaneViewNames, program.moveSideSelectionToViewportBottom, "zb"),
		keybindingActionFor(keymapScopeSide, "exit_review_mode", sidePaneViewNames, program.exitReviewMode, "esc", "ctrl+[", "q"),

		keybindingActionFor(keymapScopeUser, "open_detail", []string{viewUserName}, program.openDetail, "enter"),
		keybindingActionFor(keymapScopeUser, "copy_pull_request_url", []string{viewUserName}, program.copyPullRequestURL, "y"),
		keybindingActionFor(keymapScopeUser, "open_actions_popup", []string{viewUserName}, program.openActionsPopup, "a"),

		keybindingActionForDirectOnly(keymapScopePullRequests, "previous_tab", []string{viewPullRequestsName}, program.previousPullRequestTab, "["),
		keybindingActionForDirectOnly(keymapScopePullRequests, "next_tab", []string{viewPullRequestsName}, program.nextPullRequestTab, "]"),
		keybindingActionFor(keymapScopePullRequests, "open_detail", []string{viewPullRequestsName}, program.openDetail, "enter"),
		keybindingActionFor(keymapScopePullRequests, "copy_pull_request_url", []string{viewPullRequestsName}, program.copyPullRequestURL, "y"),
		keybindingActionFor(keymapScopePullRequests, "comment_on_pull_request", []string{viewPullRequestsName}, program.openPullRequestCommentComposer, "c"),
		keybindingActionFor(keymapScopePullRequests, "open_actions_popup", []string{viewPullRequestsName}, program.openActionsPopup, "a"),
		keybindingActionFor(keymapScopePullRequests, "toggle_fold", []string{viewPullRequestsName}, program.togglePullRequestFold, "za"),
		keybindingActionFor(keymapScopePullRequests, "close_all_folds", []string{viewPullRequestsName}, program.closeAllReviewTreeFolds, "zM"),
		keybindingActionFor(keymapScopePullRequests, "open_all_folds", []string{viewPullRequestsName}, program.openAllReviewTreeFolds, "zR"),
		keybindingActionFor(keymapScopePullRequests, "next_search_match", []string{viewPullRequestsName}, program.nextReviewFileTreeSearchMatch, "n"),
		keybindingActionFor(keymapScopePullRequests, "previous_search_match", []string{viewPullRequestsName}, program.previousReviewFileTreeSearchMatch, "N"),

		keybindingActionFor(keymapScopeNotifications, "open_detail", []string{viewNotificationsName}, program.openDetail, "enter"),
		keybindingActionFor(keymapScopeNotifications, "mark_notification_read", []string{viewNotificationsName}, program.markNotificationRead, "r"),
		keybindingActionFor(keymapScopeNotifications, "mark_notification_done", []string{viewNotificationsName}, program.markNotificationDone, "d"),
		keybindingActionFor(keymapScopeNotifications, "open_actions_popup", []string{viewNotificationsName}, program.openActionsPopup, "a"),

		keybindingActionFor(keymapScopeDetail, "move_cursor_left", []string{viewDetailName}, program.moveDetailCursorLeft, "h"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_right", []string{viewDetailName}, program.moveDetailCursorRight, "l"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_row_start", []string{viewDetailName}, program.moveDetailCursorToRowStart, "0"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_row_end", []string{viewDetailName}, program.moveDetailCursorToRowEnd, "$"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_top", []string{viewDetailName}, program.moveDetailCursorToTop, "gg"),
		keybindingActionFor(keymapScopeDetail, "open_link_under_cursor", []string{viewDetailName}, program.openLinkUnderCursor, "gx"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_bottom", []string{viewDetailName}, program.moveDetailCursorToBottom, "G"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_next_word", []string{viewDetailName}, program.moveDetailCursorToNextWord, "w"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_word_end", []string{viewDetailName}, program.moveDetailCursorToWordEnd, "e"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_previous_word", []string{viewDetailName}, program.moveDetailCursorToPreviousWord, "b"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_next_big_word", []string{viewDetailName}, program.moveDetailCursorToNextBigWord, "W"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_big_word_end", []string{viewDetailName}, program.moveDetailCursorToBigWordEnd, "E"),
		keybindingActionFor(keymapScopeDetail, "move_cursor_to_previous_big_word", []string{viewDetailName}, program.moveDetailCursorToPreviousBigWord, "B"),
		keybindingActionFor(keymapScopeDetail, "next_search_match", []string{viewDetailName}, program.nextDetailSearchMatch, "n"),
		keybindingActionFor(keymapScopeDetail, "previous_search_match", []string{viewDetailName}, program.previousDetailSearchMatch, "N"),
		keybindingActionFor(keymapScopeDetail, "enter_visual_mode", []string{viewDetailName}, program.enterDetailVisualMode, "v"),
		keybindingActionFor(keymapScopeDetail, "enter_line_visual_mode", []string{viewDetailName}, program.enterDetailLineVisualMode, "V"),
		keybindingActionForDirectOnly(keymapScopeDetail, "previous_tab", []string{viewDetailName}, program.previousDetailTab, "["),
		keybindingActionForDirectOnly(keymapScopeDetail, "next_tab", []string{viewDetailName}, program.nextDetailTab, "]"),
		keybindingActionFor(keymapScopeDetail, "copy_pull_request_url", []string{viewDetailName}, program.copyPullRequestURL, "y"),
		keybindingActionFor(keymapScopeDetail, "comment_on_pull_request", []string{viewDetailName}, program.openPullRequestCommentComposer, "c"),
		keybindingActionFor(keymapScopeDetail, "open_actions_popup", []string{viewDetailName}, program.openActionsPopup, "a"),
		keybindingActionFor(keymapScopeDetail, "recenter_cursor", []string{viewDetailName}, program.recenterDetailView, "zz"),
		keybindingActionFor(keymapScopeDetail, "place_cursor_at_viewport_top", []string{viewDetailName}, program.moveDetailCursorToViewportTop, "zt"),
		keybindingActionFor(keymapScopeDetail, "place_cursor_at_viewport_bottom", []string{viewDetailName}, program.moveDetailCursorToViewportBottom, "zb"),
		keybindingActionFor(keymapScopeDetail, "toggle_inline_conversation", []string{viewDetailName}, program.toggleInlineConversationVisibility, "enter", "za"),
		keybindingActionFor(keymapScopeDetail, "close_all_folds", []string{viewDetailName}, program.closeAllDetailFolds, "zM"),
		keybindingActionFor(keymapScopeDetail, "open_all_folds", []string{viewDetailName}, program.openAllDetailFolds, "zR"),
		keybindingActionFor(keymapScopeDetail, "close", []string{viewDetailName}, program.closeDetail, "esc", "ctrl+[", "q"),

		keybindingActionFor(keymapScopeSearch, "submit", []string{viewSearchName}, program.submitSearch, "enter", "ctrl+j", "ctrl+s"),
		keybindingActionFor(keymapScopeSearch, "cancel", []string{viewSearchName}, program.cancelSearch, "esc", "ctrl+["),

		keybindingActionFor(keymapScopeActionsPopup, "focus_search", []string{viewActionsPopupName}, program.focusActionsPopupSearch, "/"),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_down", []string{viewActionsPopupName}, program.moveActionsPopupSelectionDown, "j", "down"),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_up", []string{viewActionsPopupName}, program.moveActionsPopupSelectionUp, "k", "up"),
		keybindingActionFor(keymapScopeActionsPopup, "page_down", []string{viewActionsPopupName}, program.pageActionsPopupDown, "ctrl+d"),
		keybindingActionFor(keymapScopeActionsPopup, "page_up", []string{viewActionsPopupName}, program.pageActionsPopupUp, "ctrl+u"),
		keybindingActionFor(keymapScopeActionsPopup, "full_page_down", []string{viewActionsPopupName}, program.fullPageActionsPopupDown, "ctrl+f", "pagedown"),
		keybindingActionFor(keymapScopeActionsPopup, "full_page_up", []string{viewActionsPopupName}, program.fullPageActionsPopupUp, "ctrl+b", "pageup"),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_to_top", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToTop, "gg"),
		keybindingActionFor(keymapScopeActionsPopup, "move_selection_to_bottom", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToBottom, "G"),
		keybindingActionFor(keymapScopeActionsPopup, "recenter_selection", []string{viewActionsPopupName}, program.recenterActionsPopupSelection, "zz"),
		keybindingActionFor(keymapScopeActionsPopup, "place_selection_at_viewport_top", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToViewportTop, "zt"),
		keybindingActionFor(keymapScopeActionsPopup, "place_selection_at_viewport_bottom", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToViewportBottom, "zb"),
		keybindingActionFor(keymapScopeActionsPopup, "execute_selected_action", []string{viewActionsPopupName}, program.executeSelectedActionsPopupAction, "enter"),
		keybindingActionFor(keymapScopeActionsPopup, "submit_selected_picker", []string{viewActionsPopupName}, program.submitSelectedActionsPopupAction, "alt+enter"),
		keybindingActionFor(keymapScopeActionsPopup, "close", []string{viewActionsPopupName}, program.closeActionsPopup, "esc", "ctrl+[", "q"),

		keybindingActionFor(keymapScopeActionsPopupSearch, "focus_list", []string{viewActionsPopupSearchName}, program.focusActionsPopupList, "enter", "tab", "ctrl+s"),
		keybindingActionFor(keymapScopeActionsPopupSearch, "close", []string{viewActionsPopupSearchName}, program.closeActionsPopup, "esc", "ctrl+["),

		keybindingActionFor(keymapScopeModalEditor, "submit", []string{viewModalEditorName}, program.submitModalEditor, "alt+enter", "ctrl+s"),
		keybindingActionFor(keymapScopeModalEditor, "close", []string{viewModalEditorName}, program.closeModalEditor, "esc", "ctrl+["),

		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_left", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorLeft, "h"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_down", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorDown, "j", "down"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_up", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorUp, "k", "up"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_right", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorRight, "l"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_row_start", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToRowStart, "0"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_row_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToRowEnd, "$"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_top", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToTop, "gg"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "open_link_under_cursor", []string{viewPullRequestBuildInfoName}, program.openPullRequestBuildRunPopupLinkUnderCursor, "gx"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_bottom", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToBottom, "G"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_next_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToNextWord, "w"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_word_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToWordEnd, "e"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_previous_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToPreviousWord, "b"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_next_big_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToNextBigWord, "W"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_big_word_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToBigWordEnd, "E"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_previous_big_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToPreviousBigWord, "B"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "next_search_match", []string{viewPullRequestBuildInfoName}, program.nextPullRequestBuildRunPopupSearchMatch, "n"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "previous_search_match", []string{viewPullRequestBuildInfoName}, program.previousPullRequestBuildRunPopupSearchMatch, "N"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "enter_visual_mode", []string{viewPullRequestBuildInfoName}, program.enterPullRequestBuildRunPopupVisualMode, "v"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "enter_line_visual_mode", []string{viewPullRequestBuildInfoName}, program.enterPullRequestBuildRunPopupLineVisualMode, "V"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "open_search", []string{viewPullRequestBuildInfoName}, program.openSearch, "/"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "copy_content", []string{viewPullRequestBuildInfoName}, program.copyPullRequestBuildRunPopupContent, "y"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "open_actions_popup", []string{viewPullRequestBuildInfoName}, program.openActionsPopup, "a"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "page_down", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupDown, "ctrl+d"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "page_up", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupUp, "ctrl+u"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_down", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupDown, "ctrl+f", "pagedown"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_up", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupUp, "ctrl+b", "pageup"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "close", []string{viewPullRequestBuildInfoName}, program.closePullRequestBuildRunPopup, "esc", "ctrl+[", "q"),

		keybindingActionFor(keymapScopeHelp, "full_page_down", []string{viewHelpName}, program.fullPageHelpDown, "ctrl+f", "pagedown"),
		keybindingActionFor(keymapScopeHelp, "full_page_up", []string{viewHelpName}, program.fullPageHelpUp, "ctrl+b", "pageup"),
		keybindingActionFor(keymapScopeHelp, "close", []string{viewHelpName}, program.closeHelp, "esc", "ctrl+[", "q"),
	}
}
