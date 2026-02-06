# AGENTS.md - AI Agent Context for bb Project

This file contains important context, learnings, and guidance for AI coding agents working on the `bb` (Bitbucket Cloud CLI) project.

## Project Overview

- **Name:** bb
- **Purpose:** Bitbucket Cloud CLI tool for managing pull requests, comments, and pipelines
- **Language:** Go
- **Version:** v0.2.0
- **Key Packages:**
  - `pkg/bbcloud/` - Bitbucket Cloud API client
  - `pkg/httpx/` - HTTP client with retry logic
  - `pkg/cmdutil/` - Command utilities and factories
  - `pkg/iostreams/` - I/O stream abstractions

## CLI Structure (v0.2.0)

**Current command structure:**
```bash
# Authentication
bb auth                                        # Interactive login (default)
bb auth status                                 # Check auth status + scope check

# Discovery
bb list repos                                  # List repositories

# Review — Read
bb review list --repo <repo>                   # List PRs with stats
bb review view <pr> --repo <repo>              # Complete PR context
bb review view <pr> <file> --repo <repo>       # View file diff

# Review — Comment management
bb review comment <pr> --repo <repo> "message"                    # General comment
bb review comment <pr> <file> <line> --repo <repo> "message"      # Inline comment
bb review comment <pr> <file> <start> <end> --repo <repo> "msg"  # Line range comment
bb review comment <pr> --repo <repo> --edit <id> "new text"       # Edit comment
bb review comment <pr> --repo <repo> --delete <id>                # Delete comment
bb review comment <pr> --repo <repo> --resolve <id>               # Resolve comment (inline only)
bb review comment <pr> --repo <repo> --reopen <id>                # Reopen resolved comment
bb review reply <pr> <comment-id> --repo <repo> "message"         # Reply to comment

# Review — Actions
bb review create --repo <repo> --source <branch> [--target <branch>] --title "..." # Create PR
bb review approve <pr> --repo <repo>                # Approve PR
bb review approve <pr> --repo <repo> --undo         # Remove approval
bb review request-change <pr> --repo <repo>         # Request changes
bb review request-change <pr> --repo <repo> --undo  # Remove request-change
```

**Review subcommands (7):** list, view, comment, reply, create, approve, request-change

**Comment flags:** `--edit`, `--delete`, `--resolve`, `--reopen` are mutually exclusive; each takes a comment ID.

**BB API constraints:**
- `--resolve` only works on inline (diff) comments, not general comments
- `DELETE /approve` only undoes approvals; `DELETE /request-changes` undoes request-changes
- Line range comments use `start_to` (start) and `to` (end) in the inline object

**Design principles:**
- Simple Unix tool - fetch data, output JSON
- Token-efficient flat structures
- Agent does analysis, tool provides clean data
- Consistent positional arg pattern (PR always first)
- Separation of concerns (view vs actions)
- One command can aggregate multiple API calls

## Development Guidelines

### Git Commits

**IMPORTANT:** Follow the git-commit skill guidance precisely:

- **Keep commit messages concise** - Subject line only (≤72 chars) is usually sufficient
- **Body is OPTIONAL** - Only add a body if there's complex context that isn't obvious from the diff
- **Don't over-explain** - The file list and diff already show what changed
- Format: `<type>(<scope>): <summary>`
  - Examples: `feat(review): add unified diff format`, `fix(review): use correct build status API`
- **Do NOT add verbose bodies** unless genuinely necessary for understanding WHY the change was made

### Token Efficiency is Paramount

**For agent-facing commands (review, etc.):**
- Use raw unified diff format (not CSV or JSON) — LLMs are trained on billions of diffs
- Filter out noise: pending reviewers, header lines, unchanged context (when appropriate)
- Aggregate multiple API calls into single responses
- Use concise field names: `diff` not `difficultyAnalysisWithMetadata`
- Include only actionable metadata: author_id for mentions, line numbers for references

❌ **Wrong** (over-explained):
```
feat(bbcloud): implement Bitbucket Cloud API client

- Add comprehensive type definitions for API responses
- Implement base HTTP client with retry logic and auth
- Add user, repository, and pull request API methods
- Implement comment operations (read and write)
- Add pipeline status checking
- Support pagination, URL encoding, and plain text responses
- Total: ~1200 lines implementing complete Phase 2
```

✅ **Right** (concise):
```
feat(bbcloud): implement Bitbucket Cloud API client
```

### API Client Design

- All API methods should accept `context.Context` as first parameter
- Use proper URL encoding for all path parameters (`url.PathEscape`)
- Implement pagination for list operations
- Return structured errors with context (`fmt.Errorf("operation: %w", err)`)
- Plain text responses (like diffs) should use `io.Writer` interface

### Code Organization

- Phase 1: Foundation (factory, error types, project setup)
- Phase 2: API Client (`pkg/bbcloud/`)
- Phase 3: CLI Commands (TBD)
- Phase 4: Skipped
- Phase 5: Testing, documentation, release

## Improvements & Features

### Agent-Optimized Review Commands (2026-02-07)

