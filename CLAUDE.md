# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

Smyklot is a GitHub Actions bot for automated PR approvals and merges based on CODEOWNERS files. It's built in Go using TDD methodology with Ginkgo/Gomega.

**Current Status**: Phase 1 complete (GitHub Actions implementation)
**Test Coverage**: 98/98 tests passing (100%)

## Architecture

### Technology Stack

- **Language**: Go 1.25+
- **Testing**: Ginkgo v2 + Gomega (BDD style)
- **Tool Management**: mise
- **Task Runner**: Task (not Make)
- **Deployment**: GitHub Actions (no external hosting)

### Directory Structure

```text
smyklot/
├── cmd/github-action/       # GitHub Actions entrypoint binary
├── pkg/
│   ├── commands/            # Command parser (slash + mention)
│   ├── permissions/         # CODEOWNERS parser & permission checker
│   ├── feedback/            # User feedback system (emoji + comments)
│   └── github/              # GitHub API client
├── .github/workflows/       # CI/CD workflows
├── .mise.toml               # Tool versions (pinned)
├── Taskfile.yaml            # Task automation
└── go.mod                   # Go dependencies
```

### Key Packages

#### `pkg/commands`

- Parses commands from PR comments
- Supports: `/approve`, `/merge`, `@smyklot approve`, `@smyklot merge`
- Returns `Command` type with parsed action
- 20 tests covering all parsing scenarios

#### `pkg/permissions`

- Parses `.github/CODEOWNERS` file
- **Phase 1**: Only global owners (`*` pattern) supported
- **Phase 2**: Path-specific patterns (future)
- 42 tests (12 parser + 30 checker)

#### `pkg/feedback`

- Creates user feedback messages
- Emoji reactions: ✅ (success), ❌ (error), ⚠️ (warning)
- Comments for errors/warnings only
- 30 tests covering all feedback types

#### `pkg/github`

- GitHub API client
- Methods: `AddReaction`, `PostComment`, `ApprovePR`, `MergePR`, `GetPRInfo`
- Uses `GITHUB_TOKEN` from environment
- 18 tests with httptest mocking

## Development Workflow

### Testing

All code follows TDD (Test-Driven Development):

1. Write test first (Red)
2. Implement minimum code (Green)
3. Refactor (Refactor)

Run tests:

```bash
task test              # All tests with coverage
task test:unit         # Unit tests only
ginkgo -r pkg/commands # Specific package
```

### Linting

```bash
task lint              # All linters
task lint:go           # golangci-lint only
task lint:markdown     # markdownlint only
```

Linters used:
- `golangci-lint` v2.6.1 (Go code)
- `markdownlint` (Markdown files)

### Building

```bash
task build             # Build all binaries
task clean             # Clean artifacts
```

Binary output: `bin/smyklot-github-action`

## Code Style

### Go Conventions

Follow global `CLAUDE.md` Go style with these specifics:

**Error Handling**:
- Sentinel errors: `var ErrOpName = errors.New("msg")`
- Custom errors with `Error()`, `Unwrap()`, `Is()` methods
- See `pkg/permissions/errors.go` for patterns

**Test Organization**:
- Use Ginkgo `Describe/Context/It` BDD structure
- Tag tests with `[Unit]` or `[Integration]`
- Example: `var _ = Describe("Parser [Unit]", func() { ... })`

**Imports**:
- Standard library first
- External packages second
- Internal packages last
- Use `github.com/pkg/errors` for error wrapping

### Markdown Style

- Empty line after ALL headers
- Use tables for command lists
- Backticks for code, commands, filenames
- Emoji: ✅ (success), ❌ (failure), ⚠️ (warning)

## GitHub Actions Workflows

### `test.yml`

Runs on: push to `main`, `feat/**`, and PRs

Steps:
1. Checkout code (`actions/checkout@v5.0.0`)
2. Install mise (`jdx/mise-action@v3.4.0`)
3. Run tests (`ginkgo -r --race --cover`)
4. Run linters (`golangci-lint-action@v9.0.0`)
5. Verify modules (`go mod verify`)

### `pr-commands.yml`

Triggered by: `issue_comment` on PRs

Environment variables (from GitHub context):
- `GITHUB_TOKEN` - GitHub API token
- `COMMENT_BODY` - Comment text
- `COMMENT_ID` - Comment ID
- `PR_NUMBER` - PR number
- `REPO_OWNER` - Repository owner
- `REPO_NAME` - Repository name
- `COMMENT_AUTHOR` - Comment author username

Steps:
1. Checkout code
2. Install mise
3. Build binary: `go build -o bin/smyklot-github-action ./cmd/github-action`
4. Execute binary with environment variables

### Security Practices

