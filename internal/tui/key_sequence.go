package tui

type keySequenceTarget struct {
	viewName string
	actionID keybindingActionID
}

type keySequenceState struct {
	pendingTarget keySequenceTarget
}

func keySequenceTargetFor(viewName string, scope string, action string) keySequenceTarget {
	return keySequenceTarget{viewName: viewName, actionID: keybindingActionID{scope: scope, action: action}}
}

func (state *keySequenceState) clear() {
	state.pendingTarget = keySequenceTarget{}
}

func (state *keySequenceState) arm(target keySequenceTarget) {
	if target == (keySequenceTarget{}) {
		state.clear()
		return
	}

	state.pendingTarget = target
}

func (state *keySequenceState) consume(target keySequenceTarget) bool {
	if state.pendingTarget != target {
		return false
	}

	state.clear()
	return true
}

func (state *keySequenceState) armOrConsume(target keySequenceTarget) bool {
	if state.consume(target) {
		return true
	}

	state.arm(target)
	return false
}
