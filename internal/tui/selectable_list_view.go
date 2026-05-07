package tui

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

type selectableListViewState struct {
	focus               Focus
	query               string
	items               []Item
	selectedVisibleLine int
}

func (program *Program) configureSelectableListView(view *gocui.View, focus Focus, title string, query string) {
	program.applyViewStyle(view, focus, title, true)
	if program.usesManualSelectedLineRendering(query) {
		view.Highlight = false
		view.HighlightInactive = false
	}
	view.Wrap = false
}

func (program *Program) renderSelectableListView(view *gocui.View, state selectableListViewState) {
	view.Clear()

	if len(state.items) == 0 && strings.TrimSpace(state.query) != "" {
		fmt.Fprintln(view, searchNoMatchesMessage(state.query))
		return
	}

	showSelectedLine := program.usesManualSelectedLineRendering(state.query) && program.shouldHighlightSelection(state.focus, true)
	for visibleIndex, item := range state.items {
		program.renderItemLine(view, item, state.query, showSelectedLine && visibleIndex == state.selectedVisibleLine)
	}

	program.selectListLine(view, state.selectedVisibleLine, len(state.items))
}

func (program *Program) renderItemLine(view *gocui.View, item Item, query string, selected bool) {
	text := program.displayItemTitle(item)
	if len(item.TitleSegments) == 0 || text != item.Title {
		program.renderHighlightedLine(view, text, query, selected)
		return
	}

	fmt.Fprintln(view, renderStyledItemTitle(item.Title, item.TitleSegments, query, selected))
}

func (program *Program) renderHighlightedLine(view *gocui.View, text string, query string, selected bool) {
	var highlightedText string
	if selected {
		selectedForegroundPrefix := foregroundColorEscapeForAttribute(view.SelFgColor)
		selectedLinePrefix := ansiBold + selectedForegroundPrefix + backgroundColorEscape(theme.SelectedLineBackgroundHex)
		selectedMatchPrefix := ansiBold + selectedForegroundPrefix + backgroundColorEscape(theme.SearchHighlightHex)
		highlightedText, _ = highlightSearchMatchesWithPrefixes(text, query, selectedLinePrefix, selectedMatchPrefix)
	} else {
		foregroundPrefix := foregroundColorEscapeForAttribute(view.FgColor)
		matchPrefix := foregroundPrefix + backgroundColorEscape(theme.SearchHighlightHex)
		highlightedText, _ = highlightSearchMatchesWithPrefixes(text, query, foregroundPrefix, matchPrefix)
	}
	fmt.Fprintln(view, highlightedText)
}

func renderStyledItemTitle(title string, segments []ItemTitleSegment, query string, selected bool) string {
	if len(segments) == 0 {
		return title
	}

	matchRanges := titleMatchRanges(title, query)
	selectedPrefix := ""
	matchPrefix := backgroundColorEscape(theme.SearchHighlightHex)
	if selected {
		selectedPrefix = ansiBold + backgroundColorEscape(theme.SelectedLineBackgroundHex)
		matchPrefix = ansiBold + backgroundColorEscape(theme.SearchHighlightHex)
	}

	var builder strings.Builder
	currentPrefix := ""
	globalIndex := 0
	rangeIndex := 0
	for _, segment := range segments {
		for _, character := range segment.Text {
			for rangeIndex < len(matchRanges) && globalIndex >= matchRanges[rangeIndex].end {
				rangeIndex++
			}

			prefix := segment.Prefix + selectedPrefix
			if rangeIndex < len(matchRanges) && globalIndex >= matchRanges[rangeIndex].start && globalIndex < matchRanges[rangeIndex].end {
				prefix = segment.Prefix + matchPrefix
			}
			if prefix != currentPrefix {
				if currentPrefix != "" {
					builder.WriteString(ansiReset)
				}
				if prefix != "" {
					builder.WriteString(prefix)
				}
				currentPrefix = prefix
			}

			builder.WriteRune(character)
			globalIndex++
		}
	}
	if currentPrefix != "" {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

type titleMatchRange struct {
	start int
	end   int
}

func titleMatchRanges(title string, query string) []titleMatchRange {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil
	}

	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(trimmedQuery))
	byteRanges := pattern.FindAllStringIndex(title, -1)
	if len(byteRanges) == 0 {
		return nil
	}

	matchRanges := make([]titleMatchRange, 0, len(byteRanges))
	for _, byteRange := range byteRanges {
		matchRanges = append(matchRanges, titleMatchRange{
			start: utf8.RuneCountInString(title[:byteRange[0]]),
			end:   utf8.RuneCountInString(title[:byteRange[1]]),
		})
	}

	return matchRanges
}

