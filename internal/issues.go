package internal

import (
	"database/sql"
	"fmt"
	"strings"
)

// CreateIssue creates a new issue in a project. The slug is auto-generated
// from the project slug and the issue's auto-increment ID (e.g. "ASTEROID-GAME-42").
func (s *Store) CreateIssue(projectID int64, title, description, typ, state string, assigneeUserID int64, parentID int64, createdByUserID int64, priority int) (*Issue, error) {
	p, err := s.getProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	if state == "" {
		state = "todo"
	}
	if typ == "" {
		typ = "feature"
	}
	iss := &Issue{
		ProjectID: projectID, Slug: "_",
		Title: title, Description: description, Type: typ,
		State: state, AssigneeUserID: assigneeUserID, Priority: priority,
		ParentID: parentID, CreatedByUserID: createdByUserID,
		CreatedAt: now(), UpdatedAt: now(),
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO issues (project_id, slug, title, description, type, state, assignee_user_id, priority, parent_id, created_by_user_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		iss.ProjectID, iss.Slug, iss.Title, iss.Description,
		iss.Type, iss.State, nullableInt(iss.AssigneeUserID), iss.Priority, nullableInt(iss.ParentID),
		iss.CreatedByUserID, iss.CreatedAt, iss.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	iss.ID, err = res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	iss.Slug = fmt.Sprintf("%s-%d", p.Slug, iss.ID)
	if _, err := tx.Exec(`UPDATE issues SET slug = ? WHERE id = ?`, iss.Slug, iss.ID); err != nil {
		return nil, fmt.Errorf("update slug: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	iss.CreatedBy = s.loadUserRef(iss.CreatedByUserID)
	iss.Assignee = s.loadUserRef(iss.AssigneeUserID)
	return iss, nil
}

// IssueFilter specifies optional filters for listing issues.
type IssueFilter struct {
	Type            string
	State           string
	AssigneeUserID  int64
	CreatedByUserID int64
	Query           string
	Page            int
	PerPage         int
	OrderBy         string // e.g. "ORDER BY created_at DESC" — validated by caller
	FilterClause    string // compiled WHERE fragment from FilterGroup
	FilterArgs      []any  // args for FilterClause
}

// ListIssues returns issues for a project, optionally filtered.
func (s *Store) ListIssues(projectID int64, f IssueFilter) ([]Issue, int, error) {
	where := []string{"project_id = ?"}
	args := []any{projectID}

	if f.Type != "" {
		where = append(where, "type = ?")
		args = append(args, f.Type)
	}
	if f.State != "" {
		where = append(where, "state = ?")
		args = append(args, f.State)
	}
	if f.AssigneeUserID != 0 {
		where = append(where, "assignee_user_id = ?")
		args = append(args, f.AssigneeUserID)
	}
	if f.CreatedByUserID != 0 {
		where = append(where, "created_by_user_id = ?")
		args = append(args, f.CreatedByUserID)
	}
	if f.Query != "" {
		where = append(where, "(title LIKE ? OR description LIKE ?)")
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}
	if f.FilterClause != "" {
		where = append(where, f.FilterClause)
		args = append(args, f.FilterArgs...)
	}

	if f.PerPage <= 0 {
		f.PerPage = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PerPage

	order := "ORDER BY created_at DESC"
	if f.OrderBy != "" {
		order = f.OrderBy
	}
	q := fmt.Sprintf(
		`SELECT id, project_id, slug, title, description, type, state, assignee_user_id, priority, parent_id, created_by_user_id, created_at, updated_at
		 FROM issues WHERE %s %s LIMIT ? OFFSET ?`,
		strings.Join(where, " AND "), order,
	)
	args = append(args, f.PerPage, offset)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query: %w", err)
	}

	var out []Issue
	for rows.Next() {
		var i Issue
		var assignee, parentID sql.NullInt64
		if scanErr := rows.Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
			&i.Type, &i.State, &assignee, &i.Priority, &parentID,
			&i.CreatedByUserID, &i.CreatedAt, &i.UpdatedAt); scanErr != nil {
			rows.Close()
			return nil, 0, fmt.Errorf("scan: %w", scanErr)
		}
		i.AssigneeUserID = assignee.Int64
		i.ParentID = parentID.Int64
		out = append(out, i)
	}
	closeErr := rows.Close()
	if rowErr := rows.Err(); rowErr != nil {
		return nil, 0, fmt.Errorf("rows: %w", rowErr)
	}
	if closeErr != nil {
		return nil, 0, fmt.Errorf("rows close: %w", closeErr)
	}

	// Total count (without pagination)
	var total int
	countQ := fmt.Sprintf(
		`SELECT COUNT(*) FROM issues WHERE %s`,
		strings.Join(where, " AND "),
	)
	// args[:len(where)] are the WHERE args (exclude LIMIT/OFFSET)
	whereArgs := args[:len(args)-2]
	if err := s.db.QueryRow(countQ, whereArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	for i := range out {
		out[i].CreatedBy = s.loadUserRef(out[i].CreatedByUserID)
		out[i].Assignee = s.loadUserRef(out[i].AssigneeUserID)
	}
	return out, total, nil
}

// GetIssue returns a single issue by ID.
func (s *Store) GetIssue(id int64) (*Issue, error) {
	i := &Issue{}
	var assignee, parentID sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, project_id, slug, title, description, type, state, assignee_user_id, priority, parent_id, created_by_user_id, created_at, updated_at
		 FROM issues WHERE id = ?`, id,
	).Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
		&i.Type, &i.State, &assignee, &i.Priority, &parentID,
		&i.CreatedByUserID, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get issue %d: %w", id, err)
	}
	i.AssigneeUserID = assignee.Int64
	i.ParentID = parentID.Int64
	i.CreatedBy = s.loadUserRef(i.CreatedByUserID)
	i.Assignee = s.loadUserRef(i.AssigneeUserID)
	comments, cerr := s.ListComments(i.ID, 1, 50, "", "")
	if cerr == nil {
		i.Comments = comments
	}
	return i, nil
}

// GetIssueBySlug returns a single issue by its slug.
func (s *Store) GetIssueBySlug(slug string) (*Issue, error) {
	i := &Issue{}
	var assignee, parentID sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, project_id, slug, title, description, type, state, assignee_user_id, priority, parent_id, created_by_user_id, created_at, updated_at
		 FROM issues WHERE slug = ?`, slug,
	).Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
		&i.Type, &i.State, &assignee, &i.Priority, &parentID,
		&i.CreatedByUserID, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get issue by slug %s: %w", slug, err)
	}
	i.AssigneeUserID = assignee.Int64
	i.ParentID = parentID.Int64
	i.CreatedBy = s.loadUserRef(i.CreatedByUserID)
	i.Assignee = s.loadUserRef(i.AssigneeUserID)
	comments, cerr := s.ListComments(i.ID, 1, 50, "", "")
	if cerr == nil {
		i.Comments = comments
	}
	return i, nil
}

// UpdateIssue updates fields on an issue. Nil pointer fields are left unchanged.
func (s *Store) UpdateIssue(id int64, title, description, typ, state string, assigneeUserID *int64, parentID *int64, priority *int) (*Issue, error) {
	i, err := s.GetIssue(id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		i.Title = title
	}
	if description != "" {
		i.Description = description
	}
	if typ != "" {
		i.Type = typ
	}
	if state != "" {
		i.State = state
	}
	if assigneeUserID != nil {
		i.AssigneeUserID = *assigneeUserID
	}
	if priority != nil {
		i.Priority = *priority
	}
	if parentID != nil {
		i.ParentID = *parentID
	}
	i.UpdatedAt = now()
	_, err = s.db.Exec(
		`UPDATE issues SET title=?, description=?, type=?, state=?, assignee_user_id=?, priority=?, parent_id=?, updated_at=? WHERE id=?`,
		i.Title, i.Description, i.Type, i.State, nullableInt(i.AssigneeUserID), i.Priority, nullableInt(i.ParentID), i.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update issue %d: %w", id, err)
	}
	// Re-populate UserRefs in case assignee changed
	i.Assignee = s.loadUserRef(i.AssigneeUserID)
	i.CreatedBy = s.loadUserRef(i.CreatedByUserID)
	return i, nil
}

// DeleteIssue deletes an issue and its comments by ID.
func (s *Store) DeleteIssue(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM comments WHERE issue_id = ?`, id); err != nil {
		return fmt.Errorf("delete issue %d comments: %w", id, err)
	}
	if _, err := s.db.Exec(`DELETE FROM issues WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete issue %d: %w", id, err)
	}
	return nil
}
