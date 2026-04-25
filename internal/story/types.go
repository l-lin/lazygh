package story

import "strings"

const (
	PromptFilePlaceholder = "{{prompt_file}}"
	UnassignedChapterID   = "chapter-unassigned"
)

type Config struct {
	AgentCommand []string
	Prompt       string
}

type Metadata struct {
	Number       int
	Title        string
	Body         string
	Author       string
	Base         string
	Head         string
	URL          string
	Additions    int
	Deletions    int
	ChangedFiles int
}

type DiffItem struct {
	File string
}

type Request struct {
	Metadata  Metadata
	DiffItems []DiffItem
	DiffText  string
}

type Review struct {
	Summary  string
	Chapters []Chapter
}

type Chapter struct {
	ID        string
	Title     string
	Narrative string
	Files     []string
}

func ResolveConfig(config Config) Config {
	resolved := Config{
		AgentCommand: append([]string(nil), config.AgentCommand...),
		Prompt:       ResolvePrompt(config.Prompt),
	}
	for index, argument := range resolved.AgentCommand {
		resolved.AgentCommand[index] = strings.TrimSpace(argument)
	}
	return resolved
}

func (config Config) Configured() bool {
	return len(config.AgentCommand) > 0
}
