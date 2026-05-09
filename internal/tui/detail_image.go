package tui

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	urlpkg "net/url"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/jesseduffield/gocui"
)

const (
	kittyImagePlaceholderRune = '\U0010EEEE'
	detailImageOSCCommand     = "lazygh-detail-image"
)

type detailImagePlacement struct {
	line    int
	column  int
	imageID uint32
	columns int
	rows    int
}

type detailImageSpec struct {
	imageID uint32
	columns int
	rows    int
}

type styledTextControl struct {
	column int
	image  *detailImageSpec
}

type loadedDetailImage struct {
	imageID     uint32
	pixelWidth  int
	pixelHeight int
	pngData     []byte
}

type storedDetailImage struct {
	imageID     uint32
	pixelWidth  int
	pixelHeight int
	pngData     []byte
}

type detailImageStore interface {
	Store(source string, image loadedDetailImage) storedDetailImage
	ImageBySource(source string) (storedDetailImage, bool)
	ImageByID(imageID uint32) (storedDetailImage, bool)
}

type detailImageProtocol interface {
	Supported() bool
	Bounds(width int, pixelWidth int, pixelHeight int, cellWidth int, cellHeight int) (int, int)
	PlaceholderCell(imageID uint32, row int, column int) string
	PlaceholderForegroundHex(imageID uint32) string
	TransmitCommand(image storedDetailImage) string
	PlacementCommand(image detailImagePlacement) string
	DeleteCommand(imageID uint32) string
}

type imageProtocolSupport interface {
	Supported() bool
}

type terminalCellSizeProvider interface {
	CellSize() (int, int, bool)
}

type terminalGraphicsTerminal interface {
	WriteGraphicsCommand(command string) error
}

type detailImageManager interface {
	Sync(images []detailImagePlacement)
}

type environmentKittyGraphicsSupport struct{}

type kittyImageProtocol struct {
	support imageProtocolSupport
}

type screenTerminalCellSize struct{}

type screenTerminalGraphicsTerminal struct{}

type memoryDetailImageStore struct {
	mu             sync.Mutex
	nextImageID    uint32
	imagesBySource map[string]storedDetailImage
	imagesByID     map[uint32]storedDetailImage
}

type protocolDetailImageManager struct {
	imageStore    detailImageStore
	imageProtocol detailImageProtocol
	terminal      terminalGraphicsTerminal
	active        map[uint32]detailImagePlacement
}

var defaultDetailImageStore detailImageStore = newMemoryDetailImageStore()

var kittyPlaceholderDiacritics = []rune{
	'\u0305', '\u030D', '\u030E', '\u0310', '\u0312', '\u033D', '\u033E', '\u033F', '\u0346', '\u034A', '\u034B', '\u034C', '\u0350', '\u0351', '\u0352', '\u0357', '\u035B', '\u0363', '\u0364', '\u0365', '\u0366', '\u0367', '\u0368', '\u0369', '\u036A', '\u036B', '\u036C', '\u036D', '\u036E', '\u036F', '\u0483', '\u0484', '\u0485', '\u0486', '\u0487', '\u0592', '\u0593', '\u0594', '\u0595', '\u0597', '\u0598', '\u0599', '\u059C', '\u059D', '\u059E', '\u059F', '\u05A0', '\u05A1', '\u05A8', '\u05A9', '\u05AB', '\u05AC', '\u05AF', '\u05C4', '\u0610', '\u0611', '\u0612', '\u0613', '\u0614', '\u0615', '\u0616', '\u0617', '\u0657', '\u0658', '\u0659', '\u065A', '\u065B', '\u065D', '\u065E', '\u06D6', '\u06D7', '\u06D8', '\u06D9', '\u06DA', '\u06DB', '\u06DC', '\u06DF', '\u06E0', '\u06E1', '\u06E2', '\u06E4', '\u06E7', '\u06E8', '\u06EB', '\u06EC', '\u0730', '\u0732', '\u0733', '\u0735', '\u0736', '\u073A', '\u073D', '\u073F', '\u0740', '\u0741', '\u0743', '\u0745', '\u0747', '\u0749', '\u074A', '\u07EB', '\u07EC', '\u07ED', '\u07EE', '\u07EF', '\u07F0', '\u07F1', '\u07F3', '\u0816', '\u0817', '\u0818', '\u0819', '\u081B', '\u081C', '\u081D', '\u081E', '\u081F', '\u0820', '\u0821', '\u0822', '\u0823', '\u0825', '\u0826', '\u0827', '\u0829', '\u082A', '\u082B', '\u082C',
}

