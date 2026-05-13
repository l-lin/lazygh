package tui

import (
	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
)

const (
	keymapScopePrefix               = "_prefix"
	keymapScopeGlobal               = "global"
	keymapScopeMain                 = "main"
	keymapScopeSide                 = "side"
	keymapScopeUser                 = "user"
	keymapScopePullRequests         = "pull_requests"
	keymapScopeNotifications        = "notifications"
	keymapScopeSelection            = "selection"
	keymapScopeCursor               = "cursor"
	keymapScopeFolds                = "folds"
	keymapScopeReview               = "review"
	keymapScopeSearch               = "search"
	keymapScopeActionsPopup         = "actions_popup"
	keymapScopeModalEditor          = "modal_editor"
	keymapScopePullRequestBuildInfo = "pull_request_build_info"
)

var mainPaneViewNames = []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewDetailName}
var sidePaneViewNames = []string{viewUserName, viewPullRequestsName, viewNotificationsName}
var reviewPaneViewNames = []string{viewPullRequestsName, viewDetailName}

var sharedKeybindingDefinitions = map[string]sharedKeybindingDefinition{
	"toggle_help":                        sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "toggle_help"),
	"open_search":                        sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "open_search"),
	"move_selection_down":                sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "move_selection_down"),
	"move_selection_up":                  sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "move_selection_up"),
	"page_down":                          sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "page_down"),
	"page_up":                            sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "page_up"),
	"full_page_down":                     sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "full_page_down"),
	"full_page_up":                       sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "full_page_up"),
	"grow_focused_pane":                  sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "grow_focused_pane"),
	"shrink_focused_pane":                sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "shrink_focused_pane"),
	"open_actions_popup":                 sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "open_actions_popup"),
	"open_detail":                        sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSide, "open_detail"),
	"move_selection_to_top":              sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSelection, "move_selection_to_top"),
	"move_selection_to_bottom":           sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSelection, "move_selection_to_bottom"),
	"place_selection_at_viewport_top":    sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSelection, "place_selection_at_viewport_top"),
	"recenter_selection":                 sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSelection, "recenter_selection"),
	"place_selection_at_viewport_bottom": sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSelection, "place_selection_at_viewport_bottom"),
	"move_cursor_left":                   sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_left"),
	"move_cursor_right":                  sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_right"),
	"move_cursor_to_row_start":           sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_row_start"),
	"move_cursor_to_row_end":             sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_row_end"),
	"move_cursor_to_top":                 sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_top"),
	"open_link_under_cursor":             sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "open_link_under_cursor"),
	"move_cursor_to_bottom":              sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_bottom"),
	"move_cursor_to_next_word":           sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_next_word"),
	"move_cursor_to_word_end":            sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_word_end"),
	"move_cursor_to_previous_word":       sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_previous_word"),
	"move_cursor_to_next_big_word":       sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_next_big_word"),
	"move_cursor_to_big_word_end":        sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_big_word_end"),
	"move_cursor_to_previous_big_word":   sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "move_cursor_to_previous_big_word"),
	"enter_visual_mode":                  sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "enter_visual_mode"),
	"enter_line_visual_mode":             sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "enter_line_visual_mode"),
	"recenter_cursor":                    sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "recenter_cursor"),
	"place_cursor_at_viewport_top":       sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "place_cursor_at_viewport_top"),
	"place_cursor_at_viewport_bottom":    sharedKeybindingDefinitionWithDefaultBindings(keymapScopeCursor, "place_cursor_at_viewport_bottom"),
	"copy_pull_request_url":              sharedKeybindingDefinitionWithDefaultBindings(keymapScopePullRequests, "copy_pull_request_url"),
	"comment_on_pull_request":            sharedKeybindingDefinitionWithDefaultBindings(keymapScopePullRequests, "comment_on_pull_request"),
	"previous_tab":                       sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "previous_tab"),
	"next_tab":                           sharedKeybindingDefinitionWithDefaultBindings(keymapScopeGlobal, "next_tab"),
	"toggle_fold":                        sharedKeybindingDefinitionWithDefaultBindings(keymapScopeFolds, "toggle_fold"),
	"close_all_folds":                    sharedKeybindingDefinitionWithDefaultBindings(keymapScopeFolds, "close_all_folds"),
	"open_all_folds":                     sharedKeybindingDefinitionWithDefaultBindings(keymapScopeFolds, "open_all_folds"),
	"next_search_match":                  sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSearch, "next_search_match"),
	"previous_search_match":              sharedKeybindingDefinitionWithDefaultBindings(keymapScopeSearch, "previous_search_match"),
}

