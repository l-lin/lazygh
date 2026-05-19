package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
)

const viewHelpName = "help"

type helpSection struct {
	Title   string
	Entries []helpEntry
}

type helpEntry struct {
	Key         string
	Description string
}

func (program *Program) configureHelpView(view *gocui.View) {
	configureFramedOverlayView(view, "Keybindings", "")
	view.Wrap = false
	view.Highlight = false
}

func (program *Program) renderHelpView(view *gocui.View) {
	view.Clear()

	sections := program.helpSections()
	keyWidth := program.helpKeyWidth(sections)
	for sectionIndex, section := range sections {
		fmt.Fprintf(view, "    --- %s ---\n", section.Title)
		for _, entry := range section.Entries {
			fmt.Fprintf(view, "%-*s %s\n", keyWidth+2, entry.Key, entry.Description)
		}
		if sectionIndex < len(sections)-1 {
			fmt.Fprintln(view)
		}
	}
}

func (program *Program) helpSections() []helpSection {
	return []helpSection{
		{Title: "Local", Entries: program.localHelpEntries()},
		{Title: "Global", Entries: program.globalHelpEntries()},
	}
}

func (program *Program) localHelpEntries() []helpEntry {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		switch actionContext.ActiveView.Focus {
		case FocusDetailView:
			entries := []helpEntry{
				{Key: "h/j/k/l/<up>/<down>/<left>/<right>", Description: "Move cursor"},
				{Key: "0/$", Description: "Line start/end"},
				{Key: "gg/G", Description: "First/last line"},
				{Key: program.reviewFileMotionHelpKeys(), Description: "Previous/next file"},
				{Key: program.reviewCommentMotionHelpKeys(), Description: "Previous/next comment"},
				{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_top"}, keybindingActionID{scope: keymapScopeCursor, action: "recenter_cursor"}, keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_bottom"}), Description: "Cursor to top/center/bottom"},
				{Key: program.wordMotionHelpKeys(keymapScopeCursor), Description: "Next/end/previous word/WORD"},
				program.characterMotionHelpEntry(),
				program.repeatCharacterMotionHelpEntry(),
				{Key: "n/N", Description: "Next/previous match"},
				{Key: "v/V", Description: "Start char/line visual selection"},
				program.reviewInlineCommentHelpEntry(),
				{Key: program.inlineConversationToggleHelpKeys(), Description: "Expand/collapse conversation"},
				{Key: program.bulkFoldHelpKeys(), Description: "Close/open all folds"},
				{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}), Description: "Actions"},
				{Key: program.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopePullRequests, action: "copy_pull_request_url"}), Description: "Yank selection / PR URL"},
				{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
				{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
				{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
				{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
				{Key: "+/-", Description: "Toggle fullscreen"},
				{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search diff"},
				program.searchWordUnderCursorHelpEntry(),
				{Key: "<esc>/q", Description: "Exit visual / return"},
			}
			if program.inlineCommentReplyShortcutAvailable() {
				entries = append(entries, program.inlineCommentReplyHelpEntry())
			}
			return entries
		case FocusPullRequestsView:
			return []helpEntry{
				{Key: "j/k/<up>/<down>", Description: "Move down/up"},
				{Key: "gg/G", Description: "First/last file"},
				{Key: program.reviewFileMotionHelpKeys(), Description: "Previous/next file"},
				{Key: program.reviewCommentMotionHelpKeys(), Description: "Previous/next comment"},
				{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
				{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
				{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
				{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
				{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
				{Key: "+/-", Description: "Resize panes"},
				{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search file tree"},
				{Key: program.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopePullRequests, action: "next_search_match"}) + "/" + program.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopePullRequests, action: "previous_search_match"}), Description: "Next/previous match"},
				program.pullRequestYankHelpEntry(keymapScopePullRequests),
				program.pullRequestCommentHelpEntry(keymapScopePullRequests),
				{Key: program.reviewTreeToggleHelpKeys(), Description: "Expand/collapse fold"},
				{Key: program.reviewTreeBulkFoldHelpKeys(), Description: "Close/open all folds"},
				{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "Actions"},
				{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open diff"},
				{Key: "<esc>/q", Description: "Exit review mode"},
			}
		default:
			return []helpEntry{
				{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				program.pullRequestYankHelpEntry(keymapScopeUser),
				program.pullRequestCommentHelpEntry(keymapScopeUser),
				{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}), Description: "Actions"},
				{Key: program.helpKeysOrFallback("0", keybindingActionID{scope: keymapScopeSide, action: "focus_detail_view"}), Description: "Focus diff"},
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
			{Key: program.helpKeysOrFallback("gx", keybindingActionID{scope: keymapScopeCursor, action: "open_link_under_cursor"}), Description: "Open link under cursor"},
			{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_top"}, keybindingActionID{scope: keymapScopeCursor, action: "recenter_cursor"}, keybindingActionID{scope: keymapScopeCursor, action: "place_cursor_at_viewport_bottom"}), Description: "Cursor to top/center/bottom"},
			{Key: program.wordMotionHelpKeys(keymapScopeCursor), Description: "Next/end/previous word/WORD"},
			program.characterMotionHelpEntry(),
			program.repeatCharacterMotionHelpEntry(),
			{Key: "n/N", Description: "Next/previous match"},
			{Key: "v/V", Description: "Start char/line visual selection"},
			{Key: program.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopePullRequests, action: "copy_pull_request_url"}), Description: "Yank selection / PR URL"},
			{Key: program.helpKeysOrFallback("<c-v>", keybindingActionID{scope: keymapScopePullRequests, action: "open_pull_request_by_url"}), Description: "Open PR from clipboard"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Toggle fullscreen"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search detail"},
			program.searchWordUnderCursorHelpEntry(),
			{Key: "<esc>/q", Description: "Exit visual / return"},
		}
		entries = append(entries, helpEntry{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeGlobal, action: "open_actions_popup"}), Description: "Actions"})
		if program.shouldShowPullRequestDetailTabs() {
			entries = append(entries, program.detailPullRequestCommentHelpEntry())
			if program.inlineCommentReplyShortcutAvailable() {
				entries = append(entries, program.inlineCommentReplyHelpEntry())
			}
			entries = append(entries,
				helpEntry{Key: program.inlineConversationToggleHelpKeys(), Description: "Expand/collapse section"},
				helpEntry{Key: program.bulkFoldHelpKeys(), Description: "Close/open all folds"},
				helpEntry{Key: program.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopeGlobal, action: "previous_tab"}), Description: "Previous detail tab"},
				helpEntry{Key: program.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopeGlobal, action: "next_tab"}), Description: "Next detail tab"},
			)
		}
		return entries
	case FocusPullRequestsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last pull request"},
			{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search pull requests"},
			{Key: program.helpKeysOrFallback(":", keybindingActionID{scope: keymapScopePullRequests, action: "custom_search"}), Description: "Custom search"},
			{Key: program.helpKeysOrFallback("<c-v>", keybindingActionID{scope: keymapScopePullRequests, action: "open_pull_request_by_url"}), Description: "Open PR from clipboard"},
			{Key: program.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopePullRequests, action: "next_search_match"}) + "/" + program.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopePullRequests, action: "previous_search_match"}), Description: "Next/previous match"},
			program.pullRequestYankHelpEntry(keymapScopePullRequests),
			program.pullRequestCommentHelpEntry(keymapScopePullRequests),
			{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "Actions"},
			{Key: program.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopeGlobal, action: "previous_tab"}), Description: "Previous tab"},
			{Key: program.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopeGlobal, action: "next_tab"}), Description: "Next tab"},
			{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open detail"},
		}
	case FocusNotificationsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last notification"},
			{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search notifications"},
			{Key: program.helpKeysOrFallback("r", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_read"}), Description: "Mark notification as read"},
			{Key: program.helpKeysOrFallback("d", keybindingActionID{scope: keymapScopeNotifications, action: "mark_notification_done"}), Description: "Mark notification as done"},
			{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeNotifications, action: "open_actions_popup"}), Description: "Actions"},
			{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeNotifications, action: "open_detail"}), Description: "Open detail"},
		}
	default:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last user"},
			{Key: program.helpViewportPlacementKeysOrFallback("zt", "zz", "zb", keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_top"}, keybindingActionID{scope: keymapScopeSide, action: "recenter_selection"}, keybindingActionID{scope: keymapScopeSide, action: "place_selection_at_viewport_bottom"}), Description: "Selection to top/center/bottom"},
			{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Half-page down + recenter"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Half-page up + recenter"},
			{Key: program.helpKeysOrFallback("<c-f>/pagedown", keybindingActionID{scope: keymapScopeMain, action: "full_page_down"}), Description: "Full-page down"},
			{Key: program.helpKeysOrFallback("<c-b>/pageup", keybindingActionID{scope: keymapScopeMain, action: "full_page_up"}), Description: "Full-page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search users"},
			{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeUser, action: "open_actions_popup"}), Description: "Actions"},
			{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeUser, action: "open_detail"}), Description: "Open detail"},
		}
	}
}

