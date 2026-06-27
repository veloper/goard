# Changelog

## v0.2.0 (2026-06-27)

### 🚨 Breaking changes

- **Rename Ticketer → Goard.** Binary, CLI, module path, env vars, Docker image all updated.
- **API envelope.** All responses wrapped in `{"meta":{...},"data":...}`.
- **UserRef objects.** `assignee`, `created_by`, `author` now return `{"id":1,"username":"admin"}` instead of bare ints.
- **Column renames.** DB columns: `created_by` → `created_by_user_id`, `assignee` → `assignee_user_id`, `author` → `author_user_id`.
- **Removed `display_name`.** User model no longer has display name.
- **Env vars.** `TICKETER_*` → `GOARD_*`.
- **Default DB.** `ticketer.db` → `goard.db`.

### ✨ New features

- **Filter DSL.** All list endpoints accept `?filter=` in react-querybuilder JSON format. Nested AND/OR groups, 10 operators.
- **Pagination.** `?page=` and `?per_page=` on all list endpoints with total count in meta.
- **Sorting.** `?sort=` and `?dir=` on all list endpoints with column whitelist.
- **MCP tools.** All list tools support filter, page, per_page, sort, dir params.
- **PAT endpoints.** `GET/PUT /api/users/{id}/pat` for explicit PAT management.
- **WebSocket envelope.** All broadcast events wrapped in `{"meta":...,"data":...}` with UserRefs.
- **Cascading deletes.** Deleting a user/project/issue cascades to dependent children.

### 🐛 Bug fixes

- **N+1 queries.** List methods close result rows before loading UserRefs, preventing SQLite deadlocks.
- **Nil pointer safety.** All `CreatedBy`/`Author` fields are `*UserRef` (pointer), never dereferenced without nil check.
- **Error status codes.** DB errors map to proper HTTP codes: 404 for not found, 409 for conflicts, 500 for unknown.
- **Raw SQL leak.** All `err.Error()` passes through `cleanErr()` — no more `"sql: no rows in result set"` in responses.
- **PAT leak.** `User.PAT` is `json:"-"` — never exposed via API. Only returned on create or explicit PAT endpoint.
- **MCP user context.** Create tool handlers use authenticated user from context, not hardcoded ID 1.
- **LastInsertId checks.** All `res.LastInsertId()` errors are now surfaced.

### 🧪 Tests

- **112 tests** across 4 test files (was ~48).
- **Store tests.** User CRUD, GetIssueBySlug, CountProjects, CountComments, pagination + sort.
- **Handler tests.** CreateUser, DeleteUser, pagination/sort through HTTP, PAT endpoints.
- **MCP tests.** `mcp_test.go` covers get_info, list_users, get_user not found, list_issues pagination, create_project user context, cleanErrStatus.
- **Filter tests.** SQL compilation, nested groups, IN/NOT IN/IS operators, JSON roundtrip.

### 🔧 Infrastructure

- **GitHub Actions CI.** Tests on every push and PR.
- **GoReleaser.** Cross-compile goard + goardctl for linux/darwin × amd64/arm64 on tag push.
- **Docker auto-build.** `veloper/goard:main` and `veloper/goard:<sha>` on every push to main.
- **Binary builds.** Archives + checksums attached to GitHub Releases.
- **`scripts/release.sh`.** Semver-aware release helper with version guard.

### 🏠 Project

- Added `CONTRIBUTING.md`.
- Removed `todo` state. Pipeline is now `backlog → in_progress → review → done | cancelled`.
- Renamed `qa` state → `review`.
- Replaced `epic` type → `improvement`.
- Renamed `cmd/tktrctl/` → `cmd/goardctl/`.
