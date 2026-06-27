[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/veloper/goard/actions/workflows/ci.yml/badge.svg)](https://github.com/veloper/goard/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/docker-veloper/goard-2496ED?logo=docker)](https://hub.docker.com/r/veloper/goard)

# Goard

Project tracker with built-in MCP server, REST API, and web UI — one binary, embedded SQLite.

[Website](https://github.com/veloper/goard) •
[Docs](docs/) •
[Agent Guide](AGENTS.md)

## Quickstart

```bash
docker run -p 8300:8300 \
  -e GOARD_ADMIN_USERNAME=admin \
  -e GOARD_ADMIN_PAT=pat_admin \
  veloper/goard
```

Open http://localhost:8300, sign in with `admin` / `pat_admin`, create a project.

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

### CLI

The Docker image includes `goardctl` for scripting:

```bash
docker compose exec goard goardctl projects create "My Project" MY-PROJECT
docker compose exec goard goardctl issues create MY-PROJECT "Fix login" --type bug --priority 1
```

## Features

- **Project view** — issues grouped by state (backlog → done) with priority colors and assignee info
- **MCP server** — every operation accessible to LLMs out of the box
- **REST API** — full CRUD for projects, issues, comments, users
- **WebSocket** — real-time updates across all clients
- **goardctl CLI** — scripting and automation via Docker exec
- **Filter DSL** — nested AND/OR queries with 10 operators
- **Slug references** — issues have human-readable IDs like `ASTEROID-GAME-42`

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

## License

BSD 3-Clause. See [LICENSE](LICENSE).
