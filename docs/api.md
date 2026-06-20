# REST API

Base URL: `http://<host>:<port>/api`

All requests require `Authorization: Bearer <pat>` header. PATs are generated server-side on user creation and are returned in the create response only.

## Info

```
GET /api/info
```

Returns server metadata. Call this first to discover valid values and see all registered users and projects. Response includes the authenticated user in the `me` field.

```json
{
  "states": ["backlog", "todo", "in_progress", "qa", "done", "cancelled"],
  "types": ["epic", "feature", "bug", "chore"],
  "priority_levels": [0, 1, 2, 3, 4],
  "priority_labels": {"0": "none", "1": "urgent", "2": "high", "3": "medium", "4": "low"},
  "users": [{"id": 1, "username": "admin", ...}],
  "projects": [{"id": 1, "name": "Game", "slug": "GAME", ...}],
  "me": {"id": 1, "username": "admin", "is_admin": true}
}
```

## Users

### List users

```
GET /api/users
→ 200 [{"id": 1, "username": "admin", "display_name": "admin", "is_admin": true, ...}]
```

### Get user

```
GET /api/users/{id}
→ 200 {"id": 1, "username": "admin", ...}
→ 404 {"error": "user not found"}
```

### Create user (admin only)

```
POST /api/users
{"username": "bot", "display_name": "Bot Builder", "admin": false}
→ 201 {"id": 5, "username": "bot", "display_name": "Bot Builder", "pat": "pat_a1b2c3d4...", "is_admin": false}
```

The PAT is auto-generated and **only returned on creation**. Store it. `admin: true` makes the user an admin (admin required to create admins).

### Update user (admin only)

```
PATCH /api/users/{id}
{"display_name": "New Name", "pat": "pat_new"}
→ 200 {"id": 5, "username": "bot", "display_name": "New Name", ...}
```

### Delete user (admin only)

```
DELETE /api/users/{id}
→ 204
```

### Me

```
GET /api/me
→ 200 {"id": 1, "username": "admin", "is_admin": true, ...}
```

Returns the authenticated user based on the PAT in the Authorization header.

## Projects

### Create project

```
POST /api/projects
{"name": "Asteroid Game", "slug": "ASTEROID-GAME"}
→ 201 {"id": 1, "name": "Asteroid Game", "slug": "ASTEROID-GAME", "description": "", ...}
```

`slug` is used in issue slugs (`ASTEROID-GAME-42`). Must be unique. `description` is optional.

### List projects

```
GET /api/projects
→ 200 [{"id": 1, "name": "Asteroid Game", "slug": "ASTEROID-GAME", ...}]
```

### Get project

```
GET /api/projects/{id}
→ 200 {"id": 1, ...}
→ 404 {"error": "project not found"}
```

### Update project

```
PATCH /api/projects/{id}
{"name": "Asteroid Game v2", "description": "A space game"}
→ 200 {"id": 1, "name": "Asteroid Game v2", "slug": "ASTEROID-GAME", "description": "A space game", ...}
```

Only provided fields change. Empty strings are ignored.

### Delete project

```
DELETE /api/projects/{id}
→ 204
```

Deletes the project and all its issues permanently.

## Issues

### Create issue

```
POST /api/projects/{project_id}/issues
{"title": "Add rotation", "type": "feature", "priority": 2}
→ 201 {"id": 1, "slug": "ASTEROID-GAME-1", "state": "todo", "title": "Add rotation", "type": "feature", "priority": 2, ...}
```

`project_id` can be a numeric ID or a slug. Only `title` is required. Defaults: `state=todo`, `type=feature`, `priority=3`.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | **Required.** Issue title |
| `description` | string | Issue description |
| `type` | string | `epic`, `feature`, `bug`, or `chore` |
| `state` | string | `backlog`, `todo`, `in_progress`, `qa`, `done`, `cancelled` |
| `assignee` | int | User ID to assign to (0 = unassigned) |
| `priority` | int | `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low |

### List issues

```
GET /api/projects/{project_id}/issues
→ 200 [{"id": 1, "slug": "ASTEROID-GAME-1", ...}, ...]
```

Filters (all optional):

| Param | Example | Description |
|-------|---------|-------------|
| `state` | `?state=qa` | Filter by state |
| `assignee` | `?assignee=1` | Filter by assignee user ID |
| `assigned_to_me` | `?assigned_to_me=true` | Shorthand for current user |
| `q` | `?q=login` | Search title and description |
| `page` | `?page=2` | Page number (default 1) |
| `per_page` | `?per_page=20` | Items per page (default 50) |

### Get issue

```
GET /api/issues/{id}
→ 200 {"id": 1, "slug": "ASTEROID-GAME-1", ...}
→ 404 {"error": "issue not found"}
```

`id` can be a numeric ID or a slug (e.g. `ASTEROID-GAME-42`).

### Update issue

```
PATCH /api/issues/{id}
{"state": "qa", "assignee": 2}
→ 200 {"id": 1, "slug": "ASTEROID-GAME-1", "state": "qa", "assignee": 2, ...}
```

Only provided fields change. Send `0` to leave priority/assignee unchanged.

### Update issue state (shorthand)

```
PUT /api/issues/{id}/state
{"state": "qa"}
→ 200 {"id": 1, "slug": "ASTEROID-GAME-1", "state": "qa", ...}
→ 400 {"error": "invalid state"}
```

Validates the state against the pipeline before updating. Cleaner than PATCH when you just want to move an issue.

### Delete issue

```
DELETE /api/issues/{id}
→ 204
```

## Comments

### List comments

```
GET /api/issues/{id}/comments
→ 200 [{"id": 1, "issue_id": 1, "body": "Looking good", "author": 2, ...}]
```

Ordered oldest first.

### Add comment

```
POST /api/issues/{id}/comments
{"body": "Fixed in commit abc123"}
→ 201 {"id": 1, "issue_id": 1, "body": "Fixed in commit abc123", "author": 1, ...}
```

`author` is set from the authenticated user automatically.

## Errors

All errors return JSON:

| Status | Meaning |
|--------|---------|
| `400` | Invalid JSON or missing required field |
| `401` | Missing or invalid PAT |
| `403` | Admin-only action attempted by non-admin |
| `404` | Entity not found |
| `500` | Server error (usually a database issue) |

```json
{"error": "title is required"}
```
