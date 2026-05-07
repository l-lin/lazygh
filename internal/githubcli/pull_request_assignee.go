package githubcli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

var ErrInvalidAssignableUsersResponse = fmt.Errorf("invalid assignable users response")

const assignableUsersPerPage = 100

func (client *Client) ListAssignableUsers(repository string) ([]PullRequestAuthor, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return nil, ErrMissingPullRequestIdentity
	}

	result, err := client.runGH(
		"gh api repos/{owner}/{repo}/assignees",
		"api",
		fmt.Sprintf("repos/%s/assignees?per_page=%d", trimmedRepository, assignableUsersPerPage),
		"--paginate",
		"--slurp",
	)
	if err != nil {
		return nil, err
	}

	return parseAssignableUsers(result.Stdout)
}

func (client *Client) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	normalizedAddLogins := normalizePullRequestAssigneeLogins(addLogins)
	normalizedRemoveLogins := normalizePullRequestAssigneeLogins(removeLogins)
	if len(normalizedAddLogins) == 0 && len(normalizedRemoveLogins) == 0 {
		return nil
	}

	args := []string{"pr", "edit", strconv.Itoa(number), "-R", trimmedRepository}
	if len(normalizedAddLogins) > 0 {
		args = append(args, "--add-assignee", strings.Join(normalizedAddLogins, ","))
	}
	if len(normalizedRemoveLogins) > 0 {
		args = append(args, "--remove-assignee", strings.Join(normalizedRemoveLogins, ","))
	}

	if _, err := client.runGH("gh pr edit", args...); err != nil {
		return err
	}

	return nil
}

func parseAssignableUsers(stdout []byte) ([]PullRequestAuthor, error) {
	var pagedUsers [][]PullRequestAuthor
	if err := json.Unmarshal(stdout, &pagedUsers); err == nil {
		flattenedUsers := make([]PullRequestAuthor, 0)
		for _, page := range pagedUsers {
			flattenedUsers = append(flattenedUsers, page...)
		}
		return normalizeAssignableUsers(flattenedUsers), nil
	}

	var flatUsers []PullRequestAuthor
	if err := json.Unmarshal(stdout, &flatUsers); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAssignableUsersResponse, err)
	}

	return normalizeAssignableUsers(flatUsers), nil
}

func normalizeAssignableUsers(users []PullRequestAuthor) []PullRequestAuthor {
	normalizedUsers := make([]PullRequestAuthor, 0, len(users))
	seenLogins := map[string]bool{}
	for _, user := range users {
		normalizedUser := user.normalized()
		if normalizedUser.Login == "" || seenLogins[normalizedUser.Login] {
			continue
		}
		seenLogins[normalizedUser.Login] = true
		normalizedUsers = append(normalizedUsers, normalizedUser)
	}
	sort.SliceStable(normalizedUsers, func(i int, j int) bool {
		return normalizedUsers[i].Login < normalizedUsers[j].Login
	})
	return normalizedUsers
}

func FormatAssignableUsersCommand(repository string) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return appconfig.FormatGHCommand([]string{"api"})
	}
	return appconfig.FormatGHCommand([]string{"api", fmt.Sprintf("repos/%s/assignees?per_page=%d", trimmedRepository, assignableUsersPerPage), "--paginate", "--slurp"})
}

func normalizePullRequestAssigneeLogins(logins []string) []string {
	normalizedLogins := make([]string, 0, len(logins))
	seenLogins := map[string]bool{}
	for _, login := range logins {
		trimmedLogin := strings.TrimSpace(login)
		if trimmedLogin == "" || seenLogins[trimmedLogin] {
			continue
		}
		seenLogins[trimmedLogin] = true
		normalizedLogins = append(normalizedLogins, trimmedLogin)
	}
	return normalizedLogins
}
