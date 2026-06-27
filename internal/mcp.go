package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpCtxKey string

const mcpUserKey mcpCtxKey = "mcp_user"

// mcpUserFromCtx extracts the authenticated user from the MCP request context.
func mcpUserFromCtx(ctx context.Context) *User {
	u, _ := ctx.Value(mcpUserKey).(*User)
	return u
}

// NewMCPServer creates an MCP server with all Goard tools backed by the Store.
func NewMCPServer(store *Store, userRole bool) *server.MCPServer {
	s := server.NewMCPServer(
		"goard",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// ── Info — always available ──
	s.AddTool(mcp.NewTool("get_info",
		mcp.WithDescription("Discover the full Goard surface — valid states (backlog→done), issue types, priority levels with labels, all registered users, and all projects. Call this first to understand what values are accepted by other tools."),
	), handleGetInfo(store))

	// ── Users — always available (read-only) ──
	s.AddTool(mcp.NewTool("list_users",
		mcp.WithDescription("List all users"),
		mcp.WithString("filter", mcp.Description(`react-querybuilder filter: {"combinator":"and","rules":[{"field":"<name>","operator":"<op>","value":"<val>"}]} | fields: username, is_admin | operators: eq, neq, like, is | is_admin values: true/false`)),
		mcp.WithNumber("page", mcp.Description("Page number, default 1")),
		mcp.WithNumber("per_page", mcp.Description("Results per page, default 50")),
		mcp.WithString("sort", mcp.Description("Sort column: username, is_admin, created_at, updated_at")),
		mcp.WithString("dir", mcp.Description("Sort direction: asc or desc")),
	), handleListUsers(store))

	s.AddTool(mcp.NewTool("get_user",
		mcp.WithDescription("Get a user by ID"),
		mcp.WithNumber("id", mcp.Description("User ID"), mcp.Required()),
	), handleGetUser(store))

	// ── Projects — read-only for user role ──
	s.AddTool(mcp.NewTool("list_projects",
		mcp.WithDescription("List all projects"),
		mcp.WithString("filter", mcp.Description(`react-querybuilder filter: {"combinator":"and","rules":[{"field":"<name>","operator":"<op>","value":"<val>"}]} | fields: name, slug, created_by_user_id | operators: eq, neq, like, is, in, not_in | created_by_user_id: numeric user ID`)),
		mcp.WithNumber("page", mcp.Description("Page number, default 1")),
		mcp.WithNumber("per_page", mcp.Description("Results per page, default 50")),
		mcp.WithString("sort", mcp.Description("Sort column: name, slug, created_at, updated_at")),
		mcp.WithString("dir", mcp.Description("Sort direction: asc or desc")),
	), handleListProjects(store))

	s.AddTool(mcp.NewTool("get_project",
		mcp.WithDescription("Get a project by ID or slug"),
		mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
	), handleGetProject(store))

	if !userRole {
		s.AddTool(mcp.NewTool("create_project",
			mcp.WithDescription("Create a new project — both name and slug are required"),
			mcp.WithString("name", mcp.Description("Project name"), mcp.Required()),
			mcp.WithString("slug", mcp.Description("Project slug"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Project description")),
		), handleCreateProject(store))

		s.AddTool(mcp.NewTool("update_project",
			mcp.WithDescription("Update an existing project — only provided fields will change"),
			mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
			mcp.WithString("name", mcp.Description("New name")),
			mcp.WithString("slug", mcp.Description("New slug")),
			mcp.WithString("description", mcp.Description("New description")),
		), handleUpdateProject(store))

		s.AddTool(mcp.NewTool("delete_project",
			mcp.WithDescription("Delete a project and all its issues permanently"),
			mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
		), handleDeleteProject(store))
	}

	// ── Issues — read-only for user role ──
	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List all issues in a project, optionally filtered by state, assignee, or custom filter"),
		mcp.WithString("project_id", mcp.Description("Project ID or slug"), mcp.Required()),
		mcp.WithString("state", mcp.Description("Filter by state: backlog, in_progress, review, done, cancelled")),
		mcp.WithNumber("assignee", mcp.Description("Filter by assignee user ID")),
		mcp.WithString("filter", mcp.Description(`react-querybuilder filter: {"combinator":"and","rules":[{"field":"<name>","operator":"<op>","value":"<val>"}]} | fields: type, state, priority, assignee_user_id, created_by_user_id | operators: eq, neq, gt, gte, lt, lte, in, not_in, like, is | type: epic/feature/bug/chore | state: backlog/in_progress/review/done/cancelled | priority: 0-4 | assignee/created_by: numeric user ID`)),
		mcp.WithNumber("page", mcp.Description("Page number, default 1")),
		mcp.WithNumber("per_page", mcp.Description("Results per page, default 50")),
		mcp.WithString("sort", mcp.Description("Sort column: title, slug, type, state, priority, created_at, updated_at")),
		mcp.WithString("dir", mcp.Description("Sort direction: asc or desc")),
	), handleListIssues(store))

	s.AddTool(mcp.NewTool("get_issue",
		mcp.WithDescription("Get an issue by ID or slug"),
		mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
	), handleGetIssue(store))

	if !userRole {
		s.AddTool(mcp.NewTool("create_issue",
			mcp.WithDescription("Create a new issue in a project — only title and project_id are required"),
			mcp.WithString("project_id", mcp.Description("Project ID or slug"), mcp.Required()),
			mcp.WithString("title", mcp.Description("Issue title"), mcp.Required()),
			mcp.WithString("description", mcp.Description("Issue description")),
			mcp.WithString("type", mcp.Description("Issue type: epic, feature, bug, chore")),
			mcp.WithString("state", mcp.Description("Initial state")),
			mcp.WithNumber("priority", mcp.Description("Priority: 0=none, 1=urgent, 2=high, 3=medium, 4=low")),
		), handleCreateIssue(store))

		s.AddTool(mcp.NewTool("update_issue",
			mcp.WithDescription("Update an existing issue — only provided fields will change (pass 0 to leave priority/assignee unchanged)"),
			mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("type", mcp.Description("New type")),
			mcp.WithString("state", mcp.Description("New state")),
			mcp.WithNumber("priority", mcp.Description("New priority")),
			mcp.WithNumber("assignee", mcp.Description("New assignee user ID")),
		), handleUpdateIssue(store))

		s.AddTool(mcp.NewTool("update_issue_state",
			mcp.WithDescription("Move an issue to a new state in the pipeline — valid states: backlog, in_progress, review, done, cancelled"),
			mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
			mcp.WithString("state", mcp.Description("Target state: backlog, in_progress, review, done, cancelled"), mcp.Required()),
		), handleUpdateIssueState(store))

		s.AddTool(mcp.NewTool("delete_issue",
			mcp.WithDescription("Delete an issue permanently"),
			mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
		), handleDeleteIssue(store))
	}

	// ── Comments — always available ──
	s.AddTool(mcp.NewTool("list_comments",
		mcp.WithDescription("List all comments on an issue, ordered by creation date"),
		mcp.WithString("issue_id", mcp.Description("Issue ID or slug"), mcp.Required()),
		mcp.WithString("filter", mcp.Description(`react-querybuilder filter: {"combinator":"and","rules":[{"field":"<name>","operator":"<op>","value":"<val>"}]} | fields: author_user_id, created_by_user_id | operators: eq, neq, is | author/created_by: numeric user ID`)),
		mcp.WithNumber("page", mcp.Description("Page number, default 1")),
		mcp.WithNumber("per_page", mcp.Description("Results per page, default 50")),
		mcp.WithString("sort", mcp.Description("Sort column: created_at, updated_at")),
		mcp.WithString("dir", mcp.Description("Sort direction: asc or desc")),
	), handleListComments(store))

	s.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to an issue — body is required"),
		mcp.WithString("issue_id", mcp.Description("Issue ID or slug"), mcp.Required()),
		mcp.WithString("body", mcp.Description("Comment text"), mcp.Required()),
	), handleAddComment(store))

	return s
}