func (program *Program) usesManualSelectedLineRendering(query string) bool {
	return strings.TrimSpace(query) != ""
}

func (program *Program) selectListLine(view *gocui.View, selectedIndex int, lineCount int) {
	if view == nil || lineCount < 1 {
		return
	}

	if placement, ok := program.consumePendingListViewportPlacement(view.Name()); ok {
		originY, cursorY := placedListLinePosition(selectedIndex, view.InnerHeight(), lineCount, placement)
		view.SetOrigin(0, originY)
		view.SetCursor(0, cursorY)
		return
	}

	_, currentOriginY := view.Origin()
	originY, cursorY := visibleListLinePosition(selectedIndex, currentOriginY, view.InnerHeight(), lineCount)
	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}

type viewportPlacement int

const (
	viewportPlacementTop viewportPlacement = iota
	viewportPlacementCenter
	viewportPlacementBottom
)

func (program *Program) placeListLine(view *gocui.View, selectedIndex int, lineCount int, placement viewportPlacement) {
	if view == nil || lineCount < 1 {
		return
	}

	program.setPendingListViewportPlacement(view.Name(), placement)
	originY, cursorY := placedListLinePosition(selectedIndex, view.InnerHeight(), lineCount, placement)
	view.SetOrigin(0, originY)
	view.SetCursor(0, cursorY)
}

func (program *Program) setPendingListViewportPlacement(viewName string, placement viewportPlacement) {
	if viewName == "" {
		return
	}
	if program.pendingListViewportPlacements == nil {
		program.pendingListViewportPlacements = map[string]viewportPlacement{}
	}
	program.pendingListViewportPlacements[viewName] = placement
}

func (program *Program) consumePendingListViewportPlacement(viewName string) (viewportPlacement, bool) {
	if viewName == "" || len(program.pendingListViewportPlacements) == 0 {
		return 0, false
	}

	placement, ok := program.pendingListViewportPlacements[viewName]
	if !ok {
		return 0, false
	}
	delete(program.pendingListViewportPlacements, viewName)
	return placement, true
}

func (program *Program) centerListLine(view *gocui.View, selectedIndex int, lineCount int) {
	program.placeListLine(view, selectedIndex, lineCount, viewportPlacementCenter)
}

func visibleListLinePosition(selectedIndex int, currentOriginY int, visibleHeight int, lineCount int) (int, int) {
	visibleHeight = maxInt(1, visibleHeight)
	lineCount = maxInt(1, lineCount)
	selectedIndex = clampIndex(selectedIndex, lineCount)
	originY := visibleViewportOrigin(selectedIndex, currentOriginY, visibleHeight, lineCount)
	return originY, selectedIndex - originY
}

func placedListLinePosition(selectedIndex int, visibleHeight int, lineCount int, placement viewportPlacement) (int, int) {
	visibleHeight = maxInt(1, visibleHeight)
	lineCount = maxInt(1, lineCount)
	selectedIndex = clampIndex(selectedIndex, lineCount)
	originY := placedViewportOrigin(selectedIndex, visibleHeight, lineCount, placement)
	return originY, selectedIndex - originY
}

func centeredViewportOrigin(selectedRow int, visibleHeight int, rowCount int) int {
	return placedViewportOrigin(selectedRow, visibleHeight, rowCount, viewportPlacementCenter)
}

func placedViewportOrigin(selectedRow int, visibleHeight int, rowCount int, placement viewportPlacement) int {
	visibleHeight = maxInt(1, visibleHeight)
	rowCount = maxInt(1, rowCount)
	selectedRow = clampIndex(selectedRow, rowCount)
	maxOriginY := maxInt(0, rowCount-visibleHeight)

	originY := selectedRow
	switch placement {
	case viewportPlacementCenter:
		originY = selectedRow - visibleHeight/2
	case viewportPlacementBottom:
		originY = selectedRow - (visibleHeight - 1)
	}

	return clampInt(originY, 0, maxOriginY)
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