func (program *Program) fullPageHelpDown(gui *gocui.Gui, view *gocui.View) error {
	return program.scrollReadOnlyView(gui, view, viewHelpName, fullPageDelta(viewPageSize(program.resolveView(gui, view, viewHelpName))))
}

func (program *Program) fullPageHelpUp(gui *gocui.Gui, view *gocui.View) error {
	return program.scrollReadOnlyView(gui, view, viewHelpName, -fullPageDelta(viewPageSize(program.resolveView(gui, view, viewHelpName))))
}

func (program *Program) globalHelpEntries() []helpEntry {
	return []helpEntry{
		{Key: program.helpKeysOrFallback("?", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}), Description: "Toggle help"},
		{Key: program.helpKeysOrFallback("tab", keybindingActionID{scope: keymapScopeGlobal, action: "next_side_view"}), Description: "Switch side view"},
		{Key: program.helpKeysOrFallback("shift+tab", keybindingActionID{scope: keymapScopeGlobal, action: "previous_side_view"}), Description: "Switch side view backwards"},
		{Key: "0/1/2/3", Description: "Jump to a view"},
		{Key: program.helpKeysOrFallback("<c-c>", keybindingActionID{scope: keymapScopeGlobal, action: "quit"}), Description: "Quit"},
	}
}

