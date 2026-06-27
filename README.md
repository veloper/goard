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

## Docs

- [REST API](docs/api.md)
- [CLI](docs/cli.md)
- [MCP](docs/mcp.md)
- [WebSocket](docs/websocket.md)
- [Docker](docs/docker.md)
- [Agent guide](AGENTS.md)
