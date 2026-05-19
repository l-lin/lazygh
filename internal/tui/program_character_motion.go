package tui

import "github.com/jesseduffield/gocui"

func (program *Program) registeredKeybindingSpecs() []keybindingSpec {
	specs := append([]keybindingSpec(nil), program.keybindingSpecs()...)
	for index, spec := range specs {
		viewName := spec.viewName
		keyRune, ok := spec.key.(rune)
		if !ok || !detailCharacterMotionEnabledView(viewName) {
			continue
		}

		originalHandler := spec.handler
		specs[index].handler = func(gui *gocui.Gui, view *gocui.View) error {
			handled, actualErr := program.consumeRegisteredCharacterMotion(gui, viewName, view, keyRune)
			if handled {
				return actualErr
			}
			return originalHandler(gui, view)
		}
	}
	return specs
}

func detailCharacterMotionEnabledView(viewName string) bool {
	return viewName == viewDetailName || viewName == viewPullRequestBuildInfoName
}

func (program *Program) consumeRegisteredCharacterMotion(gui *gocui.Gui, viewName string, view *gocui.View, target rune) (bool, error) {
	switch viewName {
	case viewDetailName:
		if !program.detailViewState.hasPendingCharacterMotion() {
			return false, nil
		}
		return true, program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.consumePendingCharacterMotion(document, viewportHeight, target)
		})
	case viewPullRequestBuildInfoName:
		if program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
			return false, nil
		}
		return true, program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
			state.consumePendingCharacterMotion(document, viewportHeight, target)
		})
	default:
		return false, nil
	}
}

func (program *Program) editDetailView(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if view == nil || mod != gocui.ModNone || ch == 0 || !program.detailViewState.hasPendingCharacterMotion() {
		return false
	}

	document := program.currentDetailDocument(view)
	program.syncDetailViewState(document, viewPageSize(view))
	if !program.detailViewState.consumePendingCharacterMotion(document, viewPageSize(view), ch) {
		return false
	}
	program.renderDetailView(view)
	return true
}

func (program *Program) editPullRequestBuildRunPopup(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if view == nil || mod != gocui.ModNone || ch == 0 || program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		return false
	}

	document := program.currentPullRequestBuildRunPopupDocument(view)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(view))
	if !program.pullRequestBuildRunPopup.viewState.consumePendingCharacterMotion(document, viewPageSize(view), ch) {
		return false
	}
	program.renderPullRequestBuildRunPopupView(view)
	return true
}

func (program *Program) startDetailCharacterFindForward(gui *gocui.Gui, view *gocui.View) error {
	return program.armDetailCharacterMotion(gui, view, detailCharacterMotionDirectionForward, detailCharacterMotionMatch)
}

func (program *Program) startDetailCharacterFindBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.armDetailCharacterMotion(gui, view, detailCharacterMotionDirectionBackward, detailCharacterMotionMatch)
}

func (program *Program) startDetailCharacterTillForward(gui *gocui.Gui, view *gocui.View) error {
	return program.armDetailCharacterMotion(gui, view, detailCharacterMotionDirectionForward, detailCharacterMotionBeforeMatch)
}

func (program *Program) startDetailCharacterTillBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.armDetailCharacterMotion(gui, view, detailCharacterMotionDirectionBackward, detailCharacterMotionAfterMatch)
}

func (program *Program) armDetailCharacterMotion(gui *gocui.Gui, view *gocui.View, direction detailCharacterMotionDirection, mode detailCharacterMotionMode) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.armCharacterMotion(direction, mode)
	})
}

func (program *Program) startPullRequestBuildRunPopupCharacterFindForward(gui *gocui.Gui, view *gocui.View) error {
	return program.armPullRequestBuildRunPopupCharacterMotion(gui, view, detailCharacterMotionDirectionForward, detailCharacterMotionMatch)
}

func (program *Program) startPullRequestBuildRunPopupCharacterFindBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.armPullRequestBuildRunPopupCharacterMotion(gui, view, detailCharacterMotionDirectionBackward, detailCharacterMotionMatch)
}

func (program *Program) startPullRequestBuildRunPopupCharacterTillForward(gui *gocui.Gui, view *gocui.View) error {
	return program.armPullRequestBuildRunPopupCharacterMotion(gui, view, detailCharacterMotionDirectionForward, detailCharacterMotionBeforeMatch)
}

func (program *Program) startPullRequestBuildRunPopupCharacterTillBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.armPullRequestBuildRunPopupCharacterMotion(gui, view, detailCharacterMotionDirectionBackward, detailCharacterMotionAfterMatch)
}

func (program *Program) armPullRequestBuildRunPopupCharacterMotion(gui *gocui.Gui, view *gocui.View, direction detailCharacterMotionDirection, mode detailCharacterMotionMode) error {
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.armCharacterMotion(direction, mode)
	})
}
