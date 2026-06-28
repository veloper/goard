package internal

import "fmt"

// CreateComment adds a comment to an issue.
func (s *Store) CreateComment(issueID int64, body string, authorUserID, createdByUserID int64) (*Comment, error) {
	c := &Comment{
		IssueID: issueID, Body: body,
		AuthorUserID: authorUserID, CreatedByUserID: createdByUserID,
		CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO comments (issue_id, body, author_user_id, created_by_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		c.IssueID, c.Body, c.AuthorUserID, c.CreatedByUserID, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.ID, err = res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	c.Author = s.loadUserRef(c.AuthorUserID)
	c.CreatedBy = s.loadUserRef(c.CreatedByUserID)
	return c, nil
}

// ListComments returns all comments for an issue, oldest first.


// CountComments returns the total number of comments for an issue.
func (s *Store) CountComments(issueID int64) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE issue_id = ?`, issueID).Scan(&n)
	return n, err
}

func (s *Store) ListComments(issueID int64, page, perPage int, orderBy string, filterClause string, filterArgs ...any) ([]Comment, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	offset := (page - 1) * perPage
	if orderBy == "" {
		orderBy = "ORDER BY created_at ASC"
	}
	where := " WHERE issue_id = ?"
	baseArgs := []any{issueID}
	if filterClause != "" {
		where += " AND " + filterClause
	}
	rows, err := s.db.Query(
		`SELECT id, issue_id, body, author_user_id, created_by_user_id, created_at, updated_at
		 FROM comments`+where+` `+orderBy+` LIMIT ? OFFSET ?`, append(append(baseArgs, filterArgs...), perPage, offset)...,
	)
	if err != nil {
		return nil, err
	}
	out := make([]Comment, 0)
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.Body, &c.AuthorUserID, &c.CreatedByUserID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, c)
	}
	closeErr := rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	for i := range out {
		out[i].Author = s.loadUserRef(out[i].AuthorUserID)
		out[i].CreatedBy = s.loadUserRef(out[i].CreatedByUserID)
	}
	return out, nil
}
