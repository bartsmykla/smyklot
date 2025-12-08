# Smyklot

> GitHub App for automated PR approvals and merges based on CODEOWNERS

## Overview

Smyklot is a GitHub App that automates pull request approvals and merges by validating permissions against your repository's CODEOWNERS file. Simply comment on a PR with commands or add emoji reactions, and Smyklot handles the rest.

## Features

- **CODEOWNERS-based permissions** - Only repository owners defined in `.github/CODEOWNERS` can approve/merge
- **Multiple command formats** - Supports slash commands (`/approve`), mentions (`@smyklot approve`), and bare commands (`lgtm`, `merge`)
- **Merge method control** - Explicit `squash` and `rebase` commands with intelligent fallback for `merge`
- **Reaction-based commands** - Use 👍 to approve, 🚀 to merge, or ❤️ to cleanup
- **Cleanup command** - Remove all bot reactions, approvals, and comments with a single command
- **Approval deduplication** - Prevents duplicate approvals with smart reaction handling
- **Flexible configuration** - Configure via `SMYKLOT_CONFIG` JSON, individual variables, or environment variables
- **Emoji feedback** - Get instant visual confirmation with ✅ (success), ❌ (error), or ⚠️ (warning)
- **Comment edit/delete handling** - Reacts to command edits and removes reactions when commands are deleted
- **Reaction removal tracking** - Automatically removes approvals/merges when reactions are removed
- **Multi-command support** - Execute multiple commands in a single comment
- **Security-first design** - Follows GitHub Actions best practices with minimal permissions
- **Zero external dependencies** - Runs entirely on GitHub Actions
- **TDD implementation** - 130+ passing tests

## Quick Start

### Prerequisites

- GitHub repository with Actions enabled
- `.github/CODEOWNERS` file in your repository

### Installation

#### Option 1: GitHub App (Recommended)

