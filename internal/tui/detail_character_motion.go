package tui

type detailCharacterMotionDirection int

const (
	detailCharacterMotionDirectionForward detailCharacterMotionDirection = iota
	detailCharacterMotionDirectionBackward
)

type detailCharacterMotionMode int

const (
	detailCharacterMotionMatch detailCharacterMotionMode = iota
	detailCharacterMotionBeforeMatch
	detailCharacterMotionAfterMatch
)

type detailCharacterMotion struct {
	target    rune
	direction detailCharacterMotionDirection
	mode      detailCharacterMotionMode
}

type detailPendingCharacterMotion struct {
	active    bool
	direction detailCharacterMotionDirection
	mode      detailCharacterMotionMode
}

func (pending detailPendingCharacterMotion) motion(target rune) detailCharacterMotion {
	return detailCharacterMotion{target: target, direction: pending.direction, mode: pending.mode}
}

func (motion detailCharacterMotionDirection) opposite() detailCharacterMotionDirection {
	switch motion {
	case detailCharacterMotionDirectionBackward:
		return detailCharacterMotionDirectionForward
	default:
		return detailCharacterMotionDirectionBackward
	}
}

func (state *detailViewState) hasPendingCharacterMotion() bool {
	return state.pendingCharacterMotion.active
}

func (state *detailViewState) armCharacterMotion(direction detailCharacterMotionDirection, mode detailCharacterMotionMode) {
	state.clearPendingPrefix()
	state.pendingCharacterMotion = detailPendingCharacterMotion{active: true, direction: direction, mode: mode}
}

func (state *detailViewState) consumePendingCharacterMotion(document detailDocument, viewportHeight int, target rune) bool {
	if !state.pendingCharacterMotion.active {
		return false
	}

	motion := state.pendingCharacterMotion.motion(target)
	state.pendingCharacterMotion = detailPendingCharacterMotion{}
	state.applyCharacterMotion(document, viewportHeight, motion)
	return true
}

func (state *detailViewState) applyCharacterMotion(document detailDocument, viewportHeight int, motion detailCharacterMotion) bool {
	state.pendingKeySequence.clear()
	state.pendingCharacterMotion = detailPendingCharacterMotion{}
	state.lastCharacterMotion = motion
	state.hasLastCharacterMotion = motion.target != 0

	next, ok := document.moveToCharacter(state.cursor, motion)
	if ok {
		state.cursor = next
		state.preferredColumn = document.screenColumnForPosition(state.cursor)
	}
	state.sync(document, viewportHeight)
	return ok
}

func (document detailDocument) moveToCharacter(position detailPosition, motion detailCharacterMotion) (detailPosition, bool) {
	if motion.target == 0 {
		return detailPosition{}, false
	}

	position = document.clampPosition(position)
	if position.line < 0 || position.line >= len(document.lines) {
		return detailPosition{}, false
	}

	line := document.lines[position.line]
	if len(line) == 0 {
		return detailPosition{}, false
	}

	matchColumn, ok := detailCharacterMotionMatchColumn(line, position.column, motion.target, motion.direction)
	if !ok {
		return detailPosition{}, false
	}

	column := matchColumn
	switch motion.mode {
	case detailCharacterMotionBeforeMatch:
		column = max(column-1, 0)
	case detailCharacterMotionAfterMatch:
		column = min(column+1, len(line)-1)
	}
	return detailPosition{line: position.line, column: column}, true
}

func detailCharacterMotionMatchColumn(line []rune, column int, target rune, direction detailCharacterMotionDirection) (int, bool) {
	if len(line) == 0 {
		return 0, false
	}

	column = clampInt(column, 0, len(line)-1)
	switch direction {
	case detailCharacterMotionDirectionBackward:
		for candidate := column - 1; candidate >= 0; candidate-- {
			if line[candidate] == target {
				return candidate, true
			}
		}
		return 0, false
	default:
		for candidate := column + 1; candidate < len(line); candidate++ {
			if line[candidate] == target {
				return candidate, true
			}
		}
		return 0, false
	}
}
