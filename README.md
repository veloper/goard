[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/veloper/goard/actions/workflows/ci.yml/badge.svg)](https://github.com/veloper/goard/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/docker-veloper/goard-2496ED?logo=docker)](https://hub.docker.com/r/veloper/goard)

# Goard

![Goard](docs/banner.png)

## Quickstart

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
open http://localhost:8300

# Or install the CLI
go install github.com/veloper/goard/cmd/goardctl@latest
```

## How agents use it

| Scenario | Flow |
|---|---|
| **Sprint planning** | Agent calls `create_issue` for each task via MCP, sets priority and assignee |
| **Bug triage** | Agent calls `list_issues?filter=...`, finds open bugs, updates state |
| **CI/CD hook** | Pipeline calls REST API to file issues on build failure |
| **Human review** | Team checks the web dashboard, moves cards, leaves comments |

## What's inside

| Interface | For |
|---|---|
| **MCP server** | LLMs manage projects directly — 16 tools |
| **REST API** | Code, CI/CD, custom integrations |
| **WebSocket** | Real-time event stream |
| **Web UI** | Read-only dashboard for human oversight |
| **CLI** | Scripting and automation |

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
