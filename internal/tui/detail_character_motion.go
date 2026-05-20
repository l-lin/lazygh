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

func (motion detailCharacterMotion) reversed() detailCharacterMotion {
	motion.direction = motion.direction.opposite()
	switch motion.mode {
	case detailCharacterMotionBeforeMatch:
		motion.mode = detailCharacterMotionAfterMatch
	case detailCharacterMotionAfterMatch:
		motion.mode = detailCharacterMotionBeforeMatch
	}
	return motion
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
	state.pendingKeySequence.clear()
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

func (state *detailViewState) repeatCharacterMotion(document detailDocument, viewportHeight int, reverse bool) bool {
	if !state.hasLastCharacterMotion {
		return false
	}

	motion := state.lastCharacterMotion
	if reverse {
		motion = motion.reversed()
	}

	state.pendingKeySequence.clear()
	state.pendingCharacterMotion = detailPendingCharacterMotion{}
	next, ok := document.moveToRepeatedCharacter(state.cursor, motion)
	if ok {
		state.cursor = next
		state.preferredColumn = document.screenColumnForPosition(state.cursor)
	}
	state.sync(document, viewportHeight)
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
	return document.moveToCharacterWithRepeat(position, motion, false)
}

func (document detailDocument) moveToRepeatedCharacter(position detailPosition, motion detailCharacterMotion) (detailPosition, bool) {
	return document.moveToCharacterWithRepeat(position, motion, true)
}

func (document detailDocument) moveToCharacterWithRepeat(position detailPosition, motion detailCharacterMotion, repeating bool) (detailPosition, bool) {
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

	searchColumn := position.column
	for {
		matchColumn, ok := detailCharacterMotionMatchColumn(line, searchColumn, motion.target, motion.direction)
		if !ok {
			return detailPosition{}, false
		}

		candidate := detailCharacterMotionPositionForMatch(position.line, len(line), matchColumn, motion.mode)
		if !repeating || candidate != position {
			return candidate, true
		}
		searchColumn = matchColumn
	}
}

func detailCharacterMotionPositionForMatch(lineIndex int, lineLength int, matchColumn int, mode detailCharacterMotionMode) detailPosition {
	column := matchColumn
	switch mode {
	case detailCharacterMotionBeforeMatch:
		column = max(column-1, 0)
	case detailCharacterMotionAfterMatch:
		column = min(column+1, lineLength-1)
	}
	return detailPosition{line: lineIndex, column: column}
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