1. Install the [Smyklot GitHub App](https://github.com/apps/smyklot) on your repository
2. Create `.github/CODEOWNERS`:

   ```text
   * @username
   ```

3. Comment on PRs with commands or add reactions

#### Option 2: Manual Workflow

Copy the workflow file to your repository:

```bash
cp .github/workflows/pr-commands.yaml your-repo/.github/workflows/
```

## Usage

### Comment Commands

Smyklot responds to these commands in PR comments:

| Command     | Aliases          | Format                                    | Description                                       |
|-------------|------------------|-------------------------------------------|---------------------------------------------------|
| `approve`   | `lgtm`, `accept` | `/approve`, `@smyklot approve`, `approve` | Approve the pull request                          |
| `merge`     | -                | `/merge`, `@smyklot merge`, `merge`       | Merge the pull request (with fallback)            |
| `squash`    | -                | `/squash`, `@smyklot squash`, `squash`    | Squash merge the pull request                     |
| `rebase`    | -                | `/rebase`, `@smyklot rebase`, `rebase`    | Rebase merge the pull request                     |
| `unapprove` | -                | `/unapprove`, `@smyklot unapprove`        | Remove approval                                   |
| `cleanup`   | -                | `/cleanup`, `@smyklot cleanup`, `cleanup` | Remove all bot reactions, approvals, and comments |
| `help`      | -                | `/help`, `@smyklot help`                  | Show help information                             |

**Command Formats**:

- **Slash commands**: `/approve`, `/merge`, `/squash`, `/rebase`, `/unapprove`, `/cleanup`, `/help`
- **Mention commands**: `@smyklot approve`, `@smyklot merge`, `@smyklot squash`, `@smyklot rebase`, `@smyklot unapprove`, `@smyklot cleanup`
- **Bare commands**: `approve`, `accept`, `lgtm`, `merge`, `squash`, `rebase`, `cleanup` (exact match only)

All commands are case-insensitive.

### Reaction Commands

| Reaction | Action                                                      |
|----------|-------------------------------------------------------------|
| 👍       | Approve the pull request                                    |
| 🚀       | Merge the pull request                                      |
| ❤️       | Cleanup (remove all bot reactions, approvals, and comments) |

**Note**: Reactions must be added to the PR description, not to comments. Removing a reaction will automatically undo the corresponding action (remove approval/merge labels).

### Multiple Commands

You can use multiple non-contradicting commands in a single comment:

```text
/approve
/merge
```

or

```text
lgtm merge
```

Commands will be executed in order: approve first, then merge.

### Examples

#### Approving a PR

Any of these will approve the PR:

```text
/approve
```

```text
@smyklot approve
```

```text
lgtm
```

Or add a 👍 reaction to the PR description.

#### Merging a PR

Merge with default method (with fallback to squash/rebase if merge commits disallowed):

```text
/merge
```

Or add a 🚀 reaction to the PR description.

#### Squash Merging

```text
/squash
```

#### Rebase Merging

```text
/rebase
```

#### Removing Approval

```text
/unapprove
```

Or remove your 👍 reaction from the PR description.

#### Cleanup

Remove all bot reactions, approvals, and comments:

```text
/cleanup
```

Or add a ❤️ reaction to the PR description.

**Note**: Cleanup cannot be combined with other commands.

## Configuration

### CODEOWNERS Setup

Create `.github/CODEOWNERS` in your repository:

```text
# Global owners can approve/merge any PR
* @username1 @username2
```

Currently only global owners (`*` pattern) are supported. Path-specific owners will be added in Phase 2.

### Bot Configuration

Smyklot can be configured via repository variables (Settings → Secrets and variables → Actions → Variables).

**Configuration Precedence**: CLI flags > Environment variables > Repository variables > Defaults

#### Option 1: Full JSON Configuration (Recommended)

Set a `SMYKLOT_CONFIG` repository variable with your complete configuration:

```json
{
  "quiet_success": false,
  "quiet_reactions": false,
  "quiet_pending": false,
  "allowed_commands": ["approve", "merge"],
  "command_aliases": {
    "ok": "approve",
    "ship": "merge"
  },
  "command_prefix": "/",
  "disable_mentions": false,
  "disable_bare_commands": false,
  "disable_unapprove": false,
  "disable_reactions": false,
  "disable_deleted_comments": false,
  "allow_self_approval": false
}
```

#### Option 2: Individual Variables

Configure individual settings via repository variables or environment variables with `SMYKLOT_` prefix:

| Variable                           | Type    | Default        | Description                                                                |
|------------------------------------|---------|----------------|----------------------------------------------------------------------------|
| `SMYKLOT_QUIET_SUCCESS`            | boolean | `false`        | Disable success feedback comments                                          |
| `SMYKLOT_QUIET_REACTIONS`          | boolean | `false`        | Disable reaction-based approval/merge comments                             |
| `SMYKLOT_QUIET_PENDING`            | boolean | `false`        | Disable pending CI comments (reactions only for "merge after CI")          |
| `SMYKLOT_ALLOWED_COMMANDS`         | list    | all            | Limit which commands are allowed                                           |
| `SMYKLOT_COMMAND_ALIASES`          | map     | default        | Define custom command aliases                                              |
| `SMYKLOT_COMMAND_PREFIX`           | string  | `/`            | Custom command prefix                                                      |
| `SMYKLOT_DISABLE_MENTIONS`         | boolean | `false`        | Disable mention commands                                                   |
| `SMYKLOT_DISABLE_BARE_COMMANDS`    | boolean | `false`        | Disable bare commands                                                      |
| `SMYKLOT_DISABLE_UNAPPROVE`        | boolean | `false`        | Disable unapprove command                                                  |
| `SMYKLOT_DISABLE_REACTIONS`        | boolean | `false`        | Disable reaction-based approvals/merges                                    |
| `SMYKLOT_DISABLE_DELETED_COMMENTS` | boolean | `false`        | Disable handling of deleted comments                                       |
| `SMYKLOT_ALLOW_SELF_APPROVAL`      | boolean | `false`        | Allow PR authors to approve their own PRs                                  |
| `SMYKLOT_BOT_USERNAME`             | string  | `smyklot[bot]` | Bot username for cleanup operations (GitHub App format: `{app-slug}[bot]`) |

#### Configuration Examples

##### Example 1: Quiet Mode

Only show emoji reactions, no success comments:

```yaml
# In workflow or as repository variable
env:
  SMYKLOT_QUIET_SUCCESS: "true"
```

Result: User sees only ✅ reaction, no "PR Approved" comment.

##### Example 2: Custom Prefix

Use `!` instead of `/` for commands:

```yaml
env:
  SMYKLOT_COMMAND_PREFIX: "!"
```

Users can now use `!approve` and `!merge`.

##### Example 3: Command Aliases

Create shortcuts for commands:

```yaml
env:
  SMYKLOT_COMMAND_ALIASES: '{"app":"approve","a":"approve","m":"merge"}'
```

Users can use `/app`, `/a`, or `/m` as shortcuts.

##### Example 4: Reactions Only

Disable comment-based commands, only allow reactions:

```json
{
  "disable_mentions": true,
  "disable_bare_commands": true,
  "command_prefix": "disabled"
}
```

Only 👍 and 🚀 reactions will work.

##### Example 5: Disable Reaction Tracking

Don't remove approvals/merges when reactions are removed:

```yaml
env:
  SMYKLOT_DISABLE_DELETED_COMMENTS: "true"
```

##### Example 6: Allow Self-Approval

**⚠️ Not recommended for production** - Allow PR authors to approve their own PRs:

```yaml
env:
  SMYKLOT_ALLOW_SELF_APPROVAL: "true"
```

or via JSON:

```json
{
  "allow_self_approval": true
}
```

**Security Note**: By default, Smyklot prevents self-approval to enforce separation of duties. Only enable this in development/testing environments.

## Architecture

### How It Works

1. User comments a command or adds a reaction on a PR
2. GitHub triggers `issue_comment` webhook
3. `pr-commands.yaml` workflow starts
4. Smyklot:
   - Parses the command (supports slash, mention, and bare formats)
   - Or processes reactions (👍 for approve, 🚀 for merge, ❤️ for cleanup)
   - Validates command combinations (cleanup cannot be combined with others)
   - Checks for duplicate approvals (prevents re-approving for both commands and reactions)
   - Fetches `.github/CODEOWNERS` via GitHub API
   - Checks user permissions
   - Calls GitHub API to approve/merge/cleanup
   - For merge: tries specified method or falls back (merge → squash → rebase)
   - Posts reactions and feedback
5. On comment edit/delete or reaction removal, updates accordingly

### Permission System

**Phase 1 (Current)**:

- Only global owners (`* @username`) are supported
- Global owners can approve/merge any PR
- Reaction-based approvals/merges with tracking
- Self-approval prevention (configurable, disabled by default)
- Fail-closed CODEOWNERS parsing (returns error if file is corrupted)

**Phase 2 (Planned)**:

- Path-specific ownership patterns
- Scoped permissions based on changed files
- Team support (`@org/team-name`)
- Required approvals count

### Security

Smyklot implements defense-in-depth security practices:

**Input Validation**:

- Comment body length validation (max 10KB) - prevents DoS attacks
- Repository owner/name format validation - prevents path traversal
- All inputs passed via environment variables (no shell interpolation)

**API Security**:

- Parameterized GraphQL queries - prevents injection attacks
- HTTP client timeout (30s) - prevents hung requests
- Exponential backoff retry logic - handles rate limiting gracefully
- Connection pooling - optimizes resource usage

**Access Control**:

- CODEOWNERS-based authorization with fail-closed parsing
- Self-approval prevention (configurable, disabled by default)
- Minimal workflow permissions (contents: read, pull-requests: write)
- Token-based authentication via GitHub App

**Data Protection**:

- Sensitive data sanitization in logs (tokens, keys, secrets redacted)
- Maximum CODEOWNERS file size (1MB) - prevents memory exhaustion
- No repository checkout required (CODEOWNERS fetched via API)

**Supply Chain Security**:

- Actions pinned by commit digest
- Go dependencies verified (`go mod verify`)
- Docker images use minimal base (`FROM scratch`)

## Development

### Requirements

- Go 1.25+
- [mise](https://mise.jdx.dev/) for tool management
- [Task](https://taskfile.dev/) for task automation

### Setup

```bash
# Clone repository
git clone https://github.com/smykla-labs/smyklot.git
cd smyklot

# Install tools
mise install

# Download dependencies
go mod download

# Run tests
task test
```

### Project Structure

```text
smyklot/
├── cmd/
│   └── github-action/       # GitHub Actions entrypoint
├── pkg/
│   ├── commands/            # Command parser (slash, mention, bare)
│   ├── config/              # Configuration management (Viper)
│   ├── feedback/            # User feedback system (reactions, comments)
│   ├── github/              # GitHub API client
│   └── permissions/         # CODEOWNERS parser & permission checker
├── .github/workflows/       # GitHub Actions workflows
├── .goreleaser.yml          # GoReleaser config for releases
├── .mise.toml               # Tool versions
├── Dockerfile               # Docker image for GitHub Actions
├── Taskfile.yaml            # Task automation
└── go.mod                   # Go module definition
```

### Available Tasks

```bash
task             # Show available tasks
task test        # Run all tests with coverage
task test:unit   # Run unit tests only
task lint        # Run all linters
task build       # Build binaries
task clean       # Clean build artifacts
```

### Testing

All tests use Ginkgo/Gomega BDD framework:

```bash
# Run all tests
task test

# Run specific package
ginkgo -r pkg/commands/

# Watch mode for TDD
ginkgo watch -r
```

Current test coverage: 130+ tests passing

- 52+ command parser tests
- 12 CODEOWNERS parser tests
- 30 permission checker tests
- 30 feedback system tests
- 18+ GitHub client tests

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Write tests first (TDD)
4. Implement the feature
5. Ensure all checks pass: `task lint && task test`
6. Commit with conventional commits (`feat:`, `fix:`, `docs:`, etc.)
7. Push to your fork
8. Open a pull request

## Roadmap

### Phase 1: GitHub Actions Bot ✅

- [x] Command parser (slash, mention, bare)
- [x] Multi-command support
- [x] Merge method commands (merge, squash, rebase)
- [x] Merge method fallback (merge → squash → rebase)
- [x] Cleanup command (remove all bot reactions, approvals, comments)
- [x] Approval deduplication (prevent duplicate approvals)
- [x] Reaction-based approvals/merges/cleanup (👍, 🚀, ❤️)
- [x] Reaction removal tracking
- [x] Comment edit/delete handling
- [x] CODEOWNERS parser (global owners)
- [x] Permission checker
- [x] GitHub API client
- [x] Feedback system (emoji + comments)
- [x] Configuration system (Viper)
- [x] GitHub Actions workflows
- [x] Docker-based GitHub Action
- [x] Documentation

### Phase 2: Enhanced Permissions (Planned)

- [ ] Path-specific ownership patterns
- [ ] Scoped approval requirements based on changed files
- [ ] Team support in CODEOWNERS (`@org/team-name`)
- [x] Self-approval prevention (configurable)
- [ ] Required approvals count

### Phase 3: Kubernetes Deployment (Future)

**Prerequisites** (Security Hardening):

- [x] GraphQL injection prevention
- [x] HTTP client timeout and connection pooling
- [x] Rate limiting and retry logic
- [x] Input validation
- [x] Fail-closed CODEOWNERS parsing

**Remaining Work**:

- [ ] Refactor global mutable state to request-scoped parameters
- [ ] Add context.Context propagation throughout
- [ ] Implement HTTP webhook server
- [ ] Add concurrency tests with `-race` flag
- [ ] Implement comprehensive audit logging
- [ ] Kubernetes deployment (Helm chart)
- [ ] Prometheus metrics
- [ ] Migration strategy

**Estimated Effort**: 18-30 days

### Phase 4: Discord Integration (Future)

- [ ] Discord bot
- [ ] Unified command system
- [ ] Cross-platform notifications
- [ ] Status synchronization

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Built with:

- [Ginkgo](https://github.com/onsi/ginkgo) - BDD testing framework
- [Gomega](https://github.com/onsi/gomega) - Matcher library
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [mise](https://mise.jdx.dev/) - Tool version manager
- [Task](https://taskfile.dev/) - Task runner
- [GoReleaser](https://goreleaser.com/) - Release automation

---

Made with ❤️ by [@bartsmykla](https://github.com/bartsmykla)