// ── Tool handlers ──

var _ = fmt.Sprint // keep import

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

// mcpResult wraps data in the API envelope for MCP tool responses.
func mcpResult(v any) *mcp.CallToolResult {
	r, err := mcp.NewToolResultJSON(envelope{
		Meta: envelopeMeta{Status: 200},
		Data: v,
	})
	if err != nil {
		return mcp.NewToolResultError("marshal: " + err.Error())
	}
	return r
}

func mcpFilter(args map[string]any, allowed AllowedFields) (string, []any) {
	s, _ := args["filter"].(string)
	if s == "" {
		return "", nil
	}
	fg, err := ParseFilter(s)
	if err != nil {
		return "", nil
	}
	return fg.ToSQL(allowed)
}

func intFromArgs(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

func sortFromArgs(args map[string]any, allowed sortWhitelist) string {
	col, _ := args["sort"].(string)
	if col == "" {
		return ""
	}
	expr, ok := allowed[col]
	if !ok {
		return ""
	}
	dir, _ := args["dir"].(string)
	if dir != "asc" && dir != "desc" {
		dir = "asc"
	}
	return fmt.Sprintf("ORDER BY %s %s", expr, dir)
}

func jsonResult(v any) *mcp.CallToolResult {
	r, err := mcp.NewToolResultJSON(v)
	if err != nil {
		return mcp.NewToolResultError("marshal: " + err.Error())
	}
	// Strip StructuredContent — the Python MCP client (langchain)
	// expects a dict but NewToolResultJSON stores the raw Go value,
	// which is a slice for list responses.
	r.StructuredContent = nil
	return r
}

// cleanErr returns a user-facing message for the given error.
func cleanErr(err error) string {
	msg, _ := cleanErrStatus(err)
	return msg
}

// cleanErrStatus returns a user-facing message and an appropriate HTTP status code.
func cleanErrStatus(err error) (string, int) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"):
		if strings.Contains(msg, "projects.slug") {
			return "a project with this slug already exists", 409
		}
		if strings.Contains(msg, "users.username") {
			return "a user with this username already exists", 409
		}
		if strings.Contains(msg, "users.pat") {
			return "a user with this PAT already exists", 409
		}
		return "a record with that unique value already exists", 409
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return "referenced entity does not exist", 404
	case strings.Contains(msg, "no rows in result set"):
		return "not found", 404
	default:
		return msg, 500
	}
}

