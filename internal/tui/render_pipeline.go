package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

type ViewRenderer interface {
	Configure(*gocui.View)
	Render(*gocui.View)
}

type viewRendererFuncs struct {
	configure viewConfigurator
	render    viewRenderer
}

func (renderer viewRendererFuncs) Configure(view *gocui.View) {
	if view == nil || renderer.configure == nil {
		return
	}
	renderer.configure(view)
}

func (renderer viewRendererFuncs) Render(view *gocui.View) {
	if view == nil || renderer.render == nil {
		return
	}
	renderer.render(view)
}

type screenViewFrame struct {
	ViewName string
	Frame    paneFrame
	Visible  bool
	OnTop    bool
}

type PanelFrame struct {
	View ViewState
	screenViewFrame
}

type ScreenLayout struct {
	PanelFrames        []PanelFrame
	HiddenFrames       []screenViewFrame
	FooterFrames       []screenViewFrame
	StatusLine         screenViewFrame
	StatusLineKeyHints screenViewFrame
	OverlayFrames      []screenViewFrame
}

func (layout ScreenLayout) PanelFrameByViewNumber(number int) (PanelFrame, bool) {
	for _, frame := range layout.PanelFrames {
		if frame.View.Number == number {
			return frame, true
		}
	}
	return PanelFrame{}, false
}

func (layout ScreenLayout) OverlayFrame(viewName string) (screenViewFrame, bool) {
	for _, frame := range layout.OverlayFrames {
		if frame.ViewName == viewName {
			return frame, true
		}
	}
	return screenViewFrame{}, false
}

func (layout ScreenLayout) allFrames() []screenViewFrame {
	frames := make([]screenViewFrame, 0, len(layout.PanelFrames)+len(layout.HiddenFrames)+len(layout.FooterFrames)+len(layout.OverlayFrames)+2)
	for _, frame := range layout.PanelFrames {
		frames = append(frames, frame.screenViewFrame)
	}
	frames = append(frames, layout.HiddenFrames...)
	frames = append(frames, layout.FooterFrames...)
	frames = append(frames, layout.StatusLine)
	frames = append(frames, layout.OverlayFrames...)
	frames = append(frames, layout.StatusLineKeyHints)
	return frames
}

type screenComposition struct {
	Layout    ScreenLayout
	Renderers map[string]ViewRenderer
}

type modalEditorLayoutState struct {
	visible     bool
	totalHeight int
}

type pullRequestBuildRunPopupLayoutState struct {
	visible       bool
	body          string
	widthPercent  int
	heightPercent int
}

type transientErrorPopupLayoutState struct {
	message string
}

type screenLayoutInput struct {
	screenState               ScreenState
	contentMaxY               int
	paneLayoutSize            PaneLayoutSize
	fullscreenPane            Focus
	sidebarTopPaneHeight      int
	footer                    footerPresenter
	help                      helpPresenter
	helpVisible               bool
	searchPromptVisible       bool
	modalEditor               modalEditorLayoutState
	buildPopup                pullRequestBuildRunPopupLayoutState
	transientErrorPopup       transientErrorPopupLayoutState
	actionsPopupVisible       bool
	actionsPopupSearchVisible bool
	actionsPopup              actionsPopupPresenter
}

