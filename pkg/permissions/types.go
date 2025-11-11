package permissions

// CodeownersFile represents a parsed CODEOWNERS file
type CodeownersFile struct {
	// Path is the file path to the CODEOWNERS file
	Path string

	// Entries contains all parsed ownership entries
	Entries []CodeownersEntry
}

// CodeownersEntry represents a single line in a CODEOWNERS file
type CodeownersEntry struct {
	// Pattern is the file/directory pattern (e.g., "*", "*.js", "/docs/")
	Pattern string

	// Owners is the list of owners for this pattern (without @ prefix)
	Owners []string
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
