package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	_ "modernc.org/sqlite"
)

const sqliteDriverName = "sqlite"

type Store struct {
	db *sql.DB
}

type CachedPullRequestDetail struct {
	Detail          githubcli.PullRequestDetail
	SourceUpdatedAt string
}

type CachedPullRequestDiff struct {
	Diff            githubcli.PullRequestDiff
	SourceUpdatedAt string
}

type cachedPullRequestDetailPayload struct {
	Detail               githubcli.PullRequestDetail `json:"detail"`
	InlineComments       []githubcli.PullRequestInlineComment
	InlineCommentThreads []cachedPullRequestReviewThread
}

type cachedPullRequestDiffPayload struct {
	Diff    githubcli.PullRequestDiff `json:"diff"`
	Threads []cachedPullRequestReviewThread
}

type cachedPullRequestReviewThread struct {
	ID                 string
	IsResolved         bool
	IsOutdated         bool
	ViewerCanResolve   bool
	ViewerCanUnresolve bool
	Path               string
	Line               int
	OriginalLine       int
	StartLine          int
	OriginalStartLine  int
	DiffSide           string
	StartDiffSide      string
	Comments           []githubcli.PullRequestComment
}

func Open(path string) (*Store, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, errors.New("missing cache path")
	}

	if actualErr := os.MkdirAll(filepath.Dir(trimmedPath), 0o755); actualErr != nil {
		return nil, actualErr
	}

	database, actualErr := sql.Open(sqliteDriverName, trimmedPath)
	if actualErr != nil {
		return nil, actualErr
	}

	store := &Store{db: database}
	if actualErr := store.initialize(); actualErr != nil {
		_ = database.Close()
		return nil, actualErr
	}

	return store, nil
}

func (store *Store) Close() error {
	if store == nil || store.db == nil {
		return nil
	}

	return store.db.Close()
}

func (store *Store) PullRequests(search appconfig.PullRequestSearch) ([]githubcli.PullRequest, bool, error) {
	var payload string
	actualErr := store.db.QueryRow(`
		SELECT payload_json
		FROM pull_request_lists
		WHERE search_key = ?
	`, pullRequestSearchKey(search)).Scan(&payload)
	if errors.Is(actualErr, sql.ErrNoRows) {
		return nil, false, nil
	}
	if actualErr != nil {
		return nil, false, actualErr
	}

	var pullRequests []githubcli.PullRequest
	if actualErr := json.Unmarshal([]byte(payload), &pullRequests); actualErr != nil {
		return nil, false, fmt.Errorf("decode cached pull requests: %w", actualErr)
	}

	return pullRequests, true, nil
}

func (store *Store) SavePullRequests(search appconfig.PullRequestSearch, pullRequests []githubcli.PullRequest) error {
	payloadJSON, actualErr := marshalJSON(pullRequests)
	if actualErr != nil {
		return actualErr
	}

	transaction, actualErr := store.db.Begin()
	if actualErr != nil {
		return actualErr
	}
	defer rollbackOnFailure(transaction)

	if _, actualErr = transaction.Exec(`
		INSERT INTO pull_request_lists (search_key, search_label, command_json, payload_json)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(search_key) DO UPDATE SET
			search_label = excluded.search_label,
			command_json = excluded.command_json,
			payload_json = excluded.payload_json,
			updated_at = CURRENT_TIMESTAMP
	`, pullRequestSearchKey(search), strings.TrimSpace(search.Label), joinedCommandArguments(search.Command), payloadJSON); actualErr != nil {
		return actualErr
	}

	for _, pullRequest := range pullRequests {
		if actualErr = upsertPullRequestSummary(transaction, pullRequest); actualErr != nil {
			return actualErr
		}
	}

	return transaction.Commit()
}

func (store *Store) PullRequestDetail(repository string, number int) (CachedPullRequestDetail, bool, error) {
	var payload sql.NullString
	var sourceUpdatedAt string
	actualErr := store.db.QueryRow(`
		SELECT detail_json, detail_updated_at
		FROM pull_requests
		WHERE repository = ? AND number = ?
	`, strings.TrimSpace(repository), number).Scan(&payload, &sourceUpdatedAt)
	if errors.Is(actualErr, sql.ErrNoRows) {
		return CachedPullRequestDetail{}, false, nil
	}
	if actualErr != nil {
		return CachedPullRequestDetail{}, false, actualErr
	}
	if !payload.Valid || strings.TrimSpace(payload.String) == "" {
		return CachedPullRequestDetail{}, false, nil
	}

	detail, actualErr := unmarshalPullRequestDetail(payload.String)
	if actualErr != nil {
		return CachedPullRequestDetail{}, false, fmt.Errorf("decode cached pull request detail: %w", actualErr)
	}

	return CachedPullRequestDetail{Detail: detail, SourceUpdatedAt: strings.TrimSpace(sourceUpdatedAt)}, true, nil
}