func resolveProjectID(store *Store, id string) (int64, string) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		return n, ""
	}
	p, err := store.GetProjectBySlug(id)
	if err != nil {
		return 0, "project not found"
	}
	return p.ID, ""
}

func resolveIssueID(store *Store, id string) (int64, string) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		return n, ""
	}
	iss, err := store.GetIssueBySlug(id)
	if err != nil {
		return 0, "issue not found"
	}
	return iss.ID, ""
}

func handleGetInfo(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		users, _ := store.ListUsers(1, 9999, "", "")
		projects, _ := store.ListProjects(1, 9999, "", "")
		me := mcpUserFromCtx(ctx)
		return mcpResult(map[string]any{
			"states":          ValidStates,
			"types":           ValidTypes,
			"priority_levels": ValidPriorityLevels,
			"priority_labels": map[int]string{0: "none", 1: "urgent", 2: "high", 3: "medium", 4: "low"},
			"users":           users,
			"projects":        projects,
			"me":              me,
		}), nil
	}
}

func handleListUsers(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		filterClause, filterArgs := mcpFilter(args, userFilterFields)
		page := intFromArgs(args, "page", 1)
		perPage := intFromArgs(args, "per_page", 50)
		orderBy := sortFromArgs(args, userSorts)
		users, err := store.ListUsers(page, perPage, orderBy, filterClause, filterArgs...)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(users), nil
	}
}

func handleGetUser(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int64(req.GetFloat("id", 0))
		user, err := store.GetUser(id)
		if err != nil {
			return mcp.NewToolResultError("user not found"), nil
		}
		return mcpResult(user), nil
	}
}

func handleListProjects(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		filterClause, filterArgs := mcpFilter(args, projectFilterFields)
		page := intFromArgs(args, "page", 1)
		perPage := intFromArgs(args, "per_page", 50)
		orderBy := sortFromArgs(args, projectSorts)
		projects, err := store.ListProjects(page, perPage, orderBy, filterClause, filterArgs...)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(projects), nil
	}
}

func handleGetProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		p, err := store.GetProject(id)
		if err != nil {
			return mcp.NewToolResultError("project not found"), nil
		}
		return mcpResult(p), nil
	}
}

func handleCreateProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		name := args["name"].(string)
		slug := args["slug"].(string)
		description, _ := args["description"].(string)
		user := mcpUserFromCtx(ctx)
		p, err := store.CreateProject(name, slug, description, user.ID)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(p), nil
	}
}

func handleUpdateProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		name, _ := args["name"].(string)
		slug, _ := args["slug"].(string)
		description, _ := args["description"].(string)
		p, err := store.UpdateProject(id, name, slug, description)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(p), nil
	}
}

func handleDeleteProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		if err := store.DeleteProject(id); err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return textResult("project deleted"), nil
	}
}

