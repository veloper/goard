# Reference

## API

Base URL: `http://<host>:<port>/api` — requires `Authorization: Bearer <pat>`.

### Users

```
GET    /api/users                  List users
GET    /api/users/{id}             Get user
POST   /api/users                  Create user (admin only)
PATCH  /api/users/{id}             Update user (admin only)
DELETE /api/users/{id}             Delete user (admin only)
GET    /api/me                     Get current user
```

### Projects

```
POST   /api/projects               Create project
GET    /api/projects               List projects
GET    /api/projects/{id}          Get project (by ID or slug)
PATCH  /api/projects/{id}          Update project
DELETE /api/projects/{id}          Delete project
```

### Issues

```
GET    /api/projects/{id}/issues   List issues (filterable)
POST   /api/projects/{id}/issues   Create issue
GET    /api/issues/{id}            Get issue (by ID or slug)
PATCH  /api/issues/{id}            Update issue fields
PUT    /api/issues/{id}/state      Update issue state only
DELETE /api/issues/{id}            Delete issue
```

Filters: `?state=qa&assignee=<id>&q=search&page=1&per_page=50`

### Comments

```
GET    /api/issues/{id}/comments   List comments
POST   /api/issues/{id}/comments   Add comment
```

### Info

```
GET    /api/info                   Server metadata
```

---

## CLI (tktrctl)

Environment variables: `TICKETER_HOST` and `TICKETER_PAT`.

```
tktrctl info                          Server metadata
tktrctl users list                    List users
tktrctl users show <id>               Get user
tktrctl users create <username>       Create user (--display-name, --admin)
tktrctl users update <id>             Update user (--display-name, --pat)
tktrctl users delete <id>             Delete user
tktrctl projects list                 List projects
tktrctl projects show <id>            Get project
tktrctl projects create <name> <slug> Create project (--description)
tktrctl projects update <id>          Update project (--name, --slug, --description)
tktrctl projects delete <id>          Delete project
tktrctl issues list <project>         List issues (--state, --assignee)
tktrctl issues show <id>              Get issue
tktrctl issues create <project> <title> Create issue (--type, --state, --priority, --assignee)
tktrctl issues update <id>            Update issue (--title, --state, --type, --priority, --assignee)
tktrctl issues state <id>             Show current state
tktrctl issues state-update <id> <state>  Update state
```

---

## MCP Server

**Endpoint:** `POST /mcp?pat=pat_admin` — Streamable HTTP transport.

### Tools

| Tool | Args |
|------|------|
| `get_info` | _(none)_ |
| `list_users` | _(none)_ |
| `get_user` | `id` (number) |
| `list_projects` | _(none)_ |
| `get_project` | `id` (string) |
| `create_project` | `name`, `slug` |
| `update_project` | `id`, optional fields |
| `delete_project` | `id` |
| `list_issues` | `project_id`, optional `state`, `assignee` |
| `get_issue` | `id` (string) |
| `create_issue` | `project_id`, `title` |
| `update_issue` | `id`, optional fields |
| `update_issue_state` | `id`, `state` |
| `delete_issue` | `id` |
| `list_comments` | `issue_id` |
| `add_comment` | `issue_id`, `body` |

### Client Config

```json
{
  "mcpServers": {
    "ticketer": {
      "type": "http",
      "url": "http://localhost:8300/mcp?pat=pat_admin"
    }
  }
}
```

---

## WebSocket

**Endpoint:** `ws://<host>:<port>/api/ws?pat=pat_admin`

Events are JSON with `type` and `payload`. Self-events are suppressed.

| Type | Payload |
|------|---------|
| `project_created` | Full project |
| `project_updated` | `{id, changed}` |
| `project_deleted` | `{id}` |
| `issue_created` | Full issue |
| `issue_updated` | `{id, changed}` |
| `issue_deleted` | `{id, project_id}` |
| `comment_created` | Full comment |

Update events use `changed: {field: {before, after}}` — only fields that differ.

---

## Data Model

```
User ──┬── Project   (created_by)
       ├── Issue     (created_by + assignee)
       └── Comment   (author + created_by)
```

**States:** `backlog → todo → in_progress → qa → done → cancelled`

**Types:** `epic`, `feature`, `bug`, `chore`

**Priority:** `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low

**Slugs:** Issues get `<project-slug>-<auto-increment-id>` (e.g. `ASTEROID-GAME-42`).

---

## Docker

```bash
# Pull and run
docker pull veloper/ticketer
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  veloper/ticketer

# Build from source
docker build -t ticketer .
```

### Docker Compose

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports:
      - "8300:8300"
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

volumes:
  ticketer-data:
```

### Running tktrctl in Compose

```bash
docker compose exec ticketer tktrctl projects create "Game" GAME
```

Or as a setup service:

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports: ["8300:8300"]
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

  setup:
    image: veloper/ticketer
    profiles: ["setup"]
    environment:
      TICKETER_HOST: http://ticketer:8300
      TICKETER_PAT: pat_admin
    depends_on:
      ticketer:
        condition: service_started
    command: >
      tktrctl projects create "Asteroid Game" ASTEROID-GAME &&
      tktrctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
```

Run with: `docker compose --profile setup run setup`

---

## Architecture

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ REST API    │  │ Web UI      │  │ MCP Server  │  │ WebSocket   │
│ /api/*      │  │ /           │  │ /mcp        │  │ /api/ws     │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │                │
       └────────────────┴────────────────┴────────────────┘
                              net/http
       ┌─────────────────────────────────────────────────────┐
       │                  SQLite (WAL mode)                  │
       │             modernc.org/sqlite (no CGO)             │
       └─────────────────────────────────────────────────────┘
```

Single binary, zero runtime dependencies. All services on one port.