func (store *Store) SavePullRequestDetail(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) error {
	repository, number, actualErr := normalizePullRequestIdentity(summary.Repository.NameWithOwner, summary.Number)
	if actualErr != nil {
		return actualErr
	}

	summaryJSON, actualErr := marshalJSON(summary)
	if actualErr != nil {
		return actualErr
	}
	detailJSON, actualErr := marshalPullRequestDetail(detail)
	if actualErr != nil {
		return actualErr
	}

	_, actualErr = store.db.Exec(`
		INSERT INTO pull_requests (repository, number, summary_json, summary_updated_at, detail_json, detail_updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(repository, number) DO UPDATE SET
			summary_json = excluded.summary_json,
			summary_updated_at = excluded.summary_updated_at,
			detail_json = excluded.detail_json,
			detail_updated_at = excluded.detail_updated_at,
			updated_at = CURRENT_TIMESTAMP
	`, repository, number, summaryJSON, strings.TrimSpace(summary.UpdatedAt), detailJSON, strings.TrimSpace(summary.UpdatedAt))
	return actualErr
}

func (store *Store) PullRequestDiff(repository string, number int) (CachedPullRequestDiff, bool, error) {
	var payload sql.NullString
	var sourceUpdatedAt string
	actualErr := store.db.QueryRow(`
		SELECT diff_json, diff_updated_at
		FROM pull_requests
		WHERE repository = ? AND number = ?
	`, strings.TrimSpace(repository), number).Scan(&payload, &sourceUpdatedAt)
	if errors.Is(actualErr, sql.ErrNoRows) {
		return CachedPullRequestDiff{}, false, nil
	}
	if actualErr != nil {
		return CachedPullRequestDiff{}, false, actualErr
	}
	if !payload.Valid || strings.TrimSpace(payload.String) == "" {
		return CachedPullRequestDiff{}, false, nil
	}

	diff, actualErr := unmarshalPullRequestDiff(payload.String)
	if actualErr != nil {
		return CachedPullRequestDiff{}, false, fmt.Errorf("decode cached pull request diff: %w", actualErr)
	}

	return CachedPullRequestDiff{Diff: diff, SourceUpdatedAt: strings.TrimSpace(sourceUpdatedAt)}, true, nil
}

func (store *Store) SavePullRequestDiff(summary githubcli.PullRequest, diff githubcli.PullRequestDiff) error {
	repository, number, actualErr := normalizePullRequestIdentity(summary.Repository.NameWithOwner, summary.Number)
	if actualErr != nil {
		return actualErr
	}

	summaryJSON, actualErr := marshalJSON(summary)
	if actualErr != nil {
		return actualErr
	}
	diffJSON, actualErr := marshalPullRequestDiff(diff)
	if actualErr != nil {
		return actualErr
	}

	_, actualErr = store.db.Exec(`
		INSERT INTO pull_requests (repository, number, summary_json, summary_updated_at, diff_json, diff_updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(repository, number) DO UPDATE SET
			summary_json = excluded.summary_json,
			summary_updated_at = excluded.summary_updated_at,
			diff_json = excluded.diff_json,
			diff_updated_at = excluded.diff_updated_at,
			updated_at = CURRENT_TIMESTAMP
	`, repository, number, summaryJSON, strings.TrimSpace(summary.UpdatedAt), diffJSON, strings.TrimSpace(summary.UpdatedAt))
	return actualErr
}

func (store *Store) InvalidatePullRequest(repository string, number int) error {
	trimmedRepository, normalizedNumber, actualErr := normalizePullRequestIdentity(repository, number)
	if actualErr != nil {
		return actualErr
	}

	_, actualErr = store.db.Exec(`
		UPDATE pull_requests
		SET detail_json = NULL,
			detail_updated_at = '',
			diff_json = NULL,
			diff_updated_at = '',
			updated_at = CURRENT_TIMESTAMP
		WHERE repository = ? AND number = ?
	`, trimmedRepository, normalizedNumber)
	return actualErr
}

