# REST API

Base URL: `http://<host>:<port>/api`

All requests require `Authorization: Bearer <pat>` header. PATs are generated server-side on user creation and returned in the create response only.

All responses use a standard envelope:

```json
{"meta": {"status": 200}, "data": { ... }}
{"meta": {"status": 200}, "data": [ ... ]}
{"meta": {"status": 404, "error": "not found"}, "data": null}
```

List endpoints include pagination in meta:

```json
{"meta": {"status": 200, "page": 1, "per_page": 50, "total": 137}, "data": [...]}
```

## Info

```
GET /api/info
```

Returns server metadata. Call this first to discover valid values.

```json
{"meta": {"status": 200}, "data": {
  "states": ["backlog", "todo", "in_progress", "qa", "done", "cancelled"],
  "types": ["epic", "feature", "bug", "chore"],
  "priority_levels": [0, 1, 2, 3, 4],
  "priority_labels": {"0": "none", "1": "urgent", "2": "high", "3": "medium", "4": "low"},
  "users": [{"id": 1, "username": "admin"}],
  "projects": [{"id": 1, "name": "Game", "slug": "GAME"}],
  "me": {"id": 1, "username": "admin", "is_admin": true}
}}
```

## Users

### List users

```
GET /api/users
```

| Param | Description |
|-------|-------------|
| `page` | Page number (default 1) |
| `per_page` | Results per page (default 50) |
| `sort` | Sort column: `username`, `is_admin`, `created_at`, `updated_at` |
| `dir` | Sort direction: `asc` or `desc` (default `asc`) |
| `filter` | JSON filter (react-querybuilder format) |

All user IDs (`created_by`, `assignee`, `author`) are returned as UserRef objects:

```json
{"created_by": {"id": 1, "username": "admin"}}
```

### Get user

```
GET /api/users/{id}
```

Returns user without PAT.

### Create user (admin only)

```
POST /api/users
{"username": "bot", "admin": false}
→ {"meta": {"status": 201}, "data": {"user": {"id": 5, "username": "bot", "is_admin": false}, "pat": "pat_a1b2c3d4..."}}
```

The PAT is auto-generated and **only returned on creation**. Store it.

### Update user PAT (admin only)

```
PATCH /api/users/{id}
{"pat": "pat_new"}
```

### Get user PAT (admin only)

```
GET /api/users/{id}/pat
→ {"meta": {"status": 200}, "data": {"pat": "pat_a1b2c3d4..."}}
```

### Set user PAT (admin only)

```
PUT /api/users/{id}/pat
{"pat": "pat_new"}
→ {"meta": {"status": 200}, "data": {"pat": "pat_new"}}
```

### Delete user (admin only)

```
DELETE /api/users/{id}
→ 204
```

Deletes the user's comments, issues, projects, then the user. Unassigns them from any assigned issues.

### Me

```
GET /api/me
→ {"meta": {"status": 200}, "data": {"id": 1, "username": "admin", "is_admin": true}}
```

Returns the authenticated user based on the PAT in the Authorization header.

## Projects

### Create project

```
POST /api/projects
{"name": "Asteroid Game", "slug": "ASTEROID-GAME"}
```

`slug` is used in issue slugs (`ASTEROID-GAME-42`). Must be unique. `description` is optional.

### List projects

```
GET /api/projects
```

| Param | Description |
|-------|-------------|
| `page` | Page number (default 1) |
| `per_page` | Results per page (default 50) |
| `sort` | Sort column: `name`, `slug`, `created_at`, `updated_at` |
| `dir` | Sort direction |
| `filter` | JSON filter |

### Get project

```
GET /api/projects/{id}
```

Returns project with its issues embedded in the `issues` array.

### Update project

```
PATCH /api/projects/{id}
{"name": "Asteroid Game v2", "description": "A space game"}
```

Only provided fields change. Empty strings are ignored.

### Delete project

```
DELETE /api/projects/{id}
→ 204
```

