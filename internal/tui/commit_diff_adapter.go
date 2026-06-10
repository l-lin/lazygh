package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func buildCommitDiffReviewData(diff githubdomain.CommitDiff) reviewDiffData {
	return buildReviewDiffData(githubdomain.PullRequestDiff{
		UnifiedDiff: buildCommitDiffUnifiedDiff(diff.Files),
		Files:       append([]githubdomain.PullRequestDiffFile(nil), diff.Files...),
	})
}

func buildCommitDiffUnifiedDiff(files []githubdomain.PullRequestDiffFile) string {
	sections := make([]string, 0, len(files))
	for _, file := range files {
		section := buildCommitDiffUnifiedDiffSection(file)
		if strings.TrimSpace(section) == "" {
			continue
		}
		sections = append(sections, section)
	}
	return strings.Join(sections, "\n\n")
}

func buildCommitDiffUnifiedDiffSection(file githubdomain.PullRequestDiffFile) string {
	patch := githubdomain.NormalizePullRequestDiffText(file.Patch)
	if patch == "" {
		return ""
	}

	path := strings.TrimSpace(file.Path)
	previousPath := strings.TrimSpace(file.PreviousPath)
	if previousPath == "" {
		previousPath = path
	}
	if path == "" {
		path = previousPath
	}

	lines := []string{"diff --git a/" + previousPath + " b/" + path}
	switch reviewDiffChangeType(strings.ToLower(strings.TrimSpace(file.ChangeType))) {
	case reviewDiffChangeTypeAdded:
		lines = append(lines, "new file mode 100644", "--- /dev/null", "+++ b/"+path)
	case reviewDiffChangeTypeRemoved:
		lines = append(lines, "deleted file mode 100644", "--- a/"+previousPath, "+++ /dev/null")
	case reviewDiffChangeTypeRenamed:
		lines = append(lines, "rename from "+previousPath, "rename to "+path, "--- a/"+previousPath, "+++ b/"+path)
	default:
		lines = append(lines, "--- a/"+previousPath, "+++ b/"+path)
	}
	lines = append(lines, patch)
	return strings.Join(lines, "\n")
}
