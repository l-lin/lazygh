package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) activeSearchIsReviewFileTreeSearch() bool {
	return program.reviewModeActive() && program.model.SearchActive() && program.model.SearchTargetKind() == SearchTargetReviewTree
}

func (program *Program) reviewFileTreeSearchQuery() string {
	return program.model.ReviewTreeSearchQuery()
}

func (program *Program) nextReviewFileTreeSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgRepeatReviewFileTreeSearch{Direction: searchRepeatForward})
}

func (program *Program) previousReviewFileTreeSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgRepeatReviewFileTreeSearch{Direction: searchRepeatBackward})
}

func (program *Program) repeatReviewFileTreeSearch(gui *gocui.Gui, choose searchMatchIndexChooser) error {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView {
		return nil
	}

	query := program.model.ReviewTreeSearchQuery()
	if strings.TrimSpace(query) == "" {
		return nil
	}
	if !program.followReviewFileTreeSearch(query, choose) {
		return nil
	}

	return nil
}

func (program *Program) followSubmittedReviewFileTreeSearch(query string) bool {
	return program.followReviewFileTreeSearch(query, searchMatchIndexAtOrAfter)
}

func (program *Program) followReviewFileTreeSearch(query string, choose searchMatchIndexChooser) bool {
	matchRows := program.reviewFileTreeSearchMatchRows(query)
	matchIndex := choose(matchRows, program.reviewSessionSelectedVisibleLine())
	if matchIndex < 0 || matchIndex >= len(matchRows) {
		return false
	}

	program.navigationState.reviewSession.selectedFileTreeRow = matchRows[matchIndex]
	return true
}

func (program *Program) reviewFileTreeSearchMatchCount(query string) int {
	return len(program.reviewFileTreeSearchMatchRows(query))
}

func (program *Program) reviewFileTreeSearchMatchRows(query string) []int {
	tree, files, ok := program.reviewSessionCurrentTree()
	if !ok {
		return nil
	}

	return reviewDiffTreeSearchMatchRows(tree, files, query)
}

func reviewDiffTreeSearchMatchRows(tree reviewDiffTree, files []reviewDiffFile, query string) []int {
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
		if strings.Contains(strings.ToLower(reviewDiffTreeRowSearchText(row, files)), loweredQuery) {
			matchRows = append(matchRows, row.VisibleRowIndex)
		}
	}
	return matchRows
}
