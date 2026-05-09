package tui

import (
	"strings"
	"testing"
)

func TestGlamourMarkdownRenderer_GivenALoadedImageAndKittyProtocol_WhenRendering_ThenItTracksAnImagePlacementWithoutFallbackText(t *testing.T) {
	renderer := glamourMarkdownRenderer{
		imageStore: &fakeDetailImageStore{
			imagesBySource: map[string]storedDetailImage{"https://example.com/diagram.png": {imageID: 42, pixelWidth: 40, pixelHeight: 20, pngData: []byte("png")}},
		},
		imageProtocol:    kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}},
		terminalCellSize: fixedTerminalCellSize{width: 10, height: 10},
	}

	actual, actualErr := renderer.Render("![Architecture](https://example.com/diagram.png)", 12)

	then_noError(t, actualErr)
	actualDocument := newDetailDocumentWithWrap(actual, 12, false)
	if len(actualDocument.images) != 1 {
		t.Fatalf("expected one tracked image placement, actual %d", len(actualDocument.images))
	}
	actualImage := actualDocument.images[0]
	if actualImage.imageID != 42 {
		t.Fatalf("expected image id %d, actual %d", 42, actualImage.imageID)
	}
	if actualImage.columns != 4 || actualImage.rows != 2 {
		t.Fatalf("expected the image placement to use 4 columns and 2 rows, actual %d columns and %d rows", actualImage.columns, actualImage.rows)
	}
	if actualDocument.rowCount() != 2 {
		t.Fatalf("expected the rendered markdown to keep %d image rows, actual %d", 2, actualDocument.rowCount())
	}
	if strings.Contains(actual, "Architecture") {
		t.Fatalf("expected the rendered markdown to drop the fallback alt text once the image is displayed, actual %q", actual)
	}
	if strings.Contains(actual, "https://example.com/diagram.png") {
		t.Fatalf("expected the rendered markdown to drop the fallback URL once the image is displayed, actual %q", actual)
	}
}

func TestGlamourMarkdownRenderer_GivenAnUnloadedImage_WhenRendering_ThenItFallsBackToAnInlineLabelAndAnUnderlinedLink(t *testing.T) {
	renderer := glamourMarkdownRenderer{
		imageStore:       &fakeDetailImageStore{},
		imageProtocol:    kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}},
		terminalCellSize: fixedTerminalCellSize{width: 10, height: 10},
	}

	actual, actualErr := renderer.Render("![Architecture](https://example.com/diagram.png)", 80)

	then_noError(t, actualErr)
	actualDocument := newDetailDocumentWithWrap(actual, 80, false)
	if len(actualDocument.images) != 0 {
		t.Fatalf("expected no tracked image placements without a loaded image, actual %d", len(actualDocument.images))
	}
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, "https://example.com/diagram.png")
	expected := iconMarkdownImage + " Architecture https://example.com/diagram.png"
	if visibleLine != expected {
		t.Fatalf("expected the fallback line %q, actual %q", expected, visibleLine)
	}
	urlIndex := given_runeIndexInString(t, visibleLine, "https://example.com/diagram.png")
	if !strings.Contains(actualDocument.lineStylePrefixes[lineIndex][urlIndex], underlineEscape) {
		t.Fatalf("expected the fallback URL to be underlined, actual prefix %q", actualDocument.lineStylePrefixes[lineIndex][urlIndex])
	}
}

func TestRenderCommentBoxWithMetadata_GivenAKittyImageBody_WhenFormatting_ThenItPreservesTheImagePlacementInsideTheBoxPadding(t *testing.T) {
	renderer := glamourMarkdownRenderer{
		imageStore: &fakeDetailImageStore{
			imagesBySource: map[string]storedDetailImage{"https://example.com/diagram.png": {imageID: 42, pixelWidth: 40, pixelHeight: 20, pngData: []byte("png")}},
		},
		imageProtocol:    kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}},
		terminalCellSize: fixedTerminalCellSize{width: 10, height: 10},
	}
	body, actualErr := renderer.Render("![Architecture](https://example.com/diagram.png)", 12)
	then_noError(t, actualErr)

	actual := renderCommentBoxWithMetadata(nil, "", nil, body, 18)
	actualDocument := newDetailDocumentWithWrap(actual, 18, false)

	if len(actualDocument.images) != 1 {
		t.Fatalf("expected one tracked image placement inside the comment box, actual %d", len(actualDocument.images))
	}
	actualImage := actualDocument.images[0]
	if actualImage.line != 2 {
		t.Fatalf("expected the image to start on the first boxed body line, actual line %d", actualImage.line)
	}
	if actualImage.column != 2 {
		t.Fatalf("expected the image to start after the border and padding, actual column %d", actualImage.column)
	}
}

