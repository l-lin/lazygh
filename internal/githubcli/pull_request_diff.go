package githubcli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidPullRequestDiffFilesResponse = fmt.Errorf("invalid pull request diff files response")

type PullRequestDiff struct {
	UnifiedDiff             string
	Files                   []PullRequestDiffFile
	Threads                 []PullRequestReviewThread
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

func (client *Client) GetPullRequestDiff(repository string, number int) (PullRequestDiff, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}

	unifiedDiff, err := client.getPullRequestUnifiedDiff(trimmedRepository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}

	files, err := client.listPullRequestDiffFiles(trimmedRepository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}

	threads, err := client.listPullRequestReviewThreads(trimmedRepository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}

	return PullRequestDiff{UnifiedDiff: unifiedDiff, Files: files, Threads: threads}, nil
}

func (client *Client) getPullRequestUnifiedDiff(repository string, number int) (string, error) {
	result, err := client.runGH(
		"gh api pull request diff",
		"api",
		fmt.Sprintf("repos/%s/pulls/%d", repository, number),
		"-H",
		"Accept: application/vnd.github.v3.diff",
	)
	if err != nil {
		return "", err
	}

	return normalizePullRequestDiffText(string(result.Stdout)), nil
}

func (client *Client) listPullRequestDiffFiles(repository string, number int) ([]PullRequestDiffFile, error) {
	result, err := client.runGH(
		"gh api pull request diff files",
		"api",
		fmt.Sprintf("repos/%s/pulls/%d/files?per_page=100", repository, number),
		"--paginate",
		"--slurp",
	)
	if err != nil {
		return nil, err
	}

	var pagedFiles [][]PullRequestDiffFile
	if err := json.Unmarshal(result.Stdout, &pagedFiles); err == nil {
		files := make([]PullRequestDiffFile, 0)
		for _, page := range pagedFiles {
			for _, file := range page {
				files = append(files, file.normalized())
			}
		}
		return files, nil
	}

	var files []PullRequestDiffFile
	if err := json.Unmarshal(result.Stdout, &files); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestDiffFilesResponse, err)
	}
	for index := range files {
		files[index] = files[index].normalized()
	}
	return files, nil
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

func pullRequestDiffKey(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return ""
	}
	return trimmedRepository + "#" + strconv.Itoa(number)
}
