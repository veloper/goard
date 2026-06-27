[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![CI](https://github.com/veloper/goard/actions/workflows/ci.yml/badge.svg)](https://github.com/veloper/goard/actions/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/docker-veloper/goard-2496ED?logo=docker)](https://hub.docker.com/r/veloper/goard)

# Goard

![Goard](docs/banner.png)

Project tracker for teams that include AI agents.  
MCP server + REST API + read-only web UI + CLI. One binary, embedded SQLite.

```bash
docker run -p 8300:8300 -e GOARD_ADMIN_USERNAME=admin -e GOARD_ADMIN_PAT=pat_admin veloper/goard
```

---

## What makes it different

**MCP server, not just a REST API.**  
Every tool, every operation, available to any LLM out of the box. No custom function calling, no glue code.

**Your infrastructure is one binary.**  
No Postgres. No Redis. No container orchestration. One `docker run` and you're done. The database is a file.

**Self-hosted by design.**  
Your issue data doesn't train anyone else's model. It stays on your infrastructure, in a SQLite file.

**Built for automation first.**  
The web UI is read-only. Everything — create, update, filter, sort, paginate — is designed for programmatic access.

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
# Open http://localhost:8300  — sign in with admin / pat_admin

# Or install the CLI
go install github.com/veloper/goard/cmd/goardctl@latest
```

## Docs

| | |
|---|---|
| **API** | [`docs/api.md`](docs/api.md) |
| **CLI** | [`docs/cli.md`](docs/cli.md) |
| **MCP** | [`docs/mcp.md`](docs/mcp.md) |
| **WebSocket** | [`docs/websocket.md`](docs/websocket.md) |
| **Docker** | [`docs/docker.md`](docs/docker.md) |
| **Agent Guide** | [`AGENTS.md`](AGENTS.md) |
