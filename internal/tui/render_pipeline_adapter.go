package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) screenLayoutForSize(maxX int, maxY int) ScreenLayout {
	return buildScreenLayout(program.screenLayoutInput(maxY), maxX, maxY)
}

func (program *Program) screenCompositionForSize(maxX int, maxY int) screenComposition {
	if program == nil {
		return screenComposition{}
	}

	input := program.screenLayoutInput(maxY)
	return screenComposition{Layout: buildScreenLayout(input, maxX, maxY), Renderers: program.screenRenderers(input)}
}

func (program *Program) screenLayoutInput(maxY int) screenLayoutInput {
	if program == nil {
		return screenLayoutInput{}
	}

	modalEditor := modalEditorLayoutState{visible: program.modalEditorVisible(), totalHeight: modalEditorTotalHeight}
	if modalEditor.visible {
		modalEditor.totalHeight = program.overlayState.modalEditor.Height()
	}

	buildPopup := pullRequestBuildRunPopupLayoutState{}
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		buildPopup = pullRequestBuildRunPopupLayoutState{
			visible:       true,
			body:          popup.body,
			widthPercent:  popup.widthPercent,
			heightPercent: popup.heightPercent,
		}
	}

	paneLayoutSize := PaneLayoutDefault
	fullscreenPane := FocusPullRequestsView
	if program.model != nil {
		paneLayoutSize = program.model.PaneLayoutSize()
		fullscreenPane = program.model.FullscreenPane()
	}

	actionsPopupVisible := program.model != nil && program.model.ActionsPopupVisible()
	actionsPopupSearchVisible := program.model != nil && program.model.ActionsPopupSearchActive()

	return screenLayoutInput{
		screenState:               program.screenState(),
		contentMaxY:               program.layoutContentHeight(maxY),
		paneLayoutSize:            paneLayoutSize,
		fullscreenPane:            fullscreenPane,
		sidebarTopPaneHeight:      program.sidebarTopPaneHeight(),
		footer:                    program.footerPresenter(),
		help:                      program.helpPresenter(),
		helpVisible:               program.overlayState.helpVisible,
		searchPromptVisible:       program.searchPromptVisible(),
		modalEditor:               modalEditor,
		buildPopup:                buildPopup,
		transientErrorPopup:       transientErrorPopupLayoutState{message: program.overlayState.transientErrorPopup.message},
		actionsPopupVisible:       actionsPopupVisible,
		actionsPopupSearchVisible: actionsPopupSearchVisible,
		actionsPopup:              program.actionsPopupPresenter(),
	}
}

func (program *Program) screenRenderers(input screenLayoutInput) map[string]ViewRenderer {
	if program == nil {
		return nil
	}

	renderers := map[string]ViewRenderer{
		viewDetailName:               viewRendererFuncs{configure: program.configureDetailView, render: program.renderDetailView},
		viewUserName:                 viewRendererFuncs{configure: program.configureUserView, render: program.renderUserView},
		viewPullRequestsName:         viewRendererFuncs{configure: program.configurePullRequestsView, render: program.renderPullRequestsView},
		viewNotificationsName:        viewRendererFuncs{configure: program.configureNotificationsView, render: program.renderNotificationsView},
		viewHelpName:                 viewRendererFuncs{configure: program.configureHelpView, render: program.renderHelpView},
		viewSearchName:               viewRendererFuncs{configure: program.configureSearchView, render: program.renderSearchView},
		viewModalEditorName:          viewRendererFuncs{configure: program.configureModalEditorView, render: program.renderModalEditorView},
		viewPullRequestBuildInfoName: viewRendererFuncs{configure: program.configurePullRequestBuildRunPopupView, render: program.renderPullRequestBuildRunPopupView},
		viewTransientErrorPopupName:  viewRendererFuncs{configure: program.configureTransientErrorPopupView, render: program.renderTransientErrorPopupView},
		viewActionsPopupChromeName:   viewRendererFuncs{configure: program.configureActionsPopupChromeView, render: program.renderActionsPopupChromeView},
		viewActionsPopupName:         viewRendererFuncs{configure: program.configureActionsPopupView, render: program.renderActionsPopupView},
		viewActionsPopupSearchName:   viewRendererFuncs{configure: program.configureActionsPopupSearchView, render: program.renderActionsPopupSearchView},
		viewStatusLineName:           viewRendererFuncs{configure: program.configureStatusLineView, render: program.renderStatusLineView},
	}

	for _, footerName := range []string{viewUserFooterName, viewPullRequestsFooterName, viewNotificationsFooterName, viewDetailFooterName} {
		footerText := strings.TrimSpace(input.footer.paneFooterStateFor(focusForFooterName(footerName)).Text())
		renderers[footerName] = viewRendererFuncs{
			configure: program.configurePaneFooterView,
			render: func(text string) viewRenderer {
				return func(view *gocui.View) {
					program.renderPaneFooterView(view, text)
				}
			}(footerText),
		}
	}

	if keyHintsText := strings.TrimSpace(input.footer.statusLineKeyHintsText()); keyHintsText != "" {
		renderers[viewStatusLineKeyHintsName] = viewRendererFuncs{
			configure: program.configureStatusLineKeyHintsView,
			render: func(view *gocui.View) {
				program.renderStatusLineKeyHintsView(view, keyHintsText)
			},
		}
	}

	return renderers
}
