# Deployment Guide

<!-- markdownlint-disable MD013 -->

This guide covers deploying Smyklot to your GitHub repository.

## Prerequisites

- GitHub repository with pull requests
- Ability to create `.github/CODEOWNERS` file
- Ability to create/modify GitHub Actions workflows

## Deployment Steps

### 1. Create CODEOWNERS File

Create `.github/CODEOWNERS` in your repository root:

```text
# Global owners (can approve/merge any PR)
* @yourusername @teammate1 @teammate2
```

**Note**: Only global owners (`* @username`) are supported in Phase 1.
Path-specific patterns will be supported in Phase 2.

### 2. Create Workflow File

Create `.github/workflows/pr-commands.yaml`:

Smyklot supports two ways to pass parameters:

- **Inputs** (recommended): Cleaner syntax using action inputs
- **Environment variables**: Alternative approach, useful for
  compatibility

Both approaches support automatic fallback to environment variables when
inputs are not provided.

#### Option A: Using GITHUB_TOKEN (simpler, comments from workflow user)

Using inputs:

```yaml
name: PR Commands

on:
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  handle-command:
    name: Handle PR Command
    if: |
      github.event.issue.pull_request &&
      github.event.comment.user.type != 'Bot' &&
      (
        startsWith(github.event.comment.body, '/approve') ||
        startsWith(github.event.comment.body, '/merge') ||
        contains(github.event.comment.body, '@smyklot')
      )
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout repository
        uses: actions/checkout@08c6903cd8c0fde910a37f88322edcfb5dd907a8 # v5.0.0

      - name: Run Smyklot
        uses: smykla-labs/smyklot@v0.1.0
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          comment-body: ${{ github.event.comment.body }}
          comment-id: ${{ github.event.comment.id }}
          pr-number: ${{ github.event.issue.number }}
          repo-owner: ${{ github.repository_owner }}
          repo-name: ${{ github.event.repository.name }}
          comment-author: ${{ github.event.comment.user.login }}
```

Using environment variables:

```yaml
      - name: Run Smyklot
        uses: smykla-labs/smyklot@v0.1.0
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          COMMENT_BODY: ${{ github.event.comment.body }}
          COMMENT_ID: ${{ github.event.comment.id }}
          PR_NUMBER: ${{ github.event.issue.number }}
          REPO_OWNER: ${{ github.repository_owner }}
          REPO_NAME: ${{ github.event.repository.name }}
          COMMENT_AUTHOR: ${{ github.event.comment.user.login }}
```

#### Option B: Using GitHub App (recommended, comments from app)

Using inputs:

```yaml
name: PR Commands

on:
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  handle-command:
    name: Handle PR Command
    if: |
      github.event.issue.pull_request &&
      github.event.comment.user.type != 'Bot' &&
      (
        startsWith(github.event.comment.body, '/approve') ||
        startsWith(github.event.comment.body, '/merge') ||
        contains(github.event.comment.body, '@smyklot')
      )
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout repository
        uses: actions/checkout@08c6903cd8c0fde910a37f88322edcfb5dd907a8 # v5.0.0

      - name: Generate GitHub App token
        id: generate-token
        uses: actions/create-github-app-token@67018539274d69449ef7c02e8e71183d1719ab42 # v2.1.4
        with:
          app-id: ${{ vars.APP_ID }}
          private-key: ${{ secrets.APP_PRIVATE_KEY }}

      - name: Run Smyklot
        uses: smykla-labs/smyklot@v0.1.0
        with:
          token: ${{ steps.generate-token.outputs.token }}
          comment-body: ${{ github.event.comment.body }}
          comment-id: ${{ github.event.comment.id }}
          pr-number: ${{ github.event.issue.number }}
          repo-owner: ${{ github.repository_owner }}
          repo-name: ${{ github.event.repository.name }}
          comment-author: ${{ github.event.comment.user.login }}
```

Using environment variables:

```yaml
      - name: Run Smyklot
        uses: smykla-labs/smyklot@v0.1.0
        env:
          GITHUB_TOKEN: ${{ steps.generate-token.outputs.token }}
          COMMENT_BODY: ${{ github.event.comment.body }}
          COMMENT_ID: ${{ github.event.comment.id }}
          PR_NUMBER: ${{ github.event.issue.number }}
          REPO_OWNER: ${{ github.repository_owner }}
          REPO_NAME: ${{ github.event.repository.name }}
          COMMENT_AUTHOR: ${{ github.event.comment.user.login }}
```

### 3. (Optional) Configure GitHub App Authentication