func (program *Program) helpKeyWidth(sections []helpSection) int {
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

func (program *Program) helpKeysOrFallback(fallback string, actionIDs ...keybindingActionID) string {
	if len(actionIDs) == 0 {
		return formatKeyTextForDisplay(fallback)
	}

	actualLabels, ok, hasOverride := program.resolvedKeyLabels(actionIDs...)
	if !ok || !hasOverride || len(actualLabels) == 0 {
		return formatKeyTextForDisplay(fallback)
	}

	return strings.Join(formattedKeySequenceLabelsForDisplay(actualLabels), "/")
}

func (program *Program) resolvedKeyLabelsText(actionIDs ...keybindingActionID) string {
	actualLabels, ok, _ := program.resolvedKeyLabels(actionIDs...)
	if !ok || len(actualLabels) == 0 {
		return ""
	}
	return strings.Join(formattedKeySequenceLabelsForDisplay(actualLabels), "/")
}

func (program *Program) resolvedKeyLabels(actionIDs ...keybindingActionID) ([]string, bool, bool) {
	if len(actionIDs) == 0 {
		return nil, false, false
	}

	resolvedActions := map[keybindingActionID]resolvedKeybindingAction{}
	for _, action := range program.resolvedKeybindingActions() {
		resolvedActions[action.action.id] = action
	}

	actualLabels := make([]string, 0)
	hasOverride := false
	for _, actionID := range actionIDs {
		action, ok := resolvedActions[actionID]
		if !ok {
			return nil, false, false
		}
		if action.overridden {
			hasOverride = true
		}
		for _, binding := range action.bindings {
			actualLabels = append(actualLabels, binding.label)
		}
	}

	return actualLabels, true, hasOverride
}

func formattedKeySequenceLabelsForDisplay(labels []string) []string {
	if len(labels) == 0 {
		return nil
	}

	formatted := make([]string, 0, len(labels))
	for _, label := range labels {
		formatted = append(formatted, formatKeySequenceLabelForDisplay(label))
	}
	return formatted
}

func formatKeyTextForDisplay(text string) string {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" || trimmedText == "/" || strings.Contains(trimmedText, "<c-/>") {
		return formatKeySequenceLabelForDisplay(trimmedText)
	}

	segments := strings.Split(trimmedText, "/")
	if len(segments) <= 1 {
		return formatKeySequenceLabelForDisplay(trimmedText)
	}
	for index, segment := range segments {
		if segment == "" {
			return formatKeySequenceLabelForDisplay(trimmedText)
		}
		segments[index] = formatKeySequenceLabelForDisplay(segment)
	}
	return strings.Join(segments, "/")
}

