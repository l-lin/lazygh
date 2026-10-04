package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type given_progressQueries struct {
	fakePullRequestDetailLoader
	when_detailProgress func(func(githubdomain.PullRequestDetail)) (githubdomain.PullRequestDetail, error)
	when_diffProgress   func(func(githubdomain.PullRequestDiff)) (githubdomain.PullRequestDiff, error)
	when_teamOwners     func() (map[string][]string, error)
}

func (queries *given_progressQueries) GetPullRequestDetailProgressively(_ string, _ int, publish func(githubdomain.PullRequestDetail)) (githubdomain.PullRequestDetail, error) {
	return queries.when_detailProgress(publish)
}

func (queries *given_progressQueries) GetPullRequestDiffProgressively(_ string, _ int, publish func(githubdomain.PullRequestDiff)) (githubdomain.PullRequestDiff, error) {
	return queries.when_diffProgress(publish)
}

func (queries *given_progressQueries) GetPullRequestFileTeamOwners(string, int, []string) (map[string][]string, error) {
	return queries.when_teamOwners()
}

func TestDetailCommand_GivenProgressiveProvider_WhenLoading_ThenItPublishesBeforeCompleting(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	subject.applyPullRequestDetailLoadPlanned(MsgPullRequestDetailLoadPlanned{Key: key})
	subject.detailQueries = &given_progressQueries{when_detailProgress: func(publish func(githubdomain.PullRequestDetail)) (githubdomain.PullRequestDetail, error) {
		detail := githubdomain.PullRequestDetail{Body: "fresh", PendingReviewStateKnown: true}
		publish(detail)
		actual := subject.pullRequestDetailCache[key]
		expected := "fresh"
		if actual.detail.Body != expected || !subject.pullRequestDetailLoadInFlight[key] {
			t.Fatalf("expected visible progress before completion, actual %+v", actual)
		}
		detail.Comments = []githubdomain.PullRequestComment{{Body: "enriched"}}
		return detail, nil
	}}
	actualMessage := loadPullRequestDetailResult(newPullRequestDetailWorkflowRuntime(subject, nil), summary)
	subject.applyPullRequestDetailLoaded(actualMessage)
	actual := subject.pullRequestDetailCache[key]
	if len(actual.detail.Comments) != 1 || actual.needsRefresh || actual.sourceUpdatedAt != summary.UpdatedAt || subject.pullRequestDetailLoadInFlight[key] {
		t.Fatalf("expected complete enriched result and settled load, actual %+v", actual)
	}
}

func TestDiffCommand_GivenOptionalOwnership_WhenLoading_ThenItPublishesBeforeOwnershipLookup(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	subject.startReviewSession(summary, "")
	subject.applyPullRequestDiffLoadPlanned(MsgPullRequestDiffLoadPlanned{Key: key})
	rawDiff := githubdomain.PullRequestDiff{Files: []githubdomain.PullRequestDiffFile{{Path: "main.go"}}}
	subject.detailQueries = &given_progressQueries{
		when_diffProgress: func(publish func(githubdomain.PullRequestDiff)) (githubdomain.PullRequestDiff, error) {
			publish(rawDiff)
			actual := subject.pullRequestDiffCache[key]
			if len(actual.data.Files) != 1 || !subject.pullRequestDiffLoadInFlight[key] {
				t.Fatalf("expected visible diff before threads, actual %+v", actual)
			}
			return rawDiff, nil
		},
		when_teamOwners: func() (map[string][]string, error) {
			actual := subject.pullRequestDiffCache[key]
			if len(actual.data.Files) != 1 || !actual.needsRefresh || !subject.pullRequestDiffLoadInFlight[key] {
				t.Fatalf("expected visible diff before ownership, actual %+v", actual)
			}
			return map[string][]string{"main.go": {"platform"}}, nil
		},
	}
	actualMessage := loadPullRequestDiffResult(newPullRequestDiffWorkflowRuntime(subject, nil), summary)
	subject.applyPullRequestDiffLoaded(actualMessage)
	actual := subject.pullRequestDiffCache[key]
	if len(actual.data.Files[0].TeamOwners) != 1 || actual.needsRefresh || actual.sourceUpdatedAt != summary.UpdatedAt || subject.pullRequestDiffLoadInFlight[key] {
		t.Fatalf("expected complete owned diff and settled load, actual %+v", actual)
	}
}
