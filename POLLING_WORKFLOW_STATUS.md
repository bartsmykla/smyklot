# Polling Workflow Implementation Status

## What We Did

### 1. Implemented Polling Subcommand (v1.3.0)

**Problem**: Reaction-based commands (👍, 🚀, ❤️) didn't work in GitHub Actions because `issue_comment` events don't fire for reaction changes.

**Solution**: Created a `poll` subcommand that checks all open PRs for reactions on the PR description every 5 minutes.

**Implementation**:

- Added `poll` subcommand to CLI (`cmd/github-action/poll.go`)
- Added `GetOpenPRs()` method to GitHub client
- Created `.github/workflows/poll-reactions.yaml` workflow
- Runs on schedule: `*/5 * * * *` (every 5 minutes)
- Can be triggered manually via `workflow_dispatch`

**Files changed**:

- `cmd/github-action/poll.go` - New poll subcommand
- `pkg/github/client.go` - Added `GetOpenPRs()` method
- `.github/workflows/poll-reactions.yaml` - Scheduled workflow

**Deployed to**:

- smyklot repository (smykla-labs/smyklot)
- .dotfiles repository (smykla-labs/.dotfiles) - PR #35

### 2. Auto-Merge Support (v1.4.0)

**Problem**: When PRs have branch protection (pending checks, merge queue, required reviews), immediate merge attempts fail. Instead of failing, should enable auto-merge.

**Solution**: Detect merge queue configuration and enable auto-merge when immediate merge is not possible.

**Implementation**:

- Added `EnableAutoMerge()` method using GitHub GraphQL API
- Added `IsMergeQueueEnabled()` to detect branch protection
- Added auto-merge feedback functions
- Modified `executeMerge()` and `handleReactionMerge()` to use auto-merge on errors
- Added `BaseBranch` field to `PRInfo` for queue detection
- Fixed eyes reaction removal in `postCombinedFeedback()`

**Files changed**:

- `pkg/github/client.go` - Added `EnableAutoMerge()`, `IsMergeQueueEnabled()`
- `pkg/github/types.go` - Added `BaseBranch` to `PRInfo`
- `pkg/feedback/feedback.go` - Added auto-merge feedback functions
- `cmd/github-action/main.go` - Updated merge logic, fixed eyes removal

**How it works**:

1. Check if merge queue is enabled for base branch
2. If enabled, use auto-merge directly (no failed attempt)
3. If not enabled, try immediate merge
4. If merge fails due to protection rules, enable auto-merge
5. Auto-merge uses requested merge method (merge/squash/rebase)

**Error detection**:

- Merge queue requirements
- Pending status checks
- Required reviews
- Branch protection rules

### 3. Docker Image Workflow Optimization (v1.4.0)

**Problem**: Workflows were building smyklot from source on every run, which was inefficient. The `.dotfiles` repository workflow failed because it tried to build source code that doesn't exist there.

**Solution**: Replace source build with pre-built Docker image in all workflows.

**Implementation**:

- Updated `poll-reactions` workflow to use `ghcr.io/smykla-labs/smyklot:1.4.0`
- Updated `pr-commands` workflow to use Docker image
- Removed mise installation, Go caching, and build steps
- Simplified workflows from 6 steps to 2-3 steps
- Added GitHub App token generation to `.dotfiles` workflows

**Files changed**:

- `.github/workflows/poll-reactions.yaml` - Use Docker image
- `.github/workflows/pr-commands.yaml` - Use Docker image

**Benefits**:

- Faster workflow execution (no build time)
- Consistent execution environment
- Works in repositories without source code
- Reduced workflow complexity

**Deployed to**:

- smyklot repository (smykla-labs/smyklot)
- .dotfiles repository (smykla-labs/.dotfiles) - PR #38

### 4. PR Reaction Scope Fix (v1.7.6)

**Problem**: Poll workflow was checking reactions on all PR comments instead of the PR description itself. This caused confusion as users expected to react to the PR, not individual comments.

**Solution**: Changed poll workflow to process reactions on PR descriptions only.

**Implementation**:

- Added `GetPRReactions()` method to get reactions on PR description
- Refactored `processPR()` to process PR reactions directly instead of iterating through comments
- Updated `handleReactions()` to distinguish between PR and comment reactions
- Removed `processComment()` and `pollCommentReactions()` functions (no longer needed)
- Fixed cleanup command `/user` permission issue by adding `SMYKLOT_BOT_USERNAME` configuration
- Added `DismissReviewByUsername()` method to avoid restricted `/user` endpoint

**Files changed**:

- `cmd/github-action/main.go` - Added `BotUsername` configuration, updated cleanup logic
- `cmd/github-action/poll.go` - Rewrote to process PR reactions instead of comment reactions
- `pkg/github/client.go` - Added `GetPRReactions()` and `DismissReviewByUsername()` methods

**Benefits**:

- Clearer UX: Users react to the PR description, not comments
- Consistent with GitHub's reaction model
- Cleanup command now works with GitHub App tokens
- 60+ lines of unused code removed