func formatKeySequenceLabelForDisplay(label string) string {
	trimmedLabel := strings.TrimSpace(label)
	if trimmedLabel == "" {
		return ""
	}

	normalizedLabel := strings.ToLower(trimmedLabel)
	normalizedLabel = strings.TrimPrefix(normalizedLabel, "<")
	normalizedLabel = strings.TrimSuffix(normalizedLabel, ">")
	if after, ok := strings.CutPrefix(normalizedLabel, "control+"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "ctrl-"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "c-"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "alt-"); ok {
		normalizedLabel = "alt+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "shift-"); ok {
		normalizedLabel = "shift+" + after
	}

	switch normalizedLabel {
	case "enter":
		return "Enter"
	case "esc", "escape":
		return "Escape"
	case "tab":
		return "Tab"
	case "shift+tab", "backtab":
		return "Shift+Tab"
	case "up", "arrowup", "arrow-up":
		return "Up"
	case "down", "arrowdown", "arrow-down":
		return "Down"
	case "left", "arrowleft", "arrow-left":
		return "Left"
	case "right", "arrowright", "arrow-right":
		return "Right"
	case "pageup", "page-up", "pgup":
		return "PageUp"
	case "pagedown", "page-down", "pgdown", "pgdn":
		return "PageDown"
	case "space":
		return "Space"
	case "alt+enter":
		return "Alt+Enter"
	}

	if after, ok := strings.CutPrefix(normalizedLabel, "ctrl+"); ok {
		suffix := after
		switch suffix {
		case "space":
			return "Ctrl+Space"
		case "lsqbracket":
			suffix = "["
		case "rsqbracket":
			suffix = "]"
		case "backslash":
			suffix = "\\"
		case "slash":
			suffix = "/"
		case "underscore":
			suffix = "_"
		default:
			suffix = strings.ToUpper(suffix)
		}
		return "Ctrl+" + suffix
	}

	return trimmedLabel
}

func (program *Program) helpViewportPlacementKeysOrFallback(topFallback string, centerFallback string, bottomFallback string, topActionID keybindingActionID, centerActionID keybindingActionID, bottomActionID keybindingActionID) string {
	keys := []string{
		program.helpKeysOrFallback(topFallback, topActionID),
		program.helpKeysOrFallback(centerFallback, centerActionID),
		program.helpKeysOrFallback(bottomFallback, bottomActionID),
	}
	return strings.Join(keys, "/")
}

func (program *Program) reviewFileMotionHelpKeys() string {
	return program.helpKeysOrFallback("[[", keybindingActionID{scope: keymapScopeReview, action: "previous_file"}) + "/" + program.helpKeysOrFallback("]]", keybindingActionID{scope: keymapScopeReview, action: "next_file"})
}

