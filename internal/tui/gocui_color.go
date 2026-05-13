package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
)

func gocuiColorOrDefault(hexColor string) gocui.Attribute {
	if strings.TrimSpace(hexColor) == "" {
		return gocui.ColorDefault
	}
	return gocui.GetColor(hexColor)
}

func hexColorForAttribute(attribute gocui.Attribute) string {
	color := attribute & gocui.AttrColorBits
	red, green, blue := color.RGB()
	if red < 0 || green < 0 || blue < 0 {
		return ""
	}
	return fmt.Sprintf("#%02X%02X%02X", red, green, blue)
}

func readableForegroundHexForBackground(preferredHex string, backgroundHex string, fallbackHexes ...string) string {
	return readableForegroundHexForBackgroundWithMinimum(preferredHex, backgroundHex, 4.5, fallbackHexes...)
}

func readableForegroundHexForBackgroundWithMinimum(preferredHex string, backgroundHex string, minimumContrast float64, fallbackHexes ...string) string {
	trimmedBackgroundHex := strings.TrimSpace(backgroundHex)
	trimmedPreferredHex := strings.TrimSpace(preferredHex)
	if minimumContrast <= 0 {
		minimumContrast = 4.5
	}
	if trimmedBackgroundHex == "" {
		if trimmedPreferredHex != "" {
			return trimmedPreferredHex
		}
		for _, fallbackHex := range fallbackHexes {
			if trimmedFallbackHex := strings.TrimSpace(fallbackHex); trimmedFallbackHex != "" {
				return trimmedFallbackHex
			}
		}
		return ""
	}

	candidates := uniqueReadableForegroundCandidates(append([]string{trimmedPreferredHex}, append(fallbackHexes, markdownReadableDarkHex, markdownReadableLightHex)...))
	bestHex := ""
	bestContrast := 0.0
	for _, candidateHex := range candidates {
		actualContrast := foregroundContrastRatio(candidateHex, trimmedBackgroundHex)
		if actualContrast >= minimumContrast {
			return candidateHex
		}
		if actualContrast > bestContrast {
			bestContrast = actualContrast
			bestHex = candidateHex
		}
	}
	return bestHex
}

func uniqueReadableForegroundCandidates(candidates []string) []string {
	unique := make([]string, 0, len(candidates))
	seen := map[string]bool{}
	for _, candidate := range candidates {
		trimmedCandidate := strings.TrimSpace(candidate)
		if trimmedCandidate == "" || seen[trimmedCandidate] {
			continue
		}
		seen[trimmedCandidate] = true
		unique = append(unique, trimmedCandidate)
	}
	return unique
}

func foregroundContrastRatio(foregroundHex string, backgroundHex string) float64 {
	foregroundLuminance, ok := relativeLuminance(strings.TrimSpace(foregroundHex))
	if !ok {
		return 0
	}
	backgroundLuminance, ok := relativeLuminance(strings.TrimSpace(backgroundHex))
	if !ok {
		return 0
	}
	return contrastRatio(foregroundLuminance, backgroundLuminance)
}
