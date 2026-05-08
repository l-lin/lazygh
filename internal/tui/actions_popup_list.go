package tui

import "strings"

const (
	actionsPopupGroupPullRequest = "Pull request"
	actionsPopupGroupReview      = "Review"
	actionsPopupGroupNavigation  = "Navigation"
	actionsPopupGroupTheme       = "Theme"
	actionsPopupGroupCache       = "Cache"
)

type actionsPopupVisibleLine struct {
	text        string
	actionIndex int
	selectable  bool
}

func actionsPopupGrouped(group string, actions ...actionsPopupAction) []actionsPopupAction {
	grouped := make([]actionsPopupAction, 0, len(actions))
	for _, action := range actions {
		grouped = append(grouped, action.withGroup(group))
	}
	return grouped
}

func buildActionsPopupVisibleLines(actions []actionsPopupAction, filteredIndexes []int) []actionsPopupVisibleLine {
	visibleLines := make([]actionsPopupVisibleLine, 0, len(filteredIndexes))
	lastGroup := ""
	for _, actionIndex := range filteredIndexes {
		if actionIndex < 0 || actionIndex >= len(actions) {
			continue
		}
		action := actions[actionIndex]
		group := strings.TrimSpace(action.group)
		if group != "" && group != lastGroup {
			visibleLines = append(visibleLines, actionsPopupVisibleLine{text: group, actionIndex: -1, selectable: false})
			lastGroup = group
		}
		visibleLines = append(visibleLines, actionsPopupVisibleLine{text: action.label(), actionIndex: actionIndex, selectable: true})
	}
	return visibleLines
}

func (program *Program) currentActionsPopupVisibleLines() []actionsPopupVisibleLine {
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	return buildActionsPopupVisibleLines(program.currentActionsPopupActions(), program.model.ActionsPopupFilteredActionIndexes())
}

func (program *Program) currentActionsPopupSelectedRenderedLine() int {
	selectedActionIndex := program.model.ActionsPopupSelectedActionIndex()
	visibleLines := program.currentActionsPopupVisibleLines()
	for index, line := range visibleLines {
		if line.selectable && line.actionIndex == selectedActionIndex {
			return index
		}
	}
	return 0
}

func (program *Program) currentActionsPopupRenderedLineCount() int {
	return len(program.currentActionsPopupVisibleLines())
}
