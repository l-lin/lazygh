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
	actionsPopupGroupErrors        = "Errors"
	actionsPopupGroupDisplay       = "Display"
	actionsPopupGroupCache         = "Cache"
)

type actionsPopupVisibleLine struct {
	text          string
	actionIndex   int
	selectable    bool
	centered      bool
	separator     bool
	titleSegments []ItemTitleSegment
}

func (line actionsPopupVisibleLine) item(viewWidth int) Item {
	if line.separator {
		separatorWidth := maxInt(viewWidth, 1)
		separatorText := strings.Repeat("─", separatorWidth)
		return Item{Title: separatorText, TitleSegments: []ItemTitleSegment{{
			Text:   separatorText,
			Prefix: foregroundColorEscape(theme.InactiveBorderHex),
		}}}
	}
	if line.selectable {
		return Item{Title: line.text, TitleSegments: append([]ItemTitleSegment(nil), line.titleSegments...)}
	}
	if !line.centered {
		return Item{Title: line.text, TitleSegments: append([]ItemTitleSegment(nil), line.titleSegments...)}
	}

	centeredTitle := centeredText(line.text, viewWidth)
	return Item{
		Title: centeredTitle,
		TitleSegments: []ItemTitleSegment{{
			Text:               centeredTitle,
			ForegroundHex:      theme.ActionsPopupGroupForegroundHex,
			BackgroundHex:      theme.ActionsPopupGroupBackgroundHex,
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
	if program != nil && program.refreshReadCache.enabled && program.refreshReadCache.actionsPopupVisibleKnown {
		return program.refreshReadCache.actionsPopupVisibleLines
	}

	visibleLines := program.buildCurrentActionsPopupVisibleLines()
	program.cacheActionsPopupVisibleLines(visibleLines)
	return visibleLines
}

func (program *Program) buildCurrentActionsPopupVisibleLines() []actionsPopupVisibleLine {
	if program.assigneePickerVisible() {
		return program.currentAssigneePickerVisibleLines()
	}

	actions := program.currentActionsPopupActions()
	return buildActionsPopupVisibleLines(actions, program.currentActionsPopupFilteredIndexes())
}

func (program *Program) currentAssigneePickerVisibleLines() []actionsPopupVisibleLine {
	actions := program.currentActionsPopupActions()
	sections := program.currentAssigneePickerCandidateSections(program.model.ActionsPopupSearchQuery())
	visibleLines := make([]actionsPopupVisibleLine, 0, len(actions)+2)
	for index := range sections.pinned {
		if index >= len(actions) {
			break
		}
		visibleLines = append(visibleLines, actionsPopupVisibleLine{text: actions[index].label(), actionIndex: index, selectable: true})
	}
	if len(sections.pinned) > 0 && len(sections.searchResults) > 0 {
		visibleLines = append(visibleLines, actionsPopupVisibleLine{separator: true})
	}
	for index := range sections.searchResults {
		actionIndex := len(sections.pinned) + index
		if actionIndex >= len(actions) {
			break
		}
		visibleLines = append(visibleLines, actionsPopupVisibleLine{text: actions[actionIndex].label(), actionIndex: actionIndex, selectable: true})
	}
	if program.assigneePickerLoading() {
		visibleLines = append(visibleLines, actionsPopupVisibleLine{text: program.loadingSpinnerStatus("Fetching assignees")})
	}
	return visibleLines
}

func (program *Program) currentActionsPopupSelectedRenderedLine() int {
	selectedActionIndex := program.currentActionsPopupSelectedActionIndex()
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
