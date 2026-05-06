package tui

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestDetailDocument_LinkAt_GivenHyperlinkUnderTheCursor_WhenResolving_ThenItReturnsTheHyperlinkTarget(t *testing.T) {
	document := newDetailDocument("\x1b]8;id=1;https://example.com/docs\aDocs\x1b]8;;\a", 80)

	actual, ok := document.linkAt(detailPosition{line: 0, column: 2})

	if !ok {
		t.Fatal("expected a link under the cursor")
	}
	if actual != "https://example.com/docs" {
		t.Fatalf("expected link %q, actual %q", "https://example.com/docs", actual)
	}
}

func TestDetailDocument_LinkAt_GivenRawURLUnderTheCursor_WhenResolving_ThenItReturnsTheVisibleURL(t *testing.T) {
	document := newDetailDocument("Visit https://example.com/docs now", 80)

	actual, ok := document.linkAt(detailPosition{line: 0, column: 10})

	if !ok {
		t.Fatal("expected a link under the cursor")
	}
	if actual != "https://example.com/docs" {
		t.Fatalf("expected link %q, actual %q", "https://example.com/docs", actual)
	}
}

func TestOpenLinkUnderCursor_GivenGXOnADetailLink_WhenOpening_ThenItUsesTheConfiguredLinkOpener(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Docs https://example.com/docs"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	opener := &fakeLinkOpener{}
	subject.linkOpener = opener
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailLink(t, subject, detailView, "https://example.com/docs")
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	goHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'g')
	xHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'x')
	actualErr = goHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = xHandler(gui, detailView)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(opener.urls, []string{"https://example.com/docs"}) {
		t.Fatalf("expected opened links %v, actual %v", []string{"https://example.com/docs"}, opener.urls)
	}
	then_statusLineContains(t, gui, openLinkSuccessMessage)
}

func TestActionsPopup_GivenDetailFocusWithALinkUnderCursor_WhenExecutingOpenLinkAction_ThenItUsesTheConfiguredLinkOpenerAndClosesThePopup(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.OpenDetail()
	subject := given_pullRequestCommentProgram(model, &fakePullRequestDetailLoader{})
	opener := &fakeLinkOpener{}
	subject.linkOpener = opener
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "Docs https://example.com/docs", URL: "https://github.com/acme/widgets/pull/42"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_cursorOnDetailLink(t, subject, detailView, "https://example.com/docs")
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Open link under cursor") {
		t.Fatalf("expected popup buffer to contain %q, actual %q", "Open link under cursor", popupView.Buffer())
	}

	subject.model.UpdateActionsPopupSearch("link under cursor", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "link under cursor"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(opener.urls, []string{"https://example.com/docs"}) {
		t.Fatalf("expected opened links %v, actual %v", []string{"https://example.com/docs"}, opener.urls)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_statusLineContains(t, gui, openLinkSuccessMessage)
}

func given_cursorOnDetailLink(t *testing.T, subject *Program, view *gocui.View, expectedLink string) {
	t.Helper()

	document := subject.currentDetailDocument(view)
	lineIndex, line := given_detailDocumentLineContaining(t, document, expectedLink)
	byteIndex := strings.Index(line, expectedLink)
	if byteIndex < 0 {
		t.Fatalf("expected line %q to contain link %q", line, expectedLink)
	}

	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: utf8.RuneCountInString(line[:byteIndex])}
	subject.detailViewState.preferredColumn = subject.detailViewState.cursor.column
}

type fakeLinkOpener struct {
	urls []string
	err  error
}

func (opener *fakeLinkOpener) Open(url string) error {
	opener.urls = append(opener.urls, url)
	return opener.err
}
