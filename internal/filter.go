package internal

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ── Operators ──

// Operator is a comparison operator used in filter rules.
type Operator string

const (
	OpEq    Operator = "eq"     // field = value
	OpNeq   Operator = "neq"    // field != value
	OpGt    Operator = "gt"     // field > value
	OpGte   Operator = "gte"    // field >= value
	OpLt    Operator = "lt"     // field < value
	OpLte   Operator = "lte"    // field <= value
	OpIn    Operator = "in"     // field IN (values)
	OpNotIn Operator = "not_in" // field NOT IN (values)
	OpLike  Operator = "like"   // field LIKE value
	OpIs    Operator = "is"     // field IS NULL / IS NOT NULL
)

// ValidOperators returns all supported operators.
func ValidOperators() []Operator {
	return []Operator{OpEq, OpNeq, OpGt, OpGte, OpLt, OpLte, OpIn, OpNotIn, OpLike, OpIs}
}

// sqlOp maps an Operator to its SQL fragment.
func (o Operator) sqlOp() string {
	switch o {
	case OpEq:
		return "="
	case OpNeq:
		return "!="
	case OpGt:
		return ">"
	case OpGte:
		return ">="
	case OpLt:
		return "<"
	case OpLte:
		return "<="
	case OpIn:
		return "IN"
	case OpNotIn:
		return "NOT IN"
	case OpLike:
		return "LIKE"
	default:
		return ""
	}
}

// ── FilterRule ──

// FilterRule is a single condition: field operator value.
type FilterRule struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator"`
	Value    any      `json:"value"` // single value, or []any for in/not_in
}

// ── FilterGroup ──

// Combinator joins rules within a group.
type Combinator string

const (
	CombinatorAnd Combinator = "and"
	CombinatorOr  Combinator = "or"
)

// FilterGroup is a recursive filter tree. A group holds either rules
// (leaf conditions) or nested groups (sub-groups), joined by the combinator.
type FilterGroup struct {
	Combinator Combinator    `json:"combinator"`
	Rules      []FilterRule  `json:"rules,omitempty"`
	Groups     []FilterGroup `json:"groups,omitempty"`
}

// AllowedFields defines which column names a filter rule may target.
// The map key is the JSON-facing field name; the value is the SQL column expression.
type AllowedFields map[string]string

// ── Compilation ──

// ToSQL compiles the filter tree into a WHERE clause string and a slice of
// query arguments. Returns ("", nil) for an empty group.
// Panics at compile time if a field is not in allowed or an operator is unsupported.
func (fg FilterGroup) ToSQL(allowed AllowedFields) (string, []any) {
	if len(fg.Rules) == 0 && len(fg.Groups) == 0 {
		return "", nil
	}
	var clauses []string
	var args []any

	for _, r := range fg.Rules {
		clause, ruleArgs := compileRule(r, allowed)
		clauses = append(clauses, clause)
		args = append(args, ruleArgs...)
	}
	for _, g := range fg.Groups {
		clause, groupArgs := g.ToSQL(allowed)
		if clause == "" {
			continue
		}
		clauses = append(clauses, "("+clause+")")
		args = append(args, groupArgs...)
	}

	if len(clauses) == 0 {
		return "", nil
	}
	sep := " " + strings.ToUpper(string(fg.Combinator)) + " "
	return strings.Join(clauses, sep), args
}

func compileRule(r FilterRule, allowed AllowedFields) (string, []any) {
	col, ok := allowed[r.Field]
	if !ok {
		panic(fmt.Sprintf("filter: unknown field %q", r.Field))
	}

	switch r.Operator {
	case OpIs:
		// value is ignored — use "is null" for truthy, "is not null" for falsy
		if isTruthy(r.Value) {
			return col + " IS NULL", nil
		}
		return col + " IS NOT NULL", nil

	case OpIn, OpNotIn:
		vals, ok := r.Value.([]any)
		if !ok {
			// Accept JSON arrays parsed as []any, or a single json.RawMessage
			if jsonArr, ok2 := r.Value.(json.RawMessage); ok2 {
				json.Unmarshal(jsonArr, &vals)
			}
		}
		if len(vals) == 0 {
			// IN () is always false, NOT IN () is always true
			if r.Operator == OpIn {
				return "1=0", nil
			}
			return "1=1", nil
		}
		placeholders := make([]string, len(vals))
		anyVals := make([]any, len(vals))
		for i, v := range vals {
			placeholders[i] = "?"
			anyVals[i] = v
		}
		return fmt.Sprintf("%s %s (%s)", col, r.Operator.sqlOp(), strings.Join(placeholders, ",")), anyVals

	default:
		return fmt.Sprintf("%s %s ?", col, r.Operator.sqlOp()), []any{r.Value}
	}
}

// isTruthy returns true for values that SQL considers "set" — used by the IS operator.
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != "" && x != "false" && x != "0"
	case float64, int64, int:
		return x != 0
	default:
		return true
	}
}

// ── JSON helpers for MCP (string arg → FilterGroup) ──

// ParseFilter parses a JSON string into a FilterGroup.
func ParseFilter(s string) (FilterGroup, error) {
	if s == "" {
		return FilterGroup{}, nil
	}
	var fg FilterGroup
	if err := json.Unmarshal([]byte(s), &fg); err != nil {
		return FilterGroup{}, fmt.Errorf("parse filter: %w", err)
	}
	return fg, nil
}

// MustParseFilter parses a JSON string or panics.
func MustParseFilter(s string) FilterGroup {
	fg, err := ParseFilter(s)
	if err != nil {
		panic(err)
	}
	return fg
}
