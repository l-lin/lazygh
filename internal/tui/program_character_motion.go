package tui

import (
	"sort"
	"unicode"

	"github.com/jesseduffield/gocui"
)

const (
	detailCharacterMotionPrintableASCIIFirst = ' '
	detailCharacterMotionPrintableASCIILast  = '~'
)

func (program *Program) registeredKeybindingSpecs() []keybindingSpec {
	specs := program.activeDetailCharacterMotionTargetBindingSpecs()
	specs = append(specs, program.activePullRequestBuildRunPopupCharacterMotionTargetBindingSpecs()...)
	specs = append(specs, program.keybindingSpecs()...)
	return specs
}

func (program *Program) activeDetailCharacterMotionTargetBindingSpecs() []keybindingSpec {
	if !program.detailViewState.hasPendingCharacterMotion() {
		return nil
	}

	actualView := program.resolveView(program.gui, nil, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	bindings := characterMotionTargetRunes(detailDocumentLineAt(document, program.detailViewState.cursor.line))
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewDetailName, key: target, handler: program.detailCharacterMotionTargetHandler(target)})
	}
	return specs
}

func (program *Program) activePullRequestBuildRunPopupCharacterMotionTargetBindingSpecs() []keybindingSpec {
	popup := program.pullRequestBuildRunPopup
	if popup == nil || !popup.viewState.hasPendingCharacterMotion() {
		return nil
	}

	actualView := program.resolveView(program.gui, nil, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(actualView))
	bindings := characterMotionTargetRunes(detailDocumentLineAt(document, popup.viewState.cursor.line))
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: target, handler: program.pullRequestBuildRunPopupCharacterMotionTargetHandler(target)})
	}
	return specs
}

func detailDocumentLineAt(document detailDocument, lineIndex int) []rune {
	if lineIndex < 0 || lineIndex >= len(document.lines) {
		return nil
	}
	return document.lines[lineIndex]
}

func characterMotionTargetRunes(line []rune) []rune {
	seen := map[rune]struct{}{}
	runes := make([]rune, 0, int(detailCharacterMotionPrintableASCIILast-detailCharacterMotionPrintableASCIIFirst)+1+len(line))
	for target := detailCharacterMotionPrintableASCIIFirst; target <= detailCharacterMotionPrintableASCIILast; target++ {
		seen[target] = struct{}{}
		runes = append(runes, target)
	}
	for _, target := range line {
		if unicode.IsControl(target) {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		runes = append(runes, target)
	}
	sort.Slice(runes[int(detailCharacterMotionPrintableASCIILast-detailCharacterMotionPrintableASCIIFirst)+1:], func(left int, right int) bool {
		asciiCount := int(detailCharacterMotionPrintableASCIILast-detailCharacterMotionPrintableASCIIFirst) + 1
		return runes[asciiCount+left] < runes[asciiCount+right]
	})
	return runes
}

func (program *Program) detailCharacterMotionTargetHandler(target rune) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		if actualErr := program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.consumePendingCharacterMotion(document, viewportHeight, target)
		}); actualErr != nil {
			return actualErr
		}
		return program.reloadRegisteredKeybindings(gui)
	}
}

func (program *Program) pullRequestBuildRunPopupCharacterMotionTargetHandler(target rune) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		if actualErr := program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
			state.consumePendingCharacterMotion(document, viewportHeight, target)
		}); actualErr != nil {
			return actualErr
		}
		return program.reloadRegisteredKeybindings(gui)
	}
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

func (program *Program) repeatDetailCharacterMotionForward(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatDetailCharacterMotion(gui, view, false)
}

func (program *Program) repeatDetailCharacterMotionBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatDetailCharacterMotion(gui, view, true)
}

func (program *Program) armDetailCharacterMotion(gui *gocui.Gui, view *gocui.View, direction detailCharacterMotionDirection, mode detailCharacterMotionMode) error {
	if actualErr := program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.armCharacterMotion(direction, mode)
	}); actualErr != nil {
		return actualErr
	}
	return program.reloadRegisteredKeybindings(gui)
}

func (program *Program) repeatDetailCharacterMotion(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	if !program.detailViewState.hasLastCharacterMotion {
		return nil
	}
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.repeatCharacterMotion(document, viewportHeight, reverse)
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

func (program *Program) repeatPullRequestBuildRunPopupCharacterMotionForward(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatPullRequestBuildRunPopupCharacterMotion(gui, view, false)
}

func (program *Program) repeatPullRequestBuildRunPopupCharacterMotionBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.repeatPullRequestBuildRunPopupCharacterMotion(gui, view, true)
}

func (program *Program) armPullRequestBuildRunPopupCharacterMotion(gui *gocui.Gui, view *gocui.View, direction detailCharacterMotionDirection, mode detailCharacterMotionMode) error {
	if actualErr := program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.armCharacterMotion(direction, mode)
	}); actualErr != nil {
		return actualErr
	}
	return program.reloadRegisteredKeybindings(gui)
}

func (program *Program) repeatPullRequestBuildRunPopupCharacterMotion(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	if program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasLastCharacterMotion {
		return nil
	}
	return program.mutatePullRequestBuildRunPopupViewState(gui, view, func(state *detailViewState, document detailDocument, viewportHeight int) {
		state.repeatCharacterMotion(document, viewportHeight, reverse)
	})
}
