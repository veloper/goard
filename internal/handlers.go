package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/mark3labs/mcp-go/server"
	"strconv"
)

type Handler struct {
	store *Store
	hub   *Hub
	mcp   *server.StreamableHTTPServer
	mcpUser *server.StreamableHTTPServer
}

func NewHandler(store *Store, hub *Hub) *Handler {
	return &Handler{store: store, hub: hub}
}

type envelopeMeta struct {
	Status  int    `json:"status"`
	Error   string `json:"error,omitempty"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
	Total   int    `json:"total,omitempty"`
	Sort    string `json:"sort,omitempty"`
	Dir     string `json:"dir,omitempty"`
}

type envelope struct {
	Meta envelopeMeta `json:"meta"`
	Data any          `json:"data"`
}

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{
		Meta: envelopeMeta{Status: status},
		Data: v,
	})
}

func jsonListResp(w http.ResponseWriter, status int, v any, page, perPage, total int, sort, dir string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{
		Meta: envelopeMeta{Status: status, Page: page, PerPage: perPage, Total: total, Sort: sort, Dir: dir},
		Data: v,
	})
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{
		Meta: envelopeMeta{Status: status, Error: msg},
		Data: nil,
	})
}

// ── Info ──

type infoUser struct {
	ID int64 `json:"id"`
}

type infoProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	users, _ := h.store.ListUsers(1, 9999, "", "")
	projects, _ := h.store.ListProjects(1, 9999, "", "")

	compactUsers := make([]infoUser, 0, len(users))
	for _, u := range users {
		compactUsers = append(compactUsers, infoUser{ID: u.ID})
	}
	compactProjects := make([]infoProject, 0, len(projects))
	for _, p := range projects {
		compactProjects = append(compactProjects, infoProject{ID: p.ID, Name: p.Name, Slug: p.Slug})
	}

	jsonResp(w, 200, map[string]any{
		"states":          ValidStates,
		"types":           ValidTypes,
		"priority_levels": ValidPriorityLevels,
		"priority_labels": map[int]string{0: "none", 1: "urgent", 2: "high", 3: "medium", 4: "low"},
		"users":           compactUsers,
		"projects":        compactProjects,
	})
}

// parseInt64 parses a string as an int64, returning 0 on failure.
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parsePageParams(r *http.Request) (page, perPage int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ = strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	return
}

// allowedSorts defines which column names can appear in ?sort= for each endpoint.
// The map value is the SQL ORDER BY expression (same as key for simple cases).
type sortWhitelist map[string]string

var issueSorts = sortWhitelist{
	"created_at": "created_at", "updated_at": "updated_at",
	"priority": "priority", "state": "state", "type": "type",
	"title": "title", "slug": "slug",
}

var issueFilterFields = AllowedFields{
	"type": "type", "state": "state",
	"priority": "priority", "assignee_user_id": "assignee_user_id",
	"created_by_user_id": "created_by_user_id",
}

var projectSorts = sortWhitelist{
	"created_at": "created_at", "updated_at": "updated_at",
	"name": "name", "slug": "slug",
}

var projectFilterFields = AllowedFields{
	"name": "name", "slug": "slug",
	"created_by_user_id": "created_by_user_id",
}

var userSorts = sortWhitelist{
	"created_at": "created_at", "updated_at": "updated_at",
	"username": "username", "is_admin": "is_admin",
}

var userFilterFields = AllowedFields{
	"username": "username", "is_admin": "is_admin",
}

var commentSorts = sortWhitelist{
	"created_at": "created_at", "updated_at": "updated_at",
}

var commentFilterFields = AllowedFields{
	"author_user_id": "author_user_id",
	"created_by_user_id": "created_by_user_id",
}

// parseFilter reads ?filter= as a JSON FilterGroup and compiles it against
// the given allowed fields. Returns ("", nil) when absent or empty.
func parseFilter(r *http.Request, allowed AllowedFields) (string, []any) {
	s := r.URL.Query().Get("filter")
	if s == "" {
		return "", nil
	}
	fg, err := ParseFilter(s)
	if err != nil {
		return "", nil // silently ignore malformed filters
	}
	return fg.ToSQL(allowed)
}

// parseSort validates ?sort= and ?dir= against a whitelist and returns
// the ORDER BY clause, or empty string for default ordering.
func parseSort(r *http.Request, allowed sortWhitelist) string {
	col := r.URL.Query().Get("sort")
	if col == "" {
		return ""
	}
	expr, ok := allowed[col]
	if !ok {
		return ""
	}
	dir := r.URL.Query().Get("dir")
	if dir != "asc" && dir != "desc" {
		dir = "asc"
	}
	return fmt.Sprintf("ORDER BY %s %s", expr, dir)
}

// ── Users ──

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, 200, UserFromCtx(r.Context()))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePageParams(r)
	orderBy := parseSort(r, userSorts)
	filterClause, filterArgs := parseFilter(r, userFilterFields)
	users, err := h.store.ListUsers(page, perPage, orderBy, filterClause, filterArgs...)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	total, err := h.store.CountUsers()
	if err != nil {
		total = 0
	}
	jsonListResp(w, 200, users, page, perPage, total, r.URL.Query().Get("sort"), r.URL.Query().Get("dir"))
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user, err := h.store.GetUser(id)
	if err != nil {
		jsonErr(w, 404, "user not found")
		return
	}
	jsonResp(w, 200, user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	admin := UserFromCtx(r.Context())
	if !admin.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	var body struct {
		Username    string `json:"username"`
		Admin       bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Username == "" {
		jsonErr(w, 400, "username is required")
		return
	}
	pat := generatePAT()
	user, err := h.store.CreateUser(body.Username, pat, body.Admin)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	jsonResp(w, 201, map[string]any{"user": user, "pat": pat})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	var body struct {
		PAT         string `json:"pat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	user, err := h.store.UpdateUser(id, body.PAT)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	jsonResp(w, 200, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	if err := h.store.DeleteUser(id); err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	w.WriteHeader(204)
}