func newMemoryDetailImageStore() *memoryDetailImageStore {
	return &memoryDetailImageStore{
		nextImageID:    1,
		imagesBySource: map[string]storedDetailImage{},
		imagesByID:     map[uint32]storedDetailImage{},
	}
}

func (store *memoryDetailImageStore) Store(source string, image loadedDetailImage) storedDetailImage {
	store.mu.Lock()
	defer store.mu.Unlock()

	trimmedSource := strings.TrimSpace(source)
	if existing, ok := store.imagesBySource[trimmedSource]; ok {
		return existing
	}

	imageID := image.imageID
	if imageID == 0 {
		if store.nextImageID == 0 || store.nextImageID > 0xFFFFFF {
			store.nextImageID = 1
		}
		imageID = store.nextImageID
		store.nextImageID++
	}
	storedImage := storedDetailImage{imageID: imageID, pixelWidth: image.pixelWidth, pixelHeight: image.pixelHeight, pngData: append([]byte(nil), image.pngData...)}
	store.imagesBySource[trimmedSource] = storedImage
	store.imagesByID[imageID] = storedImage
	return storedImage
}

func (store *memoryDetailImageStore) ImageBySource(source string) (storedDetailImage, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	storedImage, ok := store.imagesBySource[strings.TrimSpace(source)]
	return storedImage, ok
}

func (store *memoryDetailImageStore) ImageByID(imageID uint32) (storedDetailImage, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	storedImage, ok := store.imagesByID[imageID]
	return storedImage, ok
}

func (environmentKittyGraphicsSupport) Supported() bool {
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}

	for _, value := range []string{strings.ToLower(os.Getenv("TERM")), strings.ToLower(os.Getenv("TERM_PROGRAM"))} {
		if strings.Contains(value, "kitty") || strings.Contains(value, "ghostty") {
			return true
		}
	}

	return false
}

func (protocol kittyImageProtocol) actualSupport() imageProtocolSupport {
	if protocol.support != nil {
		return protocol.support
	}
	return environmentKittyGraphicsSupport{}
}

func (protocol kittyImageProtocol) Supported() bool {
	return protocol.actualSupport().Supported()
}

func (kittyImageProtocol) Bounds(width int, pixelWidth int, pixelHeight int, cellWidth int, cellHeight int) (int, int) {
	const (
		minimumColumns    = 1
		minimumRows       = 1
		defaultCellWidth  = 8
		defaultCellHeight = 16
	)

	if width < minimumColumns {
		width = minimumColumns
	}
	if cellWidth < 1 {
		cellWidth = defaultCellWidth
	}
	if cellHeight < 1 {
		cellHeight = defaultCellHeight
	}
	if pixelWidth < 1 || pixelHeight < 1 {
		return minimumColumns, minimumRows
	}

	maxColumns := minInt(width, len(kittyPlaceholderDiacritics))
	naturalColumns := (pixelWidth + cellWidth - 1) / cellWidth
	imageColumns := maxInt(minimumColumns, minInt(naturalColumns, maxColumns))
	displayPixelWidth := minInt(pixelWidth, imageColumns*cellWidth)
	displayPixelHeight := maxInt(1, (pixelHeight*displayPixelWidth+pixelWidth-1)/pixelWidth)
	maxRows := len(kittyPlaceholderDiacritics)
	imageRows := maxInt(minimumRows, minInt((displayPixelHeight+cellHeight-1)/cellHeight, maxRows))
	return imageColumns, imageRows
}

func (kittyImageProtocol) PlaceholderCell(_ uint32, row int, column int) string {
	if row < 0 || row >= len(kittyPlaceholderDiacritics) || column < 0 || column >= len(kittyPlaceholderDiacritics) {
		return " "
	}
	return string([]rune{kittyImagePlaceholderRune, kittyPlaceholderDiacritics[row], kittyPlaceholderDiacritics[column]})
}

func (kittyImageProtocol) PlaceholderForegroundHex(imageID uint32) string {
	return fmt.Sprintf("#%06X", imageID&0xFFFFFF)
}

