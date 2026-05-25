package tui

import (
	"reflect"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestPlanSessionLoad_GivenAvailableSessionQueries_WhenPlanning_ThenItReturnsAnExplicitLoadStartAndCommand(t *testing.T) {
	actual := planSessionLoad(sessionLoadPlanInput{hasSessionQueries: true})

	expectedMessageTypes := []string{"tui.MsgConnectedUserLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.loadConnectedUserCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}
}

func TestPlanPullRequestListLoad_GivenTheActiveTabWithLiveQueries_WhenPlanning_ThenItHydratesTheCacheAndMarksTheLoadExplicitly(t *testing.T) {
	actual := planPullRequestListLoad(pullRequestListLoadPlanInput{activeTab: MyPullRequestsTab, targetTab: MyPullRequestsTab, hasPullRequestQueries: true})

	expectedMessageTypes := []string{"tui.MsgPullRequestsLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.hydratePullRequestsFromCacheCmd", "tui.loadPullRequestsCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgPullRequestsLoadPlanned)
	if !ok {
		t.Fatalf("expected a pull request load planned message, actual %T", actual.messages[0])
	}
	if expected := MyPullRequestsTab; actualMessage.Tab != expected {
		t.Fatalf("expected planned tab %v, actual %v", expected, actualMessage.Tab)
	}
}

func TestPlanPullRequestDetailLoad_GivenAVisibleCachedDetailNeedingRefresh_WhenPlanning_ThenItHydratesFirstAndMarksTheRefreshExplicitly(t *testing.T) {
	summary := given_workflowPlanPullRequestSummary("2026-05-05T10:05:00Z")
	actual := planPullRequestDetailLoad(pullRequestDetailLoadPlanInput{
		summary:                  summary,
		hasSelection:             true,
		key:                      pullRequestDetailKey(summary.Repository, summary.Number),
		visibleResult:            pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Cached body", State: "OPEN"}), sourceUpdatedAt: "2026-05-05T10:00:00Z"},
		visibleResultLoaded:      true,
		hydrateVisibleResult:     true,
		hasPullRequestDetailPort: true,
	})

	expectedMessageTypes := []string{"tui.MsgPullRequestDetailLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.hydratePullRequestDetailFromCacheCmd", "tui.loadPullRequestDetailCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgPullRequestDetailLoadPlanned)
	if !ok {
		t.Fatalf("expected a pull request detail load planned message, actual %T", actual.messages[0])
	}
	if expected := "acme/widgets#42"; actualMessage.Key != expected {
		t.Fatalf("expected planned detail key %q, actual %q", expected, actualMessage.Key)
	}
}

func TestPlanPullRequestDiffLoad_GivenAVisibleCachedDiffMissingTeamOwners_WhenPlanning_ThenItHydratesFirstAndMarksTheRefreshExplicitly(t *testing.T) {
	summary := given_workflowPlanPullRequestSummary("2026-05-05T10:00:00Z")
	actual := planPullRequestDiffLoad(pullRequestDiffLoadPlanInput{
		summary:                summary,
		hasSelection:           true,
		key:                    pullRequestDetailKey(summary.Repository, summary.Number),
		visibleResult:          pullRequestDiffResult{data: reviewDiffData{Files: []reviewDiffFile{{Path: "main.go"}}}},
		visibleResultLoaded:    true,
		hydrateVisibleResult:   true,
		shouldLoadTeamOwners:   true,
		hasPullRequestDiffPort: true,
	})

	expectedMessageTypes := []string{"tui.MsgPullRequestDiffLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.hydratePullRequestDiffFromCacheCmd", "tui.loadPullRequestDiffCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgPullRequestDiffLoadPlanned)
	if !ok {
		t.Fatalf("expected a pull request diff load planned message, actual %T", actual.messages[0])
	}
	if expected := "acme/widgets#42"; actualMessage.Key != expected {
		t.Fatalf("expected planned diff key %q, actual %q", expected, actualMessage.Key)
	}
}

func TestPlanNotificationDetailLoad_GivenAnUnloadedIssueNotification_WhenPlanning_ThenItMarksTheIssueRequestExplicitly(t *testing.T) {
	actual := planNotificationDetailLoad(notificationDetailLoadPlanInput{kind: notificationDetailLoadKindIssue, repository: "acme/widgets", number: 42, key: "acme/widgets#42", hasNotificationQueries: true})

	expectedMessageTypes := []string{"tui.MsgIssueDetailLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.loadIssueDetailCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgIssueDetailLoadPlanned)
	if !ok {
		t.Fatalf("expected an issue detail load planned message, actual %T", actual.messages[0])
	}
	if expected := 42; actualMessage.Number != expected {
		t.Fatalf("expected issue number %d, actual %d", expected, actualMessage.Number)
	}
}

func TestPlanCurrentDetailImageHTMLLoads_GivenDuplicateRenderableSources_WhenPlanning_ThenItMarksOneExplicitLoadPerKey(t *testing.T) {
	source := detailImageHTMLSource{
		key:          "acme/widgets#42:description",
		repository:   "acme/widgets",
		markdown:     "![Architecture](./docs/diagram.png)",
		renderedHTML: "",
		applyRenderedHTML: func(*Program, string) {
		},
	}
	actual := planCurrentDetailImageHTMLLoads(detailImageHTMLLoadPlanInput{hasMarkdownHTMLRenderer: true, sources: []detailImageHTMLSource{source, source}, loadInFlightByKey: map[string]bool{}, loadFailedByKey: map[string]bool{}})

	expectedMessageTypes := []string{"tui.MsgCurrentDetailImageHTMLLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.loadCurrentDetailImageHTMLCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgCurrentDetailImageHTMLLoadPlanned)
	if !ok {
		t.Fatalf("expected a current detail image HTML load planned message, actual %T", actual.messages[0])
	}
	if expected := source.key; actualMessage.SourceKey != expected {
		t.Fatalf("expected source key %q, actual %q", expected, actualMessage.SourceKey)
	}
}

func TestPlanCurrentDetailImageLoads_GivenDuplicateRemoteImages_WhenPlanning_ThenItMarksOneExplicitLoadPerURL(t *testing.T) {
	source := detailImageHTMLSource{key: "acme/widgets#42:description", markdown: "![Architecture](https://example.com/diagram.png)\n![Architecture](https://example.com/diagram.png)"}
	actual := planCurrentDetailImageLoads(detailImageLoadPlanInput{detailImageStoreAvailable: true, sources: []detailImageHTMLSource{source}, imageAlreadyLoadedByURL: map[string]bool{}, loadInFlightByURL: map[string]bool{}, loadFailedByURL: map[string]bool{}})

	expectedMessageTypes := []string{"tui.MsgCurrentDetailImageLoadPlanned"}
	if actualMessageTypes := given_workflowPlanMessageTypeNames(actual); !reflect.DeepEqual(actualMessageTypes, expectedMessageTypes) {
		t.Fatalf("expected message types %v, actual %v", expectedMessageTypes, actualMessageTypes)
	}
	expectedCommandTypes := []string{"tui.loadCurrentDetailImageCmd"}
	if actualCommandTypes := given_workflowPlanCommandTypeNames(actual); !reflect.DeepEqual(actualCommandTypes, expectedCommandTypes) {
		t.Fatalf("expected command types %v, actual %v", expectedCommandTypes, actualCommandTypes)
	}

	actualMessage, ok := actual.messages[0].(MsgCurrentDetailImageLoadPlanned)
	if !ok {
		t.Fatalf("expected a current detail image load planned message, actual %T", actual.messages[0])
	}
	expectedImageURL := "https://example.com/diagram.png"
	if actualMessage.ImageURL != expectedImageURL {
		t.Fatalf("expected planned image URL %q, actual %q", expectedImageURL, actualMessage.ImageURL)
	}
}

func given_workflowPlanPullRequestSummary(updatedAt string) githubdomain.PullRequest {
	return githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{
		Title:      "First PR",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		Body:       "Summary body",
		State:      "OPEN",
		UpdatedAt:  updatedAt,
	})
}

func given_workflowPlanMessageTypeNames(plan workflowPlan) []string {
	actual := make([]string, 0, len(plan.messages))
	for _, message := range plan.messages {
		actual = append(actual, reflect.TypeOf(message).String())
	}
	return actual
}

func given_workflowPlanCommandTypeNames(plan workflowPlan) []string {
	actual := make([]string, 0, len(plan.commands))
	for _, command := range plan.commands {
		actual = append(actual, reflect.TypeOf(command).String())
	}
	return actual
}