- Actions pinned by commit digest (not tags)
- Environment variables for untrusted input (not CLI args)
- Minimal permissions (contents: read, pull-requests: write)
- No shell interpolation of user data
- Ubuntu 24.04 for latest security patches

## Current Implementation

### Phase 1: GitHub Actions Bot ✅

**Completed**:
- [x] Command parser (`/approve`, `/merge`, mentions)
- [x] CODEOWNERS parser (global owners only)
- [x] Permission checker (global ownership)
- [x] Feedback system (reactions + comments)
- [x] GitHub API client (approve, merge, reactions)
- [x] GitHub Actions workflows
- [x] Documentation (README, CLAUDE.md)
- [x] 98 tests passing (100% coverage)

**Not Implemented**:
- [ ] Path-specific CODEOWNERS patterns (Phase 2)
- [ ] Self-approval prevention (Phase 2)
- [ ] Team support in CODEOWNERS (Phase 2)

### Phase 2: Enhanced Permissions (Planned)

- Path-specific ownership patterns
- Scoped approval based on changed files
- Team support (`@org/team-name`)
- Prevent self-approval

### Phase 3: Kubernetes Deployment (Future)

- HTTP webhook server
- Persistent service deployment
- Scalable architecture

### Phase 4: Discord Integration (Future)

- Discord bot
- Unified command system
- Cross-platform notifications

## Common Tasks

### Adding a New Command

1. Add test in `pkg/commands/parser_test.go`
2. Implement in `pkg/commands/parser.go`
3. Add command type to `pkg/commands/types.go`
4. Add handler in `cmd/github-action/main.go`
5. Update README command table

### Adding a New Feedback Type

1. Add test in `pkg/feedback/feedback_test.go`
2. Implement `New*` function in `pkg/feedback/feedback.go`
3. Use in command handlers

### Modifying GitHub API

1. Add/update test in `pkg/github/client_test.go`
2. Implement in `pkg/github/client.go`
3. Use httptest for mocking in tests

## Debugging

### Running Locally

The binary expects environment variables from GitHub Actions:

```bash
export GITHUB_TOKEN="your-token"
export COMMENT_BODY="/approve"
export COMMENT_ID="123"
export PR_NUMBER="45"
export REPO_OWNER="bartsmykla"
export REPO_NAME="test-repo"
export COMMENT_AUTHOR="username"

./bin/smyklot-github-action
```

### Test Output

```bash
# Verbose test output
ginkgo -v -r

# Failed tests only
ginkgo --fail-fast -r

# Watch mode
ginkgo watch -r
```

## Dependencies

### Direct Dependencies

```go
require (
    github.com/onsi/ginkgo/v2 v2.27.2   # BDD testing framework
    github.com/onsi/gomega v1.38.2      # Matcher library
    github.com/pkg/errors v0.9.1        # Error wrapping
)
```

All others are indirect (transitive dependencies).

### Tool Versions

Managed via `.mise.toml`:

```toml
go = "1.25.4"                              # Latest stable
task = "3.42.0"                            # Task runner
"npm:markdownlint-cli" = "0.45.0"          # Markdown linter
"go:github.com/onsi/ginkgo/v2/ginkgo" = "latest" # Test runner
```

## Git Workflow

- **Branching**: `feat/*` for features, `fix/*` for fixes
- **Commits**: Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`)
- **Commit flags**: Always use `-sS` (sign-off + GPG sign)
- **PR workflow**: Follow global `CLAUDE.md` guidelines

## Testing on .dotfiles Repository

The workflows can be tested on `bartsmykla/.dotfiles`:

1. Copy workflows to `.dotfiles/.github/workflows/`
2. Create `.dotfiles/.github/CODEOWNERS`:
   ```text
   * @bartsmykla
   ```
3. Open a test PR
4. Comment `/approve` or `/merge`
5. Verify reactions and API calls

## Important Notes

- **CODEOWNERS format**: Line-based, not YAML (GitHub standard)
- **Global owners only**: Phase 1 supports only `*` pattern
- **Emoji-only success**: Success feedback uses reaction only (no comment)
- **Comments for errors**: Errors/warnings post both reaction and comment
- **Environment variables**: All inputs from GitHub context, not CLI args
- **No external dependencies**: Runs entirely on GitHub Actions
- **Stateless**: Each command execution is independent

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Task Documentation](https://taskfile.dev/)
- [mise Documentation](https://mise.jdx.dev/)

## Questions?

When working on this codebase:

1. **Follow TDD**: Write tests before implementation
2. **Check existing patterns**: Look at similar code in the same package
3. **Run tests frequently**: `task test` should always pass
4. **Update documentation**: Keep README and this file in sync
5. **Ask before major changes**: Especially for architecture decisions
