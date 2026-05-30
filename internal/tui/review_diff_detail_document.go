package tui

func newReviewDiffDetailDocument(rows []reviewDiffRenderedRow, width int) detailDocument {
	return newReviewDiffDetailDocumentWithWordWrap(rows, width, false)
}

func newReviewDiffDetailDocumentWithWordWrap(rows []reviewDiffRenderedRow, width int, wordWrapEnabled bool) detailDocument {
	lines := make([]detailDocumentLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, reviewDiffDetailDocumentLine(row))
	}
	return newDetailDocumentFromLines(lines, width, wordWrapEnabled)
}

func reviewDiffDetailDocumentLine(row reviewDiffRenderedRow) detailDocumentLine {
	styledLine := firstStyledTextLineFromText(row.Text)
	if row.Kind != reviewDiffRenderedRowKindDiffLine {
		return detailDocumentLine{body: styledLine}
	}

	prefix, body := splitStyledTextLine(styledLine, reviewDiffStyledLineBodyStartColumn(styledLine))
	return detailDocumentLine{prefix: prefix, body: body}
}

func firstStyledTextLineFromText(text string) styledTextLine {
	lines := parseStyledTextLines(text)
	if len(lines) == 0 {
		return styledTextLine{}
	}
	return cloneStyledTextLine(lines[0])
}

func splitStyledTextLine(line styledTextLine, splitColumn int) (styledTextLine, styledTextLine) {
	clampedSplitColumn := clampInt(splitColumn, 0, len(line.runes))
	prefix := styledTextLine{
		runes:            append([]rune(nil), line.runes[:clampedSplitColumn]...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes[:minInt(clampedSplitColumn, len(line.stylePrefixes))]...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets[:minInt(clampedSplitColumn, len(line.hyperlinkTargets))]...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	body := styledTextLine{
		runes:            append([]rune(nil), line.runes[clampedSplitColumn:]...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes[minInt(clampedSplitColumn, len(line.stylePrefixes)):]...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets[minInt(clampedSplitColumn, len(line.hyperlinkTargets)):]...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	for _, control := range line.controls {
		switch {
		case control.column < clampedSplitColumn:
			prefix.controls = append(prefix.controls, styledTextControl{column: control.column, image: control.image})
		default:
			body.controls = append(body.controls, styledTextControl{column: max(control.column-clampedSplitColumn, 0), image: control.image})
		}
	}
	return prefix, body
}

func reviewDiffStyledLineBodyStartColumn(line styledTextLine) int {
	for index, character := range line.runes {
		if character != '│' {
			continue
		}
		return minInt(index+2, len(line.runes))
	}
	return 0
}
