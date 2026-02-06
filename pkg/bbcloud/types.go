package bbcloud

import "time"

// User represents a Bitbucket Cloud user or account
type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AccountID   string `json:"account_id"`
	Nickname    string `json:"nickname,omitempty"`
	Type        string `json:"type"`
	Links       Links  `json:"links,omitempty"`
}

// GetName returns the best available name for the user
// Prefers Username, falls back to DisplayName, then Nickname
func (u *User) GetName() string {
	if u.Username != "" {
		return u.Username
	}
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Nickname != "" {
		return u.Nickname
	}
	return ""
}

// Repository represents a Bitbucket Cloud repository
type Repository struct {
	UUID        string       `json:"uuid"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	FullName    string       `json:"full_name"`
	IsPrivate   bool         `json:"is_private"`
	Description string       `json:"description,omitempty"`
	CreatedOn   time.Time    `json:"created_on"`
	UpdatedOn   time.Time    `json:"updated_on"`
	MainBranch  *Branch      `json:"mainbranch,omitempty"`
	Owner       *User        `json:"owner,omitempty"`
	Project     *Project     `json:"project,omitempty"`
	Links       Links        `json:"links,omitempty"`
	Type        string       `json:"type"`
	SCM         string       `json:"scm,omitempty"`
	Language    string       `json:"language,omitempty"`
	Size        int64        `json:"size,omitempty"`
}

// Branch represents a repository branch
type Branch struct {
	Name   string           `json:"name"`
	Target *CommitReference `json:"target,omitempty"`
	Type   string           `json:"type"`
}

// Project represents a Bitbucket Cloud project
type Project struct {
	UUID string `json:"uuid"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PullRequest represents a Bitbucket Cloud pull request
type PullRequest struct {
	ID           int                 `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	State        string              `json:"state"`
	Author       *User               `json:"author,omitempty"`
	Source       *PullRequestBranch  `json:"source,omitempty"`
	Destination  *PullRequestBranch  `json:"destination,omitempty"`
	Reviewers    []User              `json:"reviewers,omitempty"`
	Participants []Participant       `json:"participants,omitempty"`
	CreatedOn    time.Time           `json:"created_on"`
	UpdatedOn    time.Time           `json:"updated_on"`
	MergeCommit  *CommitReference    `json:"merge_commit,omitempty"`
	Links        Links               `json:"links,omitempty"`
	Type         string              `json:"type"`
	CloseSourceBranch bool           `json:"close_source_branch"`
	CommentCount int                 `json:"comment_count,omitempty"`
	TaskCount    int                 `json:"task_count,omitempty"`
}

// PullRequestBranch represents source or destination branch in a PR
type PullRequestBranch struct {
	Branch     *Branch      `json:"branch,omitempty"`
	Commit     *CommitReference `json:"commit,omitempty"`
	Repository *Repository  `json:"repository,omitempty"`
}

// CommitReference represents a commit reference
type CommitReference struct {
	Hash  string    `json:"hash"`
	Type  string    `json:"type"`
	Links Links     `json:"links,omitempty"`
	Date  time.Time `json:"date,omitempty"`
}

// Participant represents a PR participant
type Participant struct {
	User            *User  `json:"user,omitempty"`
	Role            string `json:"role"`
	Approved        bool   `json:"approved"`
	State           string `json:"state,omitempty"`
	ParticipatedOn  time.Time `json:"participated_on,omitempty"`
}

// FileStats represents file-level diff statistics
type FileStats struct {
	Path         string    `json:"path,omitempty"` // Top-level path (rarely populated)
	LinesAdded   int       `json:"lines_added"`
	LinesRemoved int       `json:"lines_removed"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
	Old          *FileInfo `json:"old,omitempty"`
	New          *FileInfo `json:"new,omitempty"`
}

// FileInfo represents file information in diff stats
type FileInfo struct {
	Path        string `json:"path"`
	EscapedPath string `json:"escaped_path,omitempty"`
	Type        string `json:"type"`
	Links       Links  `json:"links,omitempty"`
}

// GetPath returns the file path, preferring new over old
func (fs *FileStats) GetPath() string {
	if fs.Path != "" {
		return fs.Path
	}
	if fs.New != nil && fs.New.Path != "" {
		return fs.New.Path
	}
	if fs.Old != nil && fs.Old.Path != "" {
		return fs.Old.Path
	}
	return ""
}

// Comment represents a PR comment (general or inline)
type Comment struct {
	ID        int              `json:"id"`
	Content   *Content         `json:"content,omitempty"`
	User      *User            `json:"user,omitempty"`
	CreatedOn time.Time        `json:"created_on"`
	UpdatedOn time.Time        `json:"updated_on"`
	Inline    *InlineLocation  `json:"inline,omitempty"`
	Parent    *CommentRef      `json:"parent,omitempty"`
	Links     Links            `json:"links,omitempty"`
	Type      string           `json:"type"`
	Deleted   bool             `json:"deleted,omitempty"`
}

// Content represents rich content (markdown, raw, html)
type Content struct {
	Raw    string `json:"raw"`
	Markup string `json:"markup,omitempty"`
	HTML   string `json:"html,omitempty"`
	Type   string `json:"type,omitempty"`
}

// InlineLocation represents location of an inline comment
type InlineLocation struct {
	Path    string `json:"path"`
	From    *int   `json:"from,omitempty"`
	To      *int   `json:"to,omitempty"`
	StartTo *int   `json:"start_to,omitempty"`
}

// CommentRef is a reference to a parent comment
type CommentRef struct {
	ID int `json:"id"`
}

// IsInline returns true if this is an inline comment
func (c *Comment) IsInline() bool {
	return c.Inline != nil
}

// Activity represents an activity item in a PR timeline
type Activity struct {
	Update    *ActivityUpdate  `json:"update,omitempty"`
	Comment   *Comment         `json:"comment,omitempty"`
	Approval  *ActivityApproval `json:"approval,omitempty"`
}

// ActivityUpdate represents a PR update activity
type ActivityUpdate struct {
	Date        time.Time   `json:"date"`
	Author      *User       `json:"author,omitempty"`
	State       string      `json:"state,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Source      *PullRequestBranch `json:"source,omitempty"`
	Destination *PullRequestBranch `json:"destination,omitempty"`
}

// ActivityApproval represents an approval activity
type ActivityApproval struct {
	Date time.Time `json:"date"`
	User *User     `json:"user,omitempty"`
}

// Pipeline represents a Bitbucket Pipelines build
type Pipeline struct {
	UUID         string           `json:"uuid"`
	BuildNumber  int              `json:"build_number"`
	State        *PipelineState   `json:"state,omitempty"`
	CreatedOn    time.Time        `json:"created_on"`
	CompletedOn  *time.Time       `json:"completed_on,omitempty"`
	Target       *PipelineTarget  `json:"target,omitempty"`
	Repository   *Repository      `json:"repository,omitempty"`
	Creator      *User            `json:"creator,omitempty"`
	Links        Links            `json:"links,omitempty"`
	Type         string           `json:"type"`
}

// CommitStatus represents a commit status (build/check result)
// This is different from Pipeline - it's used for the /statuses endpoint
type CommitStatus struct {
	UUID        string          `json:"uuid,omitempty"`
	Key         string          `json:"key"`
	RefName     string          `json:"refname,omitempty"`
	URL         string          `json:"url,omitempty"`
	State       string          `json:"state"` // SUCCESSFUL, FAILED, INPROGRESS, STOPPED
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	CreatedOn   time.Time       `json:"created_on"`
	UpdatedOn   time.Time       `json:"updated_on"`
	Type        string          `json:"type"`
	Links       Links           `json:"links,omitempty"`
}

// PipelineState represents the state of a pipeline
type PipelineState struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Result *PipelineResult `json:"result,omitempty"`
}

// PipelineResult represents the result of a pipeline
type PipelineResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PipelineTarget represents what the pipeline is building
type PipelineTarget struct {
	Type       string           `json:"type"`
	RefType    string           `json:"ref_type,omitempty"`
	RefName    string           `json:"ref_name,omitempty"`
	Commit     *CommitReference `json:"commit,omitempty"`
	Selector   *PipelineSelector `json:"selector,omitempty"`
}

// PipelineSelector represents pipeline selector configuration
type PipelineSelector struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern,omitempty"`
}

// Links represents HAL-style links in API responses
type Links struct {
	Self       *Link `json:"self,omitempty"`
	HTML       *Link `json:"html,omitempty"`
	Avatar     *Link `json:"avatar,omitempty"`
	Commits    *Link `json:"commits,omitempty"`
	Diff       *Link `json:"diff,omitempty"`
	Approve    *Link `json:"approve,omitempty"`
	Comments   *Link `json:"comments,omitempty"`
	Activity   *Link `json:"activity,omitempty"`
	Statuses   *Link `json:"statuses,omitempty"`
	Decline    *Link `json:"decline,omitempty"`
	Merge      *Link `json:"merge,omitempty"`
}

// Link represents a single HAL link
type Link struct {
	Href string `json:"href"`
	Name string `json:"name,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	PageLen  int    `json:"pagelen"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// RepositoryList represents a paginated list of repositories
type RepositoryList struct {
	PaginatedResponse
	Values []Repository `json:"values"`
}

// PullRequestList represents a paginated list of pull requests
type PullRequestList struct {
	PaginatedResponse
	Values []PullRequest `json:"values"`
}

// CommentList represents a paginated list of comments
type CommentList struct {
	PaginatedResponse
	Values []Comment `json:"values"`
}

// PipelineList represents a paginated list of pipelines
type PipelineList struct {
	PaginatedResponse
	Values []Pipeline `json:"values"`
}

// CommitStatusList represents a paginated list of commit statuses
type CommitStatusList struct {
	PaginatedResponse
	Values []CommitStatus `json:"values"`
}

// FileStatsList represents a paginated list of file stats
type FileStatsList struct {
	PaginatedResponse
	Values []FileStats `json:"values"`
}

// Error represents a Bitbucket API error response
type Error struct {
	Type      string       `json:"type"`
	ErrorInfo ErrorDetail  `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Message string                 `json:"message"`
	Detail  string                 `json:"detail,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

func (e *Error) Error() string {
	if e.ErrorInfo.Detail != "" {
		return e.ErrorInfo.Message + ": " + e.ErrorInfo.Detail
	}
	return e.ErrorInfo.Message
}
