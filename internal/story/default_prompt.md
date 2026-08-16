Group the changes into a logical, reviewer-friendly story.
Use a professional tone.
Prefer chapters that reflect one cohesive behavior change, refactor step, or debugging thread. For each chapter, explain what changed, why it changed.
Write in readable markdown with short paragraphs and light structure. Use lists, emphasis, links, and code fences only when they make the review clearer. Prefer whitespaces over dense prose. Keep it concise, concrete, and useful.
When a visual makes the point clearer, include lightweight cues. Use your judgment and don't overwhelm the reader.

- Show logic or an algorithm as pseudocode:

```text
on(save)
  if content is unchanged
    return cached result
  write new content
  return fresh result
```

- Show runtime control flow as a call tree:

```text
submitForm
  createSession
    persistPrompt
    launchAgent
  navigateToSession
```

- Show file responsibility or a broad refactor as a shallow file tree:

```text
src/
├── commands/       # parses user actions
├── sessions/       # owns session state
└── transport/      # sends API requests
```

- Show component interaction, control flow, or data flow with ASCII diagrams:

```text
+---------+     +-----+     +--------+
| Browser | --> | API | --> | Worker |
+---------+     +-----+     +--------+
                   |
                   v
              +----------+
              | Postgres |
              +----------+
```
