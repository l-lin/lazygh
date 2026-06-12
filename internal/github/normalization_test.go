package github

import (
	"errors"
	"reflect"
	"testing"
)

func TestNotifications_GivenWhitespaceAndDoneEntries_WhenNormalizing_ThenItTrimsNestedFieldsAndDropsDoneNotifications(t *testing.T) {
	subject := []Notification{
		{
			ID:              "  thread-1  ",
			Reason:          "  mention  ",
			UpdatedAt:       " 2026-05-25T12:00:00Z ",
			LastReadAt:      " 2026-05-25T11:00:00Z ",
			URL:             " https://api.github.com/notifications/1 ",
			SubscriptionURL: " https://api.github.com/subscriptions/1 ",
			Repository:      RepositoryRef{Name: " widgets ", NameWithOwner: " acme/widgets "},
			Subject: NotificationSubject{
				Title:            " Ship it ",
				Type:             " PullRequest ",
				URL:              " https://api.github.com/repos/acme/widgets/pulls/42 ",
				LatestCommentURL: " https://api.github.com/repos/acme/widgets/issues/comments/1 ",
			},
		},
		{
			ID:         "thread-2",
			Done:       true,
			Repository: RepositoryRef{NameWithOwner: "acme/widgets"},
			Subject:    NotificationSubject{Title: "Already done", Type: NotificationSubjectTypePullRequest},
		},
	}

	actual := normalizedNotifications(subject)

	expected := []Notification{{
		ID:              "thread-1",
		Reason:          "mention",
		UpdatedAt:       "2026-05-25T12:00:00Z",
		LastReadAt:      "2026-05-25T11:00:00Z",
		URL:             "https://api.github.com/notifications/1",
		SubscriptionURL: "https://api.github.com/subscriptions/1",
		Repository:      RepositoryRef{Name: "widgets", NameWithOwner: "acme/widgets"},
		Subject: NotificationSubject{
			Title:            "Ship it",
			Type:             "PullRequest",
			URL:              "https://api.github.com/repos/acme/widgets/pulls/42",
			LatestCommentURL: "https://api.github.com/repos/acme/widgets/issues/comments/1",
		},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized notifications %+v, actual %+v", expected, actual)
	}
}

func TestNormalizeNotificationSubjectTarget_GivenBlankRepositoryOrInvalidID_WhenNormalizing_ThenItReturnsTheMissingTargetError(t *testing.T) {
	_, actualErr := NormalizeNotificationSubjectTarget("   ", 7)
	if !errors.Is(actualErr, ErrMissingNotificationSubjectTarget) {
		t.Fatalf("expected error %v, actual %v", ErrMissingNotificationSubjectTarget, actualErr)
	}

	_, actualErr = NormalizeNotificationSubjectTarget("acme/widgets", 0)
	if !errors.Is(actualErr, ErrMissingNotificationSubjectTarget) {
		t.Fatalf("expected error %v, actual %v", ErrMissingNotificationSubjectTarget, actualErr)
	}
}

func TestNormalizeNotificationSubjectTarget_GivenTrimmedRepositoryAndPositiveID_WhenNormalizing_ThenItReturnsTheNormalizedRepository(t *testing.T) {
	actual, actualErr := NormalizeNotificationSubjectTarget(" acme/widgets ", 7)

	then_noDomainError(t, actualErr)
	expected := "acme/widgets"
	if actual != expected {
		t.Fatalf("expected repository %q, actual %q", expected, actual)
	}
}

func TestIssueDetail_GivenWhitespaceNestedFields_WhenNormalizing_ThenItTrimsAuthorLabelsAndAssignees(t *testing.T) {
	subject := IssueDetail{
		Title:     " Release 1.0 ",
		URL:       " https://github.com/acme/widgets/issues/42 ",
		Body:      " Ship it ",
		BodyHTML:  " <p>Ship it</p> ",
		State:     " OPEN ",
		CreatedAt: " 2026-05-25T09:00:00Z ",
		UpdatedAt: " 2026-05-25T10:00:00Z ",
		Author:    &PullRequestAuthor{Login: " octocat ", Name: " Octo Cat "},
		Labels:    []PullRequestLabel{{Name: " bug "}},
		Assignees: []PullRequestAuthor{{Login: " alice ", Name: " Alice "}},
	}

	actual := subject.normalized()

	expected := IssueDetail{
		Title:     "Release 1.0",
		URL:       "https://github.com/acme/widgets/issues/42",
		Body:      "Ship it",
		BodyHTML:  "<p>Ship it</p>",
		State:     "OPEN",
		CreatedAt: "2026-05-25T09:00:00Z",
		UpdatedAt: "2026-05-25T10:00:00Z",
		Author:    &PullRequestAuthor{Login: "octocat", Name: "Octo Cat"},
		Labels:    []PullRequestLabel{{Name: "bug"}},
		Assignees: []PullRequestAuthor{{Login: "alice", Name: "Alice"}},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized issue detail %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestComment_GivenWhitespaceAndReactionGroups_WhenNormalizing_ThenItTrimsNestedFieldsAndNormalizesReactions(t *testing.T) {
	subject := PullRequestComment{
		ID:        " PRRC_42 ",
		Body:      " Looks good ",
		BodyHTML:  " <p>Looks good</p> ",
		CreatedAt: " 2026-05-25T12:00:00Z ",
		URL:       " https://github.com/acme/widgets/pull/42#discussion_r1 ",
		DiffHunk:  " @@ -1 +1 @@ ",
		State:     " OPEN ",
		Author:    &PullRequestCommentAuthor{Login: " octocat "},
		ReactionGroups: []ReactionGroup{
			{Content: ReactionContent(" HEART "), TotalCount: 2},
			{Content: ReactionContent(" THUMBS_UP "), TotalCount: -1, ViewerHasReacted: true},
			{Content: ReactionContent("   "), TotalCount: 1},
		},
	}

	actual := subject.normalized()

	expected := PullRequestComment{
		ID:        "PRRC_42",
		Body:      "Looks good",
		BodyHTML:  "<p>Looks good</p>",
		CreatedAt: "2026-05-25T12:00:00Z",
		URL:       "https://github.com/acme/widgets/pull/42#discussion_r1",
		DiffHunk:  "@@ -1 +1 @@",
		State:     "OPEN",
		Author:    &PullRequestCommentAuthor{Login: "octocat"},
		ReactionGroups: []ReactionGroup{
			{Content: ReactionContentThumbsUp, TotalCount: 0, ViewerHasReacted: true},
			{Content: ReactionContentHeart, TotalCount: 2},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized pull request comment %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestDiffFile_GivenWindowsPatchAndDuplicateOwners_WhenNormalizing_ThenItTrimsAndDeduplicates(t *testing.T) {
	subject := PullRequestDiffFile{
		Path:         " internal/tui/model.go ",
		PreviousPath: " internal/tui/old_model.go ",
		ChangeType:   " Renamed ",
		Patch:        "@@ -1 +1 @@\r\n-old\r\n+new\r\n\r\n",
		TeamOwners:   []string{" @acme/reviewers ", "", "@acme/reviewers", " @acme/docs "},
	}

	actual := subject.normalized()

	expected := PullRequestDiffFile{
		Path:         "internal/tui/model.go",
		PreviousPath: "internal/tui/old_model.go",
		ChangeType:   "renamed",
		Patch:        "@@ -1 +1 @@\n-old\n+new",
		TeamOwners:   []string{"@acme/reviewers", "@acme/docs"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized diff file %+v, actual %+v", expected, actual)
	}
}

func TestNormalizePullRequestDiffText_GivenCRLFAndTrailingBlankLines_WhenNormalizing_ThenItReturnsCanonicalLFText(t *testing.T) {
	actual := NormalizePullRequestDiffText("@@ -1 +1 @@\r\n-old\r\n+new\r\n\r\n")

	expected := "@@ -1 +1 @@\n-old\n+new"
	if actual != expected {
		t.Fatalf("expected normalized diff text %q, actual %q", expected, actual)
	}
}

func TestNormalizePullRequestDiffFileTeamOwners_GivenWhitespaceAndDuplicates_WhenNormalizing_ThenItTrimsAndDeduplicatesOwners(t *testing.T) {
	actual := NormalizePullRequestDiffFileTeamOwners([]string{" @acme/reviewers ", "", "@acme/reviewers", " @acme/docs "})

	expected := []string{"@acme/reviewers", "@acme/docs"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized team owners %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestReviewRequest_GivenWhitespaceReviewerAndOrganization_WhenNormalizing_ThenItTrimsNestedFields(t *testing.T) {
	subject := PullRequestReviewRequest{RequestedReviewer: PullRequestRequestedReviewer{
		TypeName:     " User ",
		Login:        " octocat ",
		Name:         " Octo Cat ",
		Slug:         " docs-team ",
		Organization: &PullRequestReviewRequestOrganization{Login: " acme "},
	}}

	actual := subject.normalized()

	expected := PullRequestReviewRequest{RequestedReviewer: PullRequestRequestedReviewer{
		TypeName:     "User",
		Login:        "octocat",
		Name:         "Octo Cat",
		Slug:         "docs-team",
		Organization: &PullRequestReviewRequestOrganization{Login: "acme"},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized review request %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestAutoMergeRequest_GivenWhitespaceEnabledAt_WhenNormalizing_ThenItTrimsTheTimestamp(t *testing.T) {
	subject := PullRequestAutoMergeRequest{EnabledAt: " 2026-05-25T12:30:00Z "}

	actual := subject.normalized()

	expected := PullRequestAutoMergeRequest{EnabledAt: "2026-05-25T12:30:00Z"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized auto-merge request %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestMergeQueueEntry_GivenWhitespaceFields_WhenNormalizing_ThenItTrimsStringsAndKeepsNumericData(t *testing.T) {
	subject := PullRequestMergeQueueEntry{
		ID:                   " MQE_1 ",
		State:                " QUEUED ",
		Position:             2,
		EstimatedTimeToMerge: 17,
	}

	actual := subject.normalized()

	expected := PullRequestMergeQueueEntry{
		ID:                   "MQE_1",
		State:                "QUEUED",
		Position:             2,
		EstimatedTimeToMerge: 17,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected normalized merge queue entry %+v, actual %+v", expected, actual)
	}
}

func TestPullRequestURLHelpers_GivenRepositoryAndSubjectURLs_WhenNormalizing_ThenTheyReturnCanonicalValues(t *testing.T) {
	actualPullRequestURL := PullRequestHTMLURL(" acme/widgets ", 42)
	actualCanonicalURL := CanonicalPullRequestURL(" acme/widgets ", 77)
	actualTrailingID, actualOK := subjectTrailingID(" https://api.github.com/repos/acme/widgets/issues/3235 ", "/issues/")
	actualSegments := pullRequestURLPathSegments(" /acme/widgets/pull/77/files/ ")

	expectedSegments := []string{"acme", "widgets", "pull", "77", "files"}
	if actualPullRequestURL != "https://github.com/acme/widgets/pull/42" {
		t.Fatalf("expected pull request HTML URL %q, actual %q", "https://github.com/acme/widgets/pull/42", actualPullRequestURL)
	}
	if actualCanonicalURL != "https://github.com/acme/widgets/pull/77" {
		t.Fatalf("expected canonical pull request URL %q, actual %q", "https://github.com/acme/widgets/pull/77", actualCanonicalURL)
	}
	if !actualOK || actualTrailingID != 3235 {
		t.Fatalf("expected trailing id %d with ok=true, actual %d ok=%t", 3235, actualTrailingID, actualOK)
	}
	if !reflect.DeepEqual(actualSegments, expectedSegments) {
		t.Fatalf("expected path segments %v, actual %v", expectedSegments, actualSegments)
	}
}
