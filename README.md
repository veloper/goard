[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/veloper/goard/actions/workflows/ci.yml/badge.svg)](https://github.com/veloper/goard/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/docker-veloper/goard-2496ED?logo=docker)](https://hub.docker.com/r/veloper/goard)

# Goard

Kanban for engineering teams that include AI agents.

```bash
docker run -p 8300:8300 \
  -e GOARD_ADMIN_USERNAME=admin \
  -e GOARD_ADMIN_PAT=pat_admin \
  veloper/goard
```

Open http://localhost:8300, sign in, create a project, start tracking work.

---

## Features

- **Kanban board** — six columns (backlog → done), priority-colored cards, assignee avatars
- **MCP server** — every operation available to any LLM out of the box. No glue code.
- **One binary** — Go + SQLite, no Postgres, no Redis, no container orchestration. Runs anywhere.
- **Self-hosted** — your data stays on your infrastructure. Doesn't train anyone else's model.

## How it works

Create a project and get a kanban board. Each project has its own board with columns for backlog, todo, in progress, qa, done, cancelled. Issues have states, priorities (urgent → low), types (epic, feature, bug, chore), and assignees.

The web UI works for humans. The MCP server and REST API work for agents. Both hit the same data — create an issue via MCP and it appears on the board in real time.

## Quickstart

```bash
docker run -p 8300:8300 \
  -e GOARD_ADMIN_USERNAME=admin \
  -e GOARD_ADMIN_PAT=pat_admin \
  veloper/goard
```

Open http://localhost:8300, sign in with `admin` / `pat_admin`.

### Docker Compose

```yaml
services:
  goard:
    image: veloper/goard
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

```bash
docker compose up -d
```

### CLI (via Docker)

```bash
# Create a project
docker compose exec goard goardctl projects create "My Project" MY-PROJECT

# Create an issue
docker compose exec goard goardctl issues create MY-PROJECT "Fix login" --type bug --priority 1

# List everything
docker compose exec goard goardctl projects list
docker compose exec goard goardctl issues list MY-PROJECT
```

## Configuration

| Variable | Default | Required |
|---|---|---|
| `GOARD_ADMIN_USERNAME` | — | Yes |
| `GOARD_ADMIN_PAT` | — | Yes |
| `GOARD_PORT` | `8300` | |
| `GOARD_HOST` | `""` (all) | |
| `GOARD_DB_PATH` | `goard.db` | |

## Docs

| | |
|---|---|
| **API** | [`docs/api.md`](docs/api.md) |
| **CLI** | [`docs/cli.md`](docs/cli.md) |
| **MCP** | [`docs/mcp.md`](docs/mcp.md) |
| **WebSocket** | [`docs/websocket.md`](docs/websocket.md) |
| **Docker** | [`docs/docker.md`](docs/docker.md) |
| **Agent Guide** | [`AGENTS.md`](AGENTS.md) |
