package story

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed default_prompt.md
var defaultPromptMarkdown string

//go:embed prompt_template.md.tmpl
var promptTemplateMarkdown string

var defaultPrompt = strings.TrimSpace(defaultPromptMarkdown)

var buildPromptTemplate = template.Must(template.New("story-prompt").Parse(promptTemplateMarkdown))

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
	data := buildPromptTemplateData(request, prompt)
	var builder strings.Builder
	if actualErr := buildPromptTemplate.Execute(&builder, data); actualErr != nil {
		panic(actualErr)
	}
	return builder.String()
}

func buildPromptTemplateData(request Request, prompt string) promptTemplateData {
	metadata := request.Metadata
	changedFilesList := make([]string, 0, len(request.DiffItems))
	for _, item := range request.DiffItems {
		trimmedFile := strings.TrimSpace(item.File)
		if trimmedFile == "" {
			continue
		}
		changedFilesList = append(changedFilesList, trimmedFile)
	}

	return promptTemplateData{
		ResolvedPrompt:   ResolvePrompt(prompt),
		Number:           metadata.Number,
		Title:            valueOrFallback(metadata.Title, "PR"),
		Body:             strings.TrimSpace(metadata.Body),
		Author:           valueOrFallback(metadata.Author, "unknown"),
		Base:             valueOrFallback(metadata.Base, "unknown"),
		Head:             valueOrFallback(metadata.Head, "unknown"),
		URL:              strings.TrimSpace(metadata.URL),
		Additions:        metadata.Additions,
		Deletions:        metadata.Deletions,
		ChangedFiles:     metadata.ChangedFiles,
		ChangedFilesList: changedFilesList,
		DiffText:         strings.TrimSpace(request.DiffText),
	}
}

type promptTemplateData struct {
	ResolvedPrompt   string
	Number           int
	Title            string
	Body             string
	Author           string
	Base             string
	Head             string
	URL              string
	Additions        int
	Deletions        int
	ChangedFiles     int
	ChangedFilesList []string
	DiffText         string
}

func valueOrFallback(value string, fallback string) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fallback
	}
	return trimmedValue
}