func (h *Handler) GetUserPAT(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	user, err := h.store.GetUser(id)
	if err != nil {
		jsonErr(w, 404, "user not found")
		return
	}
	jsonResp(w, 200, map[string]string{"pat": user.PAT})
}

func (h *Handler) SetUserPAT(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	var body struct {
		PAT string `json:"pat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.PAT == "" {
		jsonErr(w, 400, "pat is required")
		return
	}
	user, err := h.store.UpdateUser(id, body.PAT)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	jsonResp(w, 200, map[string]string{"pat": user.PAT})
}

// ── Projects ──

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePageParams(r)
	orderBy := parseSort(r, projectSorts)
	filterClause, filterArgs := parseFilter(r, projectFilterFields)
	projects, err := h.store.ListProjects(page, perPage, orderBy, filterClause, filterArgs...)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	total, _ := h.store.CountProjects()
	jsonListResp(w, 200, projects, page, perPage, total, r.URL.Query().Get("sort"), r.URL.Query().Get("dir"))
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Name == "" {
		jsonErr(w, 400, "name is required")
		return
	}
	user := UserFromCtx(r.Context())
	p, err := h.store.CreateProject(body.Name, body.Slug, body.Description, user.ID)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectCreated, Payload: p, By: user.ID})
	jsonResp(w, 201, p)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	p, err := h.store.GetProject(id)
	if err != nil {
		jsonErr(w, 404, "project not found")
		return
	}
	jsonResp(w, 200, p)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user := UserFromCtx(r.Context())
	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	existing, err := h.store.GetProject(id)
	if err != nil {
		jsonErr(w, 404, "project not found")
		return
	}
	before := *existing
	p, err := h.store.UpdateProject(id, body.Name, body.Slug, body.Description)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectUpdated,
		Payload: map[string]any{"id": p.ID, "changed": diffProject(&before, p)},
		By:      user.ID})
	jsonResp(w, 200, p)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user := UserFromCtx(r.Context())
	if err := h.store.DeleteProject(id); err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectDeleted,
		Payload: map[string]int64{"id": id}, By: user.ID})
	w.WriteHeader(204)
}

// resolveIssue tries to find an issue by int64 ID first, then by slug.
func (h *Handler) resolveIssue(id string) (*Issue, error) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		iss, err := h.store.GetIssue(n)
		if err == nil {
			return iss, nil
		}
	}
	return h.store.GetIssueBySlug(id)
}

// ── Issues ──

func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	projectID := parseInt64(r.PathValue("id"))

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}

	orderBy := parseSort(r, issueSorts)
	filterClause, filterArgs := parseFilter(r, issueFilterFields)
	f := IssueFilter{
		Type:      r.URL.Query().Get("type"),
		State:     r.URL.Query().Get("state"),
		AssigneeUserID:  parseInt64(r.URL.Query().Get("assignee")),
		CreatedByUserID: parseInt64(r.URL.Query().Get("created_by")),
		Query:     r.URL.Query().Get("q"),
		Page:      page,
		PerPage:   perPage,
		OrderBy:   orderBy,
		FilterClause: filterClause,
		FilterArgs:   filterArgs,
	}

	// assigned_to_me convenience
	if r.URL.Query().Get("assigned_to_me") == "true" && f.AssigneeUserID == 0 {
		user := UserFromCtx(r.Context())
		f.AssigneeUserID = user.ID
	}

	issues, total, err := h.store.ListIssues(projectID, f)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	jsonListResp(w, 200, issues, page, perPage, total, r.URL.Query().Get("sort"), r.URL.Query().Get("dir"))
}

func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	projectID := parseInt64(r.PathValue("id"))
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
		Assignee    int64  `json:"assignee"`
		Priority    int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Title == "" {
		jsonErr(w, 400, "title is required")
		return
	}
	user := UserFromCtx(r.Context())
	iss, err := h.store.CreateIssue(projectID, body.Title, body.Description, body.Type, body.State, body.Assignee, 0, user.ID, body.Priority)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueCreated, Payload: iss, By: user.ID})
	jsonResp(w, 201, iss)
}

func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	jsonResp(w, 200, iss)
}

func (h *Handler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	user := UserFromCtx(r.Context())
	before := *iss
	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Type        string  `json:"type"`
		State       string  `json:"state"`
		Assignee    *int64  `json:"assignee"`
		Priority    *int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	iss, err = h.store.UpdateIssue(iss.ID, body.Title, body.Description, body.Type, body.State, body.Assignee, nil, body.Priority)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueUpdated,
		Payload: map[string]any{"id": iss.ID, "changed": diffIssue(&before, iss)},
		By:      user.ID})
	jsonResp(w, 200, iss)
}

func (h *Handler) UpdateIssueState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.State == "" {
		jsonErr(w, 400, "state is required")
		return
	}
	valid := false
	for _, s := range ValidStates {
		if s == body.State {
			valid = true
			break
		}
	}
	if !valid {
		jsonErr(w, 400, "invalid state")
		return
	}
	user := UserFromCtx(r.Context())
	before := *iss
	iss, err = h.store.UpdateIssue(iss.ID, "", "", "", body.State, nil, nil, nil)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueUpdated,
		Payload: map[string]any{"id": iss.ID, "changed": diffIssue(&before, iss)},
		By:      user.ID})
	jsonResp(w, 200, iss)
}

func (h *Handler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	user := UserFromCtx(r.Context())
	if err := h.store.DeleteIssue(iss.ID); err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueDeleted,
		Payload: map[string]int64{"id": iss.ID, "project_id": iss.ProjectID},
		By:      user.ID})
	w.WriteHeader(204)
}

// ── Comments ──

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	page, perPage := parsePageParams(r)
	orderBy := parseSort(r, commentSorts)
	filterClause, filterArgs := parseFilter(r, commentFilterFields)
	comments, err := h.store.ListComments(iss.ID, page, perPage, orderBy, filterClause, filterArgs...)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	jsonResp(w, 200, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Body == "" {
		jsonErr(w, 400, "body is required")
		return
	}
	user := UserFromCtx(r.Context())
	c, err := h.store.CreateComment(iss.ID, body.Body, user.ID, user.ID)
	if err != nil {
		errMsg, errStatus := cleanErrStatus(err); jsonErr(w, errStatus, errMsg)
		return
	}
	h.hub.Broadcast(Event{Type: EventCommentCreated, Payload: c, By: user.ID})
	jsonResp(w, 201, c)
}
