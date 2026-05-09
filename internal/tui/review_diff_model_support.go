package tui

import (
	"fmt"
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func buildReviewDiffThreadsByPath(rawThreads []githubcli.PullRequestReviewThread) map[string][]reviewDiffThread {
	threadsByPath := make(map[string][]reviewDiffThread, len(rawThreads))
	for _, rawThread := range rawThreads {
		thread := buildReviewDiffThread(rawThread)
		if strings.TrimSpace(thread.Path) == "" {
			continue
		}
		threadsByPath[thread.Path] = append(threadsByPath[thread.Path], thread)
	}
	return threadsByPath
}

func buildReviewDiffThread(rawThread githubcli.PullRequestReviewThread) reviewDiffThread {
	return reviewDiffThread{
		ID:                strings.TrimSpace(rawThread.ID),
		Path:              strings.TrimSpace(rawThread.Path),
		IsResolved:        rawThread.IsResolved,
		IsOutdated:        rawThread.IsOutdated,
		Line:              rawThread.Line,
		OriginalLine:      rawThread.OriginalLine,
		StartLine:         rawThread.StartLine,
		OriginalStartLine: rawThread.OriginalStartLine,
		Side:              reviewDiffLineSideFromGitHub(rawThread.DiffSide),
		StartSide:         reviewDiffLineSideFromGitHub(rawThread.StartDiffSide),
		Comments:          append([]githubcli.PullRequestComment(nil), rawThread.Comments...),
	}
}

func reviewDiffLineSideFromGitHub(side string) reviewDiffLineSide {
	switch strings.ToUpper(strings.TrimSpace(side)) {
	case string(reviewDiffLineSideLeft):
		return reviewDiffLineSideLeft
	case string(reviewDiffLineSideRight):
		return reviewDiffLineSideRight
	default:
		return reviewDiffLineSideNone
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
