package tui

import (
	"strings"
	"unicode/utf8"
)

type helpPresenter struct {
	actionContext                    ActionContext
	keyResolver                      keybindingLabelResolver
	inlineCommentReplyAvailable      bool
	inlineCommentResolutionHelpLabel string
	pullRequestBrowserAvailable      bool
}

func (presenter helpPresenter) sections() []helpSection {
	return []helpSection{
		{Title: "Local", Entries: presenter.localHelpEntries()},
		{Title: "Global", Entries: presenter.globalHelpEntries()},
	}
}

func (presenter helpPresenter) localHelpEntries() []helpEntry {
	actionContext := presenter.actionContext
	if actionContext.IsReviewContext() {
		switch actionContext.ActiveView.Focus {
		case FocusDetailView:
			entries := []helpEntry{
				{Key: "h/j/k/l/<up>/<down>/<left>/<right>", Description: "Move cursor"},
				{Key: "0/$", Description: "Line start/end"},
				{Key: "gg/G", Description: "First/last line"},
				{Key: presenter.reviewFileMotionHelpKeys(), Description: "Previous/next file"},
				{Key: presenter.reviewCommentMotionHelpKeys(), Description: "Previous/next comment"},
				{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_top"}, keybindingActionID{scope: keymapScopeCursor, action: "recenter_cursor"}, keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_bottom"}), Description: "Cursor to top/center/bottom"},
				{Key: presenter.wordMotionHelpKeys(keymapScopeCursor), Description: "Next/end/previous word/WORD"},
				presenter.characterMotionHelpEntry(),
				presenter.repeatCharacterMotionHelpEntry(),
				{Key: "n/N", Description: "Next/previous match"},
				{Key: presenter.helpKeysOrFallback("<c-w>", keybindingActionID{scope: keymapScopeCursor, action: "toggle_word_wrap"}), Description: detailWordWrapHelpLabel},
				{Key: "v/V", Description: "Start char/line visual selection"},
				{Key: presenter.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopeCursor, action: "start_yank"}), Description: "Start yank with motion"},
				pullRequestBrowserHelpEntry(presenter.keyResolver, keymapScopePullRequests),
				reviewInlineCommentHelpEntry(presenter.keyResolver),
				{Key: presenter.inlineConversationToggleHelpKeys(), Description: "Expand/collapse conversation"},
				{Key: presenter.bulkFoldHelpKeys(), Description: "Close/open all folds"},
				{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}), Description: "Actions"},
				presenter.refreshHelpEntry("Refresh PR"),
				pullRequestYankHelpEntry(presenter.keyResolver, keymapScopePullRequests),
				{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
				{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
				{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
				{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
				{Key: "+/-", Description: "Toggle fullscreen"},
				{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search diff"},
				presenter.searchWordUnderCursorHelpEntry(),
				{Key: "<esc>/q", Description: "Exit visual / return"},
			}
			if presenter.inlineCommentReplyAvailable {
				entries = append(entries, inlineCommentReplyHelpEntry(presenter.keyResolver))
			}
			if presenter.inlineCommentResolutionHelpLabel != "" {
				entries = append(entries, inlineCommentResolutionHelpEntry(presenter.keyResolver, presenter.inlineCommentResolutionHelpLabel))
			}
			return entries
		case FocusPullRequestsView:
			return []helpEntry{
				{Key: "j/k/<up>/<down>", Description: "Move down/up"},
				{Key: "gg/G", Description: "First/last file"},
				{Key: presenter.reviewFileMotionHelpKeys(), Description: "Previous/next file"},
				{Key: presenter.reviewCommentMotionHelpKeys(), Description: "Previous/next comment"},
				{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
				{Key: presenter.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
				{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
				{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
				{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
				{Key: "+/-", Description: "Resize panes"},
				{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search file tree"},
				{Key: presenter.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopePullRequests, action: "next_search_match"}) + "/" + presenter.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopePullRequests, action: "previous_search_match"}), Description: "Next/previous match"},
				pullRequestBrowserHelpEntry(presenter.keyResolver, keymapScopePullRequests),
				pullRequestYankHelpEntry(presenter.keyResolver, keymapScopePullRequests),
				pullRequestCommentHelpEntry(presenter.keyResolver, keymapScopePullRequests),
				{Key: presenter.reviewTreeToggleHelpKeys(), Description: "Expand/collapse fold"},
				{Key: presenter.reviewTreeBulkFoldHelpKeys(), Description: "Close/open all folds"},
				{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "Actions"},
				{Key: presenter.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open diff"},
				{Key: "<esc>/q", Description: "Exit review mode"},
			}
		default:
			return []helpEntry{
				{Key: presenter.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				pullRequestBrowserHelpEntry(presenter.keyResolver, keymapScopeUser),
				pullRequestYankHelpEntry(presenter.keyResolver, keymapScopeUser),
				pullRequestCommentHelpEntry(presenter.keyResolver, keymapScopeUser),
				{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}), Description: "Actions"},
				{Key: presenter.helpKeysOrFallback("0", keybindingActionID{scope: keymapScopeSide, action: "focus_detail_view"}), Description: "Focus diff"},
				{Key: "<esc>/q", Description: "Exit review mode"},
			}
		}
	}

	switch actionContext.ActiveView.Focus {
	case FocusDetailView:
		entries := []helpEntry{
			{Key: "h/j/k/l/<up>/<down>/<left>/<right>", Description: "Move cursor"},
			{Key: "0/$", Description: "Line start/end"},
			{Key: "gg/G", Description: "First/last line"},
			{Key: presenter.helpKeysOrFallback("gx", keybindingActionID{scope: keymapScopeCursor, action: "open_link_under_cursor"}), Description: "Open link under cursor"},
			{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_top"}, keybindingActionID{scope: keymapScopeCursor, action: "recenter_cursor"}, keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_bottom"}), Description: "Cursor to top/center/bottom"},
			{Key: presenter.wordMotionHelpKeys(keymapScopeCursor), Description: "Next/end/previous word/WORD"},
			presenter.characterMotionHelpEntry(),
			presenter.repeatCharacterMotionHelpEntry(),
			{Key: "n/N", Description: "Next/previous match"},
			{Key: presenter.helpKeysOrFallback("<c-w>", keybindingActionID{scope: keymapScopeCursor, action: "toggle_word_wrap"}), Description: detailWordWrapHelpLabel},
			{Key: "v/V", Description: "Start char/line visual selection"},
			{Key: presenter.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopeCursor, action: "start_yank"}), Description: "Start yank with motion"},
			pullRequestYankHelpEntry(presenter.keyResolver, keymapScopePullRequests),
			{Key: presenter.helpKeysOrFallback("<c-v>", keybindingActionID{scope: keymapScopePullRequests, action: "open_pull_request_by_url"}), Description: "Open PR from clipboard"},
			{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Toggle fullscreen"},
			{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search detail"},
			presenter.searchWordUnderCursorHelpEntry(),
			{Key: "<esc>/q", Description: "Exit visual / return"},
		}
		if presenter.pullRequestBrowserAvailable {
			entries = append(entries, pullRequestBrowserHelpEntry(presenter.keyResolver, keymapScopePullRequests))
		}
		entries = append(entries, helpEntry{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}), Description: "Actions"})
		if actionContext.IsPullRequestContext() {
			entries = append(entries, presenter.refreshHelpEntry("Refresh PR"))
		}
		if presenter.showsPullRequestDetailTabs() {
			entries = append(entries, detailPullRequestCommentHelpEntry(presenter.keyResolver, presenter.browserChangesInlineCommentShortcutActive()))
			if presenter.inlineCommentReplyAvailable {
				entries = append(entries, inlineCommentReplyHelpEntry(presenter.keyResolver))
			}
			if presenter.inlineCommentResolutionHelpLabel != "" {
				entries = append(entries, inlineCommentResolutionHelpEntry(presenter.keyResolver, presenter.inlineCommentResolutionHelpLabel))
			}
			entries = append(entries,
				helpEntry{Key: presenter.inlineConversationToggleHelpKeys(), Description: "Expand/collapse section"},
				helpEntry{Key: presenter.bulkFoldHelpKeys(), Description: "Close/open all folds"},
				helpEntry{Key: presenter.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopeGlobal, action: "previous_tab"}), Description: "Previous detail tab"},
				helpEntry{Key: presenter.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopeGlobal, action: "next_tab"}), Description: "Next detail tab"},
			)
		}
		return entries
	case FocusPullRequestsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last pull request"},
			{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: presenter.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search pull requests"},
			{Key: presenter.helpKeysOrFallback(":", keybindingActionID{scope: keymapScopePullRequests, action: "custom_search"}), Description: "Custom search"},
			{Key: presenter.helpKeysOrFallback("<c-v>", keybindingActionID{scope: keymapScopePullRequests, action: "open_pull_request_by_url"}), Description: "Open PR from clipboard"},
			{Key: presenter.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopePullRequests, action: "next_search_match"}) + "/" + presenter.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopePullRequests, action: "previous_search_match"}), Description: "Next/previous match"},
			pullRequestBrowserHelpEntry(presenter.keyResolver, keymapScopePullRequests),
			pullRequestYankHelpEntry(presenter.keyResolver, keymapScopePullRequests),
			pullRequestCommentHelpEntry(presenter.keyResolver, keymapScopePullRequests),
			{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "Actions"},
			presenter.refreshHelpEntry("Refresh PR list"),
			{Key: presenter.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopeGlobal, action: "previous_tab"}), Description: "Previous tab"},
			{Key: presenter.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopeGlobal, action: "next_tab"}), Description: "Next tab"},
			{Key: presenter.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open detail"},
		}
	case FocusNotificationsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last notification"},
			{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: presenter.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search notifications"},
			{Key: presenter.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopeNotifications, action: "next_search_match"}) + "/" + presenter.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopeNotifications, action: "previous_search_match"}), Description: "Next/previous match"},
			{Key: presenter.helpKeysOrFallback("r", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_read"}), Description: "Mark notification as read"},
			{Key: presenter.helpKeysOrFallback("d", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_done"}), Description: "Mark notification as done"},
			{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeNotifications, action: "open_actions_popup"}), Description: "Actions"},
			presenter.refreshHelpEntry("Refresh notifications"),
			{Key: presenter.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeNotifications, action: "open_detail"}), Description: "Open detail"},
		}
	default:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last user"},
			{Key: presenter.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: presenter.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: presenter.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: presenter.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: presenter.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: presenter.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search users"},
			{Key: presenter.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopeUser, action: "next_search_match"}) + "/" + presenter.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopeUser, action: "previous_search_match"}), Description: "Next/previous match"},
			{Key: presenter.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}), Description: "Actions"},
			{Key: presenter.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeUser, action: "open_detail"}), Description: "Open detail"},
		}
	}
}

