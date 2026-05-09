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

var sharedKeybindingDefinitions = map[string]sharedKeybindingDefinition{
	"close_all_folds":                    sharedKeybindingDefinitionWithBindings("zM"),
	"comment_on_pull_request":            sharedKeybindingDefinitionWithBindings("c"),
	"copy_pull_request_url":              sharedKeybindingDefinitionWithBindings("y"),
	"enter_line_visual_mode":             sharedKeybindingDefinitionWithBindings("V"),
	"enter_visual_mode":                  sharedKeybindingDefinitionWithBindings("v"),
	"full_page_down":                     sharedKeybindingDefinitionWithBindings("ctrl+f", "pagedown"),
	"full_page_up":                       sharedKeybindingDefinitionWithBindings("ctrl+b", "pageup"),
	"move_cursor_left":                   sharedKeybindingDefinitionWithBindings("h", "left"),
	"move_cursor_right":                  sharedKeybindingDefinitionWithBindings("l", "right"),
	"move_cursor_to_bottom":              sharedKeybindingDefinitionWithBindings("G"),
	"move_cursor_to_next_big_word":       sharedKeybindingDefinitionWithBindings("W"),
	"move_cursor_to_next_word":           sharedKeybindingDefinitionWithBindings("w"),
	"move_cursor_to_previous_big_word":   sharedKeybindingDefinitionWithBindings("B"),
	"move_cursor_to_previous_word":       sharedKeybindingDefinitionWithBindings("b"),
	"move_cursor_to_row_end":             sharedKeybindingDefinitionWithBindings("$"),
	"move_cursor_to_row_start":           sharedKeybindingDefinitionWithBindings("0"),
	"move_cursor_to_top":                 sharedKeybindingDefinitionWithBindings("gg"),
	"move_cursor_to_big_word_end":        sharedKeybindingDefinitionWithBindings("E"),
	"move_cursor_to_word_end":            sharedKeybindingDefinitionWithBindings("e"),
	"move_selection_down":                sharedKeybindingDefinitionWithBindings("j", "down"),
	"move_selection_to_bottom":           sharedKeybindingDefinitionWithBindings("G"),
	"move_selection_to_top":              sharedKeybindingDefinitionWithBindings("gg"),
	"move_selection_up":                  sharedKeybindingDefinitionWithBindings("k", "up"),
	"next_search_match":                  sharedKeybindingDefinitionWithBindings("n"),
	"next_tab":                           sharedKeybindingDefinitionWithBindings("]"),
	"open_actions_popup":                 sharedKeybindingDefinitionWithBindings("a"),
	"open_all_folds":                     sharedKeybindingDefinitionWithBindings("zR"),
	"open_detail":                        sharedKeybindingDefinitionWithBindings("enter"),
	"open_link_under_cursor":             sharedKeybindingDefinitionWithBindings("gx"),
	"open_search":                        sharedKeybindingDefinitionWithBindings("/"),
	"page_down":                          sharedKeybindingDefinitionWithBindings("ctrl+d"),
	"page_up":                            sharedKeybindingDefinitionWithBindings("ctrl+u"),
	"place_cursor_at_viewport_bottom":    sharedKeybindingDefinitionWithBindings("zb"),
	"place_cursor_at_viewport_top":       sharedKeybindingDefinitionWithBindings("zt"),
	"place_selection_at_viewport_bottom": sharedKeybindingDefinitionWithBindings("zb"),
	"place_selection_at_viewport_top":    sharedKeybindingDefinitionWithBindings("zt"),
	"previous_search_match":              sharedKeybindingDefinitionWithBindings("N"),
	"previous_tab":                       sharedKeybindingDefinitionWithBindings("["),
	"recenter_cursor":                    sharedKeybindingDefinitionWithBindings("zz"),
	"recenter_selection":                 sharedKeybindingDefinitionWithBindings("zz"),
}

type sharedKeybindingDefinition struct {
	bindings       []string
	allowSequences bool
}

func sharedKeybindingDefinitionWithBindings(bindings ...string) sharedKeybindingDefinition {
	return sharedKeybindingDefinition{bindings: append([]string(nil), bindings...), allowSequences: true}
}

func sharedKeybindingDefinitionFor(action string) (sharedKeybindingDefinition, bool) {
	definition, ok := sharedKeybindingDefinitions[action]
	if !ok {
		return sharedKeybindingDefinition{}, false
	}
	return sharedKeybindingDefinition{
		bindings:       append([]string(nil), definition.bindings...),
		allowSequences: definition.allowSequences,
	}, true
}

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

func sharedKeybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error) keybindingAction {
	definition, ok := sharedKeybindingDefinitionFor(action)
	if !ok {
		panic("missing shared keybinding definition for action " + action)
	}
	if definition.allowSequences {
		return keybindingActionFor(scope, action, viewNames, handler, definition.bindings...)
	}
	return keybindingActionForDirectOnly(scope, action, viewNames, handler, definition.bindings...)
}

func closeKeybindingActionFor(scope string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, includeQuitKey bool) keybindingAction {
	bindings := []string{"esc", "ctrl+["}
	if includeQuitKey {
		bindings = append(bindings, "q")
	}
	return keybindingActionFor(scope, "close", viewNames, handler, bindings...)
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
		sharedKeybindingActionFor(keymapScopeMain, "open_search", mainPaneViewNames, program.openSearch),
		sharedKeybindingActionFor(keymapScopeMain, "move_selection_down", mainPaneViewNames, program.moveSelectionDown),
		sharedKeybindingActionFor(keymapScopeMain, "move_selection_up", mainPaneViewNames, program.moveSelectionUp),
		keybindingActionFor(keymapScopeMain, "move_detail_view_down", mainPaneViewNames, program.moveDetailViewDown, "J"),
		keybindingActionFor(keymapScopeMain, "move_detail_view_up", mainPaneViewNames, program.moveDetailViewUp, "K"),
		sharedKeybindingActionFor(keymapScopeMain, "page_down", mainPaneViewNames, program.pageDown),
		sharedKeybindingActionFor(keymapScopeMain, "page_up", mainPaneViewNames, program.pageUp),
		sharedKeybindingActionFor(keymapScopeMain, "full_page_down", mainPaneViewNames, program.fullPageDown),
		sharedKeybindingActionFor(keymapScopeMain, "full_page_up", mainPaneViewNames, program.fullPageUp),
		keybindingActionFor(keymapScopeMain, "grow_focused_pane", mainPaneViewNames, program.growFocusedPane, "+"),
		keybindingActionFor(keymapScopeMain, "shrink_focused_pane", mainPaneViewNames, program.shrinkFocusedPane, "-"),

		keybindingActionFor(keymapScopeSide, "next_side_view", sidePaneViewNames, program.nextSideView, "l"),
		keybindingActionFor(keymapScopeSide, "previous_side_view", sidePaneViewNames, program.previousSideView, "h"),
		keybindingActionFor(keymapScopeSide, "focus_detail_view", sidePaneViewNames, program.focusDetailView, "0"),
		sharedKeybindingActionFor(keymapScopeSide, "move_selection_to_top", sidePaneViewNames, program.moveSideSelectionToTop),
		sharedKeybindingActionFor(keymapScopeSide, "move_selection_to_bottom", sidePaneViewNames, program.moveSideSelectionToBottom),
		sharedKeybindingActionFor(keymapScopeSide, "recenter_selection", sidePaneViewNames, program.recenterSideSelection),
		sharedKeybindingActionFor(keymapScopeSide, "place_selection_at_viewport_top", sidePaneViewNames, program.moveSideSelectionToViewportTop),
		sharedKeybindingActionFor(keymapScopeSide, "place_selection_at_viewport_bottom", sidePaneViewNames, program.moveSideSelectionToViewportBottom),
		keybindingActionFor(keymapScopeSide, "exit_review_mode", sidePaneViewNames, program.exitReviewMode, "esc", "ctrl+[", "q"),

		sharedKeybindingActionFor(keymapScopeUser, "open_detail", []string{viewUserName}, program.openDetail),
		sharedKeybindingActionFor(keymapScopeUser, "copy_pull_request_url", []string{viewUserName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopeUser, "open_actions_popup", []string{viewUserName}, program.openActionsPopup),

		sharedKeybindingActionFor(keymapScopePullRequests, "previous_tab", []string{viewPullRequestsName}, program.previousPullRequestTab),
		sharedKeybindingActionFor(keymapScopePullRequests, "next_tab", []string{viewPullRequestsName}, program.nextPullRequestTab),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_detail", []string{viewPullRequestsName}, program.openDetail),
		sharedKeybindingActionFor(keymapScopePullRequests, "copy_pull_request_url", []string{viewPullRequestsName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopePullRequests, "comment_on_pull_request", []string{viewPullRequestsName}, program.openPullRequestCommentComposer),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_actions_popup", []string{viewPullRequestsName}, program.openActionsPopup),
		keybindingActionFor(keymapScopePullRequests, "toggle_fold", []string{viewPullRequestsName}, program.togglePullRequestFold, "za"),
		sharedKeybindingActionFor(keymapScopePullRequests, "close_all_folds", []string{viewPullRequestsName}, program.closeAllReviewTreeFolds),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_all_folds", []string{viewPullRequestsName}, program.openAllReviewTreeFolds),
		sharedKeybindingActionFor(keymapScopePullRequests, "next_search_match", []string{viewPullRequestsName}, program.nextReviewFileTreeSearchMatch),
		sharedKeybindingActionFor(keymapScopePullRequests, "previous_search_match", []string{viewPullRequestsName}, program.previousReviewFileTreeSearchMatch),

		sharedKeybindingActionFor(keymapScopeNotifications, "open_detail", []string{viewNotificationsName}, program.openDetail),
		keybindingActionFor(keymapScopeNotifications, "mark_notification_read", []string{viewNotificationsName}, program.markNotificationRead, "r"),
		keybindingActionFor(keymapScopeNotifications, "mark_notification_done", []string{viewNotificationsName}, program.markNotificationDone, "d"),
		sharedKeybindingActionFor(keymapScopeNotifications, "open_actions_popup", []string{viewNotificationsName}, program.openActionsPopup),

		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_left", []string{viewDetailName}, program.moveDetailCursorLeft),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_right", []string{viewDetailName}, program.moveDetailCursorRight),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_row_start", []string{viewDetailName}, program.moveDetailCursorToRowStart),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_row_end", []string{viewDetailName}, program.moveDetailCursorToRowEnd),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_top", []string{viewDetailName}, program.moveDetailCursorToTop),
		sharedKeybindingActionFor(keymapScopeDetail, "open_link_under_cursor", []string{viewDetailName}, program.openLinkUnderCursor),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_bottom", []string{viewDetailName}, program.moveDetailCursorToBottom),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_next_word", []string{viewDetailName}, program.moveDetailCursorToNextWord),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_word_end", []string{viewDetailName}, program.moveDetailCursorToWordEnd),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_previous_word", []string{viewDetailName}, program.moveDetailCursorToPreviousWord),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_next_big_word", []string{viewDetailName}, program.moveDetailCursorToNextBigWord),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_big_word_end", []string{viewDetailName}, program.moveDetailCursorToBigWordEnd),
		sharedKeybindingActionFor(keymapScopeDetail, "move_cursor_to_previous_big_word", []string{viewDetailName}, program.moveDetailCursorToPreviousBigWord),
		sharedKeybindingActionFor(keymapScopeDetail, "next_search_match", []string{viewDetailName}, program.nextDetailSearchMatch),
		sharedKeybindingActionFor(keymapScopeDetail, "previous_search_match", []string{viewDetailName}, program.previousDetailSearchMatch),
		sharedKeybindingActionFor(keymapScopeDetail, "enter_visual_mode", []string{viewDetailName}, program.enterDetailVisualMode),
		sharedKeybindingActionFor(keymapScopeDetail, "enter_line_visual_mode", []string{viewDetailName}, program.enterDetailLineVisualMode),
		sharedKeybindingActionFor(keymapScopeDetail, "previous_tab", []string{viewDetailName}, program.previousDetailTab),
		sharedKeybindingActionFor(keymapScopeDetail, "next_tab", []string{viewDetailName}, program.nextDetailTab),
		sharedKeybindingActionFor(keymapScopeDetail, "copy_pull_request_url", []string{viewDetailName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopeDetail, "comment_on_pull_request", []string{viewDetailName}, program.openPullRequestCommentComposer),
		sharedKeybindingActionFor(keymapScopeDetail, "open_actions_popup", []string{viewDetailName}, program.openActionsPopup),
		sharedKeybindingActionFor(keymapScopeDetail, "recenter_cursor", []string{viewDetailName}, program.recenterDetailView),
		sharedKeybindingActionFor(keymapScopeDetail, "place_cursor_at_viewport_top", []string{viewDetailName}, program.moveDetailCursorToViewportTop),
		sharedKeybindingActionFor(keymapScopeDetail, "place_cursor_at_viewport_bottom", []string{viewDetailName}, program.moveDetailCursorToViewportBottom),
		keybindingActionFor(keymapScopeDetail, "toggle_inline_conversation", []string{viewDetailName}, program.toggleInlineConversationVisibility, "enter", "za"),
		sharedKeybindingActionFor(keymapScopeDetail, "close_all_folds", []string{viewDetailName}, program.closeAllDetailFolds),
		sharedKeybindingActionFor(keymapScopeDetail, "open_all_folds", []string{viewDetailName}, program.openAllDetailFolds),
		closeKeybindingActionFor(keymapScopeDetail, []string{viewDetailName}, program.closeDetail, true),

		keybindingActionFor(keymapScopeSearch, "submit", []string{viewSearchName}, program.submitSearch, "enter", "ctrl+j", "ctrl+s"),
		keybindingActionFor(keymapScopeSearch, "cancel", []string{viewSearchName}, program.cancelSearch, "esc", "ctrl+["),

		keybindingActionFor(keymapScopeActionsPopup, "focus_search", []string{viewActionsPopupName}, program.focusActionsPopupSearch, "/"),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "move_selection_down", []string{viewActionsPopupName}, program.moveActionsPopupSelectionDown),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "move_selection_up", []string{viewActionsPopupName}, program.moveActionsPopupSelectionUp),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "page_down", []string{viewActionsPopupName}, program.pageActionsPopupDown),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "page_up", []string{viewActionsPopupName}, program.pageActionsPopupUp),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "full_page_down", []string{viewActionsPopupName}, program.fullPageActionsPopupDown),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "full_page_up", []string{viewActionsPopupName}, program.fullPageActionsPopupUp),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "move_selection_to_top", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToTop),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "move_selection_to_bottom", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToBottom),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "recenter_selection", []string{viewActionsPopupName}, program.recenterActionsPopupSelection),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "place_selection_at_viewport_top", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToViewportTop),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "place_selection_at_viewport_bottom", []string{viewActionsPopupName}, program.moveActionsPopupSelectionToViewportBottom),
		keybindingActionFor(keymapScopeActionsPopup, "execute_selected_action", []string{viewActionsPopupName}, program.executeSelectedActionsPopupAction, "enter"),
		keybindingActionFor(keymapScopeActionsPopup, "submit_selected_picker", []string{viewActionsPopupName}, program.submitSelectedActionsPopupAction, "alt+enter"),
		closeKeybindingActionFor(keymapScopeActionsPopup, []string{viewActionsPopupName}, program.closeActionsPopup, true),

		keybindingActionFor(keymapScopeActionsPopupSearch, "focus_list", []string{viewActionsPopupSearchName}, program.focusActionsPopupList, "enter", "tab", "ctrl+s"),
		closeKeybindingActionFor(keymapScopeActionsPopupSearch, []string{viewActionsPopupSearchName}, program.closeActionsPopup, false),

		keybindingActionFor(keymapScopeModalEditor, "submit", []string{viewModalEditorName}, program.submitModalEditor, "alt+enter", "ctrl+s"),
		closeKeybindingActionFor(keymapScopeModalEditor, []string{viewModalEditorName}, program.closeModalEditor, false),

		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_left", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorLeft),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_down", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorDown, "j", "down"),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_up", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorUp, "k", "up"),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_right", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorRight),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_row_start", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToRowStart),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_row_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToRowEnd),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_top", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToTop),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "open_link_under_cursor", []string{viewPullRequestBuildInfoName}, program.openPullRequestBuildRunPopupLinkUnderCursor),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_bottom", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToBottom),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_next_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToNextWord),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_word_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToWordEnd),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_previous_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToPreviousWord),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_next_big_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToNextBigWord),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_big_word_end", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToBigWordEnd),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_to_previous_big_word", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorToPreviousBigWord),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "next_search_match", []string{viewPullRequestBuildInfoName}, program.nextPullRequestBuildRunPopupSearchMatch),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "previous_search_match", []string{viewPullRequestBuildInfoName}, program.previousPullRequestBuildRunPopupSearchMatch),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "enter_visual_mode", []string{viewPullRequestBuildInfoName}, program.enterPullRequestBuildRunPopupVisualMode),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "enter_line_visual_mode", []string{viewPullRequestBuildInfoName}, program.enterPullRequestBuildRunPopupLineVisualMode),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "open_search", []string{viewPullRequestBuildInfoName}, program.openSearch),
		keybindingActionFor(keymapScopePullRequestBuildInfo, "copy_content", []string{viewPullRequestBuildInfoName}, program.copyPullRequestBuildRunPopupContent, "y"),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "open_actions_popup", []string{viewPullRequestBuildInfoName}, program.openActionsPopup),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "page_down", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupDown),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "page_up", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupUp),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_down", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupDown),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_up", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupUp),
		closeKeybindingActionFor(keymapScopePullRequestBuildInfo, []string{viewPullRequestBuildInfoName}, program.closePullRequestBuildRunPopup, true),

		sharedKeybindingActionFor(keymapScopeHelp, "full_page_down", []string{viewHelpName}, program.fullPageHelpDown),
		sharedKeybindingActionFor(keymapScopeHelp, "full_page_up", []string{viewHelpName}, program.fullPageHelpUp),
		closeKeybindingActionFor(keymapScopeHelp, []string{viewHelpName}, program.closeHelp, true),
	}
}
