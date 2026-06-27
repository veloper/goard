package internal

// UserRef is a minimal user representation embedded in API responses
// wherever a foreign key to users exists.
type UserRef struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	PAT       string `json:"-"` // never exposed via API
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// IssueRef is a lightweight summary of an issue, used when embedding
// issues in project responses (no description/comments).
type IssueRef struct {
	ID       int64    `json:"id"`
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	State    string   `json:"state"`
	Assignee *UserRef `json:"assignee,omitempty"`
	Priority int      `json:"priority"`
}

type Project struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Description     string     `json:"description"`
	CreatedByUserID int64      `json:"-"` // internal — use CreatedBy for API
	CreatedBy       *UserRef   `json:"created_by"`
	Issues          []IssueRef `json:"issues,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

type Issue struct {
	ID              int64      `json:"id"`
	ProjectID       int64      `json:"project_id"`
	Slug            string     `json:"slug"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Type            string     `json:"type"`
	State           string     `json:"state"`
	AssigneeUserID  int64      `json:"-"` // internal — use Assignee for API
	Assignee        *UserRef   `json:"assignee,omitempty"`
	Priority        int        `json:"priority"`
	ParentID        int64      `json:"parent_id,omitempty"`
	CreatedByUserID int64      `json:"-"` // internal — use CreatedBy for API
	CreatedBy       *UserRef   `json:"created_by"`
	Comments        []Comment  `json:"comments,omitempty"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

type Comment struct {
	ID              int64    `json:"id"`
	IssueID         int64    `json:"issue_id"`
	Body            string   `json:"body"`
	AuthorUserID    int64    `json:"-"` // internal — use Author for API
	Author          *UserRef `json:"author"`
	CreatedByUserID int64    `json:"-"` // internal — use CreatedBy for API
	CreatedBy       *UserRef `json:"created_by"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type SeedUser struct {
	Username string `json:"username"`
	PAT      string `json:"-"`
	Admin    bool   `json:"admin"`
}

type Config struct {
	AdminUsername string `json:"admin_username"`
	AdminPAT      string `json:"admin_pat"`
	DBPath        string `json:"db_path"`
	Host          string `json:"host"`
	Port          string `json:"port"`
}

// Addr returns the listen address from Host and Port.
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

var ValidStates = []string{"backlog", "in_progress", "review", "done", "cancelled"}
var ValidTypes = []string{"epic", "feature", "bug", "chore"}
var ValidPriorityLevels = []int{0, 1, 2, 3, 4}
