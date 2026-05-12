package github

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrMissingPullRequestBuildLink    = errors.New("missing pull request build link")
	ErrInvalidPullRequestBuildLink    = errors.New("invalid pull request build link")
	ErrPullRequestBuildRunJobNotFound = errors.New("pull request build run job not found")

	buildRunIDPattern      = regexp.MustCompile(`(?i)(?:^|/)actions/runs/(\d+)(?:/|$)`)
	buildRunAttemptPattern = regexp.MustCompile(`(?i)/attempts/(\d+)(?:/|$)`)
	buildRunJobIDPattern   = regexp.MustCompile(`(?i)(?:^|/)job/(\d+)(?:/|$)`)
)

type BuildInfo struct {
	TypeName     string `json:"__typename"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
	Link         string `json:"link,omitempty"`
}

type PullRequestStatusCheck = BuildInfo

type BuildRunJob struct {
	DatabaseID int    `json:"databaseId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
}

type PullRequestBuildRunJob = BuildRunJob

type BuildRunReference struct {
	ID      string
	Attempt int
}

func (check BuildInfo) normalized() BuildInfo {
	check.TypeName = strings.TrimSpace(check.TypeName)
	check.Name = strings.TrimSpace(check.Name)
	check.Status = strings.TrimSpace(check.Status)
	check.Conclusion = strings.TrimSpace(check.Conclusion)
	check.WorkflowName = strings.TrimSpace(check.WorkflowName)
	check.Link = strings.TrimSpace(check.Link)
	return check
}

func (job BuildRunJob) normalized() BuildRunJob {
	job.Name = strings.TrimSpace(job.Name)
	job.Status = strings.TrimSpace(job.Status)
	job.Conclusion = strings.TrimSpace(job.Conclusion)
	job.URL = strings.TrimSpace(job.URL)
	return job
}

func ParseBuildRunReferenceFromURL(raw string) (BuildRunReference, error) {
	trimmedLink := strings.TrimSpace(raw)
	if trimmedLink == "" {
		return BuildRunReference{}, ErrMissingPullRequestBuildLink
	}

	path := BuildRunPathFromURL(trimmedLink)
	matches := buildRunIDPattern.FindStringSubmatch(path)
	if len(matches) < 2 {
		return BuildRunReference{}, ErrInvalidPullRequestBuildLink
	}

	reference := BuildRunReference{ID: strings.TrimSpace(matches[1])}
	if attemptMatches := buildRunAttemptPattern.FindStringSubmatch(path); len(attemptMatches) >= 2 {
		if attempt, err := strconv.Atoi(strings.TrimSpace(attemptMatches[1])); err == nil && attempt > 0 {
			reference.Attempt = attempt
		}
	}
	return reference, nil
}

func BuildRunJobIDFromURL(raw string) (int, bool) {
	matches := buildRunJobIDPattern.FindStringSubmatch(BuildRunPathFromURL(raw))
	if len(matches) < 2 {
		return 0, false
	}

	jobDatabaseID, err := strconv.Atoi(strings.TrimSpace(matches[1]))
	if err != nil || jobDatabaseID <= 0 {
		return 0, false
	}
	return jobDatabaseID, true
}

func BuildRunPathFromURL(raw string) string {
	trimmedLink := strings.TrimSpace(raw)
	path := trimmedLink
	if parsedURL, err := url.Parse(trimmedLink); err == nil && strings.TrimSpace(parsedURL.Path) != "" {
		path = parsedURL.Path
	}
	return path
}
