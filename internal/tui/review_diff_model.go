package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

type reviewDiffChangeType string

const (
	reviewDiffChangeTypeModified reviewDiffChangeType = "modified"
	reviewDiffChangeTypeAdded    reviewDiffChangeType = "added"
	reviewDiffChangeTypeRemoved  reviewDiffChangeType = "removed"
	reviewDiffChangeTypeRenamed  reviewDiffChangeType = "renamed"
)

type reviewDiffLineKind int

const (
	reviewDiffContextLine reviewDiffLineKind = iota
	reviewDiffDeletionLine
	reviewDiffAdditionLine
)

type reviewDiffLineSide string

const (
	reviewDiffLineSideNone  reviewDiffLineSide = ""
	reviewDiffLineSideLeft  reviewDiffLineSide = "LEFT"
	reviewDiffLineSideRight reviewDiffLineSide = "RIGHT"
	reviewDiffLineSideBoth  reviewDiffLineSide = "BOTH"
)

type reviewDiffLine struct {
	Kind      reviewDiffLineKind
	Text      string
	LeftLine  int
	RightLine int
	Side      reviewDiffLineSide
}

type reviewDiffHunk struct {
	Header string
	Lines  []reviewDiffLine
}

type reviewDiffFile struct {
	Path         string
	PreviousPath string
	ChangeType   reviewDiffChangeType
	Additions    int
	Deletions    int
	Hunks        []reviewDiffHunk
	Placeholder  string
}

type reviewDiffStats struct {
	ChangedFiles int
	Additions    int
	Deletions    int
}

type reviewDiffTreeRow struct {
	VisibleRowIndex int
	Depth           int
	Label           string
	FileIndex       int
}

type reviewDiffTree struct {
	Rows []reviewDiffTreeRow
}

type reviewDiffData struct {
	Stats    reviewDiffStats
	Files    []reviewDiffFile
	FileTree reviewDiffTree
}

type reviewDiffThreadTarget struct {
	Path        string
	Line        int
	Side        string
	StartLine   int
	StartSide   string
	SubjectType string
}

func buildReviewDiffData(raw githubcli.PullRequestDiff) reviewDiffData {
	parsedFiles := parseUnifiedReviewDiff(raw.UnifiedDiff)
	parsedFilesByPath := make(map[string]reviewDiffFile, len(parsedFiles))
	for _, file := range parsedFiles {
		parsedFilesByPath[file.Path] = file
	}

	files := make([]reviewDiffFile, 0, max(len(raw.Files), len(parsedFiles)))
	usedPaths := make(map[string]bool, len(raw.Files))
	for _, rawFile := range raw.Files {
		file := reviewDiffFile{
			Path:         strings.TrimSpace(rawFile.Path),
			PreviousPath: strings.TrimSpace(rawFile.PreviousPath),
			ChangeType:   reviewDiffChangeType(strings.ToLower(strings.TrimSpace(rawFile.ChangeType))),
			Additions:    rawFile.Additions,
			Deletions:    rawFile.Deletions,
		}
		if parsedFile, ok := parsedFilesByPath[file.Path]; ok {
			file.Hunks = parsedFile.Hunks
			if file.PreviousPath == "" {
				file.PreviousPath = parsedFile.PreviousPath
			}
			if file.ChangeType == "" {
				file.ChangeType = parsedFile.ChangeType
			}
			usedPaths[file.Path] = true
		}
		if len(file.Hunks) == 0 {
			file.Placeholder = reviewDiffPlaceholder(file)
		}
		files = append(files, file)
	}

	for _, parsedFile := range parsedFiles {
		if usedPaths[parsedFile.Path] {
			continue
		}
		if len(parsedFile.Hunks) == 0 {
			parsedFile.Placeholder = reviewDiffPlaceholder(parsedFile)
		}
		files = append(files, parsedFile)
	}

	stats := reviewDiffStats{ChangedFiles: len(files)}
	for _, file := range files {
		stats.Additions += file.Additions
		stats.Deletions += file.Deletions
	}

	return reviewDiffData{
		Stats:    stats,
		Files:    files,
		FileTree: buildReviewDiffFileTree(files),
	}
}

func reviewDiffPlaceholder(file reviewDiffFile) string {
	path := valueOrDash(strings.TrimSpace(file.Path))
	commonMessage := "GitHub did not return a textual patch for this file. It may be binary or too large."

	switch file.ChangeType {
	case reviewDiffChangeTypeRenamed:
		return strings.Join([]string{
			fmt.Sprintf("Renamed from %s to %s.", valueOrDash(strings.TrimSpace(file.PreviousPath)), path),
			"",
			commonMessage,
		}, "\n")
	case reviewDiffChangeTypeRemoved:
		return strings.Join([]string{
			fmt.Sprintf("Deleted file %s.", path),
			"",
			commonMessage,
		}, "\n")
	case reviewDiffChangeTypeAdded:
		return strings.Join([]string{
			fmt.Sprintf("Added file %s.", path),
			"",
			commonMessage,
		}, "\n")
	default:
		return strings.Join([]string{
			fmt.Sprintf("No textual diff is available for %s.", path),
			"",
			commonMessage,
		}, "\n")
	}
}

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
