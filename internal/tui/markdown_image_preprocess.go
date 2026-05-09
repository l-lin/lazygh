package tui

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	wrappedHTMLImagePattern = regexp.MustCompile(`(?is)<(?:p|div)\b[^>]*>\s*(<img\b[^>]*>)\s*</(?:p|div)>`)
	bareHTMLImagePattern    = regexp.MustCompile(`(?is)<img\b[^>]*>`)
)

type htmlImageSource struct {
	altText string
	src     string
}

func prepareMarkdownForImageRendering(markdown string, renderedHTML string) string {
	normalizedMarkdown := strings.TrimSpace(replaceHTMLImagesWithMarkdown(markdown))
	if normalizedMarkdown == "" {
		return normalizedMarkdown
	}

	occurrences := collectMarkdownImageOccurrences(normalizedMarkdown)
	if len(occurrences) == 0 {
		return normalizedMarkdown
	}

	renderedImages := renderedHTMLImageSources(renderedHTML)
	if len(renderedImages) == 0 {
		return normalizedMarkdown
	}

	var builder strings.Builder
	currentIndex := 0
	for occurrenceIndex, occurrence := range occurrences {
		builder.WriteString(normalizedMarkdown[currentIndex:occurrence.start])
		altText := strings.TrimSpace(occurrence.altText)
		imageURL := strings.TrimSpace(occurrence.imageURL)
		if occurrenceIndex < len(renderedImages) {
			renderedImage := renderedImages[occurrenceIndex]
			if strings.TrimSpace(renderedImage.src) != "" {
				imageURL = strings.TrimSpace(renderedImage.src)
			}
			if altText == "" && strings.TrimSpace(renderedImage.altText) != "" {
				altText = strings.TrimSpace(renderedImage.altText)
			}
		}
		builder.WriteString(renderMarkdownImageSyntax(altText, imageURL))
		currentIndex = occurrence.end
	}
	builder.WriteString(normalizedMarkdown[currentIndex:])
	return strings.TrimSpace(builder.String())
}

func replaceHTMLImagesWithMarkdown(markdown string) string {
	withoutWrappedImages := wrappedHTMLImagePattern.ReplaceAllStringFunc(markdown, func(fragment string) string {
		imageTag := bareHTMLImagePattern.FindString(fragment)
		if imageTag == "" {
			return fragment
		}
		altText, imageURL, ok := parseHTMLImageTag(imageTag)
		if !ok {
			return fragment
		}
		return renderMarkdownImageSyntax(altText, imageURL)
	})

	return bareHTMLImagePattern.ReplaceAllStringFunc(withoutWrappedImages, func(fragment string) string {
		altText, imageURL, ok := parseHTMLImageTag(fragment)
		if !ok {
			return fragment
		}
		return renderMarkdownImageSyntax(altText, imageURL)
	})
}

func parseHTMLImageTag(fragment string) (string, string, bool) {
	tokenizer := html.NewTokenizer(strings.NewReader(fragment))
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return "", "", false
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if !strings.EqualFold(strings.TrimSpace(token.Data), "img") {
				continue
			}
			return htmlImageAttributes(token)
		}
	}
}

func htmlImageAttributes(token html.Token) (string, string, bool) {
	altText := ""
	imageURL := ""
	for _, attribute := range token.Attr {
		switch strings.ToLower(strings.TrimSpace(attribute.Key)) {
		case "alt":
			altText = strings.TrimSpace(attribute.Val)
		case "src":
			imageURL = strings.TrimSpace(attribute.Val)
		}
	}
	if imageURL == "" {
		return "", "", false
	}
	return altText, imageURL, true
}

func renderedHTMLImageSources(renderedHTML string) []htmlImageSource {
	trimmedHTML := strings.TrimSpace(renderedHTML)
	if trimmedHTML == "" {
		return nil
	}

	tokenizer := html.NewTokenizer(strings.NewReader(trimmedHTML))
	images := make([]htmlImageSource, 0)
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return images
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if !strings.EqualFold(strings.TrimSpace(token.Data), "img") {
				continue
			}
			altText, imageURL, ok := htmlImageAttributes(token)
			if !ok {
				continue
			}
			images = append(images, htmlImageSource{altText: altText, src: imageURL})
		}
	}
}

func renderMarkdownImageSyntax(altText string, imageURL string) string {
	trimmedAltText := strings.TrimSpace(altText)
	if trimmedAltText == "" {
		trimmedAltText = "Image"
	}
	return "![" + trimmedAltText + "](" + strings.TrimSpace(imageURL) + ")"
}
