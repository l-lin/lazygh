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
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestDiff(" acme/widgets ", 42)

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42", "-H", "Accept: application/vnd.github.v3.diff"}},
		{name: "gh", args: []string{"api", "repos/acme/widgets/pulls/42/files?per_page=100", "--paginate", "--slurp"}},
	})

	expected := PullRequestDiff{
		UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\n@@ -1 +1 @@\n-old\n+new",
		Files: []PullRequestDiffFile{
			{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1, Patch: "@@ -1 +1 @@\n-old\n+new"},
			{Path: "docs/guide.md", PreviousPath: "docs/old-guide.md", ChangeType: "renamed"},
		},
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