func buildScreenLayout(input screenLayoutInput, maxX int, maxY int) ScreenLayout {
	state := input.screenState
	sideFocus := state.ActiveSideView().Focus
	_, showNotifications := state.ViewByNumber(sidePanelNotificationsViewNumber)
	mainPaneLayout := calculateMainPaneLayoutWithSidebarState(maxX, input.contentMaxY, input.paneLayoutSize, input.fullscreenPane, sideFocus, input.sidebarTopPaneHeight, showNotifications)

	layout := ScreenLayout{
		PanelFrames:  make([]PanelFrame, 0, 4),
		HiddenFrames: []screenViewFrame{},
		FooterFrames: []screenViewFrame{
			{ViewName: viewUserFooterName, Visible: false, OnTop: true},
			{ViewName: viewPullRequestsFooterName, Visible: false, OnTop: true},
			{ViewName: viewNotificationsFooterName, Visible: false, OnTop: true},
			{ViewName: viewDetailFooterName, Visible: false, OnTop: true},
		},
		StatusLine:    screenViewFrame{ViewName: viewStatusLineName, Frame: bottomPromptFrame(maxX, maxY), Visible: true, OnTop: true},
		OverlayFrames: []screenViewFrame{},
	}

	panelFramesByName := map[string]PanelFrame{}
	if view, ok := state.ViewByNumber(mainPanelViewNumber); ok {
		appendScreenPanelFrame(&layout, panelFramesByName, view, mainPaneLayout.detail, mainPaneLayout.detailVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelUserViewNumber); ok {
		appendScreenPanelFrame(&layout, panelFramesByName, view, mainPaneLayout.user, mainPaneLayout.userVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelPullRequestsViewNumber); ok {
		appendScreenPanelFrame(&layout, panelFramesByName, view, mainPaneLayout.pullRequests, mainPaneLayout.pullRequestsVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelNotificationsViewNumber); ok {
		appendScreenPanelFrame(&layout, panelFramesByName, view, mainPaneLayout.notifications, mainPaneLayout.notificationsVisible)
	}

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewDetailName} {
		if _, ok := panelFramesByName[viewName]; ok {
			continue
		}
		layout.HiddenFrames = append(layout.HiddenFrames, screenViewFrame{ViewName: viewName})
	}

	for index, footerName := range []string{viewUserFooterName, viewPullRequestsFooterName, viewNotificationsFooterName, viewDetailFooterName} {
		focus := focusForFooterName(footerName)
		parentFrame, ok := panelFramesByName[paneViewName(focus)]
		if !ok || !parentFrame.Visible {
			layout.FooterFrames[index] = screenViewFrame{ViewName: footerName, Visible: false, OnTop: true}
			continue
		}
		text := strings.TrimSpace(input.footer.paneFooterStateFor(focus).Text())
		layout.FooterFrames[index] = screenViewFrame{ViewName: footerName, Frame: paneBottomOverlayFrame(parentFrame.Frame), Visible: text != "", OnTop: true}
	}

	for _, overlayViewName := range []string{viewHelpName, viewSearchName, viewModalEditorName, viewPullRequestBuildInfoName, viewActionsPopupChromeName, viewActionsPopupName, viewActionsPopupSearchName, viewTransientErrorPopupName} {
		layout.OverlayFrames = append(layout.OverlayFrames, overlayFrameForView(input, overlayViewName, maxX, maxY))
	}

	keyHintsText := strings.TrimSpace(input.footer.statusLineKeyHintsText())
	layout.StatusLineKeyHints = screenViewFrame{ViewName: viewStatusLineKeyHintsName, Visible: false, OnTop: true}
	if keyHintsText != "" {
		layout.StatusLineKeyHints = screenViewFrame{ViewName: viewStatusLineKeyHintsName, Frame: statusLineKeyHintsFrame(maxX, maxY, keyHintsText), Visible: true, OnTop: true}
	}

	return layout
}

func appendScreenPanelFrame(layout *ScreenLayout, panelFramesByName map[string]PanelFrame, view ViewState, frame paneFrame, visible bool) {
	if layout == nil {
		return
	}
	panelFrame := PanelFrame{View: view, screenViewFrame: screenViewFrame{ViewName: paneViewName(view.Focus), Frame: frame, Visible: visible}}
	layout.PanelFrames = append(layout.PanelFrames, panelFrame)
	panelFramesByName[panelFrame.ViewName] = panelFrame
}

func overlayFrameForView(input screenLayoutInput, viewName string, maxX int, maxY int) screenViewFrame {
	switch viewName {
	case viewHelpName:
		if !input.helpVisible {
			return screenViewFrame{ViewName: viewName}
		}
		innerWidth, innerHeight := input.help.viewSize(maxX, maxY)
		return screenViewFrame{ViewName: viewName, Frame: centeredOverlayFrame(maxX, maxY, innerWidth+2, innerHeight+2), Visible: true, OnTop: true}
	case viewSearchName:
		return screenViewFrame{ViewName: viewName, Frame: bottomPromptFrame(maxX, maxY), Visible: input.searchPromptVisible, OnTop: true}
	case viewModalEditorName:
		if !input.modalEditor.visible {
			return screenViewFrame{ViewName: viewName}
		}
		totalWidth := boundedHalfWidth(maxX, modalEditorMinWidth, modalEditorFallbackWidth)
		totalHeight := input.modalEditor.totalHeight
		if totalHeight < 1 {
			totalHeight = modalEditorTotalHeight
		}
		return screenViewFrame{ViewName: viewName, Frame: centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight), Visible: true, OnTop: true}
	case viewPullRequestBuildInfoName:
		if !input.buildPopup.visible {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: pullRequestBuildRunPopupOverlayFrame(input.buildPopup, maxX, maxY), Visible: true, OnTop: true}
	case viewTransientErrorPopupName:
		if strings.TrimSpace(input.transientErrorPopup.message) == "" {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: transientErrorPopupFrameForMessage(input.transientErrorPopup.message, maxX, maxY), Visible: true, OnTop: true}
	case viewActionsPopupChromeName:
		if !input.actionsPopupVisible {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: input.actionsPopup.frame(maxX, input.contentMaxY), Visible: true, OnTop: true}
	case viewActionsPopupName:
		if !input.actionsPopupVisible {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: input.actionsPopup.listFrame(maxX, input.contentMaxY), Visible: true, OnTop: true}
	case viewActionsPopupSearchName:
		return screenViewFrame{ViewName: viewName, Frame: input.actionsPopup.searchFrame(maxX, input.contentMaxY), Visible: input.actionsPopupSearchVisible, OnTop: true}
	default:
		return screenViewFrame{ViewName: viewName}
	}
}

func pullRequestBuildRunPopupOverlayFrame(layout pullRequestBuildRunPopupLayoutState, maxX int, maxY int) paneFrame {
	totalWidth := boundedHalfWidth(maxX, pullRequestBuildRunPopupMinWidth, pullRequestBuildRunPopupFallbackWidth)
	totalHeight := pullRequestBuildRunPopupMinHeight
	if layout.widthPercent > 0 {
		totalWidth = maxInt(10, (maxX*layout.widthPercent)/100)
	}
	if layout.heightPercent > 0 {
		totalHeight = maxInt(3, (maxY*layout.heightPercent)/100)
	} else {
		totalHeight = maxInt(totalHeight, renderedTextLineCount(strings.TrimSpace(layout.body))+2)
		if totalHeight > maxY-2 {
			totalHeight = maxInt(3, maxY-2)
		}
	}
	return centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight)
}

