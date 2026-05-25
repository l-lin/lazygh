package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) toggleSelectedReviewTreeRowVisibility(gui *gocui.Gui) (bool, error) {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		return false, nil
	}

	visibleTree, _, ok := program.reviewSessionCurrentTree()
	if !ok || len(visibleTree.Rows) == 0 {
		return false, nil
	}
	selectedRowIndex := clampIndex(program.navigationState.reviewSession.selectedFileTreeRow, len(visibleTree.Rows))
	selectedRow := visibleTree.Rows[selectedRowIndex]

	rawTree, _, rawTreeOK := program.reviewSessionRawTree()
	if !rawTreeOK {
		return false, nil
	}
	targetRow, ok := reviewDiffTreeNearestFoldableRow(rawTree, selectedRow.ID)
	if !ok {
		return false, nil
	}
	program.setReviewTreeRowCollapsed(targetRow.ID, !reviewDiffTreeRowCollapsed(targetRow, program.navigationState.reviewSession.collapsedTreeRowIDs))
	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.navigationState.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, targetRow.ID)
	return true, nil
}

func (program *Program) togglePullRequestFold(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgToggleReviewTreeRowVisibility{})
}

func (program *Program) closeAllReviewTreeFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSetAllReviewTreeFolds{Collapsed: true})
}

func (program *Program) openAllReviewTreeFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSetAllReviewTreeFolds{Collapsed: false})
}

func (program *Program) setAllReviewTreeFolds(gui *gocui.Gui, _ *gocui.View, collapsed bool) error {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	rawTree, _, ok := program.reviewSessionRawTree()
	if !ok || len(rawTree.Rows) == 0 {
		return nil
	}
	currentVisibleTree, _, visibleTreeOK := program.reviewSessionCurrentTree()
	if !visibleTreeOK || len(currentVisibleTree.Rows) == 0 {
		return nil
	}
	selectedRowID := currentVisibleTree.Rows[clampIndex(program.navigationState.reviewSession.selectedFileTreeRow, len(currentVisibleTree.Rows))].ID
	if !program.setAllReviewTreeRowsCollapsed(rawTree, collapsed) {
		return nil
	}

	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.navigationState.reviewSession.collapsedTreeRowIDs)
	program.navigationState.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRowID)
	return nil
}

func (program *Program) reviewTreeFoldBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) setReviewTreeRowCollapsed(rowID string, collapsed bool) {
	trimmedRowID := strings.TrimSpace(rowID)
	if trimmedRowID == "" {
		return
	}
	if program.navigationState.reviewSession.collapsedTreeRowIDs == nil {
		program.navigationState.reviewSession.collapsedTreeRowIDs = map[string]bool{}
	}
	program.navigationState.reviewSession.collapsedTreeRowIDs[trimmedRowID] = collapsed
}

func (program *Program) setAllReviewTreeRowsCollapsed(tree reviewDiffTree, collapsed bool) bool {
	if len(tree.Rows) == 0 {
		return false
	}
	if program.navigationState.reviewSession.collapsedTreeRowIDs == nil {
		program.navigationState.reviewSession.collapsedTreeRowIDs = map[string]bool{}
	}

	changed := false
	for _, row := range tree.Rows {
		if !row.Foldable {
			continue
		}
		trimmedRowID := strings.TrimSpace(row.ID)
		if trimmedRowID == "" {
			continue
		}
		if actualCollapsed, ok := program.navigationState.reviewSession.collapsedTreeRowIDs[trimmedRowID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		program.navigationState.reviewSession.collapsedTreeRowIDs[trimmedRowID] = collapsed
	}
	return changed
}
