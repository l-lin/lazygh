package github

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrInvalidPullRequestURL      = errors.New("invalid GitHub pull request URL")
	ErrMissingPullRequestIdentity = errors.New("missing pull request identity")
)

type RepositoryRef struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

type Repository = RepositoryRef

type PullRequestSummary struct {
	ID                       string                       `json:"id"`
	Title                    string                       `json:"title"`
	Number                   int                          `json:"number"`
	Repository               RepositoryRef                `json:"repository"`
	URL                      string                       `json:"url"`
	Body                     string                       `json:"body"`
	State                    string                       `json:"state"`
	IsDraft                  bool                         `json:"isDraft"`
	UpdatedAt                string                       `json:"updatedAt"`
	ReviewDecision           string                       `json:"reviewDecision"`
	ReviewRequests           []PullRequestReviewRequest   `json:"reviewRequests"`
	MergeStateStatus         string                       `json:"mergeStateStatus"`
	Mergeable                string                       `json:"mergeable"`
	AutoMergeRequest         *PullRequestAutoMergeRequest `json:"autoMergeRequest,omitempty"`
	IsMergeQueueEnabled      bool                         `json:"isMergeQueueEnabled,omitempty"`
	IsInMergeQueue           bool                         `json:"isInMergeQueue,omitempty"`
	MergeQueueEntry          *PullRequestMergeQueueEntry  `json:"mergeQueueEntry,omitempty"`
	ViewerCanEnableAutoMerge bool                         `json:"viewerCanEnableAutoMerge,omitempty"`
	StatusCheckRollupState   string                       `json:"statusCheckRollupState"`
}

type PullRequest = PullRequestSummary

func (repository *RepositoryRef) UnmarshalJSON(data []byte) error {
	var payload struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"nameWithOwner"`
		FullName      string `json:"full_name"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	repository.Name = payload.Name
	repository.NameWithOwner = payload.NameWithOwner
	if strings.TrimSpace(repository.NameWithOwner) == "" {
		repository.NameWithOwner = payload.FullName
	}
	return nil
}

type PullRequestReviewRequest struct {
	RequestedReviewer PullRequestRequestedReviewer `json:"requestedReviewer"`
}

func (reviewRequest *PullRequestReviewRequest) UnmarshalJSON(data []byte) error {
	var wrapped struct {
		RequestedReviewer *PullRequestRequestedReviewer `json:"requestedReviewer"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.RequestedReviewer != nil {
		reviewRequest.RequestedReviewer = wrapped.RequestedReviewer.normalized()
		return nil
	}

	var direct PullRequestRequestedReviewer
	if err := json.Unmarshal(data, &direct); err != nil {
		return err
	}
	reviewRequest.RequestedReviewer = direct.normalized()
	return nil
}

type PullRequestRequestedReviewer struct {
	TypeName     string                                `json:"__typename"`
	Login        string                                `json:"login"`
	Name         string                                `json:"name"`
	Slug         string                                `json:"slug"`
	Organization *PullRequestReviewRequestOrganization `json:"organization"`
}

type PullRequestReviewRequestOrganization struct {
	Login string `json:"login"`
}

func ParsePullRequestURL(raw string) (PullRequestSummary, error) {
	trimmedURL := strings.TrimSpace(raw)
	if trimmedURL == "" {
		return PullRequestSummary{}, ErrInvalidPullRequestURL
	}

	if !strings.Contains(trimmedURL, "://") {
		normalizedPrefix := strings.ToLower(trimmedURL)
		if strings.HasPrefix(normalizedPrefix, "github.com/") || strings.HasPrefix(normalizedPrefix, "www.github.com/") {
			trimmedURL = "https://" + trimmedURL
		}
	}

	parsedURL, err := url.Parse(trimmedURL)
	if err != nil {
		return PullRequestSummary{}, ErrInvalidPullRequestURL
	}

	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	switch host {
	case "github.com", "www.github.com":
	default:
		return PullRequestSummary{}, ErrInvalidPullRequestURL
	}

	pathSegments := pullRequestURLPathSegments(parsedURL.Path)
	if len(pathSegments) < 4 || (pathSegments[2] != "pull" && pathSegments[2] != "pulls") {
		return PullRequestSummary{}, ErrInvalidPullRequestURL
	}

	owner := pathSegments[0]
	repositoryName := pathSegments[1]
	pullRequestNumber, err := strconv.Atoi(pathSegments[3])
	if owner == "" || repositoryName == "" || err != nil || pullRequestNumber <= 0 {
		return PullRequestSummary{}, ErrInvalidPullRequestURL
	}

	repository := strings.TrimSpace(owner) + "/" + strings.TrimSpace(repositoryName)
	return PullRequestSummary{
		Number: pullRequestNumber,
		Repository: RepositoryRef{
			Name:          repositoryName,
			NameWithOwner: repository,
		},
		URL: CanonicalPullRequestURL(repository, pullRequestNumber),
	}, nil
}

func NormalizePullRequestIdentity(repository string, number int) (string, int, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return "", 0, ErrMissingPullRequestIdentity
	}

	return trimmedRepository, number, nil
}

func CanonicalPullRequestURL(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return ""
	}
	return "https://github.com/" + trimmedRepository + "/pull/" + strconv.Itoa(number)
}

func PullRequestCommitChangesURL(repository Repository, number int, commitOID string) (string, bool) {
	trimmedRepository := strings.TrimSpace(repository.NameWithOwner)
	trimmedCommitOID := strings.TrimSpace(commitOID)
	if trimmedRepository == "" || number <= 0 || trimmedCommitOID == "" {
		return "", false
	}
	return CanonicalPullRequestURL(trimmedRepository, number) + "/changes/" + trimmedCommitOID, true
}

func pullRequestURLPathSegments(path string) []string {
	rawSegments := strings.Split(strings.TrimSpace(path), "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		trimmedSegment := strings.TrimSpace(segment)
		if trimmedSegment == "" {
			continue
		}
		segments = append(segments, trimmedSegment)
	}
	return segments
}

func (repository RepositoryRef) normalized() RepositoryRef {
	repository.Name = strings.TrimSpace(repository.Name)
	repository.NameWithOwner = strings.TrimSpace(repository.NameWithOwner)
	return repository
}

func (reviewRequest PullRequestReviewRequest) normalized() PullRequestReviewRequest {
	reviewRequest.RequestedReviewer = reviewRequest.RequestedReviewer.normalized()
	return reviewRequest
}

func (reviewer PullRequestRequestedReviewer) normalized() PullRequestRequestedReviewer {
	reviewer.TypeName = strings.TrimSpace(reviewer.TypeName)
	reviewer.Login = strings.TrimSpace(reviewer.Login)
	reviewer.Name = strings.TrimSpace(reviewer.Name)
	reviewer.Slug = strings.TrimSpace(reviewer.Slug)
	if reviewer.Organization != nil {
		normalizedOrganization := reviewer.Organization.normalized()
		reviewer.Organization = &normalizedOrganization
	}
	return reviewer
}

func (organization PullRequestReviewRequestOrganization) normalized() PullRequestReviewRequestOrganization {
	organization.Login = strings.TrimSpace(organization.Login)
	return organization
}
