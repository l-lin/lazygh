package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func BenchmarkRefreshViews_GivenReviewModeDescription_WhenRefreshingRepeatedly(b *testing.B) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(b)
	defer gui.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if actualErr := subject.refreshViews(gui); actualErr != nil {
			b.Fatalf("expected no refresh error, actual %v", actualErr)
		}
	}
}

func BenchmarkRefreshViews_GivenActionsPopupOpenOnReviewDescription_WhenRefreshingRepeatedly(b *testing.B) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(b)
	defer gui.Close()
	if actualErr := subject.openActionsPopup(gui, nil); actualErr != nil {
		b.Fatalf("expected no popup error, actual %v", actualErr)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if actualErr := subject.refreshViews(gui); actualErr != nil {
			b.Fatalf("expected no refresh error, actual %v", actualErr)
		}
	}
}

func BenchmarkFooterPresenter_GivenReviewModeDescription_WhenResolvingStatusLineKeyHintsRepeatedly(b *testing.B) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(b)
	defer gui.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		subject.withRefreshReadCache(func() {
			_ = subject.footerPresenter().statusLineKeyHintsText()
		})
	}
}

func BenchmarkActionsPopupSelectors_GivenReviewDescriptionPopupOpen_WhenResolvingRepeatedly(b *testing.B) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(b)
	defer gui.Close()
	if actualErr := subject.openActionsPopup(gui, nil); actualErr != nil {
		b.Fatalf("expected no popup error, actual %v", actualErr)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		subject.withRefreshReadCache(func() {
			_ = subject.actionsPopupPresenter()
			_ = subject.currentActionsPopupVisibleLines()
			_ = subject.currentActionsPopupSelectedRenderedLine()
			_ = subject.currentActionsPopupRenderedLineCount()
		})
	}
}

func BenchmarkReviewSessionReadModel_GivenReviewModeDescription_WhenResolvingMetadataAndTreeRepeatedly(b *testing.B) {
	subject, gui := given_reviewDescriptionBenchmarkProgram(b)
	defer gui.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		subject.withRefreshReadCache(func() {
			_ = subject.reviewSessionMetadataContent()
			_ = subject.reviewSessionFiles()
			_, _, _ = subject.reviewSessionCurrentTree()
		})
	}
}

func given_reviewDescriptionBenchmarkProgram(tb testing.TB) (*Program, *gocui.Gui) {
	tb.Helper()

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_benchmarkHeadlessGuiWithSize(tb, 120, 40)
	subject.configureGUI(gui)
	if actualErr := subject.layout(gui); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no layout error, actual %v", actualErr)
	}
	if actualErr := given_startingReviewModeForBenchmark(tb, gui, subject); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no review-mode startup error, actual %v", actualErr)
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no detail focus error, actual %v", actualErr)
	}
	if actualErr := subject.refreshViews(gui); actualErr != nil {
		gui.Close()
		tb.Fatalf("expected no warm refresh error, actual %v", actualErr)
	}
	return subject, gui
}

func given_startingReviewModeForBenchmark(tb testing.TB, gui *gocui.Gui, subject *Program) error {
	tb.Helper()

	if actualErr := subject.openActionsPopup(gui, nil); actualErr != nil {
		return actualErr
	}
	subject.model.UpdateActionsPopupSearch("start review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "start review"))
	if actualErr := subject.afterStateChange(gui); actualErr != nil {
		return actualErr
	}
	return subject.executeSelectedActionsPopupAction(gui, nil)
}

func given_benchmarkHeadlessGuiWithSize(tb testing.TB, width int, height int) *gocui.Gui {
	tb.Helper()

	gui, err := gocui.NewGui(gocui.NewGuiOpts{
		OutputMode: gocui.OutputTrue,
		Headless:   true,
		Width:      width,
		Height:     height,
	})
	if err != nil {
		tb.Fatalf("expected no error, actual %v", err)
	}
	return gui
}
