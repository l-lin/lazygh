package tui

import (
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

type reviewDiffThread struct {
	ID                string
	Path              string
	IsResolved        bool
	IsOutdated        bool
	Line              int
	OriginalLine      int
	StartLine         int
	OriginalStartLine int
	Side              reviewDiffLineSide
	StartSide         reviewDiffLineSide
	Comments          []githubcli.PullRequestComment
}

type reviewDiffFile struct {
	Path         string
	PreviousPath string
	ChangeType   reviewDiffChangeType
	Additions    int
	Deletions    int
	Hunks        []reviewDiffHunk
	Threads      []reviewDiffThread
	Placeholder  string
}

type reviewDiffStats struct {
	ChangedFiles int
	Additions    int
	Deletions    int
}

type reviewDiffTreeRowKind int

const (
	reviewDiffTreeRowKindDirectory reviewDiffTreeRowKind = iota
	reviewDiffTreeRowKindFile
	reviewDiffTreeRowKindChapter
)

type reviewDiffTreeRow struct {
	ID              string
	VisibleRowIndex int
	Depth           int
	Label           string
	FileIndex       int
	ChapterIndex    int
	Kind            reviewDiffTreeRowKind
	Foldable        bool
	Collapsed       bool
}

type reviewDiffTree struct {
	Rows []reviewDiffTreeRow
}

type reviewDiffData struct {
	Stats    reviewDiffStats
	Files    []reviewDiffFile
	FileTree reviewDiffTree
}

type reviewDiffThreadTarget = githubcli.PullRequestReviewThreadTarget

func buildReviewDiffData(raw githubcli.PullRequestDiff) reviewDiffData {
	parsedFiles := parseUnifiedReviewDiff(raw.UnifiedDiff)
	parsedFilesByPath := make(map[string]reviewDiffFile, len(parsedFiles))
	for _, file := range parsedFiles {
		parsedFilesByPath[file.Path] = file
	}

	threadsByPath := buildReviewDiffThreadsByPath(raw.Threads)
	files := make([]reviewDiffFile, 0, maxInt(len(raw.Files), len(parsedFiles))+len(threadsByPath))
	usedPaths := make(map[string]bool, len(raw.Files)+len(parsedFiles))
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
		}
		file.Threads = append(file.Threads, threadsByPath[file.Path]...)
		if len(file.Hunks) == 0 {
			file.Placeholder = reviewDiffPlaceholder(file)
		}
		files = append(files, file)
		usedPaths[file.Path] = true
	}

	for _, parsedFile := range parsedFiles {
		if usedPaths[parsedFile.Path] {
			continue
		}
		parsedFile.Threads = append(parsedFile.Threads, threadsByPath[parsedFile.Path]...)
		if len(parsedFile.Hunks) == 0 {
			parsedFile.Placeholder = reviewDiffPlaceholder(parsedFile)
		}
		files = append(files, parsedFile)
		usedPaths[parsedFile.Path] = true
	}

	for path, threads := range threadsByPath {
		if usedPaths[path] || strings.TrimSpace(path) == "" {
			continue
		}
		file := reviewDiffFile{Path: path, Threads: append([]reviewDiffThread(nil), threads...)}
		file.Placeholder = reviewDiffPlaceholder(file)
		files = append(files, file)
		usedPaths[path] = true
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
