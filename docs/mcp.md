# MCP Server

A Model Context Protocol server for LLM-driven project management. Uses Streamable HTTP transport.

**Endpoint:** `POST /mcp?pat=pat_admin`

The PAT is passed as a query parameter because some MCP clients cannot set HTTP headers. The endpoint is on the same port as the REST API and Web UI.

## Client Configuration

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

## Tools

### Discovery

| Tool | Description | Args |
|------|-------------|------|
| `get_info` | Server metadata: valid states/types/priorities, users, projects, authenticated user | _(none)_ |

Start here. The response tells you what values the other tools accept.

### Users

| Tool | Description | Args |
|------|-------------|------|
| `list_users` | List all registered users | _(none)_ |
| `get_user` | Get a single user by ID | `id` (number) |

User IDs are needed for assigning issues.

### Projects

| Tool | Description | Args |
|------|-------------|------|
| `list_projects` | List all projects | _(none)_ |
| `get_project` | Get a project by ID or slug | `id` (string) |
| `create_project` | Create a new project | `name`, `slug` (required), `description` (optional) |
| `update_project` | Update project fields | `id` (required), `name`/`slug`/`description` (optional) |
| `delete_project` | Delete a project permanently | `id` (required) |

### Issues

| Tool | Description | Args |
|------|-------------|------|
| `list_issues` | List issues in a project, filterable by state or assignee | `project_id` (required), `state`/`assignee` (optional) |
| `get_issue` | Get an issue by ID or slug | `id` (string) |
| `create_issue` | Create an issue | `project_id`, `title` (required), `description`/`type`/`state`/`priority` (optional) |
| `update_issue` | Update issue fields | `id` (required), optional fields |
| `update_issue_state` | Move issue through the state pipeline | `id` (required), `state` (required) |
| `delete_issue` | Delete an issue permanently | `id` (required) |

### Comments

| Tool | Description | Args |
|------|-------------|------|
| `list_comments` | List comments on an issue, oldest first | `issue_id` (required) |
| `add_comment` | Add a comment to an issue | `issue_id`, `body` (required) |

## Notes

- All `id` and `project_id` arguments accept numeric IDs or slugs (e.g. `ASTEROID-GAME-42`)
- Tool errors are returned in the MCP result with `isError: true`, not as protocol-level errors
- The server validates states against the pipeline before accepting state changes
