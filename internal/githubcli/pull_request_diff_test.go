package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPullRequestDiff_GivenValidDiffAndPagedFilesResponses_WhenFetching_ThenItReturnsTheUnifiedDiffAndAllFileStats(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte("diff --git a/internal/tui/render.go b/internal/tui/render.go\r\n@@ -1 +1 @@\r\n-old\r\n+new\r\n")},
			{stdout: []byte(`[[{"filename":" internal/tui/render.go ","status":" modified ","additions":1,"deletions":1,"patch":"@@ -1 +1 @@\n-old\n+new"}],[{"filename":" docs/guide.md ","previous_filename":" docs/old-guide.md ","status":" renamed ","additions":0,"deletions":0}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[]}}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDiff(" acme/widgets ", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42", "-H", "Accept: application/vnd.github.v3.diff"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/files?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
	})

	expected := PullRequestDiff{
		UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new",
		Files: []PullRequestDiffFile{
			{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1, Patch: "@@ -1 +1 @@\n-old\n+new"},
			{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: "renamed"},
		},
		Threads: []PullRequestReviewThread{},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected diff %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestDiff_GivenInvalidPagedFilesJSON_WhenFetching_ThenItReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte("diff --git a/internal/tui/render.go b/internal/tui/render.go\n")},
			{stdout: []byte(`[{`)},
		},
	}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetPullRequestDiff("acme/widgets", 42)

	if !errors.Is(actualErr, ErrInvalidPullRequestDiffFilesResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestDiffFilesResponse, actualErr)
	}
}

func TestGetPullRequestDiff_GivenPaginatedReviewThreads_WhenFetching_ThenItReturnsEveryThreadWithResolvedStateAndReplies(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte("diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new\n")},
			{stdout: []byte(`[[{"filename":"internal/tui/render.go","status":"modified","additions":1,"deletions":1}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":true,"endCursor":"THREAD_CURSOR_1"},"nodes":[{"id":" thread-1 ","isResolved":true,"isOutdated":false,"path":" internal/tui/render.go ","line":11,"diffSide":" RIGHT ","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":" reviewer-one "},"body":" First reply ","createdAt":" 2026-04-20T10:00:00Z "},{"author":{"login":" reviewer-two "},"body":" Second reply ","createdAt":" 2026-04-20T10:05:00Z "}]}}]}}}}}`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"thread-2","isResolved":false,"isOutdated":true,"path":"internal/tui/model.go","originalLine":21,"diffSide":"LEFT","comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"reviewer-three"},"body":"Needs work","createdAt":"2026-04-20T11:00:00Z"}]}}]}}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDiff("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42", "-H", "Accept: application/vnd.github.v3.diff"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/files?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42", "-F", "cursor=THREAD_CURSOR_1"}},
	})

	expected := PullRequestDiff{
		UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new",
		Files:       []PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		Threads: []PullRequestReviewThread{
			{
				ID:         "thread-1",
				IsResolved: true,
				Path:       "internal/tui/render.go",
				Line:       11,
				DiffSide:   "RIGHT",
				Comments: []PullRequestComment{
					{Author: &PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "First reply", CreatedAt: "2026-04-20T10:00:00Z"},
					{Author: &PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Second reply", CreatedAt: "2026-04-20T10:05:00Z"},
				},
			},
			{
				ID:           "thread-2",
				IsOutdated:   true,
				Path:         "internal/tui/model.go",
				OriginalLine: 21,
				DiffSide:     "LEFT",
				Comments:     []PullRequestComment{{Author: &PullRequestCommentAuthor{Login: "reviewer-three"}, Body: "Needs work", CreatedAt: "2026-04-20T11:00:00Z"}},
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected diff %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestDiff_GivenReviewThreadCommentPagination_WhenFetching_ThenItLoadsEveryReplyInTheThread(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte("diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new\n")},
			{stdout: []byte(`[[{"filename":"internal/tui/render.go","status":"modified","additions":1,"deletions":1}]]`)},
			{stdout: []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"id":"thread-1","isResolved":false,"isOutdated":false,"path":"internal/tui/render.go","line":11,"diffSide":"RIGHT","comments":{"pageInfo":{"hasNextPage":true,"endCursor":"COMMENT_CURSOR_1"},"nodes":[{"author":{"login":"reviewer-one"},"body":"First reply","createdAt":"2026-04-20T10:00:00Z"}]}}]}}}}}`)},
			{stdout: []byte(`{"data":{"node":{"comments":{"pageInfo":{"hasNextPage":false,"endCursor":null},"nodes":[{"author":{"login":"reviewer-two"},"body":"Second reply","createdAt":"2026-04-20T10:05:00Z"}]}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDiff("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42", "-H", "Accept: application/vnd.github.v3.diff"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/files?per_page=100", "--paginate", "--slurp"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadsQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestReviewThreadCommentsQuery, "-F", "threadID=thread-1", "-F", "cursor=COMMENT_CURSOR_1"}},
	})

	expectedComments := []PullRequestComment{
		{Author: &PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "First reply", CreatedAt: "2026-04-20T10:00:00Z"},
		{Author: &PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Second reply", CreatedAt: "2026-04-20T10:05:00Z"},
	}
	if !reflect.DeepEqual(actual.Threads[0].Comments, expectedComments) {
		t.Fatalf("expected thread comments %+v, actual %+v", expectedComments, actual.Threads[0].Comments)
	}
}
