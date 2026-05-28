package githubcli

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
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

func (client *PullRequestDetailService) GetPullRequestDiff(repository string, number int) (PullRequestDiff, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return PullRequestDiff{}, err
	}
	return newPullRequestDiffAssembler(client).Assemble(trimmedRepository, number)
}

func (client *PullRequestDetailService) getPullRequestUnifiedDiff(repository string, number int) (string, error) {
	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/pulls/%d", repository, number), Headers: []RESTHeader{{Name: "Accept", Value: "application/vnd.github.v3.diff"}}})
	if err != nil {
		return "", err
	}

	return normalizePullRequestDiffText(string(result.Stdout)), nil
}

func (client *PullRequestDetailService) listPullRequestDiffFiles(repository string, number int) ([]PullRequestDiffFile, error) {
	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/pulls/%d/files?per_page=100", repository, number), Paginate: true, Slurp: true})
	if err != nil {
		return nil, err
	}

	return parsePullRequestDiffFilesResponse(result.Stdout)
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
	return githubdomain.NormalizePullRequestDiffFileTeamOwners(teamOwners)
}

func normalizePullRequestDiffText(text string) string {
	return githubdomain.NormalizePullRequestDiffText(text)
}
