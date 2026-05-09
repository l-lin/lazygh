package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) toggleSelectedReviewTreeRowVisibility(gui *gocui.Gui) (bool, error) {
	if !program.reviewSession.active || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		return false, nil
	}

	visibleTree, _, ok := program.reviewSessionCurrentTree()
	if !ok || len(visibleTree.Rows) == 0 {
		return false, nil
	}
	selectedRowIndex := clampIndex(program.reviewSession.selectedFileTreeRow, len(visibleTree.Rows))
	selectedRow := visibleTree.Rows[selectedRowIndex]
	if !selectedRow.Foldable {
		return false, nil
	}

	rawTree, _, rawTreeOK := program.reviewSessionRawTree()
	if !rawTreeOK {
		return false, nil
	}
	program.setReviewTreeRowCollapsed(selectedRow.ID, !selectedRow.Collapsed)
	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.reviewSession.collapsedTreeRowIDs)
	program.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRow.ID)
	return true, program.refreshViewsIfGUI(gui)
}

func (program *Program) closeAllReviewTreeFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.setAllReviewTreeFolds(gui, view, true)
}

func (program *Program) openAllReviewTreeFolds(gui *gocui.Gui, view *gocui.View) error {
	return program.setAllReviewTreeFolds(gui, view, false)
}

func (program *Program) setAllReviewTreeFolds(gui *gocui.Gui, view *gocui.View, collapsed bool) error {
	if !program.reviewSession.active || program.model.Focus() != FocusPullRequestsView || program.reviewTreeFoldBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	viewName := viewPullRequestsName
	if view != nil && strings.TrimSpace(view.Name()) != "" {
		viewName = view.Name()
	}
	if !program.pendingSelectionKeySequence.consume(sideViewportPlacementTarget(viewName)) {
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
	selectedRowID := currentVisibleTree.Rows[clampIndex(program.reviewSession.selectedFileTreeRow, len(currentVisibleTree.Rows))].ID
	if !program.setAllReviewTreeRowsCollapsed(rawTree, collapsed) {
		return nil
	}

	updatedVisibleTree := reviewDiffTreeVisibleRows(rawTree, program.reviewSession.collapsedTreeRowIDs)
	program.reviewSession.selectedFileTreeRow = reviewDiffTreePreferredVisibleRowIndex(rawTree, updatedVisibleTree, selectedRowID)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) reviewTreeFoldBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) setReviewTreeRowCollapsed(rowID string, collapsed bool) {
	trimmedRowID := strings.TrimSpace(rowID)
	if trimmedRowID == "" {
		return
	}
	if program.reviewSession.collapsedTreeRowIDs == nil {
		program.reviewSession.collapsedTreeRowIDs = map[string]bool{}
	}
	program.reviewSession.collapsedTreeRowIDs[trimmedRowID] = collapsed
}

func (program *Program) setAllReviewTreeRowsCollapsed(tree reviewDiffTree, collapsed bool) bool {
	if len(tree.Rows) == 0 {
		return false
	}
	if program.reviewSession.collapsedTreeRowIDs == nil {
		program.reviewSession.collapsedTreeRowIDs = map[string]bool{}
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
		if actualCollapsed, ok := program.reviewSession.collapsedTreeRowIDs[trimmedRowID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		program.reviewSession.collapsedTreeRowIDs[trimmedRowID] = collapsed
	}
	return changed
}