func (program *Program) reviewCommentMotionHelpKeys() string {
	return program.helpKeysOrFallback("[c", keybindingActionID{scope: keymapScopeReview, action: "previous_comment"}) + "/" + program.helpKeysOrFallback("]c", keybindingActionID{scope: keymapScopeReview, action: "next_comment"})
}

func (program *Program) wordMotionHelpKeys(scope string) string {
	keys := []string{
		program.helpKeysOrFallback("w", keybindingActionID{scope: scope, action: "move_cursor_to_next_word"}),
		program.helpKeysOrFallback("e", keybindingActionID{scope: scope, action: "move_cursor_to_word_end"}),
		program.helpKeysOrFallback("b", keybindingActionID{scope: scope, action: "move_cursor_to_previous_word"}),
		program.helpKeysOrFallback("W", keybindingActionID{scope: scope, action: "move_cursor_to_next_big_word"}),
		program.helpKeysOrFallback("E", keybindingActionID{scope: scope, action: "move_cursor_to_big_word_end"}),
		program.helpKeysOrFallback("B", keybindingActionID{scope: scope, action: "move_cursor_to_previous_big_word"}),
	}
	return strings.Join(keys, "/")
}

func (program *Program) characterMotionHelpEntry() helpEntry {
	return helpEntry{Key: strings.Join([]string{
		program.helpKeysOrFallback("f", keybindingActionID{scope: keymapScopeCursor, action: "find_character_forward"}),
		program.helpKeysOrFallback("F", keybindingActionID{scope: keymapScopeCursor, action: "find_character_backward"}),
		program.helpKeysOrFallback("t", keybindingActionID{scope: keymapScopeCursor, action: "till_character_forward"}),
		program.helpKeysOrFallback("T", keybindingActionID{scope: keymapScopeCursor, action: "till_character_backward"}),
	}, "/"), Description: "Find/till character"}
}

func (program *Program) repeatCharacterMotionHelpEntry() helpEntry {
	return helpEntry{Key: strings.Join([]string{
		program.helpKeysOrFallback(";", keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_forward"}),
		program.helpKeysOrFallback(",", keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_backward"}),
	}, "/"), Description: "Repeat character motion"}
}

func (program *Program) searchWordUnderCursorHelpEntry() helpEntry {
	return helpEntry{Key: program.helpKeysOrFallback("*/#", keybindingActionID{scope: keymapScopeCursor, action: "search_word_under_cursor_forward"}, keybindingActionID{scope: keymapScopeCursor, action: "search_word_under_cursor_backward"}), Description: "Search word under cursor"}
}

func (program *Program) inlineConversationToggleHelpKeys() string {
	return program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeFolds, action: "toggle_inline_conversation"}) + "/" + program.helpKeysOrFallback("za", keybindingActionID{scope: keymapScopeFolds, action: "toggle_fold"})
}

func (program *Program) reviewTreeToggleHelpKeys() string {
	return program.helpKeysOrFallback("za", keybindingActionID{scope: keymapScopePullRequests, action: "toggle_fold"})
}

func (program *Program) bulkFoldHelpKeys() string {
	return program.helpKeysOrFallback("zM", keybindingActionID{scope: keymapScopeFolds, action: "close_all_folds"}) + "/" + program.helpKeysOrFallback("zR", keybindingActionID{scope: keymapScopeFolds, action: "open_all_folds"})
}

func (program *Program) reviewTreeBulkFoldHelpKeys() string {
	return program.helpKeysOrFallback("zM", keybindingActionID{scope: keymapScopePullRequests, action: "close_all_folds"}) + "/" + program.helpKeysOrFallback("zR", keybindingActionID{scope: keymapScopePullRequests, action: "open_all_folds"})
}

func (program *Program) helpViewSize(maxX int, maxY int) (int, int) {
	sections := program.helpSections()
	keyWidth := program.helpKeyWidth(sections)
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
