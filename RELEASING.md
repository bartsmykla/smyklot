# Releasing Smyklot

This document describes the release process for Smyklot.

## Release Process Overview

Smyklot uses **automated releases** with semantic versioning. Releases are triggered automatically based on conventional commit messages.

### Automated Release (Recommended)

The `auto-release.yaml` workflow runs **daily at 10:00 UTC** and automatically:

1. Analyzes commits since the last release
2. Determines the next version using semantic versioning
3. Creates a release commit and tag
4. Triggers the release workflow to build and publish

**Version Bumping Rules**:

- **Major** (X.0.0): Commits with `BREAKING CHANGE` or `!` suffix (e.g., `feat!: ...`)
- **Minor** (0.X.0): Commits with `feat:` or `feat(...):` prefix
- **Patch** (0.0.X): Commits with `fix:` or other types

**Example**:

```bash
# Current version: 1.7.8

# These commits trigger a minor bump (1.8.0):
git commit -sS -m "feat(api): add new approval endpoint"
git commit -sS -m "fix(parser): handle edge case"

# These commits trigger a major bump (2.0.0):
git commit -sS -m "feat!: redesign approval system"
git commit -sS -m "feat: add feature

BREAKING CHANGE: removed legacy API"
```

### Manual Release

You can trigger a release manually:

1. **Via GitHub Actions UI**:
   - Go to Actions → Auto Release
   - Click "Run workflow"
   - Optionally specify a version (e.g., `1.9.0`)
   - Click "Run workflow"

2. **Via Command Line**:

   ```bash
   # Trigger automated version bump
   gh workflow run auto-release.yaml

   # Trigger with specific version
   gh workflow run auto-release.yaml -f version=1.9.0
   ```

### Release Artifacts

Each release produces:

1. **Docker Image**: `ghcr.io/bartsmykla/smyklot:X.Y.Z`
   - Multi-architecture: `linux/amd64`, `linux/arm64`
   - Tagged with: `X.Y.Z`, `X.Y`, `X`

2. **Binary Archives**: Available on GitHub Releases
   - `smyklot_X.Y.Z_linux_amd64.tar.gz`
   - `smyklot_X.Y.Z_linux_arm64.tar.gz`
   - `smyklot_X.Y.Z_darwin_amd64.tar.gz`
   - `smyklot_X.Y.Z_darwin_arm64.tar.gz`

3. **Checksums**: `checksums.txt` for all archives

4. **Changelog**: Auto-generated from conventional commits

## Release Workflow Details

### 1. Auto-Release Workflow (`.github/workflows/auto-release.yaml`)

**Trigger**: Daily at 10:00 UTC or manual dispatch

**Steps**:

1. Generate GitHub App token for authenticated operations
2. Checkout repository with full history
3. Determine next version using semantic versioning
4. Check if release is needed (skip if no new commits)
5. Update `action.yml` with new Docker image tag
6. Create verified commit via GitHub API
7. Create signed tag via GitHub API
8. Push tag to trigger release workflow

**Secrets Required**:

- `SMYKLOT_APP_ID` - GitHub App ID
- `SMYKLOT_PRIVATE_KEY` - GitHub App private key

### 2. Release Workflow (`.github/workflows/release.yaml`)

**Trigger**: Push of tag matching `v*.*.*`

**Steps**:

1. Checkout code with full history
2. Install mise and set up toolchain
3. Set up Docker Buildx for multi-arch builds
4. Login to GitHub Container Registry
5. Run GoReleaser to:
   - Build binaries for all platforms
   - Create archives
   - Build and push Docker images
   - Generate changelog
   - Create GitHub Release

**Secrets Required**:

- `GHCR_TOKEN` - GitHub Container Registry token
- `GITHUB_TOKEN` - Auto-provided by GitHub Actions

### 3. GoReleaser Configuration (`.goreleaser.yml`)

**Builds**:

- Binary: `smyklot`
- Platforms: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
- Flags: `-s -w` (strip debug info), version injection

**Docker Images**:

- Registry: `ghcr.io/bartsmykla/smyklot`
- Tags: `X.Y.Z`, `X.Y`, `X`
- Platforms: `linux/amd64`, `linux/arm64`

**Changelog**:

- Groups commits by: Features, Bug Fixes, Performance Improvements
- Excludes: `docs:`, `test:`, `ci:`, `chore:`, `build:`

## Checking Release Status

```bash
# Check latest release
gh release view --repo bartsmykla/smyklot

# List all releases
gh release list --repo bartsmykla/smyklot

# Check latest tag
git tag --list 'v*' --sort=-version:refname | head -1

# View recent release commits
git log --oneline --grep='chore(release)' -5
```

## Troubleshooting

### Release workflow failed

1. Check GitHub Actions logs
2. Verify secrets are configured correctly
3. Ensure GoReleaser configuration is valid: `goreleaser check`

### Docker image not updated

1. Verify GHCR_TOKEN has `write:packages` permission
2. Check Docker build logs in release workflow
3. Manually trigger release workflow: `gh workflow run release.yaml`

### Version not bumped

1. Ensure commits use conventional commit format
2. Check auto-release workflow logs for version calculation
3. Manually specify version: `gh workflow run auto-release.yaml -f version=X.Y.Z`

## Version File Locations

- `action.yml` - Docker image tag (updated by auto-release)
- Git tags - Source of truth for releases (e.g., `v1.7.8`)
- GitHub Releases - Published artifacts and changelog

## Best Practices

1. **Use conventional commits**: Ensures correct version bumping
2. **Test before merging**: CI runs on every push
3. **Review changelogs**: Auto-generated but should be reviewed
4. **Monitor releases**: Check GitHub Actions for failures
5. **Update dependencies**: Keep GoReleaser and actions up to date
