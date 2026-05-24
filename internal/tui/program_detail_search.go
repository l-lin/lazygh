package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) nextDetailSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.searchWidget.detailReversed {
		return program.repeatDetailSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
			return program.detailViewState.followPreviousSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
		})
	}
	return program.repeatDetailSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
		return program.detailViewState.followNextSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
	})
}

func (program *Program) previousDetailSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	if program.searchWidget.detailReversed {
		return program.repeatDetailSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
			return program.detailViewState.followNextSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
		})
	}
	return program.repeatDetailSearch(gui, view, func(document detailDocument, viewportHeight int) bool {
		return program.detailViewState.followPreviousSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
	})
}

func (program *Program) repeatDetailSearch(gui *gocui.Gui, view *gocui.View, repeat func(detailDocument, int) bool) error {
	if program.model.Focus() != FocusDetailView || program.detailViewState.mode != detailNormalMode {
		return nil
	}
	if strings.TrimSpace(program.model.DetailSearchQuery()) == "" {
		return nil
	}

	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		repeat(document, viewportHeight)
	})
}

func (program *Program) followSubmittedDetailSearch(gui *gocui.Gui) error {
	if strings.TrimSpace(program.model.DetailSearchQuery()) == "" {
		return nil
	}

	return program.mutateDetailViewStateWithoutRefresh(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailViewState.followSubmittedSearch(document, program.model.DetailSearchQuery(), viewportHeight)
	})
}

func (program *Program) followReverseDetailSearch(gui *gocui.Gui) error {
	if strings.TrimSpace(program.model.DetailSearchQuery()) == "" {
		return nil
	}

	return program.mutateDetailViewStateWithoutRefresh(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailViewState.followPreviousSearchMatch(document, program.model.DetailSearchQuery(), viewportHeight)
	})
}
