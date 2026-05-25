package tui

import (
	"fmt"
	"strings"
)

type actionsPopupPresenter struct {
	assigneePickerVisible bool
	themePickerVisible    bool
	reactionPickerVisible bool
	errorMessage          string
	confirmationMessage   string
	searchQuery           string
	searchText            string
	searchCursor          int
	filteredActionCount   int
	totalActionCount      int
	renderedLineCount     int
}

func (presenter actionsPopupPresenter) frame(maxX int, contentMaxY int) paneFrame {
	totalWidth := boundedQuarterWidth(maxX, actionsPopupMinWidth, actionsPopupFallbackWidth)
	totalHeight := presenter.height(contentMaxY)
	return centeredOverlayFrame(maxX, contentMaxY, totalWidth, totalHeight)
}

func (presenter actionsPopupPresenter) searchFrame(maxX int, contentMaxY int) paneFrame {
	popupFrame := presenter.frame(maxX, contentMaxY)
	return paneFrame{x0: popupFrame.x0, y0: popupFrame.y0, x1: popupFrame.x1, y1: popupFrame.y0 + 2}
}

func (presenter actionsPopupPresenter) listFrame(maxX int, contentMaxY int) paneFrame {
	popupFrame := presenter.frame(maxX, contentMaxY)
	return paneFrame{x0: popupFrame.x0, y0: popupFrame.y0 + 1, x1: popupFrame.x1, y1: popupFrame.y1}
}

func (presenter actionsPopupPresenter) height(contentMaxY int) int {
	totalHeight := maxInt(actionsPopupMinHeight, presenter.renderedLineCount+3)
	if totalHeight > contentMaxY-2 {
		totalHeight = maxInt(3, contentMaxY-2)
	}
	return totalHeight
}

func (presenter actionsPopupPresenter) title() string {
	title := presenter.baseTitle()
	message := strings.TrimSpace(presenter.errorMessage)
	if message == "" {
		message = strings.TrimSpace(presenter.confirmationMessage)
	}
	if message == "" {
		return title
	}
	return fmt.Sprintf("%s · %s", title, message)
}

func (presenter actionsPopupPresenter) footer() string {
	if presenter.assigneePickerVisible {
		return assigneePickerSearchFooterHint
	}
	if strings.TrimSpace(presenter.searchQuery) == "" {
		return ""
	}
	return fmt.Sprintf("%d of %d %s", presenter.filteredActionCount, presenter.totalActionCount, presenter.itemLabel())
}

func (presenter actionsPopupPresenter) promptText() string {
	return presenter.searchText
}

func (presenter actionsPopupPresenter) promptCursor() int {
	return presenter.searchCursor
}

func (presenter actionsPopupPresenter) emptyMessage() string {
	if presenter.assigneePickerVisible {
		if strings.TrimSpace(presenter.searchQuery) == "" {
			return assigneePickerSearchFooterHint
		}
		return "No matching assignees."
	}
	return "No matching actions."
}

func (presenter actionsPopupPresenter) baseTitle() string {
	title := "Actions"
	if presenter.assigneePickerVisible {
		return assigneePickerTitle
	}
	if presenter.themePickerVisible {
		return themePickerTitle
	}
	if presenter.reactionPickerVisible {
		return reactionPickerTitle
	}
	return title
}

func (presenter actionsPopupPresenter) itemLabel() string {
	if presenter.themePickerVisible {
		return "themes"
	}
	if presenter.reactionPickerVisible {
		return "reactions"
	}
	return "actions"
}