func (presenter helpPresenter) globalHelpEntries() []helpEntry {
	return []helpEntry{
		{Key: presenter.helpKeysOrFallback("?", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}), Description: "Toggle help"},
		{Key: presenter.helpKeysOrFallback("tab", keybindingActionID{scope: keymapScopeGlobal, action: "next_side_view"}), Description: "Switch side view"},
		{Key: presenter.helpKeysOrFallback("shift+tab", keybindingActionID{scope: keymapScopeGlobal, action: "previous_side_view"}), Description: "Switch side view backwards"},
		{Key: "0/1/2/3", Description: "Jump to a view"},
		{Key: presenter.helpKeysOrFallback("<c-c>", keybindingActionID{scope: keymapScopeGlobal, action: "quit"}), Description: "Quit"},
	}
}

func (presenter helpPresenter) keyWidth(sections []helpSection) int {
	maxWidth := 0
	for _, section := range sections {
		for _, entry := range section.Entries {
			width := utf8.RuneCountInString(entry.Key)
			if width > maxWidth {
				maxWidth = width
			}
		}
	}
	return maxWidth
}

func (presenter helpPresenter) viewSize(maxX int, maxY int) (int, int) {
	sections := presenter.sections()
	keyWidth := presenter.keyWidth(sections)
	contentWidth := utf8.RuneCountInString("    --- Global ---")
	contentHeight := 0

	for sectionIndex, section := range sections {
		contentHeight++
		for _, entry := range section.Entries {
			lineWidth := keyWidth + 2 + utf8.RuneCountInString(entry.Description)
			if lineWidth > contentWidth {
				contentWidth = lineWidth
			}
			contentHeight++
		}
		if sectionIndex < len(sections)-1 {
			contentHeight++
		}
	}

	if contentWidth > maxX-6 {
		contentWidth = maxInt(20, maxX-6)
	}
	if contentHeight > maxY-6 {
		contentHeight = maxInt(5, maxY-6)
	}

	return contentWidth, contentHeight
}