func handleListIssues(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		projectStr := args["project_id"].(string)
		pid, errStr := resolveProjectID(store, projectStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		filterClause, filterArgs := mcpFilter(args, issueFilterFields)
		page := intFromArgs(args, "page", 1)
		perPage := intFromArgs(args, "per_page", 50)
		f := IssueFilter{
			Page:    page,
			PerPage: perPage,
			OrderBy: sortFromArgs(args, issueSorts),
			FilterClause: filterClause,
			FilterArgs:   filterArgs,
		}
		if s, ok := args["state"].(string); ok {
			f.State = s
		}
		if a, ok := args["assignee"].(float64); ok {
			f.AssigneeUserID = int64(a)
		}
		issues, _, err := store.ListIssues(pid, f)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(issues), nil
	}
}

func handleGetIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		iss, err := store.GetIssue(id)
		if err != nil {
			return mcp.NewToolResultError("issue not found"), nil
		}
		return mcpResult(iss), nil
	}
}

func handleCreateIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		projectStr := args["project_id"].(string)
		pid, errStr := resolveProjectID(store, projectStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		title := args["title"].(string)
		description, _ := args["description"].(string)
		typ, _ := args["type"].(string)
		state, _ := args["state"].(string)
		priority := 3
		if p, ok := args["priority"].(float64); ok {
			priority = int(p)
		}
		user := mcpUserFromCtx(ctx)
		iss, err := store.CreateIssue(pid, title, description, typ, state, 0, 0, user.ID, priority)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(iss), nil
	}
}

func handleUpdateIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		title, _ := args["title"].(string)
		description, _ := args["description"].(string)
		typ, _ := args["type"].(string)
		state, _ := args["state"].(string)
		var priority *int
		if p, ok := args["priority"].(float64); ok {
			v := int(p)
			priority = &v
		}
		var assigneeUserID *int64
		if a, ok := args["assignee"].(float64); ok {
			v := int64(a)
			assigneeUserID = &v
		}
		iss, err := store.UpdateIssue(id, title, description, typ, state, assigneeUserID, nil, priority)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(iss), nil
	}
}

func handleUpdateIssueState(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		newState := args["state"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		iss, err := store.UpdateIssue(id, "", "", "", newState, nil, nil, nil)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(iss), nil
	}
}

func handleDeleteIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		if err := store.DeleteIssue(id); err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return textResult("issue deleted"), nil
	}
}

func handleListComments(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := req.GetString("issue_id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		filterClause, filterArgs := mcpFilter(args, commentFilterFields)
		page := intFromArgs(args, "page", 1)
		perPage := intFromArgs(args, "per_page", 50)
		orderBy := sortFromArgs(args, commentSorts)
		comments, err := store.ListComments(id, page, perPage, orderBy, filterClause, filterArgs...)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(comments), nil
	}
}

func handleAddComment(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["issue_id"].(string)
		body := args["body"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		user := mcpUserFromCtx(ctx)
		c, err := store.CreateComment(id, body, user.ID, user.ID)
		if err != nil {
			return mcp.NewToolResultError(cleanErr(err)), nil
		}
		return mcpResult(c), nil
	}
}

// ── HTTP handler ──

// ServeMCP handles the POST /mcp endpoint with PAT auth via ?pat= query param.
func (h *Handler) ServeMCP(w http.ResponseWriter, r *http.Request) {
	pat := r.URL.Query().Get("pat")
	if pat == "" {
		http.Error(w, `{"error":"missing pat"}`, http.StatusUnauthorized)
		return
	}
	user, err := h.store.GetUserByPAT(pat)
	if err != nil {
		http.Error(w, `{"error":"invalid pat"}`, http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), mcpUserKey, user)
	// ?role=user restricts to read-only + comment tools. Default is manager (full access).
	mcpServer := h.mcp
	if r.URL.Query().Get("role") == "user" {
		mcpServer = h.mcpUser
	}
	mcpServer.ServeHTTP(w, r.WithContext(ctx))
}

// InitMCP sets up the MCP server and stores it on the Handler.
func (h *Handler) InitMCP(store *Store) {
	manager := NewMCPServer(store, false)
	user := NewMCPServer(store, true)
	h.mcp = server.NewStreamableHTTPServer(manager,
		server.WithEndpointPath("/mcp"),
		server.WithStateful(true),
	)
	h.mcpUser = server.NewStreamableHTTPServer(user,
		server.WithEndpointPath("/mcp"),
		server.WithStateful(true),
	)
}
