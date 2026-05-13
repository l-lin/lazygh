// Package githubcli adapts the `gh` CLI behind focused services and domain-facing
// capability adapters. Keep shell transport, command formatting, REST or
// GraphQL helpers, and GitHub API quirks here; product-facing models live in
// `internal/github`, and this package must not depend on `internal/tui` or
// `internal/cache`.
package githubcli