func (store *Store) initialize() error {
	statements := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA busy_timeout = 5000`,
		`CREATE TABLE IF NOT EXISTS pull_request_lists (
			search_key TEXT PRIMARY KEY,
			search_label TEXT NOT NULL,
			command_json TEXT NOT NULL,
			payload_json TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pull_requests (
			repository TEXT NOT NULL,
			number INTEGER NOT NULL,
			summary_json TEXT NOT NULL DEFAULT '',
			summary_updated_at TEXT NOT NULL DEFAULT '',
			detail_json TEXT,
			detail_updated_at TEXT NOT NULL DEFAULT '',
			diff_json TEXT,
			diff_updated_at TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(repository, number)
		)`,
	}

	for _, statement := range statements {
		if _, actualErr := store.db.Exec(statement); actualErr != nil {
			return actualErr
		}
	}

	return nil
}

func upsertPullRequestSummary(transaction *sql.Tx, pullRequest githubcli.PullRequest) error {
	repository, number, actualErr := normalizePullRequestIdentity(pullRequest.Repository.NameWithOwner, pullRequest.Number)
	if actualErr != nil {
		return nil
	}

	summaryJSON, actualErr := marshalJSON(pullRequest)
	if actualErr != nil {
		return actualErr
	}

	_, actualErr = transaction.Exec(`
		INSERT INTO pull_requests (repository, number, summary_json, summary_updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(repository, number) DO UPDATE SET
			summary_json = excluded.summary_json,
			summary_updated_at = excluded.summary_updated_at,
			updated_at = CURRENT_TIMESTAMP
	`, repository, number, summaryJSON, strings.TrimSpace(pullRequest.UpdatedAt))
	return actualErr
}

func marshalJSON(value any) (string, error) {
	encoded, actualErr := json.Marshal(value)
	if actualErr != nil {
		return "", actualErr
	}

	return string(encoded), nil
}

func marshalPullRequestDetail(detail githubcli.PullRequestDetail) (string, error) {
	return marshalJSON(cachedPullRequestDetailPayload{
		Detail:               detail,
		InlineComments:       append([]githubcli.PullRequestInlineComment(nil), detail.InlineComments...),
		InlineCommentThreads: cachedPullRequestReviewThreads(detail.InlineCommentThreads),
	})
}

func unmarshalPullRequestDetail(payload string) (githubcli.PullRequestDetail, error) {
	var cached cachedPullRequestDetailPayload
	if actualErr := json.Unmarshal([]byte(payload), &cached); actualErr != nil {
		return githubcli.PullRequestDetail{}, actualErr
	}

	detail := cached.Detail
	detail.InlineComments = append([]githubcli.PullRequestInlineComment(nil), cached.InlineComments...)
	detail.InlineCommentThreads = githubPullRequestReviewThreads(cached.InlineCommentThreads)
	return detail, nil
}

func marshalPullRequestDiff(diff githubcli.PullRequestDiff) (string, error) {
	return marshalJSON(cachedPullRequestDiffPayload{Diff: diff, Threads: cachedPullRequestReviewThreads(diff.Threads)})
}

func unmarshalPullRequestDiff(payload string) (githubcli.PullRequestDiff, error) {
	var cached cachedPullRequestDiffPayload
	if actualErr := json.Unmarshal([]byte(payload), &cached); actualErr != nil {
		return githubcli.PullRequestDiff{}, actualErr
	}

	diff := cached.Diff
	diff.Threads = githubPullRequestReviewThreads(cached.Threads)
	return diff, nil
}

func cachedPullRequestReviewThreads(threads []githubcli.PullRequestReviewThread) []cachedPullRequestReviewThread {
	if len(threads) == 0 {
		return nil
	}

	cachedThreads := make([]cachedPullRequestReviewThread, 0, len(threads))
	for _, thread := range threads {
		cachedThreads = append(cachedThreads, cachedPullRequestReviewThread{
			ID:                 thread.ID,
			IsResolved:         thread.IsResolved,
			IsOutdated:         thread.IsOutdated,
			ViewerCanResolve:   thread.ViewerCanResolve,
			ViewerCanUnresolve: thread.ViewerCanUnresolve,
			Path:               thread.Path,
			Line:               thread.Line,
			OriginalLine:       thread.OriginalLine,
			StartLine:          thread.StartLine,
			OriginalStartLine:  thread.OriginalStartLine,
			DiffSide:           thread.DiffSide,
			StartDiffSide:      thread.StartDiffSide,
			Comments:           append([]githubcli.PullRequestComment(nil), thread.Comments...),
		})
	}
	return cachedThreads
}

func githubPullRequestReviewThreads(threads []cachedPullRequestReviewThread) []githubcli.PullRequestReviewThread {
	if len(threads) == 0 {
		return nil
	}

	githubThreads := make([]githubcli.PullRequestReviewThread, 0, len(threads))
	for _, thread := range threads {
		githubThreads = append(githubThreads, githubcli.PullRequestReviewThread{
			ID:                 thread.ID,
			IsResolved:         thread.IsResolved,
			IsOutdated:         thread.IsOutdated,
			ViewerCanResolve:   thread.ViewerCanResolve,
			ViewerCanUnresolve: thread.ViewerCanUnresolve,
			Path:               thread.Path,
			Line:               thread.Line,
			OriginalLine:       thread.OriginalLine,
			StartLine:          thread.StartLine,
			OriginalStartLine:  thread.OriginalStartLine,
			DiffSide:           thread.DiffSide,
			StartDiffSide:      thread.StartDiffSide,
			Comments:           append([]githubcli.PullRequestComment(nil), thread.Comments...),
		})
	}
	return githubThreads
}

func rollbackOnFailure(transaction *sql.Tx) {
	_ = transaction.Rollback()
}

func pullRequestSearchKey(search appconfig.PullRequestSearch) string {
	hash := sha256.Sum256([]byte(joinedCommandArguments(search.Command)))
	return hex.EncodeToString(hash[:])
}

func joinedCommandArguments(arguments []string) string {
	trimmedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		trimmedArguments = append(trimmedArguments, strings.TrimSpace(argument))
	}

	return strings.Join(trimmedArguments, "\x00")
}

func normalizePullRequestIdentity(repository string, number int) (string, int, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return "", 0, errors.New("missing pull request identity")
	}

	return trimmedRepository, number, nil
}
