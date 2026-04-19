package tui

import (
	"errors"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestCopyPullRequestURL_GivenPullRequestsView_WhenHandlingTheAction_ThenItCopiesTheSelectedPullRequestURLAndShowsFeedback(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}),
	})
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.copyPullRequestURL(gui, nil)
	then_noError(t, actualErr)

	if len(clipboardWriter.writes) != 1 || clipboardWriter.writes[0] != "https://github.com/acme/widgets/pull/42" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/widgets/pull/42"}, clipboardWriter.writes)
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.TitlePrefix, yankSuccessMessage) {
		t.Fatalf("expected pull request title prefix to contain %q, actual %q", yankSuccessMessage, pullRequestsView.TitlePrefix)
	}
}

func TestCopyPullRequestURL_GivenDetailViewAndCachedDetailURL_WhenHandlingTheAction_ThenItPrefersTheCanonicalDetailURL(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/summary"}),
	})
	model.OpenDetail()
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	subject.pullRequestDetailCache[pullRequestDetailKey(githubcli.Repository{NameWithOwner: "acme/widgets"}, 42)] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{URL: "https://github.com/acme/widgets/pull/canonical"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.copyPullRequestURL(gui, nil)
	then_noError(t, actualErr)

	if len(clipboardWriter.writes) != 1 || clipboardWriter.writes[0] != "https://github.com/acme/widgets/pull/canonical" {
		t.Fatalf("expected clipboard writes %v, actual %v", []string{"https://github.com/acme/widgets/pull/canonical"}, clipboardWriter.writes)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.TitlePrefix, yankSuccessMessage) {
		t.Fatalf("expected detail title prefix to contain %q, actual %q", yankSuccessMessage, detailView.TitlePrefix)
	}
}

func TestCopyPullRequestURL_GivenUserView_WhenHandlingTheAction_ThenItShowsHarmlessFeedbackAndDoesNotTouchTheClipboard(t *testing.T) {
	clipboardWriter := &fakeClipboardWriter{}
	subject := NewProgramWithModel(given_model())
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.copyPullRequestURL(gui, nil)
	then_noError(t, actualErr)

	if len(clipboardWriter.writes) != 0 {
		t.Fatalf("expected no clipboard writes, actual %v", clipboardWriter.writes)
	}

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if !strings.Contains(userView.Title, yankUnavailableMessage) {
		t.Fatalf("expected user title to contain %q, actual %q", yankUnavailableMessage, userView.Title)
	}
}

func TestCopyPullRequestURL_GivenClipboardFailure_WhenHandlingTheAction_ThenItShowsFailureFeedback(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}),
	})
	clipboardWriter := &fakeClipboardWriter{writeErr: errors.New("clipboard failed")}
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = clipboardWriter
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.copyPullRequestURL(gui, nil)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.TitlePrefix, yankFailureMessage) {
		t.Fatalf("expected pull request title prefix to contain %q, actual %q", yankFailureMessage, pullRequestsView.TitlePrefix)
	}
}

type fakeClipboardWriter struct {
	writes   []string
	writeErr error
}

func (writer *fakeClipboardWriter) WriteText(text string) error {
	writer.writes = append(writer.writes, text)
	return writer.writeErr
}
