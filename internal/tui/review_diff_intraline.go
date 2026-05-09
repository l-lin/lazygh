package tui

import "github.com/l-lin/lazygh/internal/theme"

func reviewDiffChangedStyleRanges(lines []reviewDiffLine) [][]styledRuneRange {
	rangesByLine := make([][]styledRuneRange, len(lines))
	for groupStart := 0; groupStart < len(lines); {
		if lines[groupStart].Kind == reviewDiffContextLine {
			groupStart++
			continue
		}

		groupEnd := groupStart
		deletionIndexes := make([]int, 0)
		additionIndexes := make([]int, 0)
		for groupEnd < len(lines) && lines[groupEnd].Kind != reviewDiffContextLine {
			switch lines[groupEnd].Kind {
			case reviewDiffDeletionLine:
				deletionIndexes = append(deletionIndexes, groupEnd)
			case reviewDiffAdditionLine:
				additionIndexes = append(additionIndexes, groupEnd)
			}
			groupEnd++
		}

		pairCount := minInt(len(deletionIndexes), len(additionIndexes))
		for pairIndex := range pairCount {
			deletionLineIndex := deletionIndexes[pairIndex]
			additionLineIndex := additionIndexes[pairIndex]
			deletionRanges, additionRanges := reviewDiffLineChangedStyleRanges(lines[deletionLineIndex].Text, lines[additionLineIndex].Text)
			rangesByLine[deletionLineIndex] = append(rangesByLine[deletionLineIndex], deletionRanges...)
			rangesByLine[additionLineIndex] = append(rangesByLine[additionLineIndex], additionRanges...)
		}

		groupStart = groupEnd
	}

	return rangesByLine
}

func reviewDiffLineChangedStyleRanges(deletionText string, additionText string) ([]styledRuneRange, []styledRuneRange) {
	deletionRunes := []rune(deletionText)
	additionRunes := []rune(additionText)
	commonPrefixLength := reviewDiffCommonPrefixLength(deletionRunes, additionRunes)
	commonSuffixLength := reviewDiffCommonSuffixLength(deletionRunes, additionRunes, commonPrefixLength)

	deletionEnd := len(deletionRunes) - commonSuffixLength
	additionEnd := len(additionRunes) - commonSuffixLength
	deletionRanges := make([]styledRuneRange, 0, 1)
	additionRanges := make([]styledRuneRange, 0, 1)
	if commonPrefixLength < deletionEnd {
		deletionRanges = append(deletionRanges, styledRuneRange{start: commonPrefixLength, end: deletionEnd, prefix: backgroundColorEscape(theme.DiffDeletionHighlightBackgroundHex)})
	}
	if commonPrefixLength < additionEnd {
		additionRanges = append(additionRanges, styledRuneRange{start: commonPrefixLength, end: additionEnd, prefix: backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex)})
	}
	return deletionRanges, additionRanges
}

func reviewDiffCommonPrefixLength(left []rune, right []rune) int {
	prefixLength := 0
	for prefixLength < len(left) && prefixLength < len(right) && left[prefixLength] == right[prefixLength] {
		prefixLength++
	}
	return prefixLength
}

func reviewDiffCommonSuffixLength(left []rune, right []rune, prefixLength int) int {
	suffixLength := 0
	leftLimit := len(left) - prefixLength
	rightLimit := len(right) - prefixLength
	for suffixLength < leftLimit && suffixLength < rightLimit {
		leftRuneIndex := len(left) - 1 - suffixLength
		rightRuneIndex := len(right) - 1 - suffixLength
		if left[leftRuneIndex] != right[rightRuneIndex] {
			break
		}
		suffixLength++
	}
	return suffixLength
}
