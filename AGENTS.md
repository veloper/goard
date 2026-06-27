# Goard — Agent Guide

This document is for AI agents interacting with Goard. It covers everything you need to know: how the project works, how to configure it, and how to control it programmatically.

## Quick Overview

Goard is a project/issue tracker built as a single Go binary. It stores everything in SQLite and exposes four ways to interact:

- **REST API** — standard CRUD for projects, issues, comments, users
- **MCP Server** — a Model Context Protocol endpoint for LLM tool use
- **WebSocket** — real-time change broadcasts
- **CLI (`goardctl`)** — scripting and bootstrapping

All IDs are auto-increment integers. Issues have human-readable slugs like `ASTEROID-GAME-42`.

## Configuration

Required env vars on startup:

| Env var | Description |
|---------|-------------|
| `GOARD_ADMIN_USERNAME` | Admin username |
| `GOARD_ADMIN_PAT` | Admin personal access token |

Optional:

| Env var | Default | Description |
|---------|---------|-------------|
| `GOARD_HOST` | `""` | Listen host (all interfaces) |
| `GOARD_PORT` | `"8300"` | Listen port |
| `GOARD_DB_PATH` | `"goard.db"` | SQLite database path |

The admin user is created on startup. Use those credentials to create additional users via the API or CLI.

## Data Model

```
User
  ├── Project  (created_by)
  ├── Issue    (created_by + assignee)
  └── Comment  (author + created_by)
```

### States (fixed pipeline)

`backlog` → `in_progress` → `review` → `done` → `cancelled`

### Types

`epic`, `feature`, `bug`, `chore`

### Priority

| Level | Label |
|-------|-------|
| 0 | none |
| 1 | urgent |
| 2 | high |
| 3 | medium |
| 4 | low |

### Slugs

Issues get auto-generated slugs: `<project-slug>-<auto-increment-id>`.
Example: `ASTEROID-GAME-42`. Use slugs to reference issues in place of numeric IDs.

## Authentication

All API and MCP requests require a Personal Access Token (PAT). The PAT is set at user creation and cannot be changed except by an admin.

- **REST API**: `Authorization: Bearer <pat>` header
- **MCP Server**: `?pat=<pat>` query parameter on the endpoint URL
- **Web UI**: Login form at `/login`
- **CLI**: `GOARD_PAT` environment variable

The admin user is configured via `GOARD_ADMIN_USERNAME` / `GOARD_ADMIN_PAT`. Additional users can be created via `POST /api/users` (admin only).

## REST API

Base URL: `http://<host>:<port>/api`

### Users

```
GET    /api/users            List all users
GET    /api/users/{id}       Get user by ID
POST   /api/users            Create user (admin only)
PATCH  /api/users/{id}       Update user (admin only)
DELETE /api/users/{id}       Delete user (admin only)
GET    /api/me               Get current user
```

Create user example:
```json
POST /api/users
{"username": "bot", "admin": false}
→ {"meta": {"status": 201}, "data": {"user": {"id": 5, "username": "bot", "is_admin": false}, "pat": "pat_a1b2c3d4..."}}
```

The PAT is generated server-side and returned only on creation.

### Projects

```
POST   /api/projects             Create project
GET    /api/projects             List projects
GET    /api/projects/{id}        Get project by ID or slug
PATCH  /api/projects/{id}        Update project
DELETE /api/projects/{id}        Delete project
```

Create project example:
```json
POST /api/projects
{"name": "Asteroid Game", "slug": "ASTEROID-GAME"}
→ {"id": 1, "name": "Asteroid Game", "slug": "ASTEROID-GAME", ...}
```

### Issues

```
GET    /api/projects/{id}/issues  List issues (filterable)
POST   /api/projects/{id}/issues  Create issue
GET    /api/issues/{id}           Get issue (by ID or slug)
PATCH  /api/issues/{id}           Update issue fields
PUT    /api/issues/{id}/state     Update issue state only
DELETE /api/issues/{id}           Delete issue
```

Filters: `?state=review&assignee=<id>&q=search&page=1&per_page=50`

Create issue example:
```json
POST /api/projects/1/issues
{"title": "Add rotation", "type": "feature", "priority": 2}
→ {"id": 1, "slug": "ASTEROID-GAME-1", "state": "backlog", ...}
```

Update state example:
```json
PUT /api/issues/1/state
{"state": "review"}
→ {"id": 1, "slug": "ASTEROID-GAME-1", "state": "review", ...}
```

### Comments

```
GET    /api/issues/{id}/comments  List comments on an issue
POST   /api/issues/{id}/comments  Add comment
```

### Info

```
GET    /api/info
```

Returns server metadata: valid states, types, priority levels with labels, all users, all projects, and the authenticated user (`me`). Call this first to understand what values the API accepts.

## MCP Server

An MCP (Model Context Protocol) server is available for LLM-driven management.

```
POST http://<host>:<port>/mcp?pat=pat_admin
Transport: Streamable HTTP
```

### Tools

