package permissions

// OwnersFile represents the structure of an OWNERS file
type OwnersFile struct {
	// Approvers is a list of GitHub usernames who can approve changes
	Approvers []string `yaml:"approvers"`

	// Path is the directory path this OWNERS file applies to
	// An empty string indicates the root OWNERS file
	Path string `yaml:"-"`
}

// PermissionScope represents the scope of a user's permissions
type PermissionScope struct {
	// Username is the GitHub username
	Username string

	// Paths is a list of paths the user can approve
	// An empty list indicates root-level approval (all paths)
	Paths []string

	// IsRootApprover indicates whether the user is a root approver
	IsRootApprover bool
}
