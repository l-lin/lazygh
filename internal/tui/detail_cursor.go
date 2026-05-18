package tui

const (
	detailYankSuccessMessage = iconCopy + " Selection copied"
	detailYankFailureMessage = iconWarning + " Copy failed"
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
	id                   uint64
	text                 []rune
	lines                [][]rune
	lineStylePrefixes    [][]string
	lineHyperlinkTargets [][]string
	images               []detailImagePlacement
	width                int
	wrap                 bool
	lineStartOffsets     []int
	lineStartRows        []int
	rows                 []detailWrappedRow
}

type detailViewState struct {
	cursor                    detailPosition
	originRow                 int
	preferredColumn           int
	mode                      detailMode
	visualAnchor              detailPosition
	pendingKeySequence        keySequenceState
	searchMatches             []detailSearchMatch
	currentSearchMatch        int
	searchCacheDocumentID     uint64
	searchCacheQuery          string
	manualViewportScroll      bool
	preserveViewportSyncCount int
}

type detailCellStyle struct {
	selected bool
	search   bool
}
