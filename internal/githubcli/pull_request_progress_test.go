package githubcli

import (
	"errors"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestDetailProgress_GivenSlowEnrichment_WhenAssembling_ThenItPublishesBaseDescriptionFirst(t *testing.T) {
	expected := "fresh description"
	actual := ""
	enrichmentError := errors.New("comments unavailable")
	assembler := PullRequestDetailAssembler{
		LoadBaseDetail: func(string, int) (PullRequestDetail, error) {
			return PullRequestDetail{Body: expected}, nil
		},
		ListInlineComments: func(string, int) ([]PullRequestInlineComment, error) {
			if actual != expected {
				t.Fatalf("expected description before enrichment, actual %q", actual)
			}
			return nil, enrichmentError
		},
	}
	_, actualErr := assembler.AssembleProgressively("acme/widgets", 42, func(detail PullRequestDetail) { actual = detail.Body })
	if actual != expected || !errors.Is(actualErr, enrichmentError) {
		t.Fatalf("expected published description and enrichment error, actual %q, %v", actual, actualErr)
	}
}

func TestDiffProgress_GivenSlowThreads_WhenAssembling_ThenItPublishesDiffAndFilesFirst(t *testing.T) {
	expected := "fresh diff"
	actual := ""
	enrichmentError := errors.New("threads unavailable")
	assembler := PullRequestDiffAssembler{
		LoadUnifiedDiff: func(string, int) (string, error) { return expected, nil },
		LoadDiffFiles: func(string, int) ([]PullRequestDiffFile, error) {
			return []PullRequestDiffFile{{Path: "main.go"}}, nil
		},
		ListReviewThreads: func(string, int) ([]PullRequestReviewThread, error) {
			if actual != expected {
				t.Fatalf("expected diff before threads, actual %q", actual)
			}
			return nil, enrichmentError
		},
	}
	_, actualErr := assembler.AssembleProgressively("acme/widgets", 42, func(diff PullRequestDiff) {
		actual = diff.UnifiedDiff
		if len(diff.Files) != 1 {
			t.Fatalf("expected file metadata, actual %+v", diff)
		}
	})
	if actual != expected || !errors.Is(actualErr, enrichmentError) {
		t.Fatalf("expected published diff and enrichment error, actual %q, %v", actual, actualErr)
	}
}

func TestDetailProgressAdapter_GivenEnrichmentFailure_WhenFetching_ThenItPublishesDomainContentBeforeTheFailure(t *testing.T) {
	expected := "fresh description"
	enrichmentError := errors.New("offline")
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"number":42,"body":"fresh description"}`)},
		{err: enrichmentError},
	}}
	subject := NewPullRequestDetailAdapterWithRunner(runner)
	actual := ""
	_, actualErr := subject.GetPullRequestDetailProgressively("acme/widgets", 42, func(detail githubdomain.PullRequestDetail) {
		actual = detail.Body
		if len(runner.calls) != 1 {
			t.Fatalf("expected publication after base command only, actual %d calls", len(runner.calls))
		}
	})
	if actual != expected || !errors.Is(actualErr, enrichmentError) {
		t.Fatalf("expected domain content and enrichment failure, actual %q, %v", actual, actualErr)
	}
}

func TestDiffProgressAdapter_GivenEnrichmentFailure_WhenFetching_ThenItPublishesDomainDiffBeforeTheFailure(t *testing.T) {
	expected := "fresh diff"
	enrichmentError := errors.New("offline")
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(expected)},
		{stdout: []byte(`[[{"filename":"main.go"}]]`)},
		{err: enrichmentError},
	}}
	subject := NewPullRequestDetailAdapterWithRunner(runner)
	actual := ""
	_, actualErr := subject.GetPullRequestDiffProgressively("acme/widgets", 42, func(diff githubdomain.PullRequestDiff) {
		actual = diff.UnifiedDiff
		if len(runner.calls) != 2 || len(diff.Files) != 1 || diff.Files[0].Path != "main.go" {
			t.Fatalf("expected publication after diff and file commands, actual %d calls, diff %+v", len(runner.calls), diff)
		}
	})
	if actual != expected || !errors.Is(actualErr, enrichmentError) {
		t.Fatalf("expected domain diff and enrichment failure, actual %q, %v", actual, actualErr)
	}
}

func TestProgress_GivenBaseLoadFailures_WhenAssembling_ThenItPublishesNothing(t *testing.T) {
	actual := 0
	expected := 0
	loadError := errors.New("offline")
	detailAssembler := PullRequestDetailAssembler{LoadBaseDetail: func(string, int) (PullRequestDetail, error) { return PullRequestDetail{}, loadError }}
	diffAssembler := PullRequestDiffAssembler{
		LoadUnifiedDiff: func(string, int) (string, error) { return "", loadError },
		LoadDiffFiles:   func(string, int) ([]PullRequestDiffFile, error) { return nil, nil },
	}
	_, _ = detailAssembler.AssembleProgressively("acme/widgets", 42, func(PullRequestDetail) { actual++ })
	_, _ = diffAssembler.AssembleProgressively("acme/widgets", 42, func(PullRequestDiff) { actual++ })
	if actual != expected {
		t.Fatalf("expected no publications, actual %d", actual)
	}
}
