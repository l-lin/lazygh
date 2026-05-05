package tui

const (
	detailYankSuccessMessage = "󰆏 Selection copied"
	detailYankFailureMessage = "󰅚 Copy failed"
)

type detailMode int

const (
	detailNormalMode detailMode = iota
	detailVisualMode
	detailLineVisualMode
)

func (mode detailMode) isVisual() bool {
	return mode == detailVisualMode || mode == detailLineVisualMode
}

type detailPosition struct {
	line   int
	column int
}

type detailColumnRange struct {
	start int
	end   int
}

type detailWrappedRow struct {
	line        int
	startColumn int
	endColumn   int
	empty       bool
	text        string
}

type detailDocument struct {
	text              []rune
	lines             [][]rune
	lineStylePrefixes [][]string
	width             int
	wrap              bool
	lineStartOffsets  []int
	lineStartRows     []int
	rows              []detailWrappedRow
}

type detailViewState struct {
	cursor               detailPosition
	originRow            int
	preferredColumn      int
	mode                 detailMode
	visualAnchor         detailPosition
	pendingKeySequence   keySequenceState
	searchMatches        []detailSearchMatch
	currentSearchMatch   int
	manualViewportScroll bool
}

type detailCellStyle struct {
	selected bool
	search   bool
}