Deletes the project, all its issues, and all comments on those issues permanently.

## Issues

### Create issue

```
POST /api/projects/{project_id}/issues
{"title": "Add rotation", "type": "feature", "priority": 2}
```

`project_id` can be a numeric ID or a slug. Only `title` is required. Defaults: `state=todo`, `type=feature`, `priority=3`.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | **Required.** Issue title |
| `description` | string | Issue description |
| `type` | string | `epic`, `feature`, `bug`, or `chore` |
| `state` | string | `backlog`, `todo`, `in_progress`, `qa`, `done`, `cancelled` |
| `assignee` | int | User ID to assign (0 = unassigned) |
| `priority` | int | `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low |

### List issues

```
GET /api/projects/{project_id}/issues
```

| Param | Example | Description |
|-------|---------|-------------|
| `state` | `?state=qa` | Filter by state |
| `assignee` | `?assignee=1` | Filter by assignee user ID |
| `assigned_to_me` | `?assigned_to_me=true` | Shorthand for current user |
| `q` | `?q=login` | Search title and description |
| `page` | `?page=2` | Page number (default 1) |
| `per_page` | `?per_page=20` | Items per page (default 50) |
| `sort` | `?sort=priority` | Sort column: `title`, `slug`, `type`, `state`, `priority`, `created_at`, `updated_at` |
| `dir` | `?dir=desc` | Sort direction |
| `filter` | `?filter={"combinator":"and","rules":[...]}` | JSON filter (react-querybuilder format) |

### Get issue

```
GET /api/issues/{id}
```

`id` can be a numeric ID or a slug. Returns issue with comments embedded in the `comments` array.

### Update issue

```
PATCH /api/issues/{id}
{"state": "qa", "assignee": 2}
```

Only provided fields change.

### Update issue state (shorthand)

```
PUT /api/issues/{id}/state
{"state": "qa"}
```

Validates the state against the pipeline before updating.

### Delete issue

```
DELETE /api/issues/{id}
→ 204
```

Deletes the issue and all its comments.

## Comments

### List comments

```
GET /api/issues/{id}/comments
```

Ordered by creation date. Supports `page`, `per_page`, `sort`, `dir`, `filter` params.

### Add comment

```
POST /api/issues/{id}/comments
{"body": "Fixed in commit abc123"}
→ {"meta": {"status": 201}, "data": {"id": 1, "issue_id": 1, "body": "Fixed in commit abc123", "author": {"id": 1, "username": "admin"}, ...}}
```

`author` is set from the authenticated user automatically.

## Errors

All errors return JSON with the error in `meta.error`:

```json
{"meta": {"status": 400, "error": "title is required"}, "data": null}
```

| Status | Meaning |
|--------|---------|
| `400` | Invalid JSON or missing required field |
| `401` | Missing or invalid PAT |
| `403` | Admin-only action attempted by non-admin |
| `404` | Entity not found |
| `409` | Unique constraint violation (duplicate slug, username, or PAT) |
| `500` | Server error |

## Pagination

All list endpoints support:

| Param | Default | Description |
|-------|---------|-------------|
| `page` | `1` | Page number |
| `per_page` | `50` | Items per page |
| `sort` | _varies_ | Sort column (see endpoint docs) |
| `dir` | `asc` | Sort direction |

Response includes pagination in `meta`:

```json
{"meta": {"status": 200, "page": 1, "per_page": 20, "total": 137, "sort": "priority", "dir": "desc"}, "data": [...]}
```

## Filtering

All list endpoints accept a `filter` parameter in [react-querybuilder](https://react-querybuilder.js.org/) JSON format:

```json
?filter={"combinator":"and","rules":[
  {"field":"state","operator":"in","value":["todo","in_progress"]},
  {"field":"priority","operator":"gte","value":2}
]}
```

Supported operators: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `not_in`, `like`, `is` (null check).

User-typed fields (`assignee_user_id`, `created_by_user_id`, `author_user_id`) accept numeric user IDs.
