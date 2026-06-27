# Data Model

## Entity Relationships

```
User
  ├── Project   (created_by)       — project.name, project.slug
  ├── Issue     (created_by, assignee) — in a project
  └── Comment   (author, created_by) — on an issue
```

All IDs are SQLite auto-increment integers.

## States (fixed pipeline)

Issues move through a fixed state machine. No custom states:

```
backlog → in_progress → review → done → cancelled
```

- **backlog** — idea, not yet scheduled
- **in_progress** — actively being worked
- **review** — ready for testing/review
- **done** — completed
- **cancelled** — won't do (terminal state)

## Types

| Type | Description |
|------|-------------|
| `improvement` | Refactoring, performance, tech debt |
| `feature` | New capability |
| `bug` | Something broken |
| `chore` | Maintenance, non-user-facing work |

## Priority

| Level | Label | Meaning |
|-------|-------|---------|
| 0 | none | Not prioritized |
| 1 | urgent | Must fix immediately |
| 2 | high | Should be addressed soon |
| 3 | medium | Default priority |
| 4 | low | Nice to have |

## Slugs

Issues get auto-generated human-readable identifiers:

```
<project-slug>-<auto-increment-id>
```

Examples: `ASTEROID-GAME-1`, `LOGIN-PAGE-42`

The project slug is set at project creation. The issue number is its auto-increment primary key (global across all projects). Slugs can be used anywhere numeric IDs are accepted — API paths, CLI commands, MCP tools, WebSocket references.

## Example: Issue object

```json
{
  "id": 1,
  "project_id": 1,
  "slug": "ASTEROID-GAME-1",
  "title": "Add ship rotation",
  "description": "Left/right arrows rotate the ship",
  "type": "feature",
  "state": "backlog",
  "assignee": 0,
  "priority": 2,
  "parent_id": 0,
  "created_by": 1,
  "created_at": "2026-06-20T12:00:00Z",
  "updated_at": "2026-06-20T12:00:00Z"
}
```
