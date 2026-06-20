# Ticketer

A minimal, API-first project/issue tracker for AI agent teams. Built in Go, backed by SQLite, with an embedded kanban web UI.

Agents (or humans) manage issues through a REST API. Users authenticate with pre-shared PATs. No signup, no email, no complexity.

## Quickstart

```bash
# Build and run
go build -o ticketer ./cmd/ticketer
./ticketer

# Or via Docker
docker build -t ticketer .
docker run -p 8080:8080 ticketer
```

Open http://localhost:8080 for the web UI. Default users are seeded on first run:

| Username | PAT | Role |
|----------|-----|------|
| `alex_planner` | `pat_alex` | Team lead |
| `sam_builder`  | `pat_sam`  | Engineer |
| `tommy_tester` | `pat_tommy` | QA |

## API

All API requests require `Authorization: Bearer <pat>`.

### Projects

```
POST   /api/projects              Create project
GET    /api/projects              List projects
GET    /api/projects/:id          Get project
PATCH  /api/projects/:id          Update project
DELETE /api/projects/:id          Delete project
```

### Issues

```
GET    /api/projects/:id/issues   List issues (filterable)
POST   /api/projects/:id/issues   Create issue
GET    /api/issues/:id            Get issue
PATCH  /api/issues/:id            Update issue (state, assignee, etc.)
DELETE /api/issues/:id            Delete issue
```

**Query params for listing issues:** `?state=review&assignee=<id>&q=search&page=1&per_page=50`

### Comments

```
GET    /api/issues/:id/comments   List comments
POST   /api/issues/:id/comments   Add comment
```

### Users

```
GET    /api/users                 List users
GET    /api/users/:id             Get user
GET    /api/me                    Get current user (from PAT)
```

## Data Model

```
User ──┬── Project (created_by)
        ├── Issue (created_by + assignee)
        └── Comment (created_by + author)
```

**States** — system-wide, static:
`backlog` → `todo` → `in_progress` → `review` → `done` → `cancelled`

**Priority** — `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low

Issues get auto-generated sequence IDs like `ASTEROID-GAME-42` based on the project name.

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `TICKETER_ADDR` | `:8080` | Listen address |
| `TICKETER_DB_PATH` | `ticketer.db` | SQLite database file path |
| `TICKETER_DEFAULT_PAT` | `""` | PAT injected into web UI for API calls |
| `TICKETER_USERS` | (embedded defaults) | JSON array of seed users |

Custom seed users:
```json
{
  "users": [
    {"username": "alex_planner", "display_name": "Alex Planner", "pat": "pat_alex"},
    {"username": "sam_builder",  "display_name": "Sam Builder",  "pat": "pat_sam"},
    {"username": "tommy_tester", "display_name": "Tommy Tester", "pat": "pat_tommy"}
  ]
}
```

Pass as `TICKETER_USERS` env var or in a `users.json` file next to the binary.

## Web UI

The kanban board is served at `/`. No auth required — the server injects a default PAT into the page so the browser can call the API transparently.

Views:
- **`/`** — projects list
- **`/projects/:id`** — kanban board grouped by state
- **`/issues/:id`** — issue detail with comments

## Architecture

Single Go binary, zero runtime deps. SQLite with WAL mode for concurrency. Pure Go SQLite driver (`modernc.org/sqlite`) — no CGO needed.

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ REST API    │  │ Web UI      │  │ SQLite      │
│ :8080/api/* │  │ :8080/      │  │ ticketer.db │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┴────────────────┘
                   net/http
```

## Development

```bash
# Run tests
go test ./...

# Build for current platform
go build -o ticketer ./cmd/ticketer

# Build Docker image
docker build -t ticketer .
```
