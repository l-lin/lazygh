package tui

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	glamouransi "charm.land/glamour/v2/ansi"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	glamourImageAltTokenPrefix = "⟪a:"
	glamourImageURLTokenPrefix = "⟪u:"
	glamourImageTokenSuffix    = "⟫"
)

type markdownImageToken struct {
	altText  string
	imageURL string
}

type markdownImageOccurrence struct {
	start    int
	end      int
	altText  string
	imageURL string
}

func (renderer glamourMarkdownRenderer) actualImageStore() detailImageStore {
	if renderer.imageStore != nil {
		return renderer.imageStore
	}
	return defaultDetailImageStore
}

func (renderer glamourMarkdownRenderer) actualImageProtocol() detailImageProtocol {
	if renderer.imageProtocol != nil {
		return renderer.imageProtocol
	}
	return kittyImageProtocol{}
}

func (renderer glamourMarkdownRenderer) actualTerminalCellSize() terminalCellSizeProvider {
	if renderer.terminalCellSize != nil {
		return renderer.terminalCellSize
	}
	return screenTerminalCellSize{}
}

func renderMarkdownWithImageMarkers(markdown string, width int, imageStore detailImageStore, imageProtocol detailImageProtocol, terminalCellSize terminalCellSizeProvider) (string, error) {
	tokenizedMarkdown, tokens := tokenizeMarkdownImages(markdown)
	rendered, err := renderMarkdownWithGlamourStyle(tokenizedMarkdown, width, prettyMarkdownImageStyle())
	if err != nil {
		return "", err
	}

	return replaceRenderedMarkdownImages(rendered, width, imageStore, imageProtocol, terminalCellSize, tokens), nil
}

func prettyMarkdownImageStyle() glamouransi.StyleConfig {
	style := prettyMarkdownStyle()
	style.ImageText.Format = glamourImageAltTokenPrefix + "{{.text}}" + glamourImageTokenSuffix
	style.ImageText.Color = nil
	style.ImageText.Bold = nil
	style.ImageText.Underline = nil
	style.ImageText.Prefix = ""
	style.Image.Format = glamourImageURLTokenPrefix + "{{.text}}" + glamourImageTokenSuffix
	style.Image.Color = nil
	style.Image.Bold = nil
	style.Image.Underline = nil
	return style
}

func tokenizeMarkdownImages(markdown string) (string, map[string]markdownImageToken) {
	occurrences := collectMarkdownImageOccurrences(markdown)
	if len(occurrences) == 0 {
		return markdown, nil
	}

	tokens := make(map[string]markdownImageToken, len(occurrences))
	var builder strings.Builder
	currentIndex := 0
	for occurrenceIndex, occurrence := range occurrences {
		if occurrence.start < currentIndex {
			continue
		}
		builder.WriteString(markdown[currentIndex:occurrence.start])
		tokenID := fmt.Sprintf("%d", occurrenceIndex+1)
		tokens[tokenID] = markdownImageToken{altText: strings.TrimSpace(occurrence.altText), imageURL: strings.TrimSpace(occurrence.imageURL)}
		builder.WriteString("![")
		builder.WriteString(tokenID)
		builder.WriteString("](")
		builder.WriteString(tokenID)
		builder.WriteString(")")
		currentIndex = occurrence.end
	}
	builder.WriteString(markdown[currentIndex:])
	return builder.String(), tokens
}

func collectMarkdownImageOccurrences(markdown string) []markdownImageOccurrence {
	occurrences := make([]markdownImageOccurrence, 0)
	for index := 0; index < len(markdown); {
		startIndex := strings.Index(markdown[index:], "![")
		if startIndex < 0 {
			break
		}
		startIndex += index
		altEndIndex := strings.Index(markdown[startIndex+2:], "](")
		if altEndIndex < 0 {
			index = startIndex + 2
			continue
		}
		altEndIndex += startIndex + 2
		urlEndIndex := strings.Index(markdown[altEndIndex+2:], ")")
		if urlEndIndex < 0 {
			index = altEndIndex + 2
			continue
		}
		urlEndIndex += altEndIndex + 2
		imageURL := strings.TrimSpace(markdown[altEndIndex+2 : urlEndIndex])
		if fields := strings.Fields(imageURL); len(fields) > 0 {
			imageURL = fields[0]
		}
		occurrences = append(occurrences, markdownImageOccurrence{start: startIndex, end: urlEndIndex + 1, altText: markdown[startIndex+2 : altEndIndex], imageURL: imageURL})
		index = urlEndIndex + 1
	}
	return occurrences
}

