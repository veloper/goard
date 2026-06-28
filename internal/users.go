package internal

import "fmt"

// GetUserByPAT looks up a user by their personal access token.
func (s *Store) GetUserByPAT(pat string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, is_admin, pat, created_at, updated_at FROM users WHERE pat = ?`, pat,
	).Scan(&u.ID, &u.Username, &u.IsAdmin, &u.PAT, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// ListUsers returns all users ordered by username.
func (s *Store) ListUsers(page, perPage int, orderBy string, filterClause string, filterArgs ...any) ([]User, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	offset := (page - 1) * perPage
	if orderBy == "" {
		orderBy = "ORDER BY username"
	}
	where := ""
	if filterClause != "" {
		where = " WHERE " + filterClause
	}
	rows, err := s.db.Query(`SELECT id, username, is_admin, created_at, updated_at FROM users`+where+` `+orderBy+` LIMIT ? OFFSET ?`, append(filterArgs, perPage, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]User, 0)
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetUser returns a single user by ID.
func (s *Store) GetUser(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, is_admin, pat, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.IsAdmin, &u.PAT, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// CreateUser creates a new user and returns it.
func (s *Store) CreateUser(username, pat string, isAdmin bool) (*User, error) {
	adminInt := 0
	if isAdmin {
		adminInt = 1
	}
	u := &User{
		Username: username, IsAdmin: isAdmin, PAT: pat, CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO users (username, pat, is_admin, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		u.Username, pat, adminInt, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.ID, err = res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return u, nil
}

// UpdateUser updates a user's PAT.
func (s *Store) UpdateUser(id int64, pat string) (*User, error) {
	u, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}
	if pat != "" {
		_, err = s.db.Exec(`UPDATE users SET pat = ? WHERE id = ?`, pat, id)
		if err != nil {
			return nil, err
		}
		u.PAT = pat
	}
	u.UpdatedAt = now()
	_, err = s.db.Exec(`UPDATE users SET updated_at = ? WHERE id = ?`,
		u.UpdatedAt, id)
	return u, err
}

// DeleteUser deletes a user and all resources they created.
func (s *Store) DeleteUser(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM comments WHERE author_user_id = ? OR created_by_user_id = ?`, id, id); err != nil {
		return fmt.Errorf("delete user %d comments: %w", id, err)
	}
	if _, err := s.db.Exec(`UPDATE issues SET assignee_user_id = NULL WHERE assignee_user_id = ?`, id); err != nil {
		return fmt.Errorf("unassign user %d: %w", id, err)
	}
	if _, err := s.db.Exec(`DELETE FROM comments WHERE issue_id IN (SELECT id FROM issues WHERE created_by_user_id = ?)`, id); err != nil {
		return fmt.Errorf("delete user %d issue comments: %w", id, err)
	}
	if _, err := s.db.Exec(`DELETE FROM issues WHERE created_by_user_id = ?`, id); err != nil {
		return fmt.Errorf("delete user %d issues: %w", id, err)
	}
	if _, err := s.db.Exec(`DELETE FROM projects WHERE created_by_user_id = ?`, id); err != nil {
		return fmt.Errorf("delete user %d projects: %w", id, err)
	}
	_, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// GetUserByUsername returns a single user by username.


// CountUsers returns the total number of users.
func (s *Store) CountUsers() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, is_admin, pat, created_at, updated_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.IsAdmin, &u.PAT, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
