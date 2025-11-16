# Smyklot

> GitHub Actions bot for automated PR approvals and merges based on CODEOWNERS

## Overview

Smyklot is a lightweight GitHub Actions bot that enables team members to
approve and merge pull requests through simple commands, with permissions
managed through GitHub's native `.github/CODEOWNERS` file.

## Features

- Command-based PR management via issue comments
- Permission system using `.github/CODEOWNERS`
- Flexible configuration (environment variables, CLI flags)
- Automated feedback with emoji reactions
- Security-first design following GitHub Actions best practices
- Zero external dependencies - runs entirely on GitHub Actions
- TDD implementation with 130 passing tests

## Quick Start

### Prerequisites

- GitHub repository with Actions enabled
- `.github/CODEOWNERS` file in your repository

### Installation

Copy the workflow files to your repository:

```bash
# Copy workflows
cp .github/workflows/pr-commands.yml your-repo/.github/workflows/
cp .github/workflows/test.yml your-repo/.github/workflows/
```

### Configuration

#### CODEOWNERS Setup

Create `.github/CODEOWNERS` in your repository:

```text
# Global owners can approve/merge any PR
* @username1 @username2
```

Currently only global owners (`*` pattern) are supported. Path-specific
owners will be added in Phase 2.

#### Bot Configuration

Smyklot supports multiple configuration sources with the following precedence:

CLI Flags > Environment Variables > Defaults

##### Available Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `quiet_success` | boolean | `false` | Emoji reactions only |
| `allowed_commands` | string list | `[]` (all) | Allowed commands list |
| `command_aliases` | string map | `{}` | Command aliases |
| `command_prefix` | string | `/` | Slash command prefix |
| `disable_mentions` | boolean | `false` | Disable mentions |
| `disable_bare_commands` | boolean | `false` | Disable bare commands |

##### Configuration Methods

###### Environment Variables

Use `SMYKLOT_` prefix with uppercase names:

```yaml
# .github/workflows/pr-commands.yml
env:
  SMYKLOT_QUIET_SUCCESS: "true"
  SMYKLOT_ALLOWED_COMMANDS: "approve,merge"
  SMYKLOT_COMMAND_PREFIX: "!"
  SMYKLOT_DISABLE_MENTIONS: "false"
```

###### CLI Flags

Pass flags to the binary:

```bash
./smyklot-github-action \
  --quiet-success=true \
  --allowed-commands=approve,merge \
  --command-aliases='{"app":"approve","m":"merge"}' \
  --command-prefix="!" \
  --disable-mentions=false
```

##### Examples

###### Example 1: Quiet Mode

Only show emoji reactions, no success comments:

```yaml
env:
  SMYKLOT_QUIET_SUCCESS: "true"
```

Result: User sees only ✅ reaction, no "PR Approved" comment.

###### Example 2: Custom Prefix

Use `!` instead of `/` for commands:

```yaml
env:
  SMYKLOT_COMMAND_PREFIX: "!"
```

Users can now use `!approve` and `!merge`.

###### Example 3: Command Aliases

Create shortcuts for commands:

```yaml
env:
  SMYKLOT_COMMAND_ALIASES: '{"app":"approve","a":"approve","m":"merge"}'
```

Users can use `/app`, `/a`, or `/m` as shortcuts.

###### Example 4: Restrict Commands

Only allow approve command:

```yaml
env:
  SMYKLOT_ALLOWED_COMMANDS: "approve"
```

The `/merge` command will be ignored.

###### Example 5: Disable Mentions

Only allow slash commands:

```yaml
env:
  SMYKLOT_DISABLE_MENTIONS: "true"
```

`@smyklot approve` will no longer work, only `/approve`.

###### Example 6: Disable Bare Commands

Only allow slash and mention commands:

```yaml
env:
  SMYKLOT_DISABLE_BARE_COMMANDS: "true"
```

`lgtm` and `approve` will no longer work, only `/approve` and `@smyklot approve`.

## Usage

### Commands

Smyklot responds to these commands in PR comments:

| Command | Aliases | Description |
|---------|---------|-------------|
| `/approve` | `@smyklot approve`, `approve`, `accept`, `lgtm` | Approve the pull request |
| `/merge` | `@smyklot merge`, `merge` | Merge the pull request |

