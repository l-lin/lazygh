package tui

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type detailSearchMatch struct {
	start     detailPosition
	endColumn int
}

func (document detailDocument) searchMatches(query string) []detailSearchMatch {
	pattern := detailSearchPattern(query)
	if pattern == nil {
		return nil
	}

	matches := make([]detailSearchMatch, 0)
	for lineIndex, line := range document.lines {
		lineText := string(line)
		for _, match := range pattern.FindAllStringIndex(lineText, -1) {
			startColumn := utf8.RuneCountInString(lineText[:match[0]])
			endColumn := startColumn + utf8.RuneCountInString(lineText[match[0]:match[1]])
			matches = append(matches, detailSearchMatch{
				start:     detailPosition{line: lineIndex, column: startColumn},
				endColumn: endColumn,
			})
		}
	}

	return matches
}

func detailSearchPattern(query string) *regexp.Regexp {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil
	}

	return regexp.MustCompile(`(?i)` + regexp.QuoteMeta(trimmedQuery))
}

func detailSearchMatchRanges(matches []detailSearchMatch) map[int][]detailColumnRange {
	if len(matches) == 0 {
		return nil
	}

	matchRanges := make(map[int][]detailColumnRange)
	for _, match := range matches {
		matchRanges[match.start.line] = append(matchRanges[match.start.line], detailColumnRange{start: match.start.column, end: match.endColumn})
	}

	return matchRanges
}

func (state *detailViewState) syncSearch(document detailDocument, query string) {
	state.searchMatches = document.searchMatches(query)
	state.currentSearchMatch = state.searchMatchIndexAtCursor(document)
}

func (state detailViewState) searchMatchIndexAtCursor(document detailDocument) int {
	cursor := document.clampPosition(state.cursor)
	for matchIndex, match := range state.searchMatches {
		if document.comparePositions(match.start, cursor) == 0 {
			return matchIndex
		}
	}

	return -1
}

func (state *detailViewState) followSubmittedSearch(document detailDocument, query string, viewportHeight int) bool {
	state.syncSearch(document, query)
	return state.followSearchMatchByIndex(document, viewportHeight, state.searchMatchIndexAtOrAfterCursor(document))
}

func (state *detailViewState) followNextSearchMatch(document detailDocument, query string, viewportHeight int) bool {
	state.syncSearch(document, query)
	return state.followSearchMatchByIndex(document, viewportHeight, state.searchMatchIndexAfterCursor(document))
}

func (state *detailViewState) followPreviousSearchMatch(document detailDocument, query string, viewportHeight int) bool {
	state.syncSearch(document, query)
	return state.followSearchMatchByIndex(document, viewportHeight, state.searchMatchIndexBeforeCursor(document))
}

func (state detailViewState) searchMatchIndexAtOrAfterCursor(document detailDocument) int {
	if len(state.searchMatches) == 0 {
		return -1
	}

	cursor := document.clampPosition(state.cursor)
	for matchIndex, match := range state.searchMatches {
		if document.comparePositions(match.start, cursor) >= 0 {
			return matchIndex
		}
	}

	return 0
}

func (state detailViewState) searchMatchIndexAfterCursor(document detailDocument) int {
	if len(state.searchMatches) == 0 {
		return -1
	}

	cursor := document.clampPosition(state.cursor)
	for matchIndex, match := range state.searchMatches {
		if document.comparePositions(match.start, cursor) > 0 {
			return matchIndex
		}
	}

	return 0
}

func (state detailViewState) searchMatchIndexBeforeCursor(document detailDocument) int {
	if len(state.searchMatches) == 0 {
		return -1
	}

	cursor := document.clampPosition(state.cursor)
	for matchIndex := len(state.searchMatches) - 1; matchIndex >= 0; matchIndex-- {
		if document.comparePositions(state.searchMatches[matchIndex].start, cursor) < 0 {
			return matchIndex
		}
	}

	return len(state.searchMatches) - 1
}

func (state *detailViewState) followSearchMatchByIndex(document detailDocument, viewportHeight int, matchIndex int) bool {
	if matchIndex < 0 || matchIndex >= len(state.searchMatches) {
		state.currentSearchMatch = -1
		return false
	}

	state.clearPendingPrefix()
	state.cursor = state.searchMatches[matchIndex].start
	state.currentSearchMatch = matchIndex
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
	return true
}
