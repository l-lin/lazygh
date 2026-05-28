package githubcli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

var (
	ErrMissingPullRequestBuildLink    = githubdomain.ErrMissingPullRequestBuildLink
	ErrInvalidPullRequestBuildLink    = githubdomain.ErrInvalidPullRequestBuildLink
	ErrPullRequestBuildRunJobNotFound = errors.New("pull request build run job not found")
)

type pullRequestBuildRunReference struct {
	id      string
	attempt int
}

type PullRequestBuildRunJob struct {
	DatabaseID int    `json:"databaseId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	URL        string `json:"url"`
}

func (client *BuildService) GetPullRequestBuildRun(repository string, check PullRequestStatusCheck) (string, error) {
	args, err := pullRequestBuildRunCommandArguments(repository, check)
	if err != nil {
		return "", err
	}

	result, err := client.execute(rawCommand(args...))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}

func (client *BuildService) GetPullRequestBuildRunJobs(repository string, check PullRequestStatusCheck) ([]PullRequestBuildRunJob, error) {
	args, err := pullRequestBuildRunJobsCommandArguments(repository, check)
	if err != nil {
		return nil, err
	}

	result, err := client.execute(rawCommand(args...))
	if err != nil {
		return nil, err
	}

	var response struct {
		Jobs []PullRequestBuildRunJob `json:"jobs"`
	}
	if err := client.transport.decoder.DecodeJSON(result.Stdout, &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestBuildResponse, err)
	}

	jobs := make([]PullRequestBuildRunJob, 0, len(response.Jobs))
	for _, job := range response.Jobs {
		jobs = append(jobs, job.normalized())
	}
	return jobs, nil
}

func (client *BuildService) GetPullRequestBuildRunJobLog(repository string, jobDatabaseID int) (string, error) {
	args, err := pullRequestBuildRunJobLogCommandArguments(repository, jobDatabaseID)
	if err != nil {
		return "", err
	}

	result, err := client.execute(rawCommand(args...))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}

func (client *BuildService) GetPullRequestBuildRunJobLogForCheck(repository string, check PullRequestStatusCheck) (PullRequestBuildRunJob, string, error) {
	jobs, err := client.GetPullRequestBuildRunJobs(repository, check)
	if err != nil {
		return PullRequestBuildRunJob{}, "", err
	}

	job, ok := pullRequestBuildRunJobMatchingCheck(check, jobs)
	if !ok {
		return PullRequestBuildRunJob{}, "", ErrPullRequestBuildRunJobNotFound
	}

	logOutput, err := client.GetPullRequestBuildRunJobLog(repository, job.DatabaseID)
	if err != nil {
		return PullRequestBuildRunJob{}, "", err
	}
	return job, logOutput, nil
}

func FormatPullRequestBuildRunCommand(repository string, check PullRequestStatusCheck) string {
	args, err := pullRequestBuildRunCommandArguments(repository, check)
	if err != nil {
		return formatCommand("run", "view")
	}
	return formatCommandArguments(args)
}

func FormatPullRequestBuildRunJobsCommand(repository string, check PullRequestStatusCheck) string {
	args, err := pullRequestBuildRunJobsCommandArguments(repository, check)
	if err != nil {
		return formatCommand("run", "view")
	}
	return formatCommandArguments(args)
}

func FormatPullRequestBuildRunJobLogCommand(repository string, jobDatabaseID int) string {
	args, err := pullRequestBuildRunJobLogCommandArguments(repository, jobDatabaseID)
	if err != nil {
		return formatCommand("run", "view")
	}
	return formatCommandArguments(args)
}

func pullRequestBuildRunCommandArguments(repository string, check PullRequestStatusCheck) ([]string, error) {
	reference, trimmedRepository, err := pullRequestBuildRunCommandContext(repository, check)
	if err != nil {
		return nil, err
	}

	args := []string{"run", "view", reference.id, "-R", trimmedRepository}
	if reference.attempt > 0 {
		args = append(args, "--attempt", strconv.Itoa(reference.attempt))
	}
	args = append(args, "--verbose")
	return args, nil
}

func pullRequestBuildRunJobsCommandArguments(repository string, check PullRequestStatusCheck) ([]string, error) {
	reference, trimmedRepository, err := pullRequestBuildRunCommandContext(repository, check)
	if err != nil {
		return nil, err
	}

	args := []string{"run", "view", reference.id, "-R", trimmedRepository}
	if reference.attempt > 0 {
		args = append(args, "--attempt", strconv.Itoa(reference.attempt))
	}
	args = append(args, "--json", "jobs")
	return args, nil
}

func pullRequestBuildRunJobLogCommandArguments(repository string, jobDatabaseID int) ([]string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return nil, ErrMissingPullRequestIdentity
	}
	if jobDatabaseID <= 0 {
		return nil, ErrMissingPullRequestBuildLink
	}

	return []string{"run", "view", "--job=" + strconv.Itoa(jobDatabaseID), "--log", "--repo=" + trimmedRepository}, nil
}

func pullRequestBuildRunCommandContext(repository string, check PullRequestStatusCheck) (pullRequestBuildRunReference, string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return pullRequestBuildRunReference{}, "", ErrMissingPullRequestIdentity
	}

	reference, err := pullRequestBuildRunReferenceFromLink(check.Link)
	if err != nil {
		return pullRequestBuildRunReference{}, "", err
	}
	return reference, trimmedRepository, nil
}

func pullRequestBuildRunJobMatchingCheck(check PullRequestStatusCheck, jobs []PullRequestBuildRunJob) (PullRequestBuildRunJob, bool) {
	if len(jobs) == 0 {
		return PullRequestBuildRunJob{}, false
	}

	if jobDatabaseID, ok := pullRequestBuildRunJobIDFromLink(check.Link); ok {
		for _, job := range jobs {
			if job.DatabaseID == jobDatabaseID {
				return job, true
			}
		}
	}

	normalizedCheck := check.normalized()
	if normalizedCheck.Name != "" {
		for _, job := range jobs {
			if strings.EqualFold(strings.TrimSpace(job.Name), normalizedCheck.Name) {
				return job, true
			}
		}
	}

	if len(jobs) == 1 {
		return jobs[0], true
	}
	return PullRequestBuildRunJob{}, false
}

func (job PullRequestBuildRunJob) normalized() PullRequestBuildRunJob {
	job.Name = strings.TrimSpace(job.Name)
	job.Status = strings.TrimSpace(job.Status)
	job.Conclusion = strings.TrimSpace(job.Conclusion)
	job.URL = strings.TrimSpace(job.URL)
	return job
}

func pullRequestBuildRunReferenceFromLink(raw string) (pullRequestBuildRunReference, error) {
	reference, err := githubdomain.ParseBuildRunReferenceFromURL(raw)
	if err != nil {
		return pullRequestBuildRunReference{}, err
	}
	return pullRequestBuildRunReference{id: reference.ID, attempt: reference.Attempt}, nil
}

func pullRequestBuildRunJobIDFromLink(raw string) (int, bool) {
	return githubdomain.BuildRunJobIDFromURL(raw)
}

func pullRequestBuildRunPathFromLink(raw string) string {
	return githubdomain.BuildRunPathFromURL(raw)
}
