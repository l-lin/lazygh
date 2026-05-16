package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/theme"
)

const (
	actionsPopupGroupPullRequest   = "Pull request"
	actionsPopupGroupReview        = "Review"
	actionsPopupGroupNavigation    = "Navigation"
	actionsPopupGroupNotifications = "Notifications"
	actionsPopupGroupTheme         = "Theme"
	actionsPopupGroupCache         = "Cache"
)

type actionsPopupVisibleLine struct {
	text        string
	actionIndex int
	selectable  bool
	centered    bool
}

func (line actionsPopupVisibleLine) item(viewWidth int) Item {
	if line.selectable {
		return Item{Title: line.text}
	}
	if !line.centered {
		return Item{Title: line.text}
	}

	centeredTitle := centeredActionsPopupGroupTitle(line.text, viewWidth)
	return Item{
		Title: centeredTitle,
		TitleSegments: []ItemTitleSegment{{
			Text:               centeredTitle,
			ForegroundHex:      theme.ActionsPopupGroupForegroundHex,
			BackgroundHex:      theme.MarkdownHeadingBackgroundHex,
			MinimumContrast:    4.5,
			PreserveForeground: true,
		}},
	}
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
			visibleLines = append(visibleLines, actionsPopupVisibleLine{text: group, actionIndex: -1, selectable: false, centered: true})
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
	if program.assigneePickerVisible() {
		return program.currentAssigneePickerVisibleLines()
	}

	actions := program.currentActionsPopupActions()
	return buildActionsPopupVisibleLines(actions, program.model.ActionsPopupFilteredActionIndexes())
}

func (program *Program) currentAssigneePickerVisibleLines() []actionsPopupVisibleLine {
	actions := program.currentActionsPopupActions()
	visibleLines := buildActionsPopupVisibleLines(actions, actionIndexes(len(actions)))
	if program.assigneePickerLoading() {
		visibleLines = append(visibleLines, actionsPopupVisibleLine{text: strings.TrimSpace(program.loadingSpinnerFrame())})
	}
	return visibleLines
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

func centeredActionsPopupGroupTitle(title string, viewWidth int) string {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" || viewWidth <= len([]rune(trimmedTitle)) {
		return trimmedTitle
	}

	padding := viewWidth - len([]rune(trimmedTitle))
	leftPadding := padding / 2
	rightPadding := padding - leftPadding
	return strings.Repeat(" ", leftPadding) + trimmedTitle + strings.Repeat(" ", rightPadding)
}
