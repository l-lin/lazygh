package tui

import "strings"

func reviewDiffThreadTargetForLines(path string, selectedLines []reviewDiffLine) (reviewDiffThreadTarget, bool) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" || len(selectedLines) == 0 {
		return reviewDiffThreadTarget{}, false
	}

	side, ok := reviewDiffThreadTargetSide(selectedLines)
	if !ok {
		return reviewDiffThreadTarget{}, false
	}

	startLine := 0
	endLine := 0
	for _, line := range selectedLines {
		lineNumber := reviewDiffLineNumberForSide(line, side)
		if lineNumber <= 0 {
			return reviewDiffThreadTarget{}, false
		}
		if startLine == 0 {
			startLine = lineNumber
		}
		endLine = lineNumber
	}

	target := reviewDiffThreadTarget{
		Path:        trimmedPath,
		Line:        endLine,
		Side:        string(side),
		SubjectType: "LINE",
	}
	if startLine != 0 && startLine != endLine {
		target.StartLine = startLine
		target.StartSide = string(side)
	}
	return target, true
}

func reviewDiffThreadTargetSide(selectedLines []reviewDiffLine) (reviewDiffLineSide, bool) {
	hasAddition := false
	hasDeletion := false
	for _, line := range selectedLines {
		if !line.isAnchorable() {
			return reviewDiffLineSideNone, false
		}
		switch line.Kind {
		case reviewDiffAdditionLine:
			hasAddition = true
		case reviewDiffDeletionLine:
			hasDeletion = true
		}
	}
	if hasAddition && hasDeletion {
		return reviewDiffLineSideNone, false
	}
	if hasDeletion {
		return reviewDiffLineSideLeft, reviewDiffLinesSupportSide(selectedLines, reviewDiffLineSideLeft)
	}
	return reviewDiffLineSideRight, reviewDiffLinesSupportSide(selectedLines, reviewDiffLineSideRight)
}

func reviewDiffLinesSupportSide(lines []reviewDiffLine, side reviewDiffLineSide) bool {
	for _, line := range lines {
		if !line.supportsSide(side) {
			return false
		}
	}
	return true
}

func reviewDiffLineNumberForSide(line reviewDiffLine, side reviewDiffLineSide) int {
	switch side {
	case reviewDiffLineSideLeft:
		return line.LeftLine
	default:
		return line.RightLine
	}
}

func (line reviewDiffLine) isAnchorable() bool {
	return line.Side != reviewDiffLineSideNone
}

func (line reviewDiffLine) supportsSide(side reviewDiffLineSide) bool {
	switch line.Side {
	case reviewDiffLineSideBoth:
		return side == reviewDiffLineSideLeft || side == reviewDiffLineSideRight
	case reviewDiffLineSideLeft:
		return side == reviewDiffLineSideLeft
	case reviewDiffLineSideRight:
		return side == reviewDiffLineSideRight
	default:
		return false
	}
}
