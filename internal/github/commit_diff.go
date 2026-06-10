package github

type CommitDiff struct {
	Files []PullRequestDiffFile `json:"files"`
}

func (diff CommitDiff) normalized() CommitDiff {
	if len(diff.Files) == 0 {
		return CommitDiff{}
	}

	normalizedFiles := make([]PullRequestDiffFile, 0, len(diff.Files))
	for _, file := range diff.Files {
		normalizedFiles = append(normalizedFiles, file.normalized())
	}
	diff.Files = normalizedFiles
	return diff
}
