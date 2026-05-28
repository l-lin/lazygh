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

type MainPanelRenderer struct {
	program *Program
}

func (program *Program) mainPanelRenderer() MainPanelRenderer {
	return MainPanelRenderer{program: program}
}

func (renderer MainPanelRenderer) Renderer(frame PanelFrame) (ViewRenderer, bool) {
	if frame.View.Focus != FocusDetailView {
		return nil, false
	}
	return viewRendererFuncs{configure: renderer.program.configureDetailView, render: renderer.program.renderDetailView}, true
}

type SidePanelRenderer struct {
	program *Program
}

func (program *Program) sidePanelRenderer() SidePanelRenderer {
	return SidePanelRenderer{program: program}
}

func (renderer SidePanelRenderer) Renderer(frame PanelFrame) (ViewRenderer, bool) {
	switch frame.View.Focus {
	case FocusUserView:
		return viewRendererFuncs{configure: renderer.program.configureUserView, render: renderer.program.renderUserView}, true
	case FocusPullRequestsView:
		return viewRendererFuncs{configure: renderer.program.configurePullRequestsView, render: renderer.program.renderPullRequestsView}, true
	case FocusNotificationsView:
		return viewRendererFuncs{configure: renderer.program.configureNotificationsView, render: renderer.program.renderNotificationsView}, true
	default:
		return nil, false
	}
}

type OverlayRenderer struct {
	program *Program
}

func (program *Program) overlayRenderer() OverlayRenderer {
	return OverlayRenderer{program: program}
}

func (renderer OverlayRenderer) Renderer(viewName string) (ViewRenderer, bool) {
	switch viewName {
	case viewHelpName:
		return viewRendererFuncs{configure: renderer.program.configureHelpView, render: renderer.program.renderHelpView}, true
	case viewSearchName:
		return viewRendererFuncs{configure: renderer.program.configureSearchView, render: renderer.program.renderSearchView}, true
	case viewModalEditorName:
		return viewRendererFuncs{configure: renderer.program.configureModalEditorView, render: renderer.program.renderModalEditorView}, true
	case viewPullRequestBuildInfoName:
		return viewRendererFuncs{configure: renderer.program.configurePullRequestBuildRunPopupView, render: renderer.program.renderPullRequestBuildRunPopupView}, true
	case viewTransientErrorPopupName:
		return viewRendererFuncs{configure: renderer.program.configureTransientErrorPopupView, render: renderer.program.renderTransientErrorPopupView}, true
	case viewActionsPopupChromeName:
		return viewRendererFuncs{configure: renderer.program.configureActionsPopupChromeView, render: renderer.program.renderActionsPopupChromeView}, true
	case viewActionsPopupName:
		return viewRendererFuncs{configure: renderer.program.configureActionsPopupView, render: renderer.program.renderActionsPopupView}, true
	case viewActionsPopupSearchName:
		return viewRendererFuncs{configure: renderer.program.configureActionsPopupSearchView, render: renderer.program.renderActionsPopupSearchView}, true
	case viewUserFooterName, viewPullRequestsFooterName, viewNotificationsFooterName, viewDetailFooterName:
		text := renderer.program.paneFooterTextForView(viewName)
		return viewRendererFuncs{
			configure: renderer.program.configurePaneFooterView,
			render: func(view *gocui.View) {
				renderer.program.renderPaneFooterView(view, text)
			},
		}, true
	default:
		return nil, false
	}
}