func (presenter helpPresenter) helpKeysOrFallback(fallback string, actionIDs ...keybindingActionID) string {
	return presenter.keyResolver.helpKeysOrFallback(fallback, actionIDs...)
}

func (presenter helpPresenter) showsPullRequestDetailTabs() bool {
	return presenter.actionContext.Mode == ScreenModeBrowser && presenter.actionContext.MainView.ContentKind == MainContentKindPullRequestDetail
}

func (presenter helpPresenter) browserChangesInlineCommentShortcutActive() bool {
	return presenter.showsPullRequestDetailTabs() && presenter.actionContext.ActiveDetailTab == ChangesDetailTab
}

func (presenter helpPresenter) helpViewportPlacementKeysOrFallback(topFallback string, centerFallback string, bottomFallback string, topActionID keybindingActionID, centerActionID keybindingActionID, bottomActionID keybindingActionID) string {
	keys := []string{
		presenter.helpKeysOrFallback(topFallback, topActionID),
		presenter.helpKeysOrFallback(centerFallback, centerActionID),
		presenter.helpKeysOrFallback(bottomFallback, bottomActionID),
	}
	return strings.Join(keys, "/")
}

func (presenter helpPresenter) reviewFileMotionHelpKeys() string {
	return presenter.helpKeysOrFallback("[[", keybindingActionID{scope: keymapScopeReview, action: "previous_file"}) + "/" + presenter.helpKeysOrFallback("]]", keybindingActionID{scope: keymapScopeReview, action: "next_file"})
}