**Only needed if using Option B workflow above.**

To have comments appear from the GitHub App instead of the default
`GITHUB_TOKEN` user:

1. **Create or use existing GitHub App**:
   - Go to Settings → Developer settings → GitHub Apps
   - Note the App ID
   - Generate and download a private key (.pem file)
   - Install the app on your repository

2. **Add App ID as variable and private key as secret**:

   ```bash
   gh variable set APP_ID --body "1197525"
   # Private key must be in PKCS#8 format
   # (starts with "-----BEGIN PRIVATE KEY-----")
   # If your key is in OpenSSH format, convert it first:
   # ssh-keygen -p -N "" -m pem -f openssh-key.pem
   # openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt \
   #   -in openssh-key.pem -out pkcs8-key.pem
   gh secret set APP_PRIVATE_KEY < pkcs8-key.pem
   ```

**Note**: The `actions/create-github-app-token` action automatically
detects the installation ID, so you don't need to configure it
separately.

### 4. Commit and Push

```bash
git add .github/CODEOWNERS .github/workflows/pr-commands.yaml
git commit -sS -m "feat(ci): add Smyklot PR command automation"
git push
```

### 5. Test the Deployment

Create a test pull request and try the commands:

1. **Test approval**:
   - Comment: `/approve` or `@smyklot approve`
   - Expected: ✅ reaction + PR approved

2. **Test merge**:
   - Comment: `/merge` or `@smyklot merge`
   - Expected: ✅ reaction + PR merged (if mergeable)

3. **Test unauthorized user** (optional):
   - Have a non-CODEOWNER comment `/approve`
   - Expected: ❌ reaction + explanation comment

## Command Reference

### Available Commands

| Command | Alias | Action | Requirements |
|---------|-------|--------|--------------|
| `/approve` | `@smyklot approve` | Approve the PR | Listed in CODEOWNERS |
| `/merge` | `@smyklot merge` | Merge the PR | CODEOWNERS + mergeable |

### Feedback System

**Success** (emoji only):

- ✅ - Command executed successfully

**Errors** (emoji + comment):

- ❌ - Unauthorized or error
- ⚠️ - Warning (e.g., merge conflict)
- 👀 - Processing (added immediately)

## Troubleshooting

### Command Not Working

1. **Check CODEOWNERS file exists**:

   ```bash
   cat .github/CODEOWNERS
   ```

2. **Check workflow file exists**:

   ```bash
   cat .github/workflows/pr-commands.yaml
   ```

3. **Check Actions tab**:
   - Go to repository → Actions tab
   - Look for "PR Commands" workflow
   - Check for errors in workflow runs

4. **Check permissions**:
   - Verify user is listed in CODEOWNERS
   - Verify workflow has correct permissions

### No Reaction on Comment

1. Check if comment is on a pull request (not an issue)
2. Check Actions tab for workflow execution
3. Check workflow logs for errors

### Approval Not Working

1. Verify GITHUB_TOKEN has write permissions
2. Check if user is in CODEOWNERS
3. Check workflow logs for API errors

### Merge Not Working

1. Verify PR is mergeable (no conflicts)
2. Verify required checks have passed
3. Check branch protection rules
4. Check workflow logs for API errors

## Security Considerations

### Permissions

The workflow requires minimal permissions:

- `contents: read` - Read repository files
- `pull-requests: write` - Approve and merge PRs
- `issues: write` - Add reactions and comments

### GITHUB_TOKEN

The workflow uses the built-in `GITHUB_TOKEN` which:

- Is automatically created for each workflow run
- Has repository-scoped permissions
- Expires after the workflow completes
- Cannot be used outside the repository

### Input Validation

All user inputs are passed via environment variables (not shell
interpolation) to prevent injection attacks:

```yaml
env:
  COMMENT_BODY: ${{ github.event.comment.body }}
  # Not: run: ./bot "${{ github.event.comment.body }}"
```

## Updating Smyklot

To update to a new version:

1. Change the version reference in workflow:

   ```yaml
   uses: smykla-labs/smyklot@v1.0.0  # Update this line
   ```

2. Commit and push:

   ```bash
   git add .github/workflows/pr-commands.yaml
   git commit -sS -m "chore(ci): update Smyklot to v1.0.0"
   git push
   ```

## Support

- **Issues**: <https://github.com/smykla-labs/smyklot/issues>
- **Discussions**: <https://github.com/smykla-labs/smyklot/discussions>
- **Documentation**: <https://github.com/smykla-labs/smyklot/blob/main/README.md>
