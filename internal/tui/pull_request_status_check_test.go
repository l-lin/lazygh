package tui

import (
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestClassifyPullRequestStatusCheck_GivenCompletedCancelledCheck_WhenClassifying_ThenItUsesCancelledLabelMutedOverviewAndFailingSummary(t *testing.T) {
	actual := classifyPullRequestStatusCheck(githubcli.PullRequestStatusCheck{Status: "COMPLETED", Conclusion: "CANCELLED"})

	if actual.StateLabel != "Cancelled" {
		t.Fatalf("expected state label %q, actual %q", "Cancelled", actual.StateLabel)
	}
	if actual.OverviewStatus != pullRequestOverviewStatusMuted {
		t.Fatalf("expected overview status %v, actual %v", pullRequestOverviewStatusMuted, actual.OverviewStatus)
	}
	if actual.SummaryKind != pullRequestStatusCheckSummaryKindFailing {
		t.Fatalf("expected summary kind %v, actual %v", pullRequestStatusCheckSummaryKindFailing, actual.SummaryKind)
	}
}

func TestSummarizeStatusChecks_GivenCancelledCheck_WhenSummarizing_ThenItCountsAsFailing(t *testing.T) {
	actual := summarizeStatusChecks([]githubcli.PullRequestStatusCheck{{Status: "COMPLETED", Conclusion: "CANCELLED"}})

	if actual != "1 failing" {
		t.Fatalf("expected summary %q, actual %q", "1 failing", actual)
	}
}

func TestPullRequestOverviewStatusForCheck_GivenCancelledCheck_WhenResolving_ThenItTreatsTheCheckAsMuted(t *testing.T) {
	actual := pullRequestOverviewStatusForCheck(githubcli.PullRequestStatusCheck{Status: "COMPLETED", Conclusion: "CANCELLED"})

	if actual != pullRequestOverviewStatusMuted {
		t.Fatalf("expected overview status %v, actual %v", pullRequestOverviewStatusMuted, actual)
	}
}