type sharedKeybindingDefinition struct {
	scope          string
	bindings       []string
	allowSequences bool
}

func sharedKeybindingDefinitionWithBindings(scope string, bindings ...string) sharedKeybindingDefinition {
	return sharedKeybindingDefinition{scope: scope, bindings: append([]string(nil), bindings...), allowSequences: true}
}

func sharedKeybindingDefinitionWithDefaultBindings(scope string, action string) sharedKeybindingDefinition {
	return sharedKeybindingDefinitionWithBindings(scope, mustDefaultKeymapBindings(scope, action)...)
}

func sharedKeybindingDefinitionFor(action string) (sharedKeybindingDefinition, bool) {
	definition, ok := sharedKeybindingDefinitions[action]
	if !ok {
		return sharedKeybindingDefinition{}, false
	}
	return sharedKeybindingDefinition{
		scope:          definition.scope,
		bindings:       append([]string(nil), definition.bindings...),
		allowSequences: definition.allowSequences,
	}, true
}

func mustDefaultKeymapBindings(scope string, action string) []string {
	actions, ok := appconfig.DefaultKeymaps()[scope]
	if !ok {
		panic("missing default keymap scope " + scope)
	}
	bindings, ok := actions[action]
	if !ok {
		panic("missing default keymap action " + scope + "." + action)
	}
	return append([]string(nil), bindings...)
}

func configuredKeybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error) keybindingAction {
	return keybindingActionFor(scope, action, viewNames, handler, mustDefaultKeymapBindings(scope, action)...)
}

func aliasedConfiguredKeybindingActionFor(scope string, action string, configScope string, configAction string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error) keybindingAction {
	return aliasedKeybindingActionFor(scope, action, configScope, configAction, viewNames, handler, mustDefaultKeymapBindings(configScope, configAction)...)
}

func keybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...string) keybindingAction {
	id := keybindingActionID{scope: scope, action: action}
	return keybindingAction{
		id:              id,
		configID:        id,
		configurable:    true,
		viewNames:       append([]string(nil), viewNames...),
		defaultBindings: mustConfiguredKeySequences(bindings...),
		handler:         handler,
		allowSequences:  true,
		bindingSlice:    keybindingBindingSliceAll,
	}
}

func fixedKeybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...string) keybindingAction {
	definition := keybindingActionFor(scope, action, viewNames, handler, bindings...)
	definition.configurable = false
	definition.configID = keybindingActionID{}
	return definition
}

func keybindingActionWithConfigID(definition keybindingAction, scope string, action string) keybindingAction {
	definition.configID = keybindingActionID{scope: scope, action: action}
	return definition
}

func keybindingActionWithBindingSlice(definition keybindingAction, bindingSlice keybindingBindingSlice) keybindingAction {
	definition.bindingSlice = bindingSlice
	return definition
}

func sharedKeybindingActionFor(scope string, action string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error) keybindingAction {
	definition, ok := sharedKeybindingDefinitionFor(action)
	if !ok {
		panic("missing shared keybinding definition for action " + action)
	}
	keybinding := keybindingActionFor(scope, action, viewNames, handler, definition.bindings...)
	keybinding.configID = keybindingActionID{scope: definition.scope, action: action}
	keybinding.allowSequences = definition.allowSequences
	return keybinding
}

func aliasedKeybindingActionFor(scope string, action string, configScope string, configAction string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error, bindings ...string) keybindingAction {
	return keybindingActionWithConfigID(keybindingActionFor(scope, action, viewNames, handler, bindings...), configScope, configAction)
}

func closeKeybindingActionFor(scope string, viewNames []string, handler func(*gocui.Gui, *gocui.View) error) keybindingAction {
	return keybindingActionWithConfigID(keybindingActionFor(scope, "close", viewNames, handler, mustDefaultKeymapBindings(keymapScopeGlobal, "close")...), keymapScopeGlobal, "close")
}

