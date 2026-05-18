package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestPullRequestBuildRunPopup_GivenLargeLogs_WhenPagingThroughThem_ThenTheViewBufferOnlyContainsVisibleRows(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	gui := given_headlessGuiWithSize(t, 80, 12)
	defer gui.Close()
	subject.configureGUI(gui)

	contentLines := make([]string, 0, 60)
	for lineNumber := 1; lineNumber <= 60; lineNumber++ {
		contentLines = append(contentLines, fmt.Sprintf("Line %03d", lineNumber))
	}

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		body: strings.Join(contentLines, "\n"),
	}))

	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document := subject.currentPullRequestBuildRunPopupDocument(popupView)
	then_buildPopupBufferOmitsRowText(t, popupView, document, popupView.InnerHeight())

	pageDownHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestBuildInfoName, gocui.KeyPgdn)
	then_noError(t, pageDownHandler(gui, popupView))

	popupView, actualErr = gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	document = subject.currentPullRequestBuildRunPopupDocument(popupView)
	if actual := subject.pullRequestBuildRunPopup.viewState.originRow; actual <= 0 {
		t.Fatalf("expected the popup origin row to advance after paging, actual %d", actual)
	}
	then_buildPopupBufferContainsRowText(t, popupView, document, subject.pullRequestBuildRunPopup.viewState.originRow)
	then_buildPopupBufferOmitsRowText(t, popupView, document, subject.pullRequestBuildRunPopup.viewState.originRow-1)
}

func then_buildPopupBufferContainsRowText(t *testing.T, popupView *gocui.View, document detailDocument, rowIndex int) {
	t.Helper()

	if rowIndex < 0 || rowIndex >= len(document.rows) {
		t.Fatalf("expected row index %d to exist", rowIndex)
	}
	expected := document.rows[rowIndex].text
	if !strings.Contains(popupView.Buffer(), expected) {
		t.Fatalf("expected build popup buffer to contain visible row %q, actual %q", expected, popupView.Buffer())
	}
}

func then_buildPopupBufferOmitsRowText(t *testing.T, popupView *gocui.View, document detailDocument, rowIndex int) {
	t.Helper()

	if rowIndex < 0 || rowIndex >= len(document.rows) {
		t.Fatalf("expected row index %d to exist", rowIndex)
	}
	unexpected := document.rows[rowIndex].text
	if strings.Contains(popupView.Buffer(), unexpected) {
		t.Fatalf("expected build popup buffer to omit offscreen row %q, actual %q", unexpected, popupView.Buffer())
	}
}
