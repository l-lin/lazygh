package githubcli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const pullRequestBuildJSONFields = "bucket,completedAt,description,event,link,name,startedAt,state,workflow"

var (
	ErrInvalidPullRequestBuildResponse = errors.New("invalid pull request build response")
	ErrPullRequestBuildInfoNotFound    = errors.New("pull request build info not found")
)

type PullRequestBuildInfo struct {
	Bucket      string `json:"bucket"`
	CompletedAt string `json:"completedAt"`
	Description string `json:"description"`
	Event       string `json:"event"`
	Link        string `json:"link"`
	Name        string `json:"name"`
	StartedAt   string `json:"startedAt"`
	State       string `json:"state"`
	Workflow    string `json:"workflow"`
}

func (client *BuildService) GetPullRequestBuildInfo(repository string, number int, check PullRequestStatusCheck) (PullRequestBuildInfo, error) {
	buildInfos, err := client.listPullRequestBuildInfos(repository, number)
	if err != nil {
		return PullRequestBuildInfo{}, err
	}

	actual, ok := pullRequestBuildInfoMatchingCheck(check, buildInfos)
	if !ok {
		return PullRequestBuildInfo{}, ErrPullRequestBuildInfoNotFound
	}
	return actual, nil
}

func (client serviceBase) listPullRequestBuildInfos(repository string, number int) ([]PullRequestBuildInfo, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return nil, err
	}

	result, err := client.execute(rawCommand("pr", "checks", strconv.Itoa(number), "-R", trimmedRepository, "--json", pullRequestBuildJSONFields))
	if err != nil {
		return nil, err
	}

	var buildInfos []PullRequestBuildInfo
	if err := client.transport.decoder.DecodeJSON(result.Stdout, &buildInfos); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestBuildResponse, err)
	}

	for index := range buildInfos {
		buildInfos[index] = buildInfos[index].normalized()
	}
	return buildInfos, nil
}

func mergePullRequestStatusCheckLinks(checks []PullRequestStatusCheck, buildInfos []PullRequestBuildInfo) []PullRequestStatusCheck {
	if len(checks) == 0 {
		return nil
	}
	if len(buildInfos) == 0 {
		merged := make([]PullRequestStatusCheck, 0, len(checks))
		for _, check := range checks {
			merged = append(merged, check.normalized())
		}
		return merged
	}

	merged := make([]PullRequestStatusCheck, 0, len(checks))
	usedBuilds := make([]bool, len(buildInfos))
	for _, rawCheck := range checks {
		check := rawCheck.normalized()
		for index, buildInfo := range buildInfos {
			if usedBuilds[index] || !samePullRequestBuildIdentity(check, buildInfo) {
				continue
			}
			check.Link = buildInfo.Link
			usedBuilds[index] = true
			break
		}
		merged = append(merged, check)
	}
	return merged
}

func pullRequestBuildInfoMatchingCheck(check PullRequestStatusCheck, buildInfos []PullRequestBuildInfo) (PullRequestBuildInfo, bool) {
	normalizedCheck := check.normalized()
	trimmedLink := strings.TrimSpace(normalizedCheck.Link)
	if trimmedLink != "" {
		for _, buildInfo := range buildInfos {
			if strings.EqualFold(strings.TrimSpace(buildInfo.Link), trimmedLink) {
				return buildInfo, true
			}
		}
	}

	for _, buildInfo := range buildInfos {
		if samePullRequestBuildIdentity(normalizedCheck, buildInfo) {
			return buildInfo, true
		}
	}
	return PullRequestBuildInfo{}, false
}

func samePullRequestBuildIdentity(check PullRequestStatusCheck, buildInfo PullRequestBuildInfo) bool {
	return strings.EqualFold(strings.TrimSpace(check.Name), strings.TrimSpace(buildInfo.Name)) && strings.EqualFold(strings.TrimSpace(check.WorkflowName), strings.TrimSpace(buildInfo.Workflow))
}

func (buildInfo PullRequestBuildInfo) normalized() PullRequestBuildInfo {
	buildInfo.Bucket = strings.TrimSpace(buildInfo.Bucket)
	buildInfo.CompletedAt = strings.TrimSpace(buildInfo.CompletedAt)
	buildInfo.Description = strings.TrimSpace(buildInfo.Description)
	buildInfo.Event = strings.TrimSpace(buildInfo.Event)
	buildInfo.Link = strings.TrimSpace(buildInfo.Link)
	buildInfo.Name = strings.TrimSpace(buildInfo.Name)
	buildInfo.StartedAt = strings.TrimSpace(buildInfo.StartedAt)
	buildInfo.State = strings.TrimSpace(buildInfo.State)
	buildInfo.Workflow = strings.TrimSpace(buildInfo.Workflow)
	return buildInfo
}
