package github

import "strings"

type PullRequestDiff struct {
	UnifiedDiff             string
	Files                   []PullRequestDiffFile
	Threads                 []ReviewThread
	FileTeamOwnersAttempted bool `json:"fileTeamOwnersAttempted,omitempty"`
}

type PullRequestDiffFile struct {
	Path         string   `json:"filename"`
	PreviousPath string   `json:"previous_filename"`
	ChangeType   string   `json:"status"`
	Additions    int      `json:"additions"`
	Deletions    int      `json:"deletions"`
	Patch        string   `json:"patch"`
	TeamOwners   []string `json:"teamOwners,omitempty"`
}

func (file PullRequestDiffFile) normalized() PullRequestDiffFile {
	file.Path = strings.TrimSpace(file.Path)
	file.PreviousPath = strings.TrimSpace(file.PreviousPath)
	file.ChangeType = strings.ToLower(strings.TrimSpace(file.ChangeType))
	file.Patch = normalizePullRequestDiffText(file.Patch)
	file.TeamOwners = normalizePullRequestDiffFileTeamOwners(file.TeamOwners)
	return file
}

func normalizePullRequestDiffFileTeamOwners(teamOwners []string) []string {
	if len(teamOwners) == 0 {
		return nil
	}

	normalizedOwners := make([]string, 0, len(teamOwners))
	seenOwners := map[string]bool{}
	for _, teamOwner := range teamOwners {
		trimmedTeamOwner := strings.TrimSpace(teamOwner)
		if trimmedTeamOwner == "" || seenOwners[trimmedTeamOwner] {
			continue
		}
		seenOwners[trimmedTeamOwner] = true
		normalizedOwners = append(normalizedOwners, trimmedTeamOwner)
	}
	if len(normalizedOwners) == 0 {
		return nil
	}
	return normalizedOwners
}

func normalizePullRequestDiffText(text string) string {
	normalizedText := strings.ReplaceAll(text, "\r\n", "\n")
	normalizedText = strings.ReplaceAll(normalizedText, "\r", "\n")
	return strings.TrimRight(normalizedText, "\n")
}
