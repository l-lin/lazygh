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

func (program *Program) layoutHelpView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	innerWidth, innerHeight := program.helpViewSize(maxX, maxY)
	frame := centeredOverlayFrame(maxX, maxY, innerWidth+2, innerHeight+2)

	view, err := gui.SetView(viewHelpName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}

	program.configureHelpView(view)
	program.renderHelpView(view)
	_, err = gui.SetViewOnTop(viewHelpName)
	if isUnknownViewError(err) {
		return nil
	}

	return err
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
	if program.reviewSession.active {
		switch program.model.Focus() {
		case FocusDetailView:
			return []helpEntry{
				{Key: "h/j/k/l/<up>/<down>", Description: "Move cursor"},
				{Key: "0/$", Description: "Line start/end"},
				{Key: "gg/G", Description: "First/last line"},
				{Key: "w/e/b", Description: "Next/end/previous word"},
				{Key: "n/N", Description: "Next/previous match"},
				{Key: "v/V", Description: "Start char/line visual selection"},
				program.reviewInlineCommentHelpEntry(),
				{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeDetail, action: "toggle_inline_conversation"}), Description: "Expand/collapse conversation"},
				{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeDetail, action: "open_actions_popup"}), Description: "Review actions"},
				{Key: program.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopeDetail, action: "copy_pull_request_url"}), Description: "Yank selection / PR URL"},
				{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Page down"},
				{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Page up"},
				{Key: "+/-", Description: "Toggle fullscreen"},
				{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search diff"},
				{Key: "<esc>/q", Description: "Exit visual / return"},
			}
		case FocusPullRequestsView:
			return []helpEntry{
				{Key: "j/k/<up>/<down>", Description: "Move down/up"},
				{Key: "gg/G", Description: "First/last file"},
				{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Page down"},
				{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Page up"},
				{Key: "+/-", Description: "Resize panes"},
				{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search file tree"},
				{Key: program.helpKeysOrFallback("n", keybindingActionID{scope: keymapScopePullRequests, action: "next_search_match"}) + "/" + program.helpKeysOrFallback("N", keybindingActionID{scope: keymapScopePullRequests, action: "previous_search_match"}), Description: "Next/previous match"},
				program.pullRequestYankHelpEntry(keymapScopePullRequests),
				program.pullRequestCommentHelpEntry(keymapScopePullRequests),
				{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "Review actions"},
				{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open diff"},
				{Key: "<esc>/q", Description: "Exit review mode"},
			}
		default:
			return []helpEntry{
				{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
				program.pullRequestYankHelpEntry(keymapScopeUser),
				program.pullRequestCommentHelpEntry(keymapScopeUser),
				{Key: program.helpKeysOrFallback("0", keybindingActionID{scope: keymapScopeSide, action: "focus_detail_view"}), Description: "Focus diff"},
				{Key: "<esc>/q", Description: "Exit review mode"},
			}
		}
	}

	switch program.model.Focus() {
	case FocusDetailView:
		entries := []helpEntry{
			{Key: "h/j/k/l/<up>/<down>", Description: "Move cursor"},
			{Key: "0/$", Description: "Line start/end"},
			{Key: "gg/G", Description: "First/last line"},
			{Key: "w/e/b", Description: "Next/end/previous word"},
			{Key: "n/N", Description: "Next/previous match"},
			{Key: "v/V", Description: "Start char/line visual selection"},
			{Key: program.helpKeysOrFallback("y", keybindingActionID{scope: keymapScopeDetail, action: "copy_pull_request_url"}), Description: "Yank selection / PR URL"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Page down"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Page up"},
			{Key: "+/-", Description: "Toggle fullscreen"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search detail"},
			{Key: "<esc>/q", Description: "Exit visual / return"},
		}
		if program.shouldShowPullRequestDetailTabs() {
			entries = append(entries,
				program.pullRequestCommentHelpEntry(keymapScopeDetail),
				helpEntry{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopeDetail, action: "open_actions_popup"}), Description: "PR actions"},
				helpEntry{Key: program.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopeDetail, action: "previous_tab"}), Description: "Previous detail tab"},
				helpEntry{Key: program.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopeDetail, action: "next_tab"}), Description: "Next detail tab"},
			)
		}
		return entries
	case FocusPullRequestsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last pull request"},
			{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Page down"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search pull requests"},
			program.pullRequestYankHelpEntry(keymapScopePullRequests),
			program.pullRequestCommentHelpEntry(keymapScopePullRequests),
			{Key: program.helpKeysOrFallback("a", keybindingActionID{scope: keymapScopePullRequests, action: "open_actions_popup"}), Description: "PR actions"},
			{Key: program.helpKeysOrFallback("[", keybindingActionID{scope: keymapScopePullRequests, action: "previous_tab"}), Description: "Previous tab"},
			{Key: program.helpKeysOrFallback("]", keybindingActionID{scope: keymapScopePullRequests, action: "next_tab"}), Description: "Next tab"},
			{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopePullRequests, action: "open_detail"}), Description: "Open detail"},
		}
	default:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "gg/G", Description: "First/last user"},
			{Key: program.helpKeysOrFallback("h/l", keybindingActionID{scope: keymapScopeSide, action: "previous_side_view"}, keybindingActionID{scope: keymapScopeSide, action: "next_side_view"}), Description: "Switch side view"},
			{Key: program.helpKeysOrFallback("<c-d>", keybindingActionID{scope: keymapScopeMain, action: "page_down"}), Description: "Page down"},
			{Key: program.helpKeysOrFallback("<c-u>", keybindingActionID{scope: keymapScopeMain, action: "page_up"}), Description: "Page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: program.helpKeysOrFallback("/", keybindingActionID{scope: keymapScopeMain, action: "open_search"}), Description: "Search users"},
			{Key: program.helpKeysOrFallback("<enter>", keybindingActionID{scope: keymapScopeUser, action: "open_detail"}), Description: "Open detail"},
		}
	}
}

func (program *Program) globalHelpEntries() []helpEntry {
	return []helpEntry{
		{Key: program.helpKeysOrFallback("?", keybindingActionID{scope: keymapScopeMain, action: "toggle_help"}), Description: "Toggle help"},
		{Key: program.helpKeysOrFallback("tab", keybindingActionID{scope: keymapScopeGlobal, action: "next_side_view"}), Description: "Switch side view"},
		{Key: program.helpKeysOrFallback("shift+tab", keybindingActionID{scope: keymapScopeGlobal, action: "previous_side_view"}), Description: "Switch side view backwards"},
		{Key: "0/1/2", Description: "Jump to a view"},
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
		return fallback
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
			return fallback
		}
		if action.overridden {
			hasOverride = true
		}
		for _, binding := range action.bindings {
			actualLabels = append(actualLabels, binding.label)
		}
	}

	if !hasOverride || len(actualLabels) == 0 {
		return fallback
	}

	return strings.Join(actualLabels, "/")
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