func (program *Program) keybindingActions() []keybindingAction {
	return []keybindingAction{
		configuredKeybindingActionFor(keymapScopeGlobal, "quit", []string{""}, program.quit),
		keybindingActionWithBindingSlice(configuredKeybindingActionFor(keymapScopeGlobal, "next_side_view", []string{""}, program.nextSideView), keybindingBindingSliceFirst),
		keybindingActionWithBindingSlice(configuredKeybindingActionFor(keymapScopeGlobal, "previous_side_view", []string{""}, program.previousSideView), keybindingBindingSliceFirst),

		sharedKeybindingActionFor(keymapScopeMain, "toggle_help", mainPaneViewNames, program.toggleHelp),
		fixedKeybindingActionFor(keymapScopeMain, "focus_user_view", mainPaneViewNames, program.focusUserView, "1"),
		fixedKeybindingActionFor(keymapScopeMain, "focus_pull_requests_view", mainPaneViewNames, program.focusPullRequestsView, "2"),
		fixedKeybindingActionFor(keymapScopeMain, "focus_notifications_view", mainPaneViewNames, program.focusNotificationsView, "3"),
		sharedKeybindingActionFor(keymapScopeMain, "open_search", mainPaneViewNames, program.openSearch),
		sharedKeybindingActionFor(keymapScopeMain, "move_selection_down", mainPaneViewNames, program.moveSelectionDown),
		sharedKeybindingActionFor(keymapScopeMain, "move_selection_up", mainPaneViewNames, program.moveSelectionUp),
		configuredKeybindingActionFor(keymapScopeMain, "move_detail_view_down", mainPaneViewNames, program.moveDetailViewDown),
		configuredKeybindingActionFor(keymapScopeMain, "move_detail_view_up", mainPaneViewNames, program.moveDetailViewUp),
		sharedKeybindingActionFor(keymapScopeMain, "page_down", mainPaneViewNames, program.pageDown),
		sharedKeybindingActionFor(keymapScopeMain, "page_up", mainPaneViewNames, program.pageUp),
		sharedKeybindingActionFor(keymapScopeMain, "full_page_down", mainPaneViewNames, program.fullPageDown),
		sharedKeybindingActionFor(keymapScopeMain, "full_page_up", mainPaneViewNames, program.fullPageUp),
		sharedKeybindingActionFor(keymapScopeMain, "grow_focused_pane", mainPaneViewNames, program.growFocusedPane),
		sharedKeybindingActionFor(keymapScopeMain, "shrink_focused_pane", mainPaneViewNames, program.shrinkFocusedPane),

		keybindingActionWithBindingSlice(aliasedConfiguredKeybindingActionFor(keymapScopeSide, "next_side_view", keymapScopeGlobal, "next_side_view", sidePaneViewNames, program.nextSideView), keybindingBindingSliceRest),
		keybindingActionWithBindingSlice(aliasedConfiguredKeybindingActionFor(keymapScopeSide, "previous_side_view", keymapScopeGlobal, "previous_side_view", sidePaneViewNames, program.previousSideView), keybindingBindingSliceRest),
		fixedKeybindingActionFor(keymapScopeSide, "focus_detail_view", sidePaneViewNames, program.focusDetailView, "0"),
		sharedKeybindingActionFor(keymapScopeSide, "move_selection_to_top", sidePaneViewNames, program.moveSideSelectionToTop),
		sharedKeybindingActionFor(keymapScopeSide, "move_selection_to_bottom", sidePaneViewNames, program.moveSideSelectionToBottom),
		sharedKeybindingActionFor(keymapScopeSide, "recenter_selection", sidePaneViewNames, program.recenterSideSelection),
		sharedKeybindingActionFor(keymapScopeSide, "place_selection_at_viewport_top", sidePaneViewNames, program.moveSideSelectionToViewportTop),
		sharedKeybindingActionFor(keymapScopeSide, "place_selection_at_viewport_bottom", sidePaneViewNames, program.moveSideSelectionToViewportBottom),
		configuredKeybindingActionFor(keymapScopeSide, "exit_review_mode", sidePaneViewNames, program.exitReviewMode),

		sharedKeybindingActionFor(keymapScopeUser, "open_detail", []string{viewUserName}, program.openDetail),
		sharedKeybindingActionFor(keymapScopeUser, "copy_pull_request_url", []string{viewUserName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopeUser, "open_actions_popup", []string{viewUserName}, program.openActionsPopup),

		sharedKeybindingActionFor(keymapScopeGlobal, "previous_tab", []string{viewPullRequestsName}, program.previousPullRequestTab),
		sharedKeybindingActionFor(keymapScopeGlobal, "next_tab", []string{viewPullRequestsName}, program.nextPullRequestTab),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_detail", []string{viewPullRequestsName}, program.openDetail),
		sharedKeybindingActionFor(keymapScopePullRequests, "copy_pull_request_url", []string{viewPullRequestsName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopePullRequests, "comment_on_pull_request", []string{viewPullRequestsName}, program.openPullRequestCommentComposer),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_actions_popup", []string{viewPullRequestsName}, program.openActionsPopup),
		sharedKeybindingActionFor(keymapScopePullRequests, "toggle_fold", []string{viewPullRequestsName}, program.togglePullRequestFold),
		sharedKeybindingActionFor(keymapScopePullRequests, "close_all_folds", []string{viewPullRequestsName}, program.closeAllReviewTreeFolds),
		sharedKeybindingActionFor(keymapScopePullRequests, "open_all_folds", []string{viewPullRequestsName}, program.openAllReviewTreeFolds),
		sharedKeybindingActionFor(keymapScopePullRequests, "next_search_match", []string{viewPullRequestsName}, program.nextPullRequestsSearchMatch),
		sharedKeybindingActionFor(keymapScopePullRequests, "previous_search_match", []string{viewPullRequestsName}, program.previousPullRequestsSearchMatch),

		sharedKeybindingActionFor(keymapScopeNotifications, "open_detail", []string{viewNotificationsName}, program.openDetail),
		configuredKeybindingActionFor(keymapScopeNotifications, "mark_notification_read", []string{viewNotificationsName}, program.markNotificationRead),
		configuredKeybindingActionFor(keymapScopeNotifications, "mark_notification_done", []string{viewNotificationsName}, program.markNotificationDone),
		sharedKeybindingActionFor(keymapScopeNotifications, "open_actions_popup", []string{viewNotificationsName}, program.openActionsPopup),

		configuredKeybindingActionFor(keymapScopeReview, "previous_file", reviewPaneViewNames, program.previousReviewFile),
		configuredKeybindingActionFor(keymapScopeReview, "next_file", reviewPaneViewNames, program.nextReviewFile),
		configuredKeybindingActionFor(keymapScopeReview, "previous_comment", reviewPaneViewNames, program.previousReviewComment),
		configuredKeybindingActionFor(keymapScopeReview, "next_comment", reviewPaneViewNames, program.nextReviewComment),

		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_left", []string{viewDetailName}, program.moveDetailCursorLeft),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_right", []string{viewDetailName}, program.moveDetailCursorRight),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_row_start", []string{viewDetailName}, program.moveDetailCursorToRowStart),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_row_end", []string{viewDetailName}, program.moveDetailCursorToRowEnd),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_top", []string{viewDetailName}, program.moveDetailCursorToTop),
		sharedKeybindingActionFor(keymapScopeCursor, "open_link_under_cursor", []string{viewDetailName}, program.openLinkUnderCursor),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_bottom", []string{viewDetailName}, program.moveDetailCursorToBottom),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_next_word", []string{viewDetailName}, program.moveDetailCursorToNextWord),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_word_end", []string{viewDetailName}, program.moveDetailCursorToWordEnd),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_previous_word", []string{viewDetailName}, program.moveDetailCursorToPreviousWord),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_next_big_word", []string{viewDetailName}, program.moveDetailCursorToNextBigWord),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_big_word_end", []string{viewDetailName}, program.moveDetailCursorToBigWordEnd),
		sharedKeybindingActionFor(keymapScopeCursor, "move_cursor_to_previous_big_word", []string{viewDetailName}, program.moveDetailCursorToPreviousBigWord),
		sharedKeybindingActionFor(keymapScopeSearch, "next_search_match", []string{viewDetailName}, program.nextDetailSearchMatch),
		sharedKeybindingActionFor(keymapScopeSearch, "previous_search_match", []string{viewDetailName}, program.previousDetailSearchMatch),
		sharedKeybindingActionFor(keymapScopeCursor, "enter_visual_mode", []string{viewDetailName}, program.enterDetailVisualMode),
		sharedKeybindingActionFor(keymapScopeCursor, "enter_line_visual_mode", []string{viewDetailName}, program.enterDetailLineVisualMode),
		sharedKeybindingActionFor(keymapScopeGlobal, "previous_tab", []string{viewDetailName}, program.previousDetailTab),
		sharedKeybindingActionFor(keymapScopeGlobal, "next_tab", []string{viewDetailName}, program.nextDetailTab),
		sharedKeybindingActionFor(keymapScopePullRequests, "copy_pull_request_url", []string{viewDetailName}, program.copyPullRequestURL),
		sharedKeybindingActionFor(keymapScopePullRequests, "comment_on_pull_request", []string{viewDetailName}, program.openDetailPullRequestCommentShortcut),
		configuredKeybindingActionFor(keymapScopePullRequests, "reply_to_inline_comment", []string{viewDetailName}, program.replyToInlineCommentShortcut),
		sharedKeybindingActionFor(keymapScopeGlobal, "open_actions_popup", []string{viewDetailName}, program.openActionsPopup),
		sharedKeybindingActionFor(keymapScopeCursor, "recenter_cursor", []string{viewDetailName}, program.recenterDetailView),
		sharedKeybindingActionFor(keymapScopeCursor, "place_cursor_at_viewport_top", []string{viewDetailName}, program.moveDetailCursorToViewportTop),
		sharedKeybindingActionFor(keymapScopeCursor, "place_cursor_at_viewport_bottom", []string{viewDetailName}, program.moveDetailCursorToViewportBottom),
		fixedKeybindingActionFor(keymapScopeFolds, "toggle_inline_conversation", []string{viewDetailName}, program.toggleInlineConversationVisibility, "enter"),
		sharedKeybindingActionFor(keymapScopeFolds, "toggle_fold", []string{viewDetailName}, program.toggleInlineConversationVisibility),
		sharedKeybindingActionFor(keymapScopeFolds, "close_all_folds", []string{viewDetailName}, program.closeAllDetailFolds),
		sharedKeybindingActionFor(keymapScopeFolds, "open_all_folds", []string{viewDetailName}, program.openAllDetailFolds),
		closeKeybindingActionFor(keymapScopeGlobal, []string{viewDetailName}, program.closeDetail),

		configuredKeybindingActionFor(keymapScopeSearch, "submit", []string{viewSearchName}, program.submitSearch),
		configuredKeybindingActionFor(keymapScopeSearch, "cancel", []string{viewSearchName}, program.cancelSearch),

		sharedKeybindingActionFor(keymapScopeActionsPopup, "open_search", []string{viewActionsPopupName}, program.focusActionsPopupSearch),
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
		sharedKeybindingActionFor(keymapScopeActionsPopup, "next_search_match", []string{viewActionsPopupName}, program.nextActionsPopupSearchMatch),
		sharedKeybindingActionFor(keymapScopeActionsPopup, "previous_search_match", []string{viewActionsPopupName}, program.previousActionsPopupSearchMatch),
		configuredKeybindingActionFor(keymapScopeActionsPopup, "execute_selected_action", []string{viewActionsPopupName}, program.executeSelectedActionsPopupAction),
		configuredKeybindingActionFor(keymapScopeActionsPopup, "submit_selected_picker", []string{viewActionsPopupName}, program.submitSelectedActionsPopupAction),
		closeKeybindingActionFor(keymapScopeActionsPopup, []string{viewActionsPopupName}, program.closeActionsPopup),

		aliasedConfiguredKeybindingActionFor(keymapScopeSearch, "submit", keymapScopeSearch, "submit", []string{viewActionsPopupSearchName}, program.focusActionsPopupList),
		configuredKeybindingActionFor(keymapScopeSearch, "cancel", []string{viewActionsPopupSearchName}, program.closeActionsPopup),

		configuredKeybindingActionFor(keymapScopeModalEditor, "submit", []string{viewModalEditorName}, program.submitModalEditor),
		configuredKeybindingActionFor(keymapScopeModalEditor, "open_external_editor", []string{viewModalEditorName}, program.openModalEditorInExternalEditor),
		configuredKeybindingActionFor(keymapScopeModalEditor, "cancel", []string{viewModalEditorName}, program.closeModalEditor),

		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_left", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorLeft),
		aliasedConfiguredKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_down", keymapScopeGlobal, "move_selection_down", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorDown),
		aliasedConfiguredKeybindingActionFor(keymapScopePullRequestBuildInfo, "move_cursor_up", keymapScopeGlobal, "move_selection_up", []string{viewPullRequestBuildInfoName}, program.movePullRequestBuildRunPopupCursorUp),
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
		configuredKeybindingActionFor(keymapScopePullRequestBuildInfo, "copy_content", []string{viewPullRequestBuildInfoName}, program.copyPullRequestBuildRunPopupContent),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "open_actions_popup", []string{viewPullRequestBuildInfoName}, program.openActionsPopup),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "page_down", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupDown),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "page_up", []string{viewPullRequestBuildInfoName}, program.pagePullRequestBuildRunPopupUp),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_down", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupDown),
		sharedKeybindingActionFor(keymapScopePullRequestBuildInfo, "full_page_up", []string{viewPullRequestBuildInfoName}, program.fullPagePullRequestBuildRunPopupUp),
		closeKeybindingActionFor(keymapScopePullRequestBuildInfo, []string{viewPullRequestBuildInfoName}, program.closePullRequestBuildRunPopup),

		sharedKeybindingActionFor(keymapScopeGlobal, "full_page_down", []string{viewHelpName}, program.fullPageHelpDown),
		sharedKeybindingActionFor(keymapScopeGlobal, "full_page_up", []string{viewHelpName}, program.fullPageHelpUp),
		closeKeybindingActionFor(keymapScopeGlobal, []string{viewHelpName}, program.closeHelp),
	}
}