func replaceRenderedMarkdownImages(rendered string, width int, imageStore detailImageStore, imageProtocol detailImageProtocol, terminalCellSize terminalCellSizeProvider, tokens map[string]markdownImageToken) string {
	if rendered == "" || len(tokens) == 0 {
		return rendered
	}

	var builder strings.Builder
	for index := 0; index < len(rendered); {
		altTokenID, nextIndex, ok := consumeGlamourImageToken(rendered, index, glamourImageAltTokenPrefix)
		if !ok {
			character, size := utf8.DecodeRuneInString(rendered[index:])
			builder.WriteRune(character)
			index += size
			continue
		}

		urlTokenID, afterURLIndex, urlFound := consumeFollowingGlamourImageURLToken(rendered, nextIndex)
		if !urlFound {
			builder.WriteString(rendered[index:nextIndex])
			index = nextIndex
			continue
		}

		imageToken, tokenFound := tokens[strings.TrimSpace(altTokenID)]
		if !tokenFound {
			builder.WriteString(rendered[index:afterURLIndex])
			index = afterURLIndex
			continue
		}
		if strings.TrimLeft(strings.TrimSpace(urlTokenID), "/") != strings.TrimSpace(altTokenID) {
			builder.WriteString(rendered[index:afterURLIndex])
			index = afterURLIndex
			continue
		}

		builder.WriteString(renderRenderedMarkdownImageBlock(imageToken.altText, imageToken.imageURL, width, imageStore, imageProtocol, terminalCellSize))
		index = afterURLIndex
	}

	return builder.String()
}

func consumeGlamourImageToken(rendered string, startIndex int, prefix string) (string, int, bool) {
	if !strings.HasPrefix(rendered[startIndex:], prefix) {
		return "", startIndex, false
	}

	var builder strings.Builder
	for index := startIndex + len(prefix); index < len(rendered); {
		if strings.HasPrefix(rendered[index:], glamourImageTokenSuffix) {
			return builder.String(), index + len(glamourImageTokenSuffix), true
		}
		if rendered[index] == '\x1b' {
			if sequence, nextIndex, ok := consumeCSISequence(rendered, index); ok {
				if strings.HasSuffix(sequence, "m") {
					index = nextIndex
					continue
				}
			}
			if _, nextIndex, ok := consumeOSCSequence(rendered, index); ok {
				index = nextIndex
				continue
			}
		}

		character, size := utf8.DecodeRuneInString(rendered[index:])
		builder.WriteRune(character)
		index += size
	}

	return "", startIndex, false
}

func consumeFollowingGlamourImageURLToken(rendered string, startIndex int) (string, int, bool) {
	for index := startIndex; index < len(rendered); {
		if strings.HasPrefix(rendered[index:], glamourImageURLTokenPrefix) {
			return consumeGlamourImageToken(rendered, index, glamourImageURLTokenPrefix)
		}
		if rendered[index] == '\x1b' {
			if sequence, nextIndex, ok := consumeCSISequence(rendered, index); ok {
				if strings.HasSuffix(sequence, "m") {
					index = nextIndex
					continue
				}
			}
			if _, nextIndex, ok := parseDetailImageMarkerSequence(rendered, index); ok {
				index = nextIndex
				continue
			}
			if _, nextIndex, ok := consumeOSCSequence(rendered, index); ok {
				index = nextIndex
				continue
			}
		}

		character, size := utf8.DecodeRuneInString(rendered[index:])
		if unicode.IsSpace(character) {
			index += size
			continue
		}
		break
	}

	return "", startIndex, false
}

func renderRenderedMarkdownImageBlock(altText string, imageURL string, width int, imageStore detailImageStore, imageProtocol detailImageProtocol, terminalCellSize terminalCellSizeProvider) string {
	caption := renderMarkdownImageCaption(altText)
	link := renderMarkdownImageLink(imageURL)
	fallback := strings.Join([]string{caption, link}, "\n")
	if imageStore == nil || strings.TrimSpace(imageURL) == "" || imageProtocol == nil || !imageProtocol.Supported() {
		return fallback
	}

	storedImage, ok := imageStore.ImageBySource(imageURL)
	if !ok {
		return fallback
	}

	cellWidth, cellHeight, cellSizeKnown := 0, 0, false
	if terminalCellSize != nil {
		cellWidth, cellHeight, cellSizeKnown = terminalCellSize.CellSize()
	}
	if !cellSizeKnown {
		cellWidth = 8
		cellHeight = 16
	}
	imageColumns, imageRows := imageProtocol.Bounds(width, storedImage.pixelWidth, storedImage.pixelHeight, cellWidth, cellHeight)
	if imageColumns < 1 || imageRows < 1 {
		return fallback
	}

	lines := make([]string, 0, imageRows+2)
	lines = append(lines, encodeDetailImageMarker(detailImageSpec{imageID: storedImage.imageID, columns: imageColumns, rows: imageRows}))
	for row := 1; row < imageRows; row++ {
		lines = append(lines, "")
	}
	lines = append(lines, caption, link)
	return strings.Join(lines, "\n")
}

func renderMarkdownImageCaption(altText string) string {
	trimmedAltText := strings.TrimSpace(altText)
	if trimmedAltText == "" {
		trimmedAltText = "Image"
	}
	return styleText(fmt.Sprintf("[Image: %s]", trimmedAltText), foregroundColorEscape(theme.InactiveTitleHex))
}

func renderMarkdownImageLink(imageURL string) string {
	return styleText(strings.TrimSpace(imageURL), foregroundColorEscape(theme.MarkdownLinkHex))
}
