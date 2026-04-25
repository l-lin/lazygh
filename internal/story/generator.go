package story

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrAgentNotConfigured = errors.New("story review agent is not configured")
	ErrInvalidStoryOutput = errors.New("story review agent returned invalid output")
	ErrNoUsableChapters   = errors.New("story review agent returned no usable chapters")
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
}

type Runner interface {
	Run(name string, args ...string) (CommandResult, error)
}

type execRunner struct{}

type Generator struct {
	runner Runner
}

func NewGenerator(runner Runner) Generator {
	if runner == nil {
		runner = execRunner{}
	}
	return Generator{runner: runner}
}

func (execRunner) Run(name string, args ...string) (CommandResult, error) {
	command := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	actualErr := command.Run()
	return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, actualErr
}

func BuildCommand(template []string, promptPath string) []string {
	if len(template) == 0 {
		return nil
	}

	resolved := make([]string, 0, len(template)+1)
	placeholderReplaced := false
	for _, argument := range template {
		replaced := strings.ReplaceAll(argument, PromptFilePlaceholder, promptPath)
		if replaced != argument {
			placeholderReplaced = true
		}
		resolved = append(resolved, replaced)
	}
	if !placeholderReplaced {
		resolved = append(resolved, promptPath)
	}
	return resolved
}

func (generator Generator) Generate(config Config, request Request) (Review, error) {
	resolvedConfig := ResolveConfig(config)
	if !resolvedConfig.Configured() {
		return Review{}, ErrAgentNotConfigured
	}

	promptPath, actualErr := writePromptFile(BuildPrompt(request, resolvedConfig.Prompt))
	if actualErr != nil {
		return Review{}, actualErr
	}
	defer func() { _ = os.Remove(promptPath) }()

	command := BuildCommand(resolvedConfig.AgentCommand, promptPath)
	if len(command) == 0 {
		return Review{}, ErrAgentNotConfigured
	}

	result, actualErr := generator.runner.Run(command[0], command[1:]...)
	if actualErr != nil {
		stderr := strings.TrimSpace(string(result.Stderr))
		if stderr == "" {
			return Review{}, fmt.Errorf("run story review agent: %w", actualErr)
		}
		return Review{}, fmt.Errorf("run story review agent: %w: %s", actualErr, stderr)
	}

	review, actualErr := DecodeReview(string(result.Stdout))
	if actualErr != nil {
		return Review{}, actualErr
	}
	return NormalizeReview(review, request.DiffItems)
}

func writePromptFile(prompt string) (string, error) {
	file, actualErr := os.CreateTemp("", "lazygh-story-review-*.md")
	if actualErr != nil {
		return "", actualErr
	}
	path := file.Name()
	if _, actualErr = file.WriteString(prompt); actualErr != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", actualErr
	}
	if actualErr = file.Close(); actualErr != nil {
		_ = os.Remove(path)
		return "", actualErr
	}
	return path, nil
}

func DecodeReview(raw string) (Review, error) {
	decoded := decodeJSONReview(strings.TrimSpace(raw))
	if decoded == nil {
		return Review{}, ErrInvalidStoryOutput
	}
	return *decoded, nil
}