func (kittyImageProtocol) TransmitCommand(image storedDetailImage) string {
	encodedPNG := base64.StdEncoding.EncodeToString(image.pngData)
	if encodedPNG == "" {
		return ""
	}

	var builder strings.Builder
	for startIndex := 0; startIndex < len(encodedPNG); startIndex += 4096 {
		endIndex := minInt(len(encodedPNG), startIndex+4096)
		moreChunks := 0
		if endIndex < len(encodedPNG) {
			moreChunks = 1
		}
		if startIndex == 0 {
			fmt.Fprintf(&builder, "\x1b_Ga=t,q=2,f=100,i=%d,m=%d;%s\x1b\\", image.imageID, moreChunks, encodedPNG[startIndex:endIndex])
			continue
		}
		fmt.Fprintf(&builder, "\x1b_Gm=%d;%s\x1b\\", moreChunks, encodedPNG[startIndex:endIndex])
	}
	return builder.String()
}

func (kittyImageProtocol) PlacementCommand(image detailImagePlacement) string {
	return fmt.Sprintf("\x1b_Ga=p,U=1,q=2,i=%d,c=%d,r=%d\x1b\\", image.imageID, image.columns, image.rows)
}

func (kittyImageProtocol) DeleteCommand(imageID uint32) string {
	return fmt.Sprintf("\x1b_Ga=d,d=I,q=2,i=%d\x1b\\", imageID)
}

func (screenTerminalCellSize) CellSize() (int, int, bool) {
	if gocui.Screen == nil {
		return 0, 0, false
	}

	tty, ok := gocui.Screen.Tty()
	if !ok {
		return 0, 0, false
	}
	windowSize, actualErr := tty.WindowSize()
	if actualErr != nil {
		return 0, 0, false
	}
	cellWidth, cellHeight := windowSize.CellDimensions()
	if cellWidth < 1 || cellHeight < 1 {
		return 0, 0, false
	}

	return cellWidth, cellHeight, true
}

func (screenTerminalGraphicsTerminal) WriteGraphicsCommand(command string) error {
	if gocui.Screen == nil || command == "" {
		return nil
	}

	tty, ok := gocui.Screen.Tty()
	if !ok {
		return nil
	}
	if os.Getenv("TMUX") != "" {
		command = wrapGraphicsCommandForTmux(command)
	}
	_, actualErr := io.WriteString(tty, command)
	return actualErr
}

func wrapGraphicsCommandForTmux(command string) string {
	if strings.TrimSpace(command) == "" {
		return command
	}
	return ansi.TmuxPassthrough(command)
}

func (manager *protocolDetailImageManager) actualImageProtocol() detailImageProtocol {
	if manager.imageProtocol != nil {
		return manager.imageProtocol
	}
	return kittyImageProtocol{}
}

func (manager *protocolDetailImageManager) Sync(images []detailImagePlacement) {
	protocol := manager.actualImageProtocol()
	if manager.imageStore == nil || manager.terminal == nil || !protocol.Supported() {
		return
	}
	if manager.active == nil {
		manager.active = map[uint32]detailImagePlacement{}
	}

	current := map[uint32]detailImagePlacement{}
	for _, image := range images {
		if image.imageID == 0 || image.columns < 1 || image.rows < 1 {
			continue
		}
		current[image.imageID] = image
	}

	for imageID := range manager.active {
		if _, ok := current[imageID]; ok {
			continue
		}
		_ = manager.terminal.WriteGraphicsCommand(protocol.DeleteCommand(imageID))
		delete(manager.active, imageID)
	}

	for imageID, image := range current {
		activeImage, ok := manager.active[imageID]
		if ok && activeImage.columns == image.columns && activeImage.rows == image.rows {
			continue
		}
		if ok {
			_ = manager.terminal.WriteGraphicsCommand(protocol.DeleteCommand(imageID))
		}
		storedImage, imageFound := manager.imageStore.ImageByID(imageID)
		if !imageFound {
			continue
		}
		_ = manager.terminal.WriteGraphicsCommand(protocol.TransmitCommand(storedImage))
		_ = manager.terminal.WriteGraphicsCommand(protocol.PlacementCommand(image))
		manager.active[imageID] = image
	}
}

func encodeDetailImageMarker(spec detailImageSpec) string {
	return fmt.Sprintf("\x1b]%s;i=%d;c=%d;r=%d\x1b\\", detailImageOSCCommand, spec.imageID, spec.columns, spec.rows)
}

