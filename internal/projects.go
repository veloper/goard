package internal

import "fmt"

// CreateProject creates a new project for the given user.
func (s *Store) CreateProject(name, slug, description string, createdByUserID int64) (*Project, error) {
	if slug == "" {
		slug = slugify(name)
	}
	p := &Project{
		Name: name, Slug: slug, Description: description,
		CreatedByUserID: createdByUserID, CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO projects (name, slug, description, created_by_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.Name, p.Slug, p.Description, p.CreatedByUserID, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	p.ID, err = res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	p.CreatedBy = s.loadUserRef(p.CreatedByUserID)
	return p, nil
}

// ListProjects returns all projects, newest first. Issues are NOT loaded.
func (s *Store) ListProjects(page, perPage int, orderBy string, filterClause string, filterArgs ...any) ([]Project, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	offset := (page - 1) * perPage
	if orderBy == "" {
		orderBy = "ORDER BY created_at DESC"
	}
	where := ""
	if filterClause != "" {
		where = " WHERE " + filterClause
	}
	rows, err := s.db.Query(`SELECT id, name, slug, description, created_by_user_id, created_at, updated_at FROM projects`+where+` `+orderBy+` LIMIT ? OFFSET ?`, append(filterArgs, perPage, offset)...)
	if err != nil {
		return nil, err
	}
	out := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedByUserID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, p)
	}
	closeErr := rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	for i := range out {
		out[i].CreatedBy = s.loadUserRef(out[i].CreatedByUserID)
	}
	return out, nil
}

// getProjectByID is the lean internal lookup — no issue refs loaded.
func (s *Store) getProjectByID(id int64) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, created_by_user_id, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedByUserID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project %d: %w", id, err)
	}
	p.CreatedBy = s.loadUserRef(p.CreatedByUserID)
	return p, nil
}

// GetProject returns a project by ID with its issues loaded.
func (s *Store) GetProject(id int64) (*Project, error) {
	p, err := s.getProjectByID(id)
	if err != nil {
		return nil, err
	}
	p.Issues = s.loadIssueRefs(p.ID)
	return p, nil
}

// GetProjectBySlug returns a project by slug with its issues loaded.
func (s *Store) GetProjectBySlug(slug string) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, created_by_user_id, created_at, updated_at FROM projects WHERE slug = ?`, slug,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedByUserID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project by slug %s: %w", slug, err)
	}
	p.CreatedBy = s.loadUserRef(p.CreatedByUserID)
	p.Issues = s.loadIssueRefs(p.ID)
	return p, nil
}

// UpdateProject updates a project's name, slug, and/or description.
func (s *Store) UpdateProject(id int64, name, slug, description string) (*Project, error) {
	p, err := s.getProjectByID(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if slug != "" {
		p.Slug = slug
	}
	if description != "" {
		p.Description = description
	}
	p.UpdatedAt = now()
	_, err = s.db.Exec(`UPDATE projects SET name=?, slug=?, description=?, updated_at=? WHERE id=?`,
		p.Name, p.Slug, p.Description, p.UpdatedAt, id)
	if err != nil {
		return nil, err
	}
	p.Issues = s.loadIssueRefs(p.ID)
	return p, nil
}

// DeleteProject deletes a project by ID.


// CountProjects returns the total number of projects.
func (s *Store) CountProjects() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&n)
	return n, err
}

func (s *Store) DeleteProject(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM comments WHERE issue_id IN (SELECT id FROM issues WHERE project_id = ?)`, id); err != nil {
		return fmt.Errorf("delete project %d comments: %w", id, err)
	}
	if _, err := s.db.Exec(`DELETE FROM issues WHERE project_id = ?`, id); err != nil {
		return fmt.Errorf("delete project %d issues: %w", id, err)
	}
	_, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}
