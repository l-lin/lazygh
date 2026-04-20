package tui

import (
	"strings"
	"unicode/utf8"
)

type styledTextLine struct {
	runes         []rune
	stylePrefixes []string
}

func splitStyledTextLines(text string) []styledTextLine {
	lines := []styledTextLine{{}}
	currentStylePrefix := ""

	for index := 0; index < len(text); {
		if text[index] == '\x1b' {
			if sequence, nextIndex, ok := consumeCSISequence(text, index); ok {
				if strings.HasSuffix(sequence, "m") {
					currentStylePrefix = updatedANSIStylePrefix(currentStylePrefix, sequence)
				}
				index = nextIndex
				continue
			}
			if nextIndex, ok := consumeOSCSequence(text, index); ok {
				index = nextIndex
				continue
			}
		}

		if text[index] == '\r' {
			index++
			continue
		}

		character, size := utf8.DecodeRuneInString(text[index:])
		if character == '\n' {
			lines = append(lines, styledTextLine{})
			index += size
			continue
		}

		line := &lines[len(lines)-1]
		line.runes = append(line.runes, character)
		line.stylePrefixes = append(line.stylePrefixes, currentStylePrefix)
		index += size
	}

	for index := range lines {
		lines[index] = trimTrailingStyledSpaces(lines[index])
	}

	return lines
}

func trimTrailingStyledSpaces(line styledTextLine) styledTextLine {
	trimmedLength := len(line.runes)
	for trimmedLength > 0 && line.runes[trimmedLength-1] == ' ' {
		trimmedLength--
	}

	line.runes = line.runes[:trimmedLength]
	line.stylePrefixes = line.stylePrefixes[:trimmedLength]
	return line
}

func renderStyledTextLine(line styledTextLine) string {
	if len(line.runes) == 0 {
		return ""
	}

	var builder strings.Builder
	currentPrefix := ""
	for index, character := range line.runes {
		prefix := ""
		if index < len(line.stylePrefixes) {
			prefix = line.stylePrefixes[index]
		}
		if prefix != currentPrefix {
			if currentPrefix != "" {
				builder.WriteString(ansiReset)
			}
			if prefix != "" {
				builder.WriteString(prefix)
			}
			currentPrefix = prefix
		}
		builder.WriteRune(character)
	}
	if currentPrefix != "" {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

func consumeCSISequence(text string, startIndex int) (string, int, bool) {
	if startIndex+1 >= len(text) || text[startIndex] != '\x1b' || text[startIndex+1] != '[' {
		return "", startIndex, false
	}

	for index := startIndex + 2; index < len(text); index++ {
		if text[index] >= 0x40 && text[index] <= 0x7e {
			return text[startIndex : index+1], index + 1, true
		}
	}

	return text[startIndex:], len(text), true
}

func consumeOSCSequence(text string, startIndex int) (int, bool) {
	if startIndex+1 >= len(text) || text[startIndex] != '\x1b' || text[startIndex+1] != ']' {
		return startIndex, false
	}

	for index := startIndex + 2; index < len(text); index++ {
		switch text[index] {
		case '\a':
			return index + 1, true
		case '\x1b':
			if index+1 < len(text) && text[index+1] == '\\' {
				return index + 2, true
			}
		}
	}

	return len(text), true
}

func updatedANSIStylePrefix(currentPrefix string, sequence string) string {
	parameters := ansiSequenceParameters(sequence)
	if len(parameters) == 0 {
		return ""
	}

	strippedParameters, containsReset := stripLeadingANSIReset(parameters)
	if containsReset {
		if len(strippedParameters) == 0 {
			return ""
		}
		return "\x1b[" + strings.Join(strippedParameters, ";") + "m"
	}
	if currentPrefix == "" {
		return sequence
	}

	return currentPrefix + sequence
}

func stripLeadingANSIReset(parameters []string) ([]string, bool) {
	firstNonResetIndex := 0
	containsReset := false
	for firstNonResetIndex < len(parameters) {
		if parameters[firstNonResetIndex] != "" && parameters[firstNonResetIndex] != "0" {
			break
		}
		containsReset = true
		firstNonResetIndex++
	}

	return parameters[firstNonResetIndex:], containsReset
}

func ansiSequenceParameters(sequence string) []string {
	if !strings.HasPrefix(sequence, "\x1b[") || !strings.HasSuffix(sequence, "m") {
		return nil
	}

	parameters := strings.TrimSuffix(strings.TrimPrefix(sequence, "\x1b["), "m")
	if parameters == "" {
		return nil
	}

	return strings.Split(parameters, ";")
}