func parseDetailImageMarkerSequence(text string, startIndex int) (detailImageSpec, int, bool) {
	sequence, nextIndex, ok := consumeOSCSequence(text, startIndex)
	if !ok {
		return detailImageSpec{}, startIndex, false
	}

	trimmedSequence := strings.TrimSuffix(sequence, "\a")
	trimmedSequence = strings.TrimSuffix(trimmedSequence, "\x1b\\")
	sequenceBody := strings.TrimPrefix(trimmedSequence, "\x1b]")
	if !strings.HasPrefix(sequenceBody, detailImageOSCCommand+";") {
		return detailImageSpec{}, startIndex, false
	}

	spec := detailImageSpec{}
	for _, part := range strings.Split(strings.TrimPrefix(sequenceBody, detailImageOSCCommand+";"), ";") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}
		parsedValue, actualErr := parseInteger(value)
		if actualErr != nil {
			continue
		}
		switch key {
		case "i":
			spec.imageID = uint32(parsedValue)
		case "c":
			spec.columns = parsedValue
		case "r":
			spec.rows = parsedValue
		}
	}

	if spec.imageID == 0 || spec.columns < 1 || spec.rows < 1 {
		return detailImageSpec{}, startIndex, false
	}

	return spec, nextIndex, true
}

func parseInteger(value string) (int, error) {
	parsedValue := 0
	for _, character := range strings.TrimSpace(value) {
		if character < '0' || character > '9' {
			return 0, errors.New("invalid integer")
		}
		parsedValue = (parsedValue * 10) + int(character-'0')
	}
	if parsedValue == 0 {
		return 0, nil
	}
	return parsedValue, nil
}

func loadDetailImage(source string, client *http.Client, githubToken string) (loadedDetailImage, error) {
	imageBytes, actualErr := loadDetailImageBytes(source, client, githubToken)
	if actualErr != nil {
		return loadedDetailImage{}, actualErr
	}

	decodedImage, _, actualErr := image.Decode(bytes.NewReader(imageBytes))
	if actualErr != nil {
		return loadedDetailImage{}, actualErr
	}
	var pngBuffer bytes.Buffer
	if actualErr = png.Encode(&pngBuffer, decodedImage); actualErr != nil {
		return loadedDetailImage{}, actualErr
	}

	bounds := decodedImage.Bounds()
	return loadedDetailImage{pixelWidth: bounds.Dx(), pixelHeight: bounds.Dy(), pngData: pngBuffer.Bytes()}, nil
}

func loadDetailImageBytes(source string, client *http.Client, githubToken string) ([]byte, error) {
	if strings.HasPrefix(source, "data:") {
		return decodeDetailImageDataURL(source)
	}

	parsedURL, actualErr := urlpkg.Parse(strings.TrimSpace(source))
	if actualErr != nil {
		return nil, actualErr
	}

	switch parsedURL.Scheme {
	case "http", "https":
		if client == nil {
			client = http.DefaultClient
		}
		request, actualErr := http.NewRequest(http.MethodGet, source, nil)
		if actualErr != nil {
			return nil, actualErr
		}
		if shouldAuthorizeGitHubImageRequest(parsedURL, githubToken) {
			request.Header.Set("Authorization", "token "+strings.TrimSpace(githubToken))
		}
		response, actualErr := client.Do(request)
		if actualErr != nil {
			return nil, actualErr
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil, fmt.Errorf("image request failed with status %s", response.Status)
		}
		return io.ReadAll(response.Body)
	case "file":
		return os.ReadFile(parsedURL.Path)
	case "":
		if strings.HasPrefix(source, "/") {
			return os.ReadFile(source)
		}
	}

	return nil, errors.New("unsupported image source")
}

func shouldAuthorizeGitHubImageRequest(parsedURL *urlpkg.URL, githubToken string) bool {
	if strings.TrimSpace(githubToken) == "" {
		return false
	}
	return isGitHubImageRequest(parsedURL)
}

func isGitHubImageSource(source string) bool {
	parsedURL, err := urlpkg.Parse(strings.TrimSpace(source))
	if err != nil {
		return false
	}
	return isGitHubImageRequest(parsedURL)
}

func isGitHubImageRequest(parsedURL *urlpkg.URL) bool {
	if parsedURL == nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	switch host {
	case "github.com", "raw.githubusercontent.com", "private-user-images.githubusercontent.com", "user-images.githubusercontent.com", "githubusercontent.com":
		return true
	default:
		return strings.HasSuffix(host, ".githubusercontent.com")
	}
}

func decodeDetailImageDataURL(source string) ([]byte, error) {
	metadata, payload, found := strings.Cut(strings.TrimPrefix(source, "data:"), ",")
	if !found {
		return nil, errors.New("invalid data url")
	}
	if strings.HasSuffix(metadata, ";base64") {
		return base64.StdEncoding.DecodeString(payload)
	}
	decodedPayload, actualErr := urlpkg.PathUnescape(payload)
	if actualErr != nil {
		return nil, actualErr
	}
	return []byte(decodedPayload), nil
}
