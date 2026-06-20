# Architecture

## Diagram

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

## How It Works

### Single Binary

All services compile into one Go binary. Nothing else needs to be installed — no Node.js, no Python, no nginx, no Postgres. The binary is ~20 MB.

### SQLite

Data is stored in a single SQLite file. WAL mode is enabled for concurrent reads during writes. The database driver is `modernc.org/sqlite` — a pure Go port of SQLite that requires no C compiler (no CGO).

### Authentication

All requests to `/api/*` pass through a PAT (Personal Access Token) middleware. The middleware extracts the `Bearer <pat>` header, looks up the user, and injects them into the request context. WebSocket and MCP endpoints use `?pat=` query parameters instead, because browser APIs don't support custom headers on WebSocket upgrades or EventSource connections.

### Source File Layout

```
internal/
├── store.go          Database setup, migration, helpers
├── users.go          User CRUD store methods
├── projects.go       Project CRUD store methods
├── issues.go         Issue CRUD + filtering + slug generation
├── comments.go       Comment store methods
├── handlers.go       HTTP handlers for all REST endpoints
├── middleware.go     PAT authentication middleware
├── ws.go             WebSocket hub, client, diff helpers
├── mcp.go            MCP server + 16 tool handlers
└── models.go         Data types and static values
cmd/
├── ticketer/main.go  Server entry point, routing
└── tktrctl/          CLI binary (Cobra commands)
```

### Serving All Interfaces on One Port

The Go 1.22 `http.ServeMux` routes by method and path pattern:

- `GET /api/*` — REST API handlers
- `POST /mcp` — MCP streamable HTTP
- `GET /api/ws` — WebSocket upgrade
- `/login`, `/`, `/projects/*`, `/issues/*` — embedded web UI

Everything shares port 8300 by default. No separate processes, no reverse proxy needed.
