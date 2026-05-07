package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	pullRequestBuildInfoPopupFallbackWidth = 70
	pullRequestBuildInfoPopupMinWidth      = 50
	pullRequestBuildInfoPopupMinHeight     = 8
)

type pullRequestBuildInfoPopupState struct {
	title   string
	content string
}

func (program *Program) pullRequestBuildInfoPopupVisible() bool {
	return program != nil && program.pullRequestBuildInfoPopup != nil
}

func (program *Program) openPullRequestBuildInfoPopup(gui *gocui.Gui, buildInfo githubcli.PullRequestBuildInfo) error {
	program.pullRequestBuildInfoPopup = &pullRequestBuildInfoPopupState{
		title:   pullRequestBuildInfoPopupTitle(buildInfo),
		content: renderPullRequestBuildInfoPopupContent(buildInfo),
	}
	if gui == nil {
		return nil
	}
	return program.layout(gui)
}

func (program *Program) closePullRequestBuildInfoPopup(gui *gocui.Gui, _ *gocui.View) error {
	program.pullRequestBuildInfoPopup = nil
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) layoutPullRequestBuildInfoPopupView(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()
	totalWidth := boundedHalfWidth(maxX, pullRequestBuildInfoPopupMinWidth, pullRequestBuildInfoPopupFallbackWidth)
	totalHeight := pullRequestBuildInfoPopupMinHeight
	if popup := program.pullRequestBuildInfoPopup; popup != nil {
		totalHeight = maxInt(totalHeight, renderedTextLineCount(strings.TrimSpace(popup.content))+2)
	}
	if totalHeight > maxY-2 {
		totalHeight = maxInt(3, maxY-2)
	}
	frame := centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight)

	view, err := gui.SetView(viewPullRequestBuildInfoName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return err
	}

	program.configurePullRequestBuildInfoPopupView(view)
	program.renderPullRequestBuildInfoPopupView(view)
	_, err = gui.SetViewOnTop(viewPullRequestBuildInfoName)
	if isUnknownViewError(err) {
		return nil
	}
	return err
}

func (program *Program) configurePullRequestBuildInfoPopupView(view *gocui.View) {
	title := ""
	if program.pullRequestBuildInfoPopup != nil {
		title = program.pullRequestBuildInfoPopup.title
	}
	configureFramedOverlayView(view, title, "")
	view.Wrap = true
	view.Editable = false
	view.Highlight = false
}

func (program *Program) renderPullRequestBuildInfoPopupView(view *gocui.View) {
	if view == nil || program.pullRequestBuildInfoPopup == nil {
		return
	}
	renderReadOnlyTextView(view, program.pullRequestBuildInfoPopup.content)
}

func pullRequestBuildInfoPopupTitle(buildInfo githubcli.PullRequestBuildInfo) string {
	title := pullRequestBuildInfoDisplayName(buildInfo)
	if title == "" {
		return "Build info"
	}
	return "Build info · " + title
}

func pullRequestBuildInfoDisplayName(buildInfo githubcli.PullRequestBuildInfo) string {
	workflow := strings.TrimSpace(buildInfo.Workflow)
	name := strings.TrimSpace(buildInfo.Name)
	switch {
	case workflow != "" && name != "" && !strings.EqualFold(workflow, name):
		return workflow + " / " + name
	case workflow != "":
		return workflow
	case name != "":
		return name
	default:
		return ""
	}
}

func renderPullRequestBuildInfoPopupContent(buildInfo githubcli.PullRequestBuildInfo) string {
	lines := []string{fmt.Sprintf("Status: %s", pullRequestBuildInfoStateLabel(buildInfo))}
	if workflow := strings.TrimSpace(buildInfo.Workflow); workflow != "" {
		lines = append(lines, "Workflow: "+workflow)
	}
	if name := strings.TrimSpace(buildInfo.Name); name != "" {
		lines = append(lines, "Name: "+name)
	}
	if event := strings.TrimSpace(buildInfo.Event); event != "" {
		lines = append(lines, "Event: "+event)
	}
	if startedAt := strings.TrimSpace(buildInfo.StartedAt); startedAt != "" {
		lines = append(lines, "Started: "+formatTimestamp(startedAt))
	}
	if completedAt := strings.TrimSpace(buildInfo.CompletedAt); completedAt != "" {
		lines = append(lines, "Completed: "+formatTimestamp(completedAt))
	}
	if description := strings.TrimSpace(buildInfo.Description); description != "" {
		lines = append(lines, "Description: "+description)
	}
	if link := strings.TrimSpace(buildInfo.Link); link != "" {
		lines = append(lines, "Link: "+link)
	}
	return strings.Join(lines, "\n")
}

func pullRequestBuildInfoStateLabel(buildInfo githubcli.PullRequestBuildInfo) string {
	switch strings.ToLower(strings.TrimSpace(buildInfo.Bucket)) {
	case "pass":
		return "Successful"
	case "fail":
		return "Failed"
	case "pending":
		return "Pending"
	case "skipping":
		return "Skipped"
	case "cancel":
		return "Cancelled"
	}

	switch strings.ToUpper(strings.TrimSpace(buildInfo.State)) {
	case "SUCCESS":
		return "Successful"
	case "FAILURE", "ERROR":
		return "Failed"
	case "PENDING", "IN_PROGRESS", "QUEUED", "REQUESTED", "WAITING":
		return "Pending"
	case "SKIPPED":
		return "Skipped"
	case "CANCELLED":
		return "Cancelled"
	}

	trimmedState := strings.TrimSpace(buildInfo.State)
	if trimmedState == "" {
		return "-"
	}
	return trimmedState
}
