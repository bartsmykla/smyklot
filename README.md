# Smyklot

GitHub App for automated PR approvals and merges based on CODEOWNERS files.

## Features

- **Command-based PR management** via issue comments
- **Permission system** using `.github/CODEOWNERS`
- **Automated feedback** with reactions and comments
- **Security-first design** following GitHub Actions best practices

## Commands

Smyklot responds to commands in PR comments:

| Command | Aliases | Description |
|---------|---------|-------------|
| `/approve` | `@smyklot approve` | Approve the pull request |
| `/merge` | `@smyklot merge` | Merge the pull request |

## Setup

### Prerequisites

- Go 1.23+
- GitHub repository with Actions enabled
- `.github/CODEOWNERS` file

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/bartsmykla/smyklot.git
   cd smyklot
   ```

2. Install dependencies:

   ```bash
   task deps
   ```

3. Run tests:

   ```bash
   task test
   ```

### Configuration

Create a `.github/CODEOWNERS` file in your repository:

```text
# Global owners (can approve/merge any PR)
* @username1 @username2

# Path-specific owners (Phase 2)
/docs/ @doc-team
*.js @js-team
```

**Phase 1**: Only global owners (`*` pattern) are supported.

### GitHub Actions Setup

Add the workflows to your repository:

1. **Test workflow** (`.github/workflows/test.yml`):

   Runs tests and linters on push/PR.

2. **PR Commands workflow** (`.github/workflows/pr-commands.yml`):

   Handles `/approve` and `/merge` commands in PR comments.

The action binary is built automatically and executes commands based on
permissions defined in `.github/CODEOWNERS`.

## Usage

### Approving a PR

Comment on a PR:

```text
/approve
```

or

```text
@smyklot approve
```

Smyklot will:

1. Check if you're listed as a global owner in `.github/CODEOWNERS`
2. Approve the PR via GitHub API
3. Add ✅ reaction and post success comment

### Merging a PR

Comment on a PR:

```text
/merge
```

or

```text
@smyklot merge
```

Smyklot will:

1. Check if you're listed as a global owner
2. Verify the PR is mergeable (no conflicts, checks passed)
3. Merge the PR via GitHub API
4. Add ✅ reaction and post success comment

## Development

### Project Structure

```text
smyklot/
├── cmd/
│   └── github-action/      # Action entrypoint binary
├── pkg/
│   ├── commands/           # Command parser
│   ├── feedback/           # User feedback messages
│   ├── github/             # GitHub API client
│   └── permissions/        # CODEOWNERS parser and checker
├── .github/workflows/      # GitHub Actions workflows
├── Taskfile.yaml           # Task automation
└── go.mod                  # Go module definition
```

### Available Tasks

```bash
task                 # Show available tasks
task test            # Run all tests
task test:unit       # Run unit tests only
task lint            # Run all linters
task build           # Build all binaries
task clean           # Clean build artifacts
```

### Running Tests

All tests use Ginkgo/Gomega:

```bash
# Run all tests with coverage
task test

# Run specific package tests
ginkgo -r pkg/commands/

# Watch mode for TDD
task test:watch
```

### Linting

```bash
# Run all linters
task lint

# Individual linters
task lint:go         # golangci-lint
task lint:markdown   # markdownlint
task lint:mod        # go mod verify
```

## Architecture

### Permission System

Smyklot uses GitHub's `.github/CODEOWNERS` file for permissions:

- **Global owners** (`* @username`): Can approve/merge any PR
- **Path-specific owners** (Phase 2): Can approve/merge changes in their scope

### Security

- All GitHub Actions inputs passed as environment variables
- No shell interpolation of user-controlled data
- Token-based authentication with GitHub API
- Comprehensive error handling and validation

### Workflow

1. User comments `/approve` or `/merge` on a PR
2. GitHub Actions triggers the `pr-commands.yml` workflow
3. Action binary reads environment variables
4. Parses command and checks permissions
5. Executes GitHub API operations
6. Posts feedback comments and reactions

## Testing

Current test coverage: **98 tests passing**

- 20 command parser tests
- 30 feedback system tests
- 18 GitHub client tests
- 30 permission system tests

All tests follow TDD methodology with Ginkgo BDD-style specs.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests first (TDD)
4. Implement the feature
5. Ensure `task lint && task test` passes
6. Submit a pull request

## License

MIT License - see LICENSE file for details
