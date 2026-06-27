package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// resultText extracts the text from a CallToolResult's content array.
func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("content is not TextContent")
	}
	return text.Text
}

func mcpArgs(args map[string]any) mcp.CallToolParams {
	return mcp.CallToolParams{Arguments: args}
}

func TestMCP_get_info(t *testing.T) {
	s := newTestStore(t)
	handler := handleGetInfo(s)
	ctx := context.WithValue(context.Background(), mcpUserKey, &User{ID: 1})

	req := mcp.CallToolRequest{Params: mcpArgs(map[string]any{})}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handleGetInfo: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success, got error")
	}

	var env struct {
		Meta struct {
			Status int `json:"status"`
		} `json:"meta"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(resultText(t, result)), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Meta.Status != 200 {
		t.Errorf("status = %d, want 200", env.Meta.Status)
	}
	if env.Data["users"] == nil {
		t.Error("expected users in data")
	}
	if env.Data["me"] == nil {
		t.Error("expected me in data")
	}
}

func TestMCP_list_users(t *testing.T) {
	s := newTestStore(t)
	handler := handleListUsers(s)
	ctx := context.WithValue(context.Background(), mcpUserKey, &User{ID: 1})

	req := mcp.CallToolRequest{Params: mcpArgs(map[string]any{})}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handleListUsers: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success, got error")
	}

	var env struct {
		Data []map[string]any `json:"data"`
	}
	json.Unmarshal([]byte(resultText(t, result)), &env)
	if len(env.Data) != 3 {
		t.Errorf("expected 3 users, got %d", len(env.Data))
	}
}

func TestMCP_get_user_notFound(t *testing.T) {
	s := newTestStore(t)
	handler := handleGetUser(s)
	ctx := context.Background()

	req := mcp.CallToolRequest{Params: mcpArgs(map[string]any{"id": float64(999)})}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handleGetUser: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestMCP_list_issues_pagination(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Test", "TEST", "", alice.ID)
	for i := 0; i < 5; i++ {
		s.CreateIssue(p.ID, fmt.Sprintf("Issue %d", i), "", "feature", "", 0, 0, alice.ID, 0)
	}

	handler := handleListIssues(s)
	ctx := context.WithValue(context.Background(), mcpUserKey, &User{ID: 1})

	req := mcp.CallToolRequest{Params: mcpArgs(map[string]any{
		"project_id": "TEST",
		"per_page":   float64(2),
	})}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handleListIssues: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	var env struct {
		Data []map[string]any `json:"data"`
	}
	json.Unmarshal([]byte(resultText(t, result)), &env)
	if len(env.Data) != 2 {
		t.Errorf("expected 2 issues, got %d", len(env.Data))
	}
}

func TestMCP_create_project_usesUserContext(t *testing.T) {
	s := newTestStore(t)
	s.EnsureAdmin("admin", "pat_admin")
	handler := handleCreateProject(s)
	ctx := context.WithValue(context.Background(), mcpUserKey, &User{ID: 1, Username: "admin"})

	req := mcp.CallToolRequest{Params: mcpArgs(map[string]any{
		"name": "From MCP",
		"slug": "FROM-MCP",
	})}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatalf("handleCreateProject: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(t, result))
	}

	var env struct {
		Data map[string]any `json:"data"`
	}
	json.Unmarshal([]byte(resultText(t, result)), &env)
	if env.Data["name"] != "From MCP" {
		t.Errorf("name = %v, want 'From MCP'", env.Data["name"])
	}
	if env.Data["slug"] != "FROM-MCP" {
		t.Errorf("slug = %v, want 'FROM-MCP'", env.Data["slug"])
	}
}

func TestMCP_cleanErrStatus(t *testing.T) {
	tests := []struct {
		input    string
		want     string
		wantCode int
	}{
		{"UNIQUE constraint failed: projects.slug", "a project with this slug already exists", 409},
		{"UNIQUE constraint failed: users.username", "a user with this username already exists", 409},
		{"UNIQUE constraint failed: users.pat", "a user with this PAT already exists", 409},
		{"FOREIGN KEY constraint failed", "referenced entity does not exist", 404},
		{"sql: no rows in result set", "not found", 404},
		{"some random error", "some random error", 500},
	}
	for _, tt := range tests {
		got, code := cleanErrStatus(fmt.Errorf("%s", tt.input))
		if got != tt.want {
			t.Errorf("cleanErrStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
		if code != tt.wantCode {
			t.Errorf("cleanErrStatus(%q) status = %d, want %d", tt.input, code, tt.wantCode)
		}
	}
}
