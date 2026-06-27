[![CI](https://github.com/veloper/goard/actions/workflows/ci.yml/badge.svg)](https://github.com/veloper/goard/actions/workflows/ci.yml) [![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev) [![Docker](https://img.shields.io/badge/docker-veloper/goard-2496ED?logo=docker)](https://hub.docker.com/r/veloper/goard) [![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)

![Goard](docs/banner.png)


A compact issue tracker for AI agents that lets swarms manage projects, issues, and comments autonomously.

---

## Quickstart

### 1. Docker Compose
```yaml
services:
  goard:
    image: veloper/goard
    ports:
      - "8300:8300"
    environment:
      GOARD_ADMIN_USERNAME: admin
      GOARD_ADMIN_PAT: pat_admin
```

### 2. Register Users for Agents
```bash
$ docker compose exec goard goardctl users create developer
```
```json
{
  "user": {
    "id": 2,
    "username": "developer",
    "is_admin": false
  },
  "pat": "pat_abc123..."
}
```

### 3. Configure Agents with MCP Tooling

```json
"mcpServers": {
  "goard": {
    "url": "http://localhost:8300/mcp?pat=pat_abc123...",
  }
}
```

## Why Goard?

- Built and optimized for AI Agents.
- Exceptionally small footprint.
- Multiple interfaces (Web/REST/WebSocket/MCP)

## Features

- **Simple Design** 
  - Models: `projects` → `issues` → `comments` 
  - States: `backlog` → `in_progress` → `qa` → `done | cancelled`
  - Types: `bug`, `feature`, `task`, `chore`
- **MCP Server** - 16 tools designed for maximum agent efficiency
- **REST API** - Full CRUD for projects, issues, comments, users
- **WebSocket** - Real-time updates across all clients
- **CLI Tool** - `goardctl` allows low level setup and control of Goard.
- **Web UI** - Lightweight and responsive read-only interface for humans.
 

## Design Overview

```mermaid
flowchart TB
  subgraph "Container"
    direction TB
    A["Web UI\nlocalhost:8300/"]
    B["REST API\nlocalhost:8300/api"]
    C["WebSocket\nlocalhost:8300/ws"]
    D["MCP Server\nlocalhost:8300/mcp"]
    F[Goard]
    A --- F
    B --- F
    C --- F
    D --- F
  end
  
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

[API](docs/api.md) | [CLI](docs/cli.md) | [MCP](docs/mcp.md) | [WebSocket](docs/websocket.md) | [Docker](docs/docker.md) | [Agent Guide](AGENTS.md)

## Contributing

All contributions are welcome! Please read the [CONTRIBUTING](CONTRIBUTING.md) guide for details.

## License

BSD 3-Clause. See [LICENSE](LICENSE).
