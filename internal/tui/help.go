package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
)

const viewHelpName = "help"

type helpSection struct {
	Title   string
	Entries []helpEntry
}

type helpEntry struct {
	Key         string
	Description string
}

func (program *Program) configureHelpView(view *gocui.View) {
	configureFramedOverlayView(view, "Keybindings", "")
	view.Wrap = false
	view.Highlight = false
}

func (program *Program) renderHelpView(view *gocui.View) {
	view.Clear()

	presenter := program.helpPresenter()
	sections := presenter.sections()
	keyWidth := presenter.keyWidth(sections)
	for sectionIndex, section := range sections {
		fmt.Fprintf(view, "    --- %s ---\n", section.Title)
		for _, entry := range section.Entries {
			fmt.Fprintf(view, "%-*s %s\n", keyWidth+2, entry.Key, entry.Description)
		}
		if sectionIndex < len(sections)-1 {
			fmt.Fprintln(view)
		}
	}
}

func (program *Program) fullPageHelpDown(gui *gocui.Gui, view *gocui.View) error {
	return program.scrollReadOnlyView(gui, view, viewHelpName, fullPageDelta(viewPageSize(program.resolveView(gui, view, viewHelpName))))
}

func (program *Program) fullPageHelpUp(gui *gocui.Gui, view *gocui.View) error {
	return program.scrollReadOnlyView(gui, view, viewHelpName, -fullPageDelta(viewPageSize(program.resolveView(gui, view, viewHelpName))))
}

func formattedKeySequenceLabelsForDisplay(labels []string) []string {
	if len(labels) == 0 {
		return nil
	}

	formatted := make([]string, 0, len(labels))
	for _, label := range labels {
		formatted = append(formatted, formatKeySequenceLabelForDisplay(label))
	}
	return formatted
}

func formatKeyTextForDisplay(text string) string {
	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" || trimmedText == "/" || strings.Contains(trimmedText, "<c-/>") {
		return formatKeySequenceLabelForDisplay(trimmedText)
	}

	segments := strings.Split(trimmedText, "/")
	if len(segments) <= 1 {
		return formatKeySequenceLabelForDisplay(trimmedText)
	}
	for index, segment := range segments {
		if segment == "" {
			return formatKeySequenceLabelForDisplay(trimmedText)
		}
		segments[index] = formatKeySequenceLabelForDisplay(segment)
	}
	return strings.Join(segments, "/")
}

func formatKeySequenceLabelForDisplay(label string) string {
	trimmedLabel := strings.TrimSpace(label)
	if trimmedLabel == "" {
		return ""
	}

	normalizedLabel := strings.ToLower(trimmedLabel)
	normalizedLabel = strings.TrimPrefix(normalizedLabel, "<")
	normalizedLabel = strings.TrimSuffix(normalizedLabel, ">")
	if after, ok := strings.CutPrefix(normalizedLabel, "control+"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "ctrl-"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "c-"); ok {
		normalizedLabel = "ctrl+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "alt-"); ok {
		normalizedLabel = "alt+" + after
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "shift-"); ok {
		normalizedLabel = "shift+" + after
	}

	switch normalizedLabel {
	case "enter":
		return "Enter"
	case "esc", "escape":
		return "Escape"
	case "tab":
		return "Tab"
	case "shift+tab", "backtab":
		return "Shift+Tab"
	case "up", "arrowup", "arrow-up":
		return "Up"
	case "down", "arrowdown", "arrow-down":
		return "Down"
	case "left", "arrowleft", "arrow-left":
		return "Left"
	case "right", "arrowright", "arrow-right":
		return "Right"
	case "pageup", "page-up", "pgup":
		return "PageUp"
	case "pagedown", "page-down", "pgdown", "pgdn":
		return "PageDown"
	case "space":
		return "Space"
	case "alt+enter":
		return "Alt+Enter"
	}

	if after, ok := strings.CutPrefix(normalizedLabel, "ctrl+"); ok {
		suffix := after
		switch suffix {
		case "space":
			return "Ctrl+Space"
		case "lsqbracket":
			suffix = "["
		case "rsqbracket":
			suffix = "]"
		case "backslash":
			suffix = "\\"
		case "slash":
			suffix = "/"
		case "underscore":
			suffix = "_"
		default:
			suffix = strings.ToUpper(suffix)
		}
		return "Ctrl+" + suffix
	}
	if after, ok := strings.CutPrefix(normalizedLabel, "alt+"); ok {
		return "Alt+" + formattedModifiedKeySuffix(after)
	}

	return trimmedLabel
}

func formattedModifiedKeySuffix(label string) string {
	if utf8.RuneCountInString(label) == 1 {
		return strings.ToUpper(label)
	}
	return formatKeySequenceLabelForDisplay(label)
}