func TestRenderDetailRow_GivenAnEmptyRowWithAKittyImagePlacement_WhenRendering_ThenItEmitsUnicodePlaceholders(t *testing.T) {
	document := detailDocument{
		images: []detailImagePlacement{{line: 0, column: 0, imageID: 42, columns: 2, rows: 2}},
		rows:   []detailWrappedRow{{line: 0, startColumn: 0, endColumn: 0, empty: true}},
	}

	actual := renderDetailRow(document, document.rows[0], nil, newDetailViewState())

	if !strings.Contains(actual, string(kittyImagePlaceholderRune)) {
		t.Fatalf("expected kitty unicode placeholders in the rendered row, actual %q", actual)
	}
}

func TestProtocolDetailImageManager_GivenCurrentImages_WhenSyncing_ThenItTransmitsNewImagesAndDeletesStaleOnes(t *testing.T) {
	store := &fakeDetailImageStore{imagesByID: map[uint32]storedDetailImage{
		42: {imageID: 42, pixelWidth: 40, pixelHeight: 20, pngData: []byte("first")},
		77: {imageID: 77, pixelWidth: 20, pixelHeight: 20, pngData: []byte("second")},
	}}
	terminal := &fakeGraphicsTerminal{}
	subject := protocolDetailImageManager{
		imageStore:    store,
		imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: true}},
		terminal:      terminal,
	}

	subject.Sync([]detailImagePlacement{{imageID: 42, columns: 4, rows: 2}, {imageID: 77, columns: 2, rows: 2}})
	subject.Sync([]detailImagePlacement{{imageID: 77, columns: 2, rows: 2}})

	actual := strings.Join(terminal.commands, "\n")
	for _, expected := range []string{"a=t", "i=42", "i=77", "U=1", "d=I"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected image sync commands to contain %q, actual %q", expected, actual)
		}
	}
}

func TestWrapGraphicsCommandForTmux_GivenAGraphicsCommand_WhenWrapping_ThenItUsesTmuxPassthrough(t *testing.T) {
	actual := wrapGraphicsCommandForTmux("\x1b_Ga=t,q=2,f=100,i=42,m=0;Zm9v\x1b\\")

	if !strings.HasPrefix(actual, "\x1bPtmux;") {
		t.Fatalf("expected tmux passthrough prefix, actual %q", actual)
	}
	if !strings.HasSuffix(actual, "\x1b\\") {
		t.Fatalf("expected tmux passthrough suffix, actual %q", actual)
	}
	if strings.Count(actual, "\x1b") < 4 {
		t.Fatalf("expected tmux passthrough to escape embedded ESC bytes, actual %q", actual)
	}
}

type fakeDetailImageStore struct {
	imagesBySource map[string]storedDetailImage
	imagesByID     map[uint32]storedDetailImage
}

func (store *fakeDetailImageStore) Store(source string, image loadedDetailImage) storedDetailImage {
	if store.imagesBySource == nil {
		store.imagesBySource = map[string]storedDetailImage{}
	}
	if store.imagesByID == nil {
		store.imagesByID = map[uint32]storedDetailImage{}
	}
	storedImage := storedDetailImage{imageID: image.imageID, pixelWidth: image.pixelWidth, pixelHeight: image.pixelHeight, pngData: image.pngData}
	store.imagesBySource[source] = storedImage
	store.imagesByID[storedImage.imageID] = storedImage
	return storedImage
}

func (store *fakeDetailImageStore) ImageBySource(source string) (storedDetailImage, bool) {
	image, ok := store.imagesBySource[source]
	return image, ok
}

func (store *fakeDetailImageStore) ImageByID(imageID uint32) (storedDetailImage, bool) {
	image, ok := store.imagesByID[imageID]
	return image, ok
}

type fakeImageProtocolSupport struct {
	supported bool
}

func (support fakeImageProtocolSupport) Supported() bool {
	return support.supported
}

type fixedTerminalCellSize struct {
	width  int
	height int
}

func (size fixedTerminalCellSize) CellSize() (int, int, bool) {
	return size.width, size.height, true
}

type fakeGraphicsTerminal struct {
	commands []string
}

func (terminal *fakeGraphicsTerminal) WriteGraphicsCommand(command string) error {
	terminal.commands = append(terminal.commands, command)
	return nil
}