**Feature:** Implemented `bb review` commands designed specifically for AI agent code review workflows with extreme token efficiency.

**Commands:**
- `bb review list --repo <repo>` - List PRs with stats (files, additions, deletions, approvals)
- `bb review view <pr> --repo <repo>` - Complete PR context in one call (metadata + files + build + reviewers + comments)
- `bb review view <pr> <file> --repo <repo>` - File diff with inline comments (unified diff format)

**Design Principles:**
- ✅ **Extreme token efficiency**: Raw unified diff format avoids escaping overhead
- ✅ **One command = multiple API calls**: Aggregates data from 4-5 Bitbucket endpoints
- ✅ **Flat JSON**: Easy parsing, minimal nesting (diff is a string field)
- ✅ **No fancy analysis**: Agent does the thinking, tool provides clean data
- ✅ **Actionable data**: Includes author_id for @mentions, line numbers for precise references

**Key Optimizations:**
1. **Raw Unified Diff**: Diffs returned as raw unified diff strings (not CSV or JSON). ~80% token reduction vs JSON objects.

2. **Reviewer Filtering**: Only shows reviewers who have taken action (approved or changes_requested), ignoring pending reviewers

3. **Author IDs for @mentions**: All comments include `author_id` (UUID) for proper Bitbucket @mentions

4. **Build Status**: Uses `GetPRPipelines()` for actual build status, not deployment status

**Example Workflow:**
```bash
# 1. Find PRs
bb review list --repo test_repo

# 2. Get PR context (includes reviewers, build status, file stats, comment counts)
bb review view 450 --repo test_repo

# 3. Review specific file (unified diff + inline comments with author_id)
bb review view 450 src/auth/jwt.ts --repo test_repo

# 4. Extract diff
jq -r '.diff' | head -50
```

**Output Format Examples:**

*PR Overview:*
```json
{
  "id": 450,
  "title": "feat: add authentication",
  "build_status": "SUCCESSFUL",
  "reviewers": [
    {"username": "Alice", "state": "approved"},
    {"username": "Bob", "state": "changes_requested"}
  ],
  "files": [{"path": "auth.ts", "additions": 150, "comments": 2}],
  "total_comments": 5
}
```

*File Review with CSV Diff:*
```json
{
  "pr": 450,
  "file": "auth.ts",
  "diff": "@@ -225,7 +225,7 @@\n-                            [:div\n+                            [:div.checkbox\n",
  "comments": [
    {
      "id": 123,
      "line": 227,
      "author": "Alice",
      "author_id": "{uuid-here}",
      "text": "Use semantic class name"
    }
  ]
}
```

**Files:**
- `pkg/cmd/review/` - Review command package
- `pkg/cmd/review/list.go` - PR list command  
- `pkg/cmd/review/view.go` - PR view and file diff commands
- `pkg/cmd/review/comment.go` - Comment CRUD + resolve/reopen
- `pkg/cmd/review/reply.go` - Reply to comments
- `pkg/cmd/review/approve.go` - Approve/unapprove + `friendlyError()` helper
- `pkg/cmd/review/request_change.go` - Request/unrequest changes
- `pkg/cmd/review/create.go` - Create new PRs

## Key Patterns & Learnings

### Client Initialization
All command `RunE` functions must initialize the client before use:
```go
client, err := opts.factory.NewBBCloudClient("")
```
Then pass `client` to run functions. Never store client in opts struct without initializing.

### Credential Loading Order
`Factory.loadCredentials()` checks env vars (`BB_WORKSPACE`, `BB_USERNAME`, `BB_TOKEN`) first, then falls back to keyring. All 3 must be set to skip keyring.

### Auth Status Scope Detection
`bb auth status` parses the `x-oauth-scopes` response header from `GET /user` to check granted scopes against required scopes. Uses `DoWithHeaders()` in httpx — `Do()` is a thin wrapper around it.

Required scopes: `read:user`, `read:workspace`, `read:repository`, `read:pullrequest`, `write:pullrequest`, `read:pipeline` (all `:bitbucket` suffix).

### Structured Error Handling for LLMs
Write commands (approve, request-change) output structured JSON errors instead of raw Go errors:
```json
{"pr": 253, "repo": "team007", "action": "request-change", "error": "PR is already merged"}
```
The `friendlyError()` helper in `approve.go` maps BB API error messages to clean strings using `strings.Contains`.

### BB API Inline Comment Fields
- Single line: `{"inline": {"path": "file.py", "to": 50}}`
- Line range: `{"inline": {"path": "file.py", "start_to": 16, "to": 38}}`
- `from` / `start_from` exist but are for "old file" side (not commonly used)

### Keyring Storage
Credentials stored as single JSON blob at key `bb/credentials` to minimize keyring access. Environment variables: `BB_ALLOW_INSECURE_STORE`, `BB_KEYRING_PASSPHRASE`, `BB_KEYRING_TIMEOUT`, `BB_HTTP_DEBUG`.

## Meta-Instructions

**ANY learning or guidance received during development MUST be added to this file immediately.**

This ensures:
- Consistent behavior across development sessions
- Knowledge persistence for future agents
- Clear project conventions and patterns
