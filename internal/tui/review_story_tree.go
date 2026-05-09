package tui

import (
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/story"
)

type reviewStoryChapter struct {
	ID          string
	Title       string
	Narrative   string
	Files       []string
	FileIndexes []int
}

type reviewStoryData struct {
	Summary  string
	Chapters []reviewStoryChapter
	Tree     reviewDiffTree
}

func buildReviewStoryData(review story.Review, files []reviewDiffFile) reviewStoryData {
	fileIndexesByPath := make(map[string]int, len(files))
	for index, file := range files {
		trimmedPath := strings.TrimSpace(file.Path)
		if trimmedPath == "" {
			continue
		}
		fileIndexesByPath[trimmedPath] = index
	}

	chapters := make([]reviewStoryChapter, 0, len(review.Chapters))
	rows := make([]reviewDiffTreeRow, 0)
	for chapterIndex, chapter := range review.Chapters {
		reviewChapter := reviewStoryChapter{
			ID:        strings.TrimSpace(chapter.ID),
			Title:     strings.TrimSpace(chapter.Title),
			Narrative: strings.TrimSpace(chapter.Narrative),
			Files:     append([]string(nil), chapter.Files...),
		}
		for _, file := range chapter.Files {
			if fileIndex, ok := fileIndexesByPath[strings.TrimSpace(file)]; ok {
				reviewChapter.FileIndexes = append(reviewChapter.FileIndexes, fileIndex)
			}
		}
		chapters = append(chapters, reviewChapter)
		chapterRowID := reviewDiffTreeRowIDForChapter(reviewChapter.ID, chapterIndex)
		appendReviewDiffTreeRow(&rows, chapterRowID, 0, reviewStoryChapterLabel(chapterIndex, reviewChapter.Title, len(reviewChapter.FileIndexes)), -1, chapterIndex, reviewDiffTreeRowKindChapter, len(reviewChapter.FileIndexes) > 0)
		for _, fileIndex := range reviewChapter.FileIndexes {
			appendReviewDiffTreeRow(&rows, chapterRowID+":"+reviewDiffTreeRowIDForFile(files[fileIndex].Path), 1, files[fileIndex].Path, fileIndex, 0, reviewDiffTreeRowKindFile, false)
		}
	}

	return reviewStoryData{
		Summary:  strings.TrimSpace(review.Summary),
		Chapters: chapters,
		Tree:     reviewDiffTree{Rows: rows},
	}
}

func reviewStoryChapterLabel(index int, title string, fileCount int) string {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		trimmedTitle = fmt.Sprintf("Chapter %d", index+1)
	} else if !strings.HasPrefix(strings.ToLower(trimmedTitle), "chapter ") {
		trimmedTitle = fmt.Sprintf("Chapter %d - %s", index+1, trimmedTitle)
	}

	suffix := ""
	switch fileCount {
	case 1:
		suffix = " (1 file)"
	case 0:
	default:
		suffix = fmt.Sprintf(" (%d files)", fileCount)
	}
	return trimmedTitle + suffix
}