## What Needs to Be Done Next

### Short Term (Phase 1 Completion)

- [ ] **Test polling workflow on .dotfiles**
  - Merge PR #35
  - Test 👍 reaction for approve
  - Test 🚀 reaction for merge
  - Test ❤️ reaction for cleanup
  - Verify reactions are processed within 5 minutes

- [ ] **Test auto-merge on .dotfiles PR #34**
  - Comment `merge` on PR #34
  - Verify auto-merge is enabled (not immediate merge)
  - Verify merge happens after checks pass
  - Verify eyes reaction is removed

- [ ] **Update documentation**
  - Update README.md with auto-merge behavior
  - Document merge queue detection
  - Add troubleshooting section for common issues

- [ ] **Monitor workflow performance**
  - Check GitHub Actions usage (polling every 5 minutes)
  - Adjust polling frequency if needed (can reduce to 10-15 min)
  - Monitor for API rate limiting

### Medium Term (Phase 1 Enhancements)

- [ ] **Improve polling efficiency**
  - Only poll PRs with recent activity (last 24h)
  - Skip PRs that haven't changed since last poll
  - Add caching to reduce API calls

- [ ] **Better merge queue detection**
  - Cache branch protection rules (avoid API call every merge)
  - Support multiple base branches (main, master, develop)
  - Handle edge cases (branch protection disabled mid-merge)

- [ ] **Enhanced feedback**
  - Show estimated merge time when auto-merge enabled
  - Notify when checks complete and merge happens
  - Provide better error messages for merge failures

- [ ] **Configuration options**
  - Allow custom polling interval via config
  - Option to disable polling (use only on-demand)
  - Configure which reactions to poll for

### Long Term (Phase 2+)

- [ ] **Path-specific CODEOWNERS support**
  - Parse path patterns from CODEOWNERS
  - Check changed files in PR
  - Require approvals from path-specific owners
  - See Phase 2 in README.md for full scope

- [ ] **Webhook-based deployment (Phase 3)**
  - **Platform**: Fly.io (decided)
  - Real-time reaction processing (no polling delay)
  - HTTP webhook server
  - No cold starts for instant webhook response
  - Simple Dockerfile-based deployment
  - Built-in observability
  - Free tier sufficient for personal use
  - See Phase 3 in README.md for architecture
  - **Setup**:
    1. Create `fly.toml` configuration
    2. Deploy: `fly launch --name smyklot-webhook && fly deploy`
    3. Configure webhook URL: `https://smyklot-webhook.fly.dev`

- [ ] **Advanced auto-merge features**
  - Smart merge method selection based on PR size
  - Automatic rebase when base branch updates
  - Conflict detection and notification
  - Integration with merge queue positions

## Current Issues and Known Limitations

1. **Polling delay**: Reactions processed every 5 minutes (not real-time)
   - Workaround: Use comment commands for immediate action
   - Long-term fix: Phase 3 webhook deployment

2. **API rate limiting**: Polling uses GitHub API calls
   - Current: ~12 calls/hour (5 min interval)
   - Risk: High activity repos might hit limits
   - Mitigation: Adjust polling frequency, implement caching

3. **Merge queue detection**: Limited branch protection API support
   - May not detect all merge queue configurations
   - False negatives possible (tries immediate merge when queue required)
   - Testing needed on various repo configurations

4. **Base branch assumption**: Defaults to checking detected base branch
   - Works for standard workflows
   - May need enhancement for complex branching strategies

## Testing Checklist

### Polling Workflow

- [ ] Workflow triggers on schedule
- [ ] Workflow can be manually triggered
- [ ] Processes 👍 reactions (approve)
- [ ] Processes 🚀 reactions (merge)
- [ ] Processes ❤️ reactions (cleanup)
- [ ] Removes reactions when undone
- [ ] Handles multiple PRs correctly
- [ ] Handles PRs with no comments
- [ ] Handles permission errors gracefully

### Auto-Merge

- [ ] Detects merge queue configuration
- [ ] Enables auto-merge when queue detected
- [ ] Enables auto-merge on pending checks
- [ ] Enables auto-merge on required reviews
- [ ] Uses correct merge method (merge/squash/rebase)
- [ ] Falls back to squash/rebase when merge disallowed
- [ ] Removes eyes reaction before final status
- [ ] Posts appropriate feedback messages
- [ ] Works with comment commands
- [ ] Works with reaction commands

## Version History

- **v1.3.0** (2025-11-16): Polling workflow implementation
- **v1.4.0** (2025-11-16): Auto-merge support with queue detection
- **v1.7.6** (2025-11-16): PR reaction scope fix and cleanup command fix

## References

- [GitHub Actions Cron Syntax](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#schedule)
- [GitHub GraphQL API - Auto-Merge](https://docs.github.com/en/graphql/reference/mutations#enablepullrequestautomerge)
- [GitHub Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)
- [GitHub Merge Queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue)
