package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

type reviewDiffTreeSearchRowChooser func([]int, int) int

func (program *Program) startReviewFileTreeSearch() {
	program.model.searchActive = true
	program.model.searchTarget = FocusPullRequestsView
	program.model.searchTargetPullRequestTab = program.model.activePullRequestTab
	program.model.clearAppliedSearchQueriesForOtherViews(FocusPullRequestsView)
	program.model.searchDraft = ""
}

func (program *Program) activeSearchIsReviewFileTreeSearch() bool {
	return program.reviewSession.active && program.model.SearchActive() && program.model.SearchTarget() == FocusPullRequestsView
}

func (program *Program) updateActiveSearchDraft(query string) {
	if program.activeSearchIsReviewFileTreeSearch() {
		program.model.searchDraft = query
		return
	}

	program.model.UpdateSearchDraft(query)
}

func (program *Program) reviewFileTreeSearchQuery() string {
	if program.activeSearchIsReviewFileTreeSearch() {
		return program.model.SearchDraft()
	}

	return program.reviewSession.fileTreeSearchQuery
}

func (program *Program) submitReviewFileTreeSearch() {
	query := program.model.searchDraft
	program.reviewSession.fileTreeSearchQuery = query
	program.model.searchActive = false
	program.model.searchDraft = ""
	program.followSubmittedReviewFileTreeSearch(query)
}

func (program *Program) cancelReviewFileTreeSearch() {
	program.model.searchActive = false
	program.model.searchDraft = ""
}

func (program *Program) nextReviewFileTreeSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.repeatReviewFileTreeSearch(gui, reviewDiffTreeSearchMatchIndexAfterRow)
}

func (program *Program) previousReviewFileTreeSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.repeatReviewFileTreeSearch(gui, reviewDiffTreeSearchMatchIndexBeforeRow)
}

func (program *Program) repeatReviewFileTreeSearch(gui *gocui.Gui, choose reviewDiffTreeSearchRowChooser) error {
	if !program.reviewSession.active || program.model.Focus() != FocusPullRequestsView {
		return nil
	}

	query := program.reviewSession.fileTreeSearchQuery
	if strings.TrimSpace(query) == "" {
		return nil
	}
	if !program.followReviewFileTreeSearch(query, choose) {
		return nil
	}

	return program.refreshViewsIfGUI(gui)
}

func (program *Program) followSubmittedReviewFileTreeSearch(query string) bool {
	return program.followReviewFileTreeSearch(query, reviewDiffTreeSearchMatchIndexAtOrAfterRow)
}

func (program *Program) followReviewFileTreeSearch(query string, choose reviewDiffTreeSearchRowChooser) bool {
	matchRows := program.reviewFileTreeSearchMatchRows(query)
	matchIndex := choose(matchRows, program.reviewSessionSelectedVisibleLine())
	if matchIndex < 0 || matchIndex >= len(matchRows) {
		return false
	}

	program.reviewSession.selectedFileTreeRow = matchRows[matchIndex]
	return true
}

func (program *Program) reviewFileTreeSearchMatchCount(query string) int {
	return len(program.reviewFileTreeSearchMatchRows(query))
}

func (program *Program) reviewFileTreeSearchMatchRows(query string) []int {
	tree, _, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil
	}

	return reviewDiffTreeSearchMatchRows(tree, query)
}

func reviewDiffTreeSearchMatchRows(tree reviewDiffTree, query string) []int {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil
	}

	matchRows := make([]int, 0, len(tree.Rows))
	loweredQuery := strings.ToLower(trimmedQuery)
	for _, row := range tree.Rows {
		if !row.Foldable && row.FileIndex < 0 {
			continue
		}
		if strings.Contains(strings.ToLower(row.Label), loweredQuery) {
			matchRows = append(matchRows, row.VisibleRowIndex)
		}
	}
	return matchRows
}

func reviewDiffTreeSearchMatchIndexAtOrAfterRow(matchRows []int, currentRow int) int {
	if len(matchRows) == 0 {
		return -1
	}

	for matchIndex, matchRow := range matchRows {
		if matchRow >= currentRow {
			return matchIndex
		}
	}

	return 0
}

func reviewDiffTreeSearchMatchIndexAfterRow(matchRows []int, currentRow int) int {
	if len(matchRows) == 0 {
		return -1
	}

	for matchIndex, matchRow := range matchRows {
		if matchRow > currentRow {
			return matchIndex
		}
	}

	return 0
}

func reviewDiffTreeSearchMatchIndexBeforeRow(matchRows []int, currentRow int) int {
	if len(matchRows) == 0 {
		return -1
	}

	for matchIndex := len(matchRows) - 1; matchIndex >= 0; matchIndex-- {
		if matchRows[matchIndex] < currentRow {
			return matchIndex
		}
	}

	return len(matchRows) - 1
}
