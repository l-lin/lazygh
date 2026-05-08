# Notification fixtures

These fixtures were captured from real GitHub notification and REST payloads on `2026-05-08`, then sanitized:

- private repository names, URLs, titles, ids, and owners were replaced with placeholders
- unused fields were removed to keep only the shapes this package depends on
- issue and release detail fixtures preserve the real REST field names returned by `gh api`

Capture commands used:

```sh
gh api '/notifications?all=true&per_page=100' --paginate --slurp
gh api 'repos/<owner>/<repo>/issues/<number>'
gh api 'repos/<owner>/<repo>/releases/<id>'
```
