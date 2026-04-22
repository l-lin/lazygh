package tui

import (
	"fmt"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
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
	totalWidth := innerWidth + 2
	totalHeight := innerHeight + 2

	x0 := clampCoordinate((maxX-totalWidth)/2, maxX)
	y0 := clampCoordinate((maxY-totalHeight)/2, maxY)
	x1 := x0 + totalWidth - 1
	y1 := y0 + totalHeight - 1
	if x1 >= maxX {
		x1 = maxX - 1
		x0 = clampCoordinate(x1-totalWidth+1, maxX)
	}
	if y1 >= maxY {
		y1 = maxY - 1
		y0 = clampCoordinate(y1-totalHeight+1, maxY)
	}

	view, err := gui.SetView(viewHelpName, x0, y0, x1, y1, 0)
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
	view.Title = "Keybindings"
	view.Frame = true
	view.FrameRunes = roundFrameRunes
	view.FrameColor = gocui.GetColor(theme.ActiveBorderHex)
	view.TitleColor = gocui.GetColor(theme.ActiveTextHex)
	view.FgColor = gocui.GetColor(theme.ActiveTextHex)
	view.BgColor = gocui.ColorDefault
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
		{Title: "Global", Entries: globalHelpEntries()},
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
				reviewInlineCommentHelpEntry(),
				{Key: "a", Description: "Review actions"},
				{Key: "y", Description: "Yank selection / PR URL"},
				{Key: "<c-d>", Description: "Page down"},
				{Key: "<c-u>", Description: "Page up"},
				{Key: "+/-", Description: "Toggle fullscreen"},
				{Key: "/", Description: "Search diff"},
				{Key: "<esc>", Description: "Exit visual / return"},
			}
		case FocusPullRequestsView:
			return []helpEntry{
				{Key: "j/k/<up>/<down>", Description: "Move down/up"},
				{Key: "h/l", Description: "Switch side view"},
				{Key: "<c-d>", Description: "Page down"},
				{Key: "<c-u>", Description: "Page up"},
				{Key: "+/-", Description: "Resize panes"},
				pullRequestYankHelpEntry(),
				pullRequestCommentHelpEntry(),
				{Key: "a", Description: "Review actions"},
				{Key: "<enter>", Description: "Open diff"},
				{Key: "<esc>", Description: "Exit review mode"},
			}
		default:
			return []helpEntry{
				{Key: "h/l", Description: "Switch side view"},
				pullRequestYankHelpEntry(),
				pullRequestCommentHelpEntry(),
				{Key: "0", Description: "Focus diff"},
				{Key: "<esc>", Description: "Exit review mode"},
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
			{Key: "y", Description: "Yank selection / PR URL"},
			{Key: "<c-d>", Description: "Page down"},
			{Key: "<c-u>", Description: "Page up"},
			{Key: "+/-", Description: "Toggle fullscreen"},
			{Key: "/", Description: "Search detail"},
			{Key: "<esc>", Description: "Exit visual / return"},
		}
		if program.shouldShowPullRequestDetailTabs() {
			entries = append(entries,
				pullRequestCommentHelpEntry(),
				helpEntry{Key: "a", Description: "PR actions"},
				helpEntry{Key: "[", Description: "Previous detail tab"},
				helpEntry{Key: "]", Description: "Next detail tab"},
			)
		}
		return entries
	case FocusPullRequestsView:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "h/l", Description: "Switch side view"},
			{Key: "<c-d>", Description: "Page down"},
			{Key: "<c-u>", Description: "Page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: "/", Description: "Search pull requests"},
			pullRequestYankHelpEntry(),
			pullRequestCommentHelpEntry(),
			{Key: "a", Description: "PR actions"},
			{Key: "[", Description: "Previous tab"},
			{Key: "]", Description: "Next tab"},
			{Key: "<enter>", Description: "Open detail"},
		}
	default:
		return []helpEntry{
			{Key: "j/k/<up>/<down>", Description: "Move down/up"},
			{Key: "h/l", Description: "Switch side view"},
			{Key: "<c-d>", Description: "Page down"},
			{Key: "<c-u>", Description: "Page up"},
			{Key: "+/-", Description: "Resize panes"},
			{Key: "/", Description: "Search users"},
			{Key: "<enter>", Description: "Open detail"},
		}
	}
}

func globalHelpEntries() []helpEntry {
	return []helpEntry{
		{Key: "?", Description: "Toggle help"},
		{Key: "tab", Description: "Switch side view"},
		{Key: "shift+tab", Description: "Switch side view backwards"},
		{Key: "0/1/2", Description: "Jump to a view"},
		{Key: "<c-c>", Description: "Quit"},
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
		contentWidth = max(20, maxX-6)
	}
	if contentHeight > maxY-6 {
		contentHeight = max(5, maxY-6)
	}

	return contentWidth, contentHeight
}

func clampCoordinate(value int, maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= maxValue {
		return maxValue - 1
	}
	return value
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
