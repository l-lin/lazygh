package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestStatusCheckSummaryKind int

const (
	pullRequestStatusCheckSummaryKindPending pullRequestStatusCheckSummaryKind = iota
	pullRequestStatusCheckSummaryKindPassing
	pullRequestStatusCheckSummaryKindFailing
)

type pullRequestStatusCheckClassification struct {
	StateLabel     string
	OverviewStatus pullRequestOverviewStatus
	SummaryKind    pullRequestStatusCheckSummaryKind
}

func classifyPullRequestStatusCheck(check githubdomain.PullRequestStatusCheck) pullRequestStatusCheckClassification {
	status := strings.ToUpper(strings.TrimSpace(check.Status))
	conclusion := strings.ToUpper(strings.TrimSpace(check.Conclusion))
	if status != "COMPLETED" {
		return pullRequestStatusCheckClassification{
			StateLabel:     "Pending",
			OverviewStatus: pullRequestOverviewStatusPending,
			SummaryKind:    pullRequestStatusCheckSummaryKindPending,
		}
	}

	switch conclusion {
	case "SUCCESS":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Successful",
			OverviewStatus: pullRequestOverviewStatusSuccess,
			SummaryKind:    pullRequestStatusCheckSummaryKindPassing,
		}
	case "NEUTRAL":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Neutral",
			OverviewStatus: pullRequestOverviewStatusSuccess,
			SummaryKind:    pullRequestStatusCheckSummaryKindPassing,
		}
	case "SKIPPED":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Skipped",
			OverviewStatus: pullRequestOverviewStatusSuccess,
			SummaryKind:    pullRequestStatusCheckSummaryKindPassing,
		}
	case "CANCELLED":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Cancelled",
			OverviewStatus: pullRequestOverviewStatusMuted,
			SummaryKind:    pullRequestStatusCheckSummaryKindFailing,
		}
	case "TIMED_OUT":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Timed out",
			OverviewStatus: pullRequestOverviewStatusFailure,
			SummaryKind:    pullRequestStatusCheckSummaryKindFailing,
		}
	case "STARTUP_FAILURE":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Startup failure",
			OverviewStatus: pullRequestOverviewStatusFailure,
			SummaryKind:    pullRequestStatusCheckSummaryKindFailing,
		}
	case "ACTION_REQUIRED":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Action required",
			OverviewStatus: pullRequestOverviewStatusFailure,
			SummaryKind:    pullRequestStatusCheckSummaryKindFailing,
		}
	case "FAILURE":
		return pullRequestStatusCheckClassification{
			StateLabel:     "Failed",
			OverviewStatus: pullRequestOverviewStatusFailure,
			SummaryKind:    pullRequestStatusCheckSummaryKindFailing,
		}
	default:
		return pullRequestStatusCheckClassification{
			StateLabel:     "Pending",
			OverviewStatus: pullRequestOverviewStatusPending,
			SummaryKind:    pullRequestStatusCheckSummaryKindPending,
		}
	}
}
