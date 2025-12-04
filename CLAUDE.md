# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

Smyklot is a GitHub App for automated PR approvals and merges based on
CODEOWNERS files. It supports comment-based commands, merge method control,
cleanup operations, and reaction-based approvals (👍/🚀/❤️). Built in Go
using TDD methodology with Ginkgo/Gomega.

**Current Status**: Phase 1 complete (Docker-based GitHub Action)
**Test Coverage**: 181 tests passing (59.5% composite coverage)
**Deployment**: Docker image published to ghcr.io
**Security**: Hardened with defense-in-depth practices

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
│   ├── commands/            # Command parser (slash, mention, bare)
│   ├── config/              # Configuration management (Viper)
│   ├── permissions/         # CODEOWNERS parser & permission checker
│   ├── feedback/            # User feedback system (emoji + comments)
│   └── github/              # GitHub API client
├── .github/workflows/       # CI/CD workflows
├── .goreleaser.yml          # GoReleaser v2 configuration
├── .mise.toml               # Tool versions (pinned)
├── Dockerfile               # Docker image for GitHub Action
├── Taskfile.yaml            # Task automation
└── go.mod                   # Go dependencies
```

### Key Packages

#### `pkg/commands`

- Parses commands from PR comments
- **Commands**: `approve` (aliases: `lgtm`, `accept`), `merge`, `squash`, `rebase`, `unapprove`, `cleanup`, `help`
- **Formats**: Slash (`/approve`), mention (`@smyklot approve`), bare (`lgtm`, `merge`)
- **Multi-command support**: Parse multiple commands in single comment
- **Command validation**: Prevents cleanup from being combined with other commands
- **Approval deduplication**: Checks existing approvals to prevent duplicates
- **Configurable**: Custom aliases, prefix, disable specific formats
- Returns `Command` type with parsed actions
- 78+ tests covering all parsing scenarios including multi-command

#### `pkg/permissions`

- Parses `.github/CODEOWNERS` file
- **Phase 1**: Only global owners (`*` pattern) supported
- **Phase 2**: Path-specific patterns (future)
- 42 tests (12 parser + 30 checker)

#### `pkg/config`

- Configuration management using Viper
- **Sources**: CLI flags > Environment variables > Repository variables > Defaults
- **JSON support**: Full configuration via `SMYKLOT_CONFIG` variable
- **Individual variables**: `SMYKLOT_*` prefix for each setting
- **Options**: quiet modes, command restrictions, custom aliases, format toggles

#### `pkg/feedback`

- Creates user feedback messages
- Emoji reactions: ✅ (success), ❌ (error), ⚠️ (warning), 👀 (processing)
- Comments for errors/warnings only
- 30 tests covering all feedback types

#### `pkg/github`

- GitHub API client
- **Methods**: `AddReaction`, `RemoveReaction`, `PostComment`, `DeleteComment`,
  `ApprovePR`, `DismissReview`, `MergePR`, `GetPRInfo`, `GetPRComments`,
  `GetCodeowners`, `GetCommentReactions`, `GetLabels`, `GetAuthenticatedUser`
- **GetPRInfo**: Fetches PR data and populates `ApprovedBy` from reviews
- **Merge methods**: Supports merge, squash, and rebase with fallback logic
- **Reaction support**: 👍 (approve), 🚀 (merge), ❤️ (cleanup) with removal tracking
- **Approval deduplication**: Both command and reaction handlers check existing approvals
- Uses GitHub App token from environment
- 18+ tests with httptest mocking

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

**Note**: All workflow files use `.yaml` extension (not `.yml`) for consistency.

### `test.yaml`

Runs on: push to `main`, `feat/**`, and PRs

Steps:

1. Checkout code (`actions/checkout@v5.0.0`)
2. Install mise (`jdx/mise-action@v3.4.0`)
3. Run tests (`ginkgo -r --race --cover`)
4. Run linters (`golangci-lint-action@v9.0.0`)
5. Verify modules (`go mod verify`)

### `pr-commands.yaml`

Triggered by: `issue_comment` on PRs

Uses: Local action reference (`uses: ./`) for latest version

Environment variables (from GitHub context):

- `GITHUB_TOKEN` - GitHub API token
- `COMMENT_BODY` - Comment text
- `COMMENT_ID` - Comment ID
- `PR_NUMBER` - PR number
- `REPO_OWNER` - Repository owner
- `REPO_NAME` - Repository name
- `COMMENT_AUTHOR` - Comment author username

Steps:

1. Checkout full repository
2. Run local action (builds and executes binary)
3. Action uses environment variables for all inputs

### Security Practices

- Actions pinned by commit digest (not tags)
- Environment variables for untrusted input (not CLI args)
- Minimal permissions (contents: read, pull-requests: write)
- No shell interpolation of user data
- Ubuntu 24.04 for latest security patches

## Current Implementation

### Phase 1: Docker-based GitHub Action ✅

**Completed**:

- [x] Command parser (slash, mention, bare formats)
- [x] Commands: `approve`, `merge`, `squash`, `rebase`, `unapprove`, `cleanup`, `help` with aliases
- [x] Merge method control (explicit squash/rebase with fallback)
- [x] Cleanup command (remove all bot reactions, approvals, comments)
- [x] Approval deduplication (prevent duplicate approvals for both commands and reactions)
- [x] Multi-command support (multiple commands in single comment)
- [x] Command validation (cleanup cannot be combined with others)
- [x] Reaction-based approvals/merges/cleanup (👍 approve, 🚀 merge, ❤️ cleanup)
- [x] Reaction removal tracking (auto-remove approvals/merges)
- [x] Reaction-based approval deduplication (GetPRInfo populates ApprovedBy)
- [x] Comment edit/delete handling
- [x] CODEOWNERS parser (global owners only)
- [x] CODEOWNERS API fetching (no repository checkout)
- [x] Permission checker (global ownership with fail-closed parsing)
- [x] Configuration system (Viper with JSON/individual variables)
- [x] Self-approval prevention (configurable, default: disabled)
- [x] Security hardening (GraphQL injection prevention, rate limiting, input validation)
- [x] Feedback system (reactions + comments)
- [x] GitHub API client (full CRUD operations with retry logic)
- [x] GitHub App integration with token generation
- [x] Docker-based GitHub Action (ghcr.io)
- [x] GoReleaser v2 automated releases
- [x] GitHub Actions workflows (test, release)
- [x] Documentation (README, CLAUDE.md, GitHub App description)
- [x] 181 tests passing

**Not Implemented**:

- [ ] Path-specific CODEOWNERS patterns (Phase 2)
- [ ] Team support in CODEOWNERS (Phase 2)
- [ ] Required approvals count (Phase 2)

### Phase 2: Enhanced Permissions (Planned)

- Path-specific ownership patterns
- Scoped approval based on changed files
- Team support (`@org/team-name`)
- Required approvals count

### Phase 3: Kubernetes Deployment (Future)

**Prerequisites** (from security investigation):

- [x] GraphQL injection prevention (COMPLETED - parameterized queries)
- [x] HTTP client timeout and connection pooling (COMPLETED)
- [x] Rate limiting and retry logic (COMPLETED)
- [x] Input validation (COMPLETED)
- [x] Fail-closed CODEOWNERS parsing (COMPLETED)
- [x] Refactor global mutable state to request-scoped parameters (COMPLETED - minimal state found)
- [x] Add context.Context propagation throughout (COMPLETED - all 35+ client methods + handlers)
- [ ] Implement structured logging with slog
- [ ] Add request ID propagation through context
- [ ] Add concurrency tests with `-race` flag
- [ ] Implement comprehensive audit logging

**Deployment**:

- [ ] HTTP webhook server
- [ ] Persistent service deployment
- [ ] Scalable architecture
- [ ] Prometheus metrics
- [ ] Helm chart

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

- **Commits**: Conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`)
- **Commit flags**: Always use `-sS` (sign-off + GPG sign)
- **Branching and PRs**: Follow global `CLAUDE.md` guidelines

## Releases

**Process**: Fully automated via GitHub Actions (see `RELEASING.md` for details)

- **Auto-release**: Runs daily at 10:00 UTC, analyzes commits, bumps version
- **Versioning**: Semantic versioning based on conventional commits
  - `feat:` → minor bump (0.X.0)
  - `fix:` → patch bump (0.0.X)
  - `feat!:` or `BREAKING CHANGE` → major bump (X.0.0)
- **Artifacts**: Docker image (`ghcr.io/smykla-labs/smyklot:X.Y.Z`), binaries
- **Manual trigger**: `gh workflow run auto-release.yaml`
- **Version files**: `action.yml` (Docker image tag), Git tags (source of truth)

**Important**: Use conventional commit format to ensure proper version bumping.

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

## Security Features

**Input Validation**:

- Comment body length: max 10KB (prevents DoS)
- Repository name format: alphanumeric + hyphens (prevents path traversal)
- CODEOWNERS file size: max 1MB (prevents memory exhaustion)

**API Security**:

- GraphQL queries: parameterized (prevents injection)
- HTTP client: 30s timeout with connection pooling
- Retry logic: exponential backoff for 429/5xx errors
- Rate limiting: automatic handling with backoff

**Access Control**:

- CODEOWNERS parsing: fail-closed (errors on parse failure)
- Self-approval: configurable prevention (default: disabled)
- Logging: WARNING when falling back to admin permissions

**Data Protection**:

- Step summary: sanitizes secrets (token, key, secret, password)
- Comment body: truncated and redacted in logs

## Configuration

All configuration options support three input methods with precedence:
**CLI flags > Environment variables > JSON config > Defaults**

### JSON Configuration

```json
{
  "quiet_success": false,
  "quiet_reactions": false,
  "allowed_commands": [],
  "command_aliases": {},
  "command_prefix": "/",
  "disable_mentions": false,
  "disable_bare_commands": false,
  "disable_unapprove": false,
  "disable_reactions": false,
  "disable_deleted_comments": false,
  "allow_self_approval": false
}
```

### Environment Variables

All variables prefixed with `SMYKLOT_`:

- `SMYKLOT_QUIET_SUCCESS`
- `SMYKLOT_QUIET_REACTIONS`
- `SMYKLOT_ALLOWED_COMMANDS`
- `SMYKLOT_COMMAND_ALIASES`
- `SMYKLOT_COMMAND_PREFIX`
- `SMYKLOT_DISABLE_MENTIONS`
- `SMYKLOT_DISABLE_BARE_COMMANDS`
- `SMYKLOT_DISABLE_UNAPPROVE`
- `SMYKLOT_DISABLE_REACTIONS`
- `SMYKLOT_DISABLE_DELETED_COMMENTS`
- `SMYKLOT_ALLOW_SELF_APPROVAL`
- `SMYKLOT_BOT_USERNAME`

### CLI Flags

Use `--` prefix with kebab-case, e.g., `--allow-self-approval`

## Important Notes

- **CODEOWNERS format**: Line-based, not YAML (GitHub standard)
- **Global owners only**: Phase 1 supports only `*` pattern
- **Self-approval**: Disabled by default (configurable)
- **Emoji-only success**: Success feedback uses reaction only (no comment)
- **Comments for errors**: Errors/warnings post both reaction and comment
- **Environment variables**: All inputs from GitHub context, not CLI args
- **No external dependencies**: Runs entirely on GitHub Actions
- **Stateless**: Each command execution is independent
- **Phase 3 readiness**: Requires refactoring for concurrent execution

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