func (program *Program) applyScreenComposition(gui *gocui.Gui, composition screenComposition) error {
	if gui == nil {
		return nil
	}

	for _, frame := range composition.Layout.allFrames() {
		view, actualErr := setScreenView(gui, frame)
		if actualErr != nil {
			return actualErr
		}
		if view != nil {
			if renderer, ok := composition.Renderers[frame.ViewName]; ok {
				renderer.Configure(view)
				program.syncViewShellState(frame.ViewName, view)
				program.prepareViewRenderState(frame.ViewName, view)
				renderer.Render(view)
			}
		}
		if frame.Visible && frame.OnTop {
			if _, actualErr = gui.SetViewOnTop(frame.ViewName); actualErr != nil && !isUnknownViewError(actualErr) {
				return actualErr
			}
		}
	}

	program.clearVisibleListViewportPlacements(composition.Layout)
	program.syncRenderedDetailImages(gui)
	return nil
}

func setScreenView(gui *gocui.Gui, frame screenViewFrame) (*gocui.View, error) {
	if !frame.Visible {
		return nil, deleteViewIfPresent(gui, frame.ViewName)
	}

	view, actualErr := gui.SetView(frame.ViewName, frame.Frame.x0, frame.Frame.y0, frame.Frame.x1, frame.Frame.y1, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return nil, actualErr
	}
	return view, nil
}

func bottomPromptFrame(maxX int, maxY int) paneFrame {
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	return paneFrame{x0: -1, y0: maxY - 2, x1: maxX, y1: maxY}
}

func statusLineKeyHintsFrame(maxX int, maxY int, text string) paneFrame {
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}
	width := maxInt(1, runeCountInt(strings.TrimSpace(text)))
	x0 := max(maxX-width-1, -1)
	return paneFrame{x0: x0, y0: maxY - 2, x1: maxX, y1: maxY}
}

func paneBottomOverlayFrame(parent paneFrame) paneFrame {
	return paneFrame{x0: parent.x0, y0: parent.y1 - 1, x1: parent.x1, y1: parent.y1 + 1}
}

func focusForFooterName(viewName string) Focus {
	switch viewName {
	case viewPullRequestsFooterName:
		return FocusPullRequestsView
	case viewNotificationsFooterName:
		return FocusNotificationsView
	case viewDetailFooterName:
		return FocusDetailView
	default:
		return FocusUserView
	}
}
