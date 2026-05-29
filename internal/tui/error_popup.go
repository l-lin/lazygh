package tui

import (
	"strings"
	"time"

	"github.com/jesseduffield/gocui"
)

const (
	defaultTransientErrorPopupDuration = 5 * time.Second
	transientErrorPopupTitle           = iconWarning + " Error"
	transientErrorPopupMinWidth        = 30
	transientErrorPopupFallbackWidth   = 50
	transientErrorPopupReservedRows    = 2
	maxRecordedErrorMessages           = 100
	recentErrorsActionTitle            = "View recent errors"
	recentErrorsPopupTitle             = "Recent errors"
	recentErrorsPopupWidthPercent      = 90
	recentErrorsPopupHeightPercent     = 90
)

type transientErrorPopupState struct {
	message    string
	expiresAt  time.Time
	generation uint64
}

func recordedErrorMessagesWithAppended(existing []string, message string) []string {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return append([]string(nil), existing...)
	}

	updated := append(append([]string(nil), existing...), trimmedMessage)
	if len(updated) <= maxRecordedErrorMessages {
		return updated
	}
	return append([]string(nil), updated[len(updated)-maxRecordedErrorMessages:]...)
}

func newTransientErrorPopupState(previous transientErrorPopupState, message string, now time.Time, duration time.Duration) transientErrorPopupState {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return previous
	}

	popup := transientErrorPopupState{message: trimmedMessage, generation: previous.generation + 1}
	if duration > 0 {
		popup.expiresAt = now.Add(duration)
	}
	return popup
}

func (program *Program) hasRecordedErrors() bool {
	return program != nil && len(program.overlayState.errorMessages) > 0
}

func (program *Program) currentRecentErrorsActionsPopupAction() (actionsPopupAction, bool) {
	if !program.hasRecordedErrors() {
		return actionsPopupAction{}, false
	}
	return program.recentErrorsActionsPopupAction(), true
}

func (program *Program) recentErrorsActionsPopupAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if program.hasRecordedErrors() {
		requested = MsgPullRequestBuildRunPopupOpened{Content: pullRequestBuildRunPopupContent{
			title:         recentErrorsPopupTitle,
			body:          program.renderRecentErrorsPopupBody(),
			widthPercent:  recentErrorsPopupWidthPercent,
			heightPercent: recentErrorsPopupHeightPercent,
		}}
	}
	return actionsPopupAction{
		id:        "view-recent-errors",
		title:     recentErrorsActionTitle,
		icon:      actionsPopupRecentErrorsIcon,
		requested: requested,
	}
}

func (program *Program) renderRecentErrorsPopupBody() string {
	if !program.hasRecordedErrors() {
		return "No recent errors recorded."
	}
	messages := make([]string, 0, len(program.overlayState.errorMessages))
	for index := len(program.overlayState.errorMessages) - 1; index >= 0; index-- {
		messages = append(messages, program.overlayState.errorMessages[index])
	}
	return strings.Join(messages, "\n\n")
}

func (program *Program) clearExpiredTransientErrorPopup(now time.Time) bool {
	if program == nil {
		return false
	}
	if !program.transientErrorPopupVisible() {
		return false
	}
	if program.overlayState.transientErrorPopup.expiresAt.IsZero() || now.Before(program.overlayState.transientErrorPopup.expiresAt) {
		return false
	}

	program.clearTransientErrorPopupState()
	return true
}

func (program *Program) transientErrorPopupVisible() bool {
	return program != nil && strings.TrimSpace(program.overlayState.transientErrorPopup.message) != ""
}

func (program *Program) configureTransientErrorPopupView(view *gocui.View) {
	configureFramedOverlayView(view, transientErrorPopupTitle, "")
	view.Wrap = false
	view.Editable = false
	view.Editor = nil
	view.Highlight = false
	view.HighlightInactive = false
}

func (program *Program) renderTransientErrorPopupView(view *gocui.View) {
	if view == nil || !program.transientErrorPopupVisible() {
		return
	}

	view.Clear()
	for index, line := range program.transientErrorPopupLines(maxInt(1, view.InnerWidth())) {
		if index > 0 {
			_, _ = view.Write([]byte("\n"))
		}
		_, _ = view.Write([]byte(line))
	}
	view.SetOrigin(0, 0)
	view.SetCursor(0, 0)
}

func (program *Program) transientErrorPopupFrame(maxX int, maxY int) paneFrame {
	return transientErrorPopupFrameForMessage(program.overlayState.transientErrorPopup.message, maxX, maxY)
}

func transientErrorPopupFrameForMessage(message string, maxX int, maxY int) paneFrame {
	trimmedMessage := strings.TrimSpace(message)
	maxWidth := boundedHalfWidth(maxX, transientErrorPopupMinWidth, transientErrorPopupFallbackWidth)
	longestLineWidth := transientErrorPopupMinWidth
	for _, line := range strings.Split(trimmedMessage, "\n") {
		lineWidth := runeCountInt(strings.TrimRight(line, "\r"))
		if lineWidth > longestLineWidth {
			longestLineWidth = lineWidth
		}
	}
	contentWidth := clampInt(longestLineWidth+2, transientErrorPopupMinWidth, maxWidth)
	lines := transientErrorPopupLinesForMessage(trimmedMessage, maxInt(1, contentWidth-2))
	totalHeight := len(lines) + 2
	if totalHeight < 3 {
		totalHeight = 3
	}
	maxHeight := maxInt(3, maxY-transientErrorPopupReservedRows)
	if totalHeight > maxHeight {
		totalHeight = maxHeight
	}
	return bottomRightOverlayFrame(maxX, maxY, contentWidth, totalHeight, transientErrorPopupReservedRows)
}

func (program *Program) transientErrorPopupLines(innerWidth int) []string {
	return transientErrorPopupLinesForMessage(program.overlayState.transientErrorPopup.message, innerWidth)
}

func transientErrorPopupLinesForMessage(message string, innerWidth int) []string {
	trimmedMessage := strings.TrimSpace(message)
	if trimmedMessage == "" {
		return []string{""}
	}

	document := newDetailDocumentWithWrap(trimmedMessage, maxInt(1, innerWidth), true)
	lines := make([]string, 0, document.rowCount())
	for _, row := range document.rows {
		if row.empty {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, row.text)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func bottomRightOverlayFrame(maxX int, maxY int, totalWidth int, totalHeight int, reservedBottomRows int) paneFrame {
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	if reservedBottomRows < 0 {
		reservedBottomRows = 0
	}

	totalWidth = clampOverlayDimension(totalWidth, maxX)
	availableHeight := maxInt(1, maxY-reservedBottomRows)
	totalHeight = clampOverlayDimension(totalHeight, availableHeight)

	x1 := maxX - 1
	x0 := clampCoordinate(x1-totalWidth+1, maxX)
	y1 := maxY - reservedBottomRows - 1
	if y1 < 0 {
		y1 = 0
	}
	y0 := y1 - totalHeight + 1
	if y0 < 0 {
		y0 = 0
		y1 = minInt(maxY-1, y0+totalHeight-1)
	}

	return paneFrame{x0: x0, y0: y0, x1: x1, y1: y1}
}