**Command Formats**:

- **Slash commands**: `/approve`, `/merge`
- **Mention commands**: `@smyklot approve`, `@smyklot merge`
- **Bare commands**: `approve`, `accept`, `lgtm`, `merge`

All commands are case-insensitive.

**Multiple Commands**:

You can use multiple non-contradicting commands in a single comment:

```text
approve merge
```

or

```text
lgtm
merge
```

Commands will be executed in order: approve first, then merge.

### Example: Approving a PR

Comment on any pull request using any of these formats:

```text
/approve
```

or

```text
@smyklot approve
```

or

```text
lgtm
```

Smyklot will:

1. Check if you're a global owner in `.github/CODEOWNERS`
2. Approve the PR via GitHub API
3. Add ✅ reaction to your comment

### Example: Merging a PR

Comment on any pull request:

```text
/merge
```

Smyklot will:

1. Check if you're a global owner
2. Verify the PR is mergeable
3. Merge the PR via GitHub API
4. Add ✅ reaction to your comment

### Error Handling

If you're not authorized, Smyklot will:

- Add ❌ reaction to your comment
- Post a comment explaining who can approve/merge

## Development

### Requirements

- Go 1.25+
- [mise](https://mise.jdx.dev/) for tool management
- [Task](https://taskfile.dev/) for task automation

### Setup

```bash
# Clone repository
git clone https://github.com/bartsmykla/smyklot.git
cd smyklot

# Install tools
mise install

# Download dependencies
task deps

# Run tests
task test
```

### Project Structure

```text
smyklot/
├── cmd/
│   └── github-action/       # GitHub Actions entrypoint
├── pkg/
│   ├── commands/            # Command parser
│   ├── feedback/            # User feedback system
│   ├── github/              # GitHub API client
│   └── permissions/         # CODEOWNERS parser & checker
├── .github/workflows/       # GitHub Actions workflows
├── .mise.toml               # Tool versions
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
task test:watch
```

Current test coverage: 130 tests passing

- 52 command parser tests
- 12 CODEOWNERS parser tests
- 30 permission checker tests
- 30 feedback system tests
- 18 GitHub client tests

## Architecture

### How It Works

1. User comments a command on a PR (e.g., `/approve`, `@smyklot merge`, `lgtm`)
2. GitHub triggers `issue_comment` webhook
3. `pr-commands.yml` workflow starts
4. Action binary:
   - Parses the command (supports slash, mention, and bare formats)
   - Reads `.github/CODEOWNERS`
   - Checks user permissions
   - Calls GitHub API
   - Posts reactions and feedback

### Permission System

Phase 1 (Current):

- Only global owners (`* @username`) are supported
- Global owners can approve/merge any PR

Phase 2 (Planned):

- Path-specific ownership patterns
- Scoped permissions based on changed files

### Security

- All inputs passed via environment variables
- No shell interpolation of user data
- Actions pinned by commit digest
- Minimal workflow permissions
- Token-based authentication

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

- [x] Command parser
- [x] CODEOWNERS parser
- [x] Permission checker
- [x] GitHub API client
- [x] Feedback system
- [x] GitHub Actions workflows
- [x] Documentation

### Phase 2: Enhanced Permissions (Planned)

- [ ] Path-specific ownership
- [ ] Scoped approval requirements
- [ ] Team support in CODEOWNERS
- [ ] Self-approval prevention

### Phase 3: Kubernetes Deployment (Future)

- [ ] HTTP webhook server
- [ ] Persistent service
- [ ] Scalable deployment
- [ ] Prometheus metrics

### Phase 4: Discord Integration (Future)

- [ ] Discord bot
- [ ] Unified command system
- [ ] Cross-platform notifications

## License

MIT License - see [LICENSE](LICENSE) for details

## Acknowledgments

Built with:

- [Ginkgo](https://github.com/onsi/ginkgo) - BDD testing framework
- [Gomega](https://github.com/onsi/gomega) - Matcher library
- [mise](https://mise.jdx.dev/) - Tool version manager
- [Task](https://taskfile.dev/) - Task runner
