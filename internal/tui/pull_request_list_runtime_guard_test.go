package tui

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenPullRequestListWorkflowFiles_WhenScanning_ThenNoThinProgramListWrapperRemains(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.listPullRequests`,
		`return program\.pullRequestListQueries\.ListPullRequests`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "workflow_pull_request_list_commands.go" || base == "program_loading.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected the pull-request list workflow to avoid the thin Program list wrapper, actual %v", actualMatches)
	}
}
