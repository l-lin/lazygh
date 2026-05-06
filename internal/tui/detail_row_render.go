package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func renderDetailRow(document detailDocument, row detailWrappedRow, searchMatchRanges map[int][]detailColumnRange, state detailViewState) string {
	if row.empty {
		return ""
	}

	line := document.lines[row.line]
	rowRunes := line[row.startColumn : row.endColumn+1]
	lineStylePrefixes := document.lineStylePrefixes[row.line]
	lineMatchRanges := searchMatchRanges[row.line]
	paddingPrefix := markdownFullWidthLinePaddingPrefix(document.width, lineStylePrefixes, row.startColumn, row.endColumn)
	if len(lineMatchRanges) == 0 && !state.mode.isVisual() && !detailLineHasStylePrefixes(lineStylePrefixes, row.startColumn, row.endColumn) && paddingPrefix == "" {
		return row.text
	}

	var builder strings.Builder
	currentPrefix := ""
	for offset, character := range rowRunes {
		column := row.startColumn + offset
		prefix := detailCellStylePrefix(detailCellStyle{
			selected: state.isPositionSelected(document, detailPosition{line: row.line, column: column}),
			search:   detailColumnInRanges(column, lineMatchRanges),
		}, detailLineStylePrefix(lineStylePrefixes, column))
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
	}
	for paddingWidth := len(rowRunes); paddingWidth < document.width && paddingPrefix != ""; paddingWidth++ {
		if paddingPrefix != currentPrefix {
			if currentPrefix != "" {
				builder.WriteString(ansiReset)
			}
			builder.WriteString(paddingPrefix)
			currentPrefix = paddingPrefix
		}
		builder.WriteRune('⠀')
	}
	if currentPrefix != "" {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

func detailCellStylePrefix(style detailCellStyle, basePrefix string) string {
	if style.selected {
		return basePrefix + ansiBold + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	}
	if style.search {
		return basePrefix + backgroundColorEscape(theme.SearchHighlightHex)
	}

	return basePrefix
}

func detailLineHasStylePrefixes(prefixes []string, startColumn int, endColumn int) bool {
	if len(prefixes) == 0 {
		return false
	}

	for column := startColumn; column <= endColumn && column < len(prefixes); column++ {
		if prefixes[column] != "" {
			return true
		}
	}

	return false
}

func detailLineStylePrefix(prefixes []string, column int) string {
	if column < 0 || column >= len(prefixes) {
		return ""
	}

	return prefixes[column]
}

func detailColumnInRanges(column int, ranges []detailColumnRange) bool {
	for _, matchRange := range ranges {
		if column >= matchRange.start && column < matchRange.end {
			return true
		}
	}

	return false
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