func (presenter helpPresenter) reviewCommentMotionHelpKeys() string {
	return presenter.helpKeysOrFallback("[c", keybindingActionID{scope: keymapScopeReview, action: "previous_comment"}) + "/" + presenter.helpKeysOrFallback("]c", keybindingActionID{scope: keymapScopeReview, action: "next_comment"})
}

func (presenter helpPresenter) wordMotionHelpKeys(scope string) string {
	keys := []string{
		presenter.helpKeysOrFallback("w", keybindingActionID{scope: scope, action: "move_cursor_to_next_word"}),
		presenter.helpKeysOrFallback("e", keybindingActionID{scope: scope, action: "move_cursor_to_word_end"}),
		presenter.helpKeysOrFallback("b", keybindingActionID{scope: scope, action: "move_cursor_to_previous_word"}),
		presenter.helpKeysOrFallback("W", keybindingActionID{scope: scope, action: "move_cursor_to_next_big_word"}),
		presenter.helpKeysOrFallback("E", keybindingActionID{scope: scope, action: "move_cursor_to_big_word_end"}),
		presenter.helpKeysOrFallback("B", keybindingActionID{scope: scope, action: "move_cursor_to_previous_big_word"}),
	}
	return strings.Join(keys, "/")
}

func (presenter helpPresenter) characterMotionHelpEntry() helpEntry {
	return helpEntry{Key: strings.Join([]string{
		presenter.helpKeysOrFallback("f", keybindingActionID{scope: keymapScopeCursor, action: "find_character_forward"}),
		presenter.helpKeysOrFallback("F", keybindingActionID{scope: keymapScopeCursor, action: "find_character_backward"}),
		presenter.helpKeysOrFallback("t", keybindingActionID{scope: keymapScopeCursor, action: "till_character_forward"}),
		presenter.helpKeysOrFallback("T", keybindingActionID{scope: keymapScopeCursor, action: "till_character_backward"}),
	}, "/"), Description: "Find/till character"}
}

func (presenter helpPresenter) repeatCharacterMotionHelpEntry() helpEntry {
	return helpEntry{Key: strings.Join([]string{
		presenter.helpKeysOrFallback(";", keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_forward"}),
		presenter.helpKeysOrFallback(",", keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_backward"}),
	}, "/"), Description: "Repeat character motion"}
}

func (presenter helpPresenter) searchWordUnderCursorHelpEntry() helpEntry {
	return helpEntry{Key: presenter.helpKeysOrFallback("*/#", keybindingActionID{scope: keymapScopeCursor, action: "search_word_under_cursor_forward"}, keybindingActionID{scope: keymapScopeCursor, action: "search_word_under_cursor_backward"}), Description: "Search word under cursor"}
}

func (presenter helpPresenter) refreshHelpEntry(description string) helpEntry {
	return helpEntry{Key: presenter.helpKeysOrFallback("alt+r", keybindingActionID{scope: keymapScopeMain, action: "refresh"}), Description: description}
}

func (presenter helpPresenter) inlineConversationToggleHelpKeys() string {
	return presenter.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeFolds, action: "toggle_inline_conversation"}) + "/" + presenter.helpKeysOrFallback("za", keybindingActionID{scope: keymapScopeFolds, action: "toggle_fold"})
}

func (presenter helpPresenter) reviewTreeToggleHelpKeys() string {
	return presenter.helpKeysOrFallback("za", keybindingActionID{scope: keymapScopePullRequests, action: "toggle_fold"})
}

func (presenter helpPresenter) bulkFoldHelpKeys() string {
	return presenter.helpKeysOrFallback("zM", keybindingActionID{scope: keymapScopeFolds, action: "close_all_folds"}) + "/" + presenter.helpKeysOrFallback("zR", keybindingActionID{scope: keymapScopeFolds, action: "open_all_folds"})
}

func (presenter helpPresenter) reviewTreeBulkFoldHelpKeys() string {
	return presenter.helpKeysOrFallback("zM", keybindingActionID{scope: keymapScopePullRequests, action: "close_all_folds"}) + "/" + presenter.helpKeysOrFallback("zR", keybindingActionID{scope: keymapScopePullRequests, action: "open_all_folds"})
}
