package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pullRequestBuildRunPopupFallbackWidth  = 90
	pullRequestBuildRunPopupMinWidth       = 60
	pullRequestBuildRunPopupMinHeight      = 16
	pullRequestBuildLogsPopupWidthPercent  = 90
	pullRequestBuildLogsPopupHeightPercent = 90
	pullRequestBuildRunUnknownStepLabel    = "UNKNOWN STEP"
)

type pullRequestBuildRunLoadState struct {
	command string
}

type pullRequestBuildRunPopupContent struct {
	title         string
	checkTitle    string
	runURL        string
	repository    string
	body          string
	jobs          []githubdomain.PullRequestBuildRunJob
	previousPopup *pullRequestBuildRunPopupState
	widthPercent  int
	heightPercent int
}

type pullRequestBuildRunPopupState struct {
	title         string
	runURL        string
	repository    string
	body          string
	jobs          []githubdomain.PullRequestBuildRunJob
	previousPopup *pullRequestBuildRunPopupState
	widthPercent  int
	heightPercent int
	searchQuery   string
	searchActive  bool
	viewState     detailViewState
	documents     map[int]detailDocument
}

func (program *Program) pullRequestBuildRunPopupVisible() bool {
	return program != nil && program.pullRequestBuildRunPopup != nil
}

func (program *Program) openPullRequestBuildRunPopup(gui *gocui.Gui, content pullRequestBuildRunPopupContent) error {
	return program.dispatch(gui, MsgPullRequestBuildRunPopupOpened{Content: content})
}

func (program *Program) closePullRequestBuildRunPopup(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgPullRequestBuildRunPopupClosed{})
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
	program.pullRequestBuildRunPopup.viewState.syncSearch(document, program.pullRequestBuildRunPopup.searchQuery)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(view))
	renderVisibleDetailDocumentView(view, document, program.pullRequestBuildRunPopup.viewState)
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

func pullRequestBuildRunLogsPopupTitle(jobName string) string {
	trimmedName := strings.TrimSpace(jobName)
	if trimmedName == "" {
		return "Build logs"
	}
	return "Build logs · " + trimmedName
}

func renderPullRequestBuildRunPopupContent(content pullRequestBuildRunPopupContent) string {
	sections := make([]string, 0, 3)
	if runURL := strings.TrimSpace(content.runURL); runURL != "" {
		sections = append(sections, "Run: "+runURL)
	}

	body := strings.TrimSpace(content.body)
	if body == "" {
		body = "No build run details available."
	}
	sections = append(sections, body)

	if renderedJobs := renderPullRequestBuildRunPopupJobs(content.jobs); renderedJobs != "" {
		sections = append(sections, renderedJobs)
	}
	return strings.Join(sections, "\n\n")
}

func renderPullRequestBuildRunPopupJobs(jobs []githubdomain.PullRequestBuildRunJob) string {
	if len(jobs) == 0 {
		return ""
	}

	lines := []string{"Jobs"}
	for _, job := range jobs {
		lines = append(lines, renderPullRequestBuildRunPopupJobLine(job))
	}
	return strings.Join(lines, "\n")
}

func renderPullRequestBuildRunPopupJobLine(job githubdomain.PullRequestBuildRunJob) string {
	jobName := strings.TrimSpace(job.Name)
	if jobName == "" {
		jobName = "Job"
	}
	if job.DatabaseID > 0 {
		return fmt.Sprintf("Job: %s (#%d)", jobName, job.DatabaseID)
	}
	return "Job: " + jobName
}

func sanitizePullRequestBuildRunLog(raw string) string {
	trimmedRaw := strings.TrimSpace(strings.ReplaceAll(raw, "\r", ""))
	if trimmedRaw == "" {
		return "No build logs available."
	}

	lines := strings.Split(trimmedRaw, "\n")
	for index, line := range lines {
		if markerIndex := strings.Index(line, pullRequestBuildRunUnknownStepLabel); markerIndex >= 0 {
			suffixStart := min(markerIndex+len(pullRequestBuildRunUnknownStepLabel), len(line))
			lines[index] = strings.TrimSpace(line[suffixStart:])
			continue
		}
		lines[index] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}