func decodeJSONReview(raw string) *Review {
	if raw == "" {
		return nil
	}
	decode := func(candidate string) *Review {
		var payload struct {
			Summary  string `json:"summary"`
			Chapters []struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				Narrative string `json:"narrative"`
				Review    string `json:"review"`
				Story     string `json:"story"`
				Files     []any  `json:"files"`
			} `json:"chapters"`
		}
		if actualErr := json.Unmarshal([]byte(candidate), &payload); actualErr != nil {
			return nil
		}
		review := &Review{Summary: strings.TrimSpace(payload.Summary)}
		for index, chapter := range payload.Chapters {
			narrative := strings.TrimSpace(chapter.Narrative)
			if narrative == "" {
				narrative = strings.TrimSpace(chapter.Review)
			}
			if narrative == "" {
				narrative = strings.TrimSpace(chapter.Story)
			}
			files := make([]string, 0, len(chapter.Files))
			for _, rawFile := range chapter.Files {
				switch actual := rawFile.(type) {
				case string:
					trimmedFile := strings.TrimSpace(actual)
					if trimmedFile != "" {
						files = append(files, trimmedFile)
					}
				case map[string]any:
					for _, key := range []string{"path", "file", "filename"} {
						value, ok := actual[key].(string)
						if !ok {
							continue
						}
						trimmedFile := strings.TrimSpace(value)
						if trimmedFile != "" {
							files = append(files, trimmedFile)
							break
						}
					}
				}
			}
			review.Chapters = append(review.Chapters, Chapter{
				ID:        valueOrFallback(chapter.ID, fmt.Sprintf("chapter-%d", index+1)),
				Title:     valueOrFallback(chapter.Title, fmt.Sprintf("Chapter %d", index+1)),
				Narrative: narrative,
				Files:     files,
			})
		}
		return review
	}

	if review := decode(raw); review != nil {
		return review
	}
	if fenced := extractFencedBlock(raw); fenced != "" {
		if review := decode(fenced); review != nil {
			return review
		}
	}
	if fragment := extractJSONObject(raw); fragment != "" {
		return decode(fragment)
	}
	return nil
}

func extractFencedBlock(raw string) string {
	for _, marker := range []string{"```json", "```"} {
		start := strings.Index(raw, marker)
		if start < 0 {
			continue
		}
		remaining := raw[start+len(marker):]
		end := strings.Index(remaining, "```")
		if end < 0 {
			continue
		}
		return strings.TrimSpace(remaining[:end])
	}
	return ""
}

func extractJSONObject(raw string) string {
	firstBrace := strings.Index(raw, "{")
	lastBrace := strings.LastIndex(raw, "}")
	if firstBrace < 0 || lastBrace < 0 || lastBrace < firstBrace {
		return ""
	}
	return raw[firstBrace : lastBrace+1]
}

func NormalizeReview(review Review, diffItems []DiffItem) (Review, error) {
	diffFiles := make([]string, 0, len(diffItems))
	diffFileSet := map[string]bool{}
	for _, item := range diffItems {
		trimmedFile := strings.TrimSpace(item.File)
		if trimmedFile == "" || diffFileSet[trimmedFile] {
			continue
		}
		diffFileSet[trimmedFile] = true
		diffFiles = append(diffFiles, trimmedFile)
	}

	normalized := Review{Summary: strings.TrimSpace(review.Summary)}
	assigned := map[string]bool{}
	for index, chapter := range review.Chapters {
		files := make([]string, 0, len(chapter.Files))
		for _, file := range chapter.Files {
			trimmedFile := strings.TrimSpace(file)
			if !diffFileSet[trimmedFile] || assigned[trimmedFile] {
				continue
			}
			assigned[trimmedFile] = true
			files = append(files, trimmedFile)
		}
		if len(diffFiles) > 0 && len(files) == 0 {
			continue
		}
		normalized.Chapters = append(normalized.Chapters, Chapter{
			ID:        valueOrFallback(chapter.ID, fmt.Sprintf("chapter-%d", index+1)),
			Title:     valueOrFallback(chapter.Title, fmt.Sprintf("Chapter %d", index+1)),
			Narrative: strings.TrimSpace(chapter.Narrative),
			Files:     files,
		})
	}

	if len(diffFiles) > 0 && len(normalized.Chapters) == 0 {
		return Review{}, ErrNoUsableChapters
	}

	unassigned := make([]string, 0, len(diffFiles))
	for _, file := range diffFiles {
		if !assigned[file] {
			unassigned = append(unassigned, file)
		}
	}
	if len(unassigned) > 0 {
		normalized.Chapters = append(normalized.Chapters, Chapter{
			ID:        UnassignedChapterID,
			Title:     "Unassigned Changes",
			Narrative: "These files did not fit cleanly into the named chapters. Review them together as the remaining thread.",
			Files:     unassigned,
		})
	}

	return normalized, nil
}