| Tool | Description | Key Args |
|------|-------------|----------|
| `get_info` | Discover server surface | _(none)_ |
| `list_users` | List all users | _(none)_ |
| `get_user` | Get user by ID | `id` (number) |
| `list_projects` | List all projects | _(none)_ |
| `get_project` | Get project by ID or slug | `id` (string) |
| `create_project` | Create a project | `name`, `slug` |
| `update_project` | Update a project | `id`, optional fields |
| `delete_project` | Delete a project | `id` |
| `list_issues` | List issues in a project | `project_id`, optional `state`, `assignee` |
| `get_issue` | Get issue by ID or slug | `id` (string) |
| `create_issue` | Create an issue | `project_id`, `title` |
| `update_issue` | Update issue fields | `id`, optional fields |
| `update_issue_state` | Move issue to new state | `id`, `state` |
| `delete_issue` | Delete an issue | `id` |
| `list_comments` | List comments on an issue | `issue_id` |
| `add_comment` | Add a comment | `issue_id`, `body` |

All tools accept `project_id` and `id` as either numeric IDs or slugs (e.g. `ASTEROID-GAME-42`).

### Client Configuration

```json
{
  "mcpServers": {
    "goard": {
      "type": "http",
      "url": "http://localhost:8300/mcp?pat=pat_admin"
    }
  }
}
```

## CLI (goardctl)

The `goardctl` binary is included in the Docker image. Configure via environment variables:

```bash
export GOARD_HOST=http://localhost:8300
export GOARD_PAT=pat_admin
```

### Commands

```
goardctl info                          # Server metadata
goardctl users list                    # List users
goardctl users show <id>               # Get user
goardctl users create <username>       # Create user (--admin)
goardctl users update <id>             # Update user (--pat)
goardctl users delete <id>             # Delete user
goardctl projects list                 # List projects
goardctl projects show <id>            # Get project
goardctl projects create <name> <slug> # Create project (--description)
goardctl projects update <id>          # Update project (--name, --slug, --description)
goardctl projects delete <id>          # Delete project
goardctl issues list <project>         # List issues (--state, --assignee)
goardctl issues show <id>              # Get issue
goardctl issues create <project> <title> # Create issue (--type, --state, --priority, --assignee)
goardctl issues update <id>            # Update issue (--title, --state, --type, --priority, --assignee)
goardctl issues state <id>             # Show current state
goardctl issues state-update <id> <state>  # Update state
```

Projects and issues can be referenced by numeric ID or slug.

## WebSocket

A WebSocket endpoint is available for real-time updates:

```
ws://<host>:<port>/api/ws?pat=pat_admin
```

The server broadcasts JSON events when data changes. Events are suppressed for the user who caused the change.

### Event Types

| Type | Description |
|------|-------------|
| `project_created` / `project_updated` / `project_deleted` | Project changes |
| `issue_created` / `issue_updated` / `issue_deleted` | Issue changes |
| `comment_created` | New comment |

### Update Event Format

```json
{
  "type": "issue_updated",
  "payload": {
    "id": 1,
    "changed": {
      "state": {"before": "backlog", "after": "review"},
      "assignee": {"before": null, "after": 2}
    }
  }
}
```

Only changed fields are included in update events. Create events include the full entity. Delete events include just the ID.

## Docker

```bash
# Build
docker build -t goard .

# Run
docker run -p 8300:8300 \
  -e GOARD_ADMIN_USERNAME=admin \
  -e GOARD_ADMIN_PAT=pat_admin \
  goard
```

The Docker image includes both `goard` and `goardctl` binaries. The database defaults to `/data/goard.db`.

### Docker Compose

```yaml
services:
  goard:
    build: .
    ports:
      - "8300:8300"
    environment:
      GOARD_ADMIN_USERNAME: admin
      GOARD_ADMIN_PAT: pat_admin
    volumes:
      - goard-data:/data

volumes:
  goard-data:
```

### Running goardctl in Compose

```bash
# One-off commands
docker compose exec goard goardctl projects create "Game" GAME

# Interactive setup service
docker compose --profile setup run setup
```

## Common Workflows

### Starting fresh

1. Set `GOARD_ADMIN_USERNAME` and `GOARD_ADMIN_PAT`
2. Start the server
3. Call `GET /api/info` to verify connectivity and see server state
4. Create users via `POST /api/users` (as admin)
5. Create projects via `POST /api/projects`
6. Create issues via `POST /api/projects/{id}/issues`

### Moving an issue through the pipeline

```
POST /api/projects/1/issues   → {"slug": "GAME-1", "state": "backlog"}
PUT  /api/issues/GAME-1/state → {"state": "in_progress"}
PUT  /api/issues/GAME-1/state → {"state": "review"}
PUT  /api/issues/GAME-1/state → {"state": "done"}
```

### Finding an issue by slug

Issues with slug `ASTEROID-GAME-42` can be accessed via:

- `GET /api/issues/ASTEROID-GAME-42` (slug)
- `GET /api/issues/42` (numeric ID)
- `PUT /api/issues/ASTEROID-GAME-42/state` (state endpoint with slug)
- `goardctl issues show ASTEROID-GAME-42`

### Quick reference for valid values

```
States:    backlog, in_progress, review, done, cancelled
Types:     epic, feature, bug, chore
Priority:  0=none, 1=urgent, 2=high, 3=medium, 4=low
```

Call `GET /api/info` to get these dynamically along with all registered users and projects.
