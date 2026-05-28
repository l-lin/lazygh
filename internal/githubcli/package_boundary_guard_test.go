package githubcli

import "testing"

func TestArchitectureGuard_GivenGithubcliFiles_WhenScanning_ThenTheyDoNotDependOnTUIOrCache(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGithubcliSourceGoFiles(t, ".", []string{
		"github.com/l-lin/lazygh/internal/tui",
		"github.com/l-lin/lazygh/internal/cache",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected githubcli to stay adapter-only and not depend on tui or cache, actual %v", actualMatches)
	}
}

func TestArchitectureGuard_GivenGithubcliFiles_WhenScanning_ThenProviderNeutralHelperRulesStayInGithubDomain(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGithubcliSourceGoFiles(t, ".", []string{
		"func ParsePullRequestURL(",
		"func (notification Notification) PullRequestSummary(",
		"func (notification Notification) IssueIdentity(",
		"func (notification Notification) ReleaseIdentity(",
		"if trimmedRepository == \"\" || trimmedRepository == \"-\" || number <= 0 {",
		"if trimmedRepository == \"\" || trimmedRepository == \"-\" || id <= 0 {",
		"fmt.Sprintf(\"https://github.com/%s/pull/%d\"",
		"switch strings.ToUpper(trimmedValue) {",
		"normalizedEvent := PullRequestReviewEvent(strings.ToUpper(strings.TrimSpace(string(event))))",
		"normalized := PullRequestReviewThreadTarget{",
		"pullRequestBuildRunIDPattern",
		"strings.ReplaceAll(text, \"\\r\\n\", \"\\n\")",
		"seenOwners := map[string]bool{}",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected githubcli to delegate provider-neutral helper rules to internal/github, actual %v", actualMatches)
	}
}
