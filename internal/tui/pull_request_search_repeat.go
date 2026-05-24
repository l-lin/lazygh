package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) nextPullRequestsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewModeActive() {
		return program.nextReviewFileTreeSearchMatch(gui, view)
	}
	return program.repeatPullRequestSearch(gui, searchMatchIndexAfter)
}

func (program *Program) previousPullRequestsSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewModeActive() {
		return program.previousReviewFileTreeSearchMatch(gui, view)
	}
	return program.repeatPullRequestSearch(gui, searchMatchIndexBefore)
}

func (program *Program) repeatPullRequestSearch(gui *gocui.Gui, choose searchMatchIndexChooser) error {
	if program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView {
		return nil
	}

	tab := program.model.ActivePullRequestTab()
	query := program.model.PullRequestSearchQuery(tab)
	if strings.TrimSpace(query) == "" {
		return nil
	}

	matchIndexes := program.model.visiblePullRequestIndexes(tab)
	matchIndex := choose(matchIndexes, program.model.SelectedPullRequestIndex(tab))
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return nil
	}

	program.model.SelectPullRequestIndex(tab, matchIndexes[matchIndex])
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) followSubmittedPullRequestSearch(tab PullRequestTab, startIndex int) {
	if program.reviewModeActive() {
		return
	}

	query := program.model.PullRequestSearchQuery(tab)
	if strings.TrimSpace(query) == "" {
		return
	}

	matchIndexes := program.model.visiblePullRequestIndexes(tab)
	matchIndex := searchMatchIndexAtOrAfter(matchIndexes, startIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return
	}

	program.model.SelectPullRequestIndex(tab, matchIndexes[matchIndex])
}
