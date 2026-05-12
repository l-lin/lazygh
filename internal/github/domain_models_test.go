package github

import (
	"errors"
	"testing"
)

func TestParsePullRequestURL_GivenCanonicalAndFilesURLs_WhenParsing_ThenItReturnsTheNormalizedIdentity(t *testing.T) {
	canonical, actualErr := ParsePullRequestURL(" https://github.com/acme/widgets/pull/42 ")
	then_noDomainError(t, actualErr)
	if canonical.Repository.NameWithOwner != "acme/widgets" || canonical.Number != 42 || canonical.URL != "https://github.com/acme/widgets/pull/42" {
		t.Fatalf("expected canonical identity for acme/widgets#42, actual %+v", canonical)
	}

	filesURL, actualErr := ParsePullRequestURL("https://github.com/acme/widgets/pull/77/files#diff-1")
	then_noDomainError(t, actualErr)
	if filesURL.Repository.NameWithOwner != "acme/widgets" || filesURL.Number != 77 || filesURL.URL != "https://github.com/acme/widgets/pull/77" {
		t.Fatalf("expected normalized files-tab identity for acme/widgets#77, actual %+v", filesURL)
	}
}

func TestNormalizePullRequestIdentity_GivenBlankRepositoryOrInvalidNumber_WhenNormalizing_ThenItRejectsTheIdentity(t *testing.T) {
	_, _, actualErr := NormalizePullRequestIdentity("   ", 42)
	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}

	_, _, actualErr = NormalizePullRequestIdentity("acme/widgets", 0)
	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestNotificationIdentities_GivenPullRequestIssueAndReleaseSubjects_WhenResolving_ThenItExtractsProviderNeutralIdentities(t *testing.T) {
	pullRequestNotification := Notification{
		Repository: RepositoryRef{NameWithOwner: "acme/widgets"},
		Subject:    NotificationSubject{Type: NotificationSubjectTypePullRequest, Title: "Ship it", URL: "https://api.github.com/repos/acme/widgets/pulls/42"},
	}
	pullRequest, ok := pullRequestNotification.PullRequestSummary()
	if !ok {
		t.Fatal("expected pull request notification to resolve a pull request summary")
	}
	if pullRequest.Repository.NameWithOwner != "acme/widgets" || pullRequest.Number != 42 {
		t.Fatalf("expected pull request identity acme/widgets#42, actual %+v", pullRequest)
	}

	issueNotification := Notification{
		Repository: RepositoryRef{NameWithOwner: "acme/opencode"},
		Subject:    NotificationSubject{Type: NotificationSubjectTypeIssue, URL: "https://api.github.com/repos/acme/opencode/issues/3235"},
	}
	issueRepository, issueNumber, ok := issueNotification.IssueIdentity()
	if !ok || issueRepository != "acme/opencode" || issueNumber != 3235 {
		t.Fatalf("expected issue identity acme/opencode#3235, actual %s#%d ok=%t", issueRepository, issueNumber, ok)
	}

	releaseNotification := Notification{
		Repository: RepositoryRef{NameWithOwner: "acme/doctoboot"},
		Subject:    NotificationSubject{Type: NotificationSubjectTypeRelease, URL: "https://api.github.com/repos/acme/doctoboot/releases/317927281"},
	}
	releaseRepository, releaseID, ok := releaseNotification.ReleaseIdentity()
	if !ok || releaseRepository != "acme/doctoboot" || releaseID != 317927281 {
		t.Fatalf("expected release identity acme/doctoboot#317927281, actual %s#%d ok=%t", releaseRepository, releaseID, ok)
	}
}

func TestBuildRunLinks_GivenRunAndJobURLs_WhenParsing_ThenItExtractsTheRunAttemptAndJobIdentity(t *testing.T) {
	reference, actualErr := ParseBuildRunReferenceFromURL("https://github.com/acme/widgets/actions/runs/42/attempts/3/job/99")
	then_noDomainError(t, actualErr)
	if reference.ID != "42" || reference.Attempt != 3 {
		t.Fatalf("expected run reference {ID:42 Attempt:3}, actual %+v", reference)
	}

	jobID, ok := BuildRunJobIDFromURL("https://github.com/acme/widgets/actions/runs/42/attempts/3/job/99")
	if !ok || jobID != 99 {
		t.Fatalf("expected build run job id %d, actual %d ok=%t", 99, jobID, ok)
	}
}

func TestNormalizeReviewContracts_GivenReviewEventsAndTargets_WhenNormalizing_ThenItAcceptsValidValuesAndRejectsInvalidOnes(t *testing.T) {
	actualEvent, actualErr := NormalizeReviewEvent(" approve ")
	then_noDomainError(t, actualErr)
	if actualEvent != ReviewEventApprove {
		t.Fatalf("expected normalized review event %q, actual %q", ReviewEventApprove, actualEvent)
	}

	actualTarget, actualErr := NormalizeReviewThreadTarget(ReviewThreadTarget{Path: " internal/tui/model.go ", Line: 12, Side: "right", StartLine: 10, StartSide: "right", SubjectType: "line"})
	then_noDomainError(t, actualErr)
	if actualTarget.Path != "internal/tui/model.go" || actualTarget.Line != 12 || actualTarget.Side != "RIGHT" || actualTarget.StartLine != 10 || actualTarget.StartSide != "RIGHT" || actualTarget.SubjectType != "LINE" {
		t.Fatalf("expected normalized review thread target, actual %+v", actualTarget)
	}

	_, actualErr = NormalizeReviewThreadTarget(ReviewThreadTarget{Path: "internal/tui/model.go", Line: 12, Side: "up", SubjectType: "line"})
	if !errors.Is(actualErr, ErrInvalidReviewThreadTarget) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidReviewThreadTarget, actualErr)
	}
}

func then_noDomainError(t *testing.T, actualErr error) {
	t.Helper()
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
