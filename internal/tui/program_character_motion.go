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
	specs = append(specs, program.activeDetailCharacterMotionRepeatBindingSpecs()...)
	specs = append(specs, program.activePullRequestBuildRunPopupCharacterMotionRepeatBindingSpecs()...)
	specs = append(specs, program.keybindingSpecs()...)
	return specs
}

func (program *Program) activeDetailCharacterMotionRepeatBindingSpecs() []keybindingSpec {
	if !program.detailState.viewState.hasLastCharacterMotion {
		return nil
	}

	return program.characterMotionRepeatBindingSpecs(
		viewDetailName,
		keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_forward"},
		keybindingActionID{scope: keymapScopeCursor, action: "repeat_character_motion_backward"},
		program.repeatDetailCharacterMotionForward,
		program.repeatDetailCharacterMotionBackward,
	)
}

func (program *Program) activePullRequestBuildRunPopupCharacterMotionRepeatBindingSpecs() []keybindingSpec {
	if program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasLastCharacterMotion {
		return nil
	}

	return program.characterMotionRepeatBindingSpecs(
		viewPullRequestBuildInfoName,
		keybindingActionID{scope: keymapScopePullRequestBuildInfo, action: "repeat_character_motion_forward"},
		keybindingActionID{scope: keymapScopePullRequestBuildInfo, action: "repeat_character_motion_backward"},
		program.repeatPullRequestBuildRunPopupCharacterMotionForward,
		program.repeatPullRequestBuildRunPopupCharacterMotionBackward,
	)
}

func (program *Program) characterMotionRepeatBindingSpecs(viewName string, forwardActionID keybindingActionID, backwardActionID keybindingActionID, forwardHandler func(*gocui.Gui, *gocui.View) error, backwardHandler func(*gocui.Gui, *gocui.View) error) []keybindingSpec {
	specs := make([]keybindingSpec, 0, 2)
	for _, binding := range program.resolvedBindingsForActionID(forwardActionID) {
		if len(binding.keys) != 1 {
			continue
		}
		specs = append(specs, keybindingSpec{viewName: viewName, key: binding.keys[0].value, mod: binding.keys[0].mod, handler: forwardHandler})
	}
	for _, binding := range program.resolvedBindingsForActionID(backwardActionID) {
		if len(binding.keys) != 1 {
			continue
		}
		specs = append(specs, keybindingSpec{viewName: viewName, key: binding.keys[0].value, mod: binding.keys[0].mod, handler: backwardHandler})
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
		program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationConsumePendingCharacter, View: view, SelectionKind: detailYankMotionCharacterInclusive, Rune: target}})
		return nil
	}
}

func (program *Program) pullRequestBuildRunPopupCharacterMotionTargetHandler(target rune) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationConsumePendingCharacter, View: view, SelectionKind: detailYankMotionCharacterInclusive, Rune: target}})
		return nil
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
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationArmCharacter, View: view, Direction: direction, Mode: mode}})
	return nil
}

func (program *Program) repeatDetailCharacterMotion(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	if !program.detailState.viewState.hasLastCharacterMotion {
		return nil
	}
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationRepeatCharacter, View: view, Reverse: reverse, SelectionKind: detailYankMotionCharacterInclusive}})
	return nil
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
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationArmCharacter, View: view, Direction: direction, Mode: mode}})
	return nil
}

func (program *Program) repeatPullRequestBuildRunPopupCharacterMotion(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	if program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasLastCharacterMotion {
		return nil
	}
	program.executeCmds(gui, []Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationRepeatCharacter, View: view, Reverse: reverse, SelectionKind: detailYankMotionCharacterInclusive}})
	return nil
}
