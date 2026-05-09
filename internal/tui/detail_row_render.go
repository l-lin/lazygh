package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func renderDetailRow(document detailDocument, row detailWrappedRow, searchMatchRanges map[int][]detailColumnRange, state detailViewState) string {
	lineMatchRanges := searchMatchRanges[row.line]
	rowImages := detailImagesOnRow(document.images, row.line)
	if row.empty && len(rowImages) == 0 {
		return ""
	}

	line := []rune(nil)
	if row.line >= 0 && row.line < len(document.lines) {
		line = document.lines[row.line]
	}
	lineStylePrefixes := []string(nil)
	if row.line >= 0 && row.line < len(document.lineStylePrefixes) {
		lineStylePrefixes = document.lineStylePrefixes[row.line]
	}
	rowRunes := []rune(nil)
	if !row.empty && len(line) > 0 {
		rowRunes = line[row.startColumn : row.endColumn+1]
	}
	paddingPrefix := markdownFullWidthLinePaddingPrefix(document.width, lineStylePrefixes, row.startColumn, row.endColumn)
	if len(rowImages) == 0 && len(lineMatchRanges) == 0 && !state.mode.isVisual() && !detailLineHasStylePrefixes(lineStylePrefixes, row.startColumn, row.endColumn) && paddingPrefix == "" {
		return row.text
	}

	visibleWidth := len(rowRunes)
	for _, image := range rowImages {
		imageEnd := image.column + image.columns
		if imageEnd <= row.startColumn {
			continue
		}
		visibleWidth = maxInt(visibleWidth, imageEnd-row.startColumn)
	}
	if paddingPrefix != "" {
		visibleWidth = maxInt(visibleWidth, document.width)
	}

	protocol := kittyImageProtocol{}
	var builder strings.Builder
	currentPrefix := ""
	for offset := 0; offset < visibleWidth; offset++ {
		column := row.startColumn + offset
		if imageCell, ok := detailImageCellAt(document.images, row.line, column); ok {
			prefix := detailCellStylePrefix(detailCellStyle{
				selected: state.isPositionSelected(document, detailPosition{line: row.line, column: column}),
				search:   detailColumnInRanges(column, lineMatchRanges),
			}, foregroundColorEscape(protocol.PlaceholderForegroundHex(imageCell.imageID)))
			if prefix != currentPrefix {
				if currentPrefix != "" {
					builder.WriteString(ansiReset)
				}
				if prefix != "" {
					builder.WriteString(prefix)
				}
				currentPrefix = prefix
			}
			builder.WriteString(protocol.PlaceholderCell(imageCell.imageID, imageCell.row, imageCell.column))
			continue
		}
		if offset < len(rowRunes) {
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
			builder.WriteRune(rowRunes[offset])
			continue
		}
		if paddingPrefix != "" {
			if paddingPrefix != currentPrefix {
				if currentPrefix != "" {
					builder.WriteString(ansiReset)
				}
				builder.WriteString(paddingPrefix)
				currentPrefix = paddingPrefix
			}
			builder.WriteRune('⠀')
		}
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

type detailImageCell struct {
	imageID uint32
	row     int
	column  int
}

func detailImagesOnRow(images []detailImagePlacement, line int) []detailImagePlacement {
	rowImages := make([]detailImagePlacement, 0, len(images))
	for _, image := range images {
		if line < image.line || line >= image.line+image.rows {
			continue
		}
		rowImages = append(rowImages, image)
	}
	return rowImages
}

func detailImageCellAt(images []detailImagePlacement, line int, column int) (detailImageCell, bool) {
	for _, image := range images {
		if line < image.line || line >= image.line+image.rows {
			continue
		}
		if column < image.column || column >= image.column+image.columns {
			continue
		}
		return detailImageCell{imageID: image.imageID, row: line - image.line, column: column - image.column}, true
	}

	return detailImageCell{}, false
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
