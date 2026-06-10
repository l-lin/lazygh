package githubcli

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

var ErrInvalidCommitDiffResponse = fmt.Errorf("invalid commit diff response")

type CommitDiff struct {
	Files []PullRequestDiffFile `json:"files"`
}

func (client *PullRequestDetailService) GetCommitDiff(repository string, commitOID string) (CommitDiff, error) {
	trimmedRepository := strings.TrimSpace(repository)
	trimmedCommitOID := strings.TrimSpace(commitOID)
	if trimmedRepository == "" {
		return CommitDiff{}, ErrMissingPullRequestIdentity
	}
	if trimmedCommitOID == "" {
		return CommitDiff{}, ErrInvalidCommitDiffResponse
	}

	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/commits/%s", trimmedRepository, trimmedCommitOID)})
	if err != nil {
		return CommitDiff{}, err
	}
	return parseCommitDiffResponse(result.Stdout)
}

func parseCommitDiffResponse(stdout []byte) (CommitDiff, error) {
	var response struct {
		Files []PullRequestDiffFile `json:"files"`
	}
	if err := decodeEndpointJSONResponse(stdout, &response, ErrInvalidCommitDiffResponse); err != nil {
		return CommitDiff{}, err
	}

	normalizedFiles := make([]PullRequestDiffFile, 0, len(response.Files))
	for _, file := range response.Files {
		normalizedFiles = append(normalizedFiles, file.normalized())
	}
	return CommitDiff{Files: normalizedFiles}, nil
}

func ToDomainCommitDiff(diff CommitDiff) githubdomain.CommitDiff {
	return githubdomain.CommitDiff{Files: toDomainPullRequestDiffFiles(diff.Files)}
}

func CommitDiffFromDomain(diff githubdomain.CommitDiff) CommitDiff {
	return CommitDiff{Files: pullRequestDiffFilesFromDomain(diff.Files)}
}