func (renderer OverlayRenderer) Frame(viewName string, maxX int, maxY int) screenViewFrame {
	switch viewName {
	case viewHelpName:
		if !renderer.program.overlayState.helpVisible {
			return screenViewFrame{ViewName: viewName}
		}
		innerWidth, innerHeight := renderer.program.helpViewSize(maxX, maxY)
		return screenViewFrame{ViewName: viewName, Frame: centeredOverlayFrame(maxX, maxY, innerWidth+2, innerHeight+2), Visible: true, OnTop: true}
	case viewSearchName:
		return screenViewFrame{ViewName: viewName, Frame: bottomPromptFrame(maxX, maxY), Visible: renderer.program.searchPromptVisible(), OnTop: true}
	case viewModalEditorName:
		if !renderer.program.modalEditorVisible() {
			return screenViewFrame{ViewName: viewName}
		}
		totalWidth := boundedHalfWidth(maxX, modalEditorMinWidth, modalEditorFallbackWidth)
		totalHeight := modalEditorTotalHeight
		if renderer.program.modalEditorVisible() {
			totalHeight = renderer.program.overlayState.modalEditor.Height()
		}
		return screenViewFrame{ViewName: viewName, Frame: centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight), Visible: true, OnTop: true}
	case viewPullRequestBuildInfoName:
		if !renderer.program.pullRequestBuildRunPopupVisible() {
			return screenViewFrame{ViewName: viewName}
		}
		totalWidth := boundedHalfWidth(maxX, pullRequestBuildRunPopupMinWidth, pullRequestBuildRunPopupFallbackWidth)
		totalHeight := pullRequestBuildRunPopupMinHeight
		if popup := renderer.program.pullRequestBuildRunPopup; popup != nil {
			if popup.widthPercent > 0 {
				totalWidth = maxInt(10, (maxX*popup.widthPercent)/100)
			}
			if popup.heightPercent > 0 {
				totalHeight = maxInt(3, (maxY*popup.heightPercent)/100)
			} else {
				totalHeight = maxInt(totalHeight, renderedTextLineCount(strings.TrimSpace(popup.body))+2)
				if totalHeight > maxY-2 {
					totalHeight = maxInt(3, maxY-2)
				}
			}
		}
		return screenViewFrame{ViewName: viewName, Frame: centeredOverlayFrame(maxX, maxY, totalWidth, totalHeight), Visible: true, OnTop: true}
	case viewTransientErrorPopupName:
		if !renderer.program.transientErrorPopupVisible() {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: renderer.program.transientErrorPopupFrame(maxX, maxY), Visible: true, OnTop: true}
	case viewActionsPopupChromeName:
		if !renderer.program.model.ActionsPopupVisible() {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: renderer.program.actionsPopupFrame(maxX, maxY), Visible: true, OnTop: true}
	case viewActionsPopupName:
		if !renderer.program.model.ActionsPopupVisible() {
			return screenViewFrame{ViewName: viewName}
		}
		return screenViewFrame{ViewName: viewName, Frame: renderer.program.actionsPopupListFrame(maxX, maxY), Visible: true, OnTop: true}
	case viewActionsPopupSearchName:
		return screenViewFrame{ViewName: viewName, Frame: renderer.program.actionsPopupSearchFrame(maxX, maxY), Visible: renderer.program.model.ActionsPopupSearchActive(), OnTop: true}
	default:
		return screenViewFrame{ViewName: viewName}
	}
}

type StatusLinePresenter struct {
	program *Program
}

func (program *Program) statusLinePresenter() StatusLinePresenter {
	return StatusLinePresenter{program: program}
}

func (presenter StatusLinePresenter) Text() string {
	return strings.TrimSpace(presenter.program.statusLineText())
}

func (presenter StatusLinePresenter) Renderer() ViewRenderer {
	return viewRendererFuncs{configure: presenter.program.configureStatusLineView, render: presenter.program.renderStatusLineView}
}

type KeyHintPresenter struct {
	program *Program
	footer  footerPresenter
}

func (program *Program) keyHintPresenter() KeyHintPresenter {
	return KeyHintPresenter{program: program, footer: program.footerPresenter()}
}

func (presenter KeyHintPresenter) Text() string {
	return strings.TrimSpace(presenter.footer.statusLineKeyHintsText())
}

func (presenter KeyHintPresenter) Renderer() (ViewRenderer, bool) {
	text := presenter.Text()
	if text == "" {
		return nil, false
	}
	return viewRendererFuncs{
		configure: presenter.program.configureStatusLineKeyHintsView,
		render: func(view *gocui.View) {
			presenter.program.renderStatusLineKeyHintsView(view, text)
		},
	}, true
}

