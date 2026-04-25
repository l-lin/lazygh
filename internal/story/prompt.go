package story

import (
	"fmt"
	"strings"
)

const defaultPrompt = "Group the changes into a logical, reviewer-friendly story. Use a professional tone. Prefer chapters that reflect one cohesive behavior change, refactor step, or debugging thread. Explain what each chapter is doing, why it exists, and what a reviewer should mentally connect across the listed files. Keep the narrative concise, concrete, and useful for code review."

func DefaultPrompt() string {
	return defaultPrompt
}

func ResolvePrompt(prompt string) string {
	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		return defaultPrompt
	}
	return trimmedPrompt
}

func BuildPrompt(request Request, prompt string) string {
	resolvedPrompt := ResolvePrompt(prompt)
	metadata := request.Metadata
	parts := []string{
		"# Goal",
		"Split this pull request diff into a reviewer-friendly story told as chapters.",
		"",
		"## Voice and guidance",
		resolvedPrompt,
		"",
		"## Output",
		"Return JSON only. No markdown fences or commentary outside the JSON. Markdown inside the `narrative` field is allowed.",
		"Use this exact schema:",
		`{`,
		`  "summary": "<2-3 sentence overview of the PR arc>",`,
		`  "chapters": [`,
		`    {`,
		`      "id": "chapter-1",`,
		`      "title": "<short chapter title>",`,
		`      "narrative": "<markdown chapter narrative>",`,
		`      "files": ["path/to/file.go"]`,
		`    }`,
		`  ]`,
		`}`,
		"",
		"## Rules",
		"- Use exact repo-relative file paths from the changed-files list.",
		"- Every changed file should appear exactly once across the named chapters.",
		"- Prefer 2-7 chapters unless the pull request is tiny.",
		"- Group files by reviewer intent, behavior change, refactor step, or shared debugging thread.",
		"- Keep titles concise and specific.",
		"- Explain what each chapter is about and why it matters for review.",
		"- The narrative is guidance for a reviewer, not a code review verdict.",
		"",
		fmt.Sprintf("# Pull Request #%d: %s", metadata.Number, valueOrFallback(metadata.Title, "PR")),
		fmt.Sprintf("Author: %s", valueOrFallback(metadata.Author, "unknown")),
		fmt.Sprintf("Branch: %s <- %s", valueOrFallback(metadata.Base, "unknown"), valueOrFallback(metadata.Head, "unknown")),
		fmt.Sprintf("URL: %s", strings.TrimSpace(metadata.URL)),
		fmt.Sprintf("Changes: +%d -%d across %d files", metadata.Additions, metadata.Deletions, metadata.ChangedFiles),
		"",
	}

	if strings.TrimSpace(metadata.Body) != "" {
		parts = append(parts, "## PR Description", strings.TrimSpace(metadata.Body), "")
	}

	parts = append(parts, "## Changed Files")
	for _, item := range request.DiffItems {
		trimmedFile := strings.TrimSpace(item.File)
		if trimmedFile == "" {
			continue
		}
		parts = append(parts, "- "+trimmedFile)
	}
	parts = append(parts, "", "## Diff", "```diff", strings.TrimSpace(request.DiffText), "```")
	return strings.Join(parts, "\n")
}

func valueOrFallback(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}
	return trimmedValue
}
