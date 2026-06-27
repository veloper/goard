package internal

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterGroup_ToSQL_simple(t *testing.T) {
	allowed := AllowedFields{
		"state":            "state",
		"priority":         "priority",
		"assignee_user_id": "assignee_user_id",
	}

	raw := `{"combinator":"and","rules":[
		{"field":"state","operator":"eq","value":"done"},
		{"field":"priority","operator":"gte","value":2}
	]}`

	var fg FilterGroup
	if err := json.Unmarshal([]byte(raw), &fg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	clause, args := fg.ToSQL(allowed)
	wantClause := "state = ? AND priority >= ?"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "done" {
		t.Errorf("arg[0] = %v, want done", args[0])
	}
}

func TestFilterGroup_ToSQL_nested(t *testing.T) {
	allowed := AllowedFields{
		"state":    "state",
		"priority": "priority",
	}

	raw := `{"combinator":"and","rules":[
		{"field":"state","operator":"neq","value":"cancelled"}
	],"groups":[
		{"combinator":"or","rules":[
			{"field":"priority","operator":"eq","value":1},
			{"field":"priority","operator":"eq","value":2}
		]}
	]}`

	var fg FilterGroup
	json.Unmarshal([]byte(raw), &fg)

	clause, args := fg.ToSQL(allowed)
	wantClause := "state != ? AND (priority = ? OR priority = ?)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d", len(args))
	}
}

func TestFilterGroup_ToSQL_in(t *testing.T) {
	allowed := AllowedFields{"state": "state"}

	raw := `{"combinator":"or","rules":[
		{"field":"state","operator":"in","value":["backlog","in_progress"]}
	]}`

	var fg FilterGroup
	json.Unmarshal([]byte(raw), &fg)

	clause, args := fg.ToSQL(allowed)
	wantClause := "state IN (?,?)"
	if clause != wantClause {
		t.Errorf("clause = %q, want %q", clause, wantClause)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
}

func TestFilterGroup_ToSQL_is(t *testing.T) {
	allowed := AllowedFields{"assignee_user_id": "assignee_user_id"}

	tests := []struct {
		value   any
		want    string
	}{
		{nil, "assignee_user_id IS NOT NULL"},
		{true, "assignee_user_id IS NULL"},
		{false, "assignee_user_id IS NOT NULL"},
		{"true", "assignee_user_id IS NULL"},
	}

	for _, tt := range tests {
		r := FilterRule{Field: "assignee_user_id", Operator: OpIs, Value: tt.value}
		clause, _ := compileRule(r, allowed)
		if clause != tt.want {
			t.Errorf("is(%v) = %q, want %q", tt.value, clause, tt.want)
		}
	}
}

func TestFilterGroup_ToSQL_empty(t *testing.T) {
	fg := FilterGroup{Combinator: CombinatorAnd}
	clause, args := fg.ToSQL(nil)
	if clause != "" || args != nil {
		t.Error("expected empty group to produce no clause")
	}
}

func TestFilterGroup_JSON_roundtrip(t *testing.T) {
	raw := `{"combinator":"and","rules":[{"field":"state","operator":"eq","value":"done"}]}`
	var fg FilterGroup
	if err := json.Unmarshal([]byte(raw), &fg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	out, err := json.Marshal(fg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var roundtripped FilterGroup
	json.Unmarshal(out, &roundtripped)

	if !reflect.DeepEqual(fg, roundtripped) {
		t.Errorf("roundtrip changed: got %+v, want %+v", roundtripped, fg)
	}
}

func TestParseFilter(t *testing.T) {
	fg, err := ParseFilter(`{"combinator":"and","rules":[{"field":"state","operator":"eq","value":"done"}]}`)
	if err != nil {
		t.Fatalf("ParseFilter: %v", err)
	}
	if fg.Combinator != CombinatorAnd {
		t.Errorf("combinator = %s, want and", fg.Combinator)
	}
	if len(fg.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(fg.Rules))
	}
	if fg.Rules[0].Field != "state" {
		t.Errorf("field = %s, want state", fg.Rules[0].Field)
	}

	// empty string returns zero-value group
	empty, _ := ParseFilter("")
	if empty.Combinator != "" {
		t.Error("expected empty group for empty string")
	}
}

func TestFilterGroup_ToSQL_panic_on_unknown_field(t *testing.T) {
	fg := FilterGroup{
		Combinator: CombinatorAnd,
		Rules: []FilterRule{
			{Field: "nonexistent", Operator: OpEq, Value: "x"},
		},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown field")
		}
	}()
	fg.ToSQL(AllowedFields{"valid": "valid"})
}
