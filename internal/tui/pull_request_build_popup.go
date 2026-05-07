package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	pullRequestBuildRunPopupFallbackWidth = 90
	pullRequestBuildRunPopupMinWidth      = 60
	pullRequestBuildRunPopupMinHeight     = 16
)

type pullRequestBuildRunLoadState struct {
	command string
}

type pullRequestBuildRunPopupContent struct {
	checkTitle string
	runURL     string
	body       string
}

type pullRequestBuildRunPopupState struct {
	title     string
	runURL    string
	body      string
	viewState detailViewState
	documents map[int]detailDocument
}

func (program *Program) pullRequestBuildRunPopupVisible() bool {
	return program != nil && program.pullRequestBuildRunPopup != nil
}

func (program *Program) openPullRequestBuildRunPopup(gui *gocui.Gui, content pullRequestBuildRunPopupContent) error {
	program.pullRequestBuildRunPopup = &pullRequestBuildRunPopupState{
		title:     pullRequestBuildRunPopupTitle(content.checkTitle),
		runURL:    strings.TrimSpace(content.runURL),
		body:      renderPullRequestBuildRunPopupContent(content),
		viewState: newDetailViewState(),
		documents: map[int]detailDocument{},
	}
	if gui == nil {
		return nil
	}
	return program.layout(gui)
}

func (program *Program) closePullRequestBuildRunPopup(gui *gocui.Gui, _ *gocui.View) error {
	if popup := program.pullRequestBuildRunPopup; popup != nil && popup.viewState.mode.isVisual() {
		popup.viewState.exitVisualMode()
		return program.refreshViewsIfGUI(gui)
	}

	program.pullRequestBuildRunPopup = nil
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) layoutPullRequestBuildRunPopupView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	totalWidth := boundedHalfWidth(maxX, pullRequestBuildRunPopupMinWidth, pullRequestBuildRunPopupFallbackWidth)
	totalHeight := pullRequestBuildRunPopupMinHeight
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		totalHeight = maxInt(totalHeight, renderedTextLineCount(strings.TrimSpace(popup.body))+2)
	}
	if totalHeight > maxY-2 {
		totalHeight = maxInt(3, maxY-2)
	}
	frame := centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight)

	view, err := gui.SetView(viewPullRequestBuildInfoName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}

	program.configurePullRequestBuildRunPopupView(view)
	program.renderPullRequestBuildRunPopupView(view)
	_, err = gui.SetViewOnTop(viewPullRequestBuildInfoName)
	if isUnknownViewError(err) {
		return nil
	}
	return err
}

func (program *Program) configurePullRequestBuildRunPopupView(view *gocui.View) {
	title := ""
	if program.pullRequestBuildRunPopup != nil {
		title = program.pullRequestBuildRunPopup.title
	}
	configureFramedOverlayView(view, title, "")
	view.Wrap = false
	view.Editable = false
	view.Editor = nil
	view.Highlight = false
}

func (program *Program) renderPullRequestBuildRunPopupView(view *gocui.View) {
	if view == nil || program.pullRequestBuildRunPopup == nil {
		return
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(view))
	renderDetailDocumentView(view, document, program.pullRequestBuildRunPopup.viewState)
}

func (program *Program) currentPullRequestBuildRunPopupDocument(view *gocui.View) detailDocument {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return detailDocument{}
	}

	width := 1
	if view != nil && view.InnerWidth() > 0 {
		width = view.InnerWidth()
	}
	if width < 1 {
		width = 1
	}
	if popup.documents != nil {
		if document, ok := popup.documents[width]; ok {
			return document
		}
	} else {
		popup.documents = map[int]detailDocument{}
	}

	document := newDetailDocumentWithWrap(popup.body, width, false)
	popup.documents[width] = document
	return document
}

func (program *Program) syncPullRequestBuildRunPopupViewState(document detailDocument, viewportHeight int) {
	if program.pullRequestBuildRunPopup == nil {
		return
	}
	program.pullRequestBuildRunPopup.viewState.sync(document, viewportHeight)
}

func (program *Program) currentPullRequestBuildRunPopupLink(view *gocui.View) (string, bool) {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return "", false
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(view))
	return document.linkAt(popup.viewState.cursor)
}

func pullRequestBuildRunPopupTitle(checkTitle string) string {
	trimmedTitle := strings.TrimSpace(checkTitle)
	if trimmedTitle == "" {
		return "Build run"
	}
	return "Build run · " + trimmedTitle
}

func renderPullRequestBuildRunPopupContent(content pullRequestBuildRunPopupContent) string {
	lines := make([]string, 0, 3)
	if runURL := strings.TrimSpace(content.runURL); runURL != "" {
		lines = append(lines, "Run: "+runURL, "")
	}

	body := strings.TrimSpace(content.body)
	if body == "" {
		body = "No build run details available."
	}
	lines = append(lines, body)
	return strings.Join(lines, "\n")
}