func (program *Program) screenLayoutForSize(maxX int, maxY int) ScreenLayout {
	state := program.screenState()
	contentMaxY := program.layoutContentHeight(maxY)
	sideFocus := state.ActiveSideView().Focus
	_, showNotifications := state.ViewByNumber(sidePanelNotificationsViewNumber)
	mainPaneLayout := calculateMainPaneLayoutWithSidebarState(maxX, contentMaxY, program.model.PaneLayoutSize(), program.model.FullscreenPane(), sideFocus, program.sidebarTopPaneHeight(), showNotifications)

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
	appendPanelFrame := func(view ViewState, frame paneFrame, visible bool) {
		panelFrame := PanelFrame{View: view, screenViewFrame: screenViewFrame{ViewName: paneViewName(view.Focus), Frame: frame, Visible: visible}}
		layout.PanelFrames = append(layout.PanelFrames, panelFrame)
		panelFramesByName[panelFrame.ViewName] = panelFrame
	}

	if view, ok := state.ViewByNumber(mainPanelViewNumber); ok {
		appendPanelFrame(view, mainPaneLayout.detail, mainPaneLayout.detailVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelUserViewNumber); ok {
		appendPanelFrame(view, mainPaneLayout.user, mainPaneLayout.userVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelPullRequestsViewNumber); ok {
		appendPanelFrame(view, mainPaneLayout.pullRequests, mainPaneLayout.pullRequestsVisible)
	}
	if view, ok := state.ViewByNumber(sidePanelNotificationsViewNumber); ok {
		appendPanelFrame(view, mainPaneLayout.notifications, mainPaneLayout.notificationsVisible)
	}

	for _, viewName := range []string{viewUserName, viewPullRequestsName, viewNotificationsName, viewDetailName} {
		if _, ok := panelFramesByName[viewName]; ok {
			continue
		}
		layout.HiddenFrames = append(layout.HiddenFrames, screenViewFrame{ViewName: viewName})
	}

	footerPresenter := program.footerPresenter()
	for index, footerName := range []string{viewUserFooterName, viewPullRequestsFooterName, viewNotificationsFooterName, viewDetailFooterName} {
		focus := focusForFooterName(footerName)
		parentName := paneViewName(focus)
		parentFrame, ok := panelFramesByName[parentName]
		if !ok || !parentFrame.Visible {
			layout.FooterFrames[index] = screenViewFrame{ViewName: footerName, Visible: false, OnTop: true}
			continue
		}
		text := strings.TrimSpace(footerPresenter.paneFooterStateFor(focus).Text())
		layout.FooterFrames[index] = screenViewFrame{ViewName: footerName, Frame: paneBottomOverlayFrame(parentFrame.Frame), Visible: text != "", OnTop: true}
	}

	overlayRenderer := program.overlayRenderer()
	for _, overlayViewName := range []string{viewHelpName, viewSearchName, viewModalEditorName, viewPullRequestBuildInfoName, viewActionsPopupChromeName, viewActionsPopupName, viewActionsPopupSearchName, viewTransientErrorPopupName} {
		layout.OverlayFrames = append(layout.OverlayFrames, overlayRenderer.Frame(overlayViewName, maxX, maxY))
	}

	keyHintsText := strings.TrimSpace(footerPresenter.statusLineKeyHintsText())
	layout.StatusLineKeyHints = screenViewFrame{ViewName: viewStatusLineKeyHintsName, Visible: false, OnTop: true}
	if keyHintsText != "" {
		layout.StatusLineKeyHints = screenViewFrame{ViewName: viewStatusLineKeyHintsName, Frame: statusLineKeyHintsFrame(maxX, maxY, keyHintsText), Visible: true, OnTop: true}
	}

	return layout
}

func (program *Program) screenCompositionForSize(maxX int, maxY int) screenComposition {
	layout := program.screenLayoutForSize(maxX, maxY)
	renderers := map[string]ViewRenderer{}
	mainPanelRenderer := program.mainPanelRenderer()
	sidePanelRenderer := program.sidePanelRenderer()
	overlayRenderer := program.overlayRenderer()

	for _, frame := range layout.PanelFrames {
		if frame.View.Number == mainPanelViewNumber {
			if renderer, ok := mainPanelRenderer.Renderer(frame); ok {
				renderers[frame.ViewName] = renderer
			}
			continue
		}
		if renderer, ok := sidePanelRenderer.Renderer(frame); ok {
			renderers[frame.ViewName] = renderer
		}
	}
	for _, frame := range layout.FooterFrames {
		if renderer, ok := overlayRenderer.Renderer(frame.ViewName); ok {
			renderers[frame.ViewName] = renderer
		}
	}
	for _, frame := range layout.OverlayFrames {
		if renderer, ok := overlayRenderer.Renderer(frame.ViewName); ok {
			renderers[frame.ViewName] = renderer
		}
	}
	renderers[viewStatusLineName] = program.statusLinePresenter().Renderer()
	if renderer, ok := program.keyHintPresenter().Renderer(); ok {
		renderers[viewStatusLineKeyHintsName] = renderer
	}

	return screenComposition{Layout: layout, Renderers: renderers}
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

func (program *Program) paneFooterTextForView(viewName string) string {
	return strings.TrimSpace(program.footerPresenter().paneFooterStateFor(focusForFooterName(viewName)).Text())
}
