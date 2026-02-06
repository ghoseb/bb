# Review Commands - Simple & Token-Efficient

## Philosophy

- **Simple Unix tool**: Fetch data, output JSON, that's it
- **Token-efficient**: Flat structures, no redundancy
- **Agent does the thinking**: Tool just provides clean data
- **One command can make multiple API calls**: That's fine, it's just data fetching

---

## Commands

### `bb review list [--repo REPO] [--state open|merged|declined]`

**Purpose**: List PRs

**Output**:
```json
{
  "prs": [
    {
      "id": 450,
      "title": "feat: add user authentication",
      "author": "alice",
      "state": "OPEN",
      "source": "feature/auth",
      "target": "main",
      "created": "2024-01-15T10:30:00Z",
      "updated": "2024-01-15T12:00:00Z",
      "files": 12,
      "additions": 450,
      "deletions": 120,
      "approved": 2,
      "declined": 0
    }
  ]
}
```

---

### `bb review view <pr-number> [--repo REPO]`

**Purpose**: Get complete PR context

**What it fetches**:
- PR metadata
- File list with stats
- Build status
- Reviewer status (only those who took action: approved/declined)
- All comments (general + inline)

**Output**:
```json
{
  "id": 450,
  "title": "feat: add user authentication",
  "description": "Adds JWT-based auth with refresh tokens...",
  "author": "alice",
  "state": "OPEN",
  "source": "feature/auth",
  "target": "main",
  "created": "2024-01-15T10:30:00Z",
  "updated": "2024-01-15T12:00:00Z",
  "reviewers": [
    {"username": "bob", "approved": true},
    {"username": "charlie", "approved": true},
    {"username": "diane", "approved": false}
  ],
  "build_status": "SUCCESSFUL",
  "files": [
    {
      "path": "src/auth/jwt.ts",
      "status": "added",
      "additions": 150,
      "deletions": 0,
      "comments": 2
    },
    {
      "path": "src/auth/middleware.ts",
      "status": "modified",
      "additions": 45,
      "deletions": 12,
      "comments": 1
    },
    {
      "path": "tests/auth.test.ts",
      "status": "added",
      "additions": 200,
      "deletions": 0,
      "comments": 0
    },
    {
      "path": "README.md",
      "status": "modified",
      "additions": 5,
      "deletions": 2,
      "comments": 0
    }
  ],
  "total_files": 12,
  "total_additions": 450,
  "total_deletions": 120,
  "total_comments": 8
}
```

---

### `bb review view <pr-number> <file-path> [--repo REPO]`

**Purpose**: Get file diff + inline comments with unified diff format

**Output**:
```json
{
  "pr": 450,
  "file": "src/auth/jwt.ts",
  "status": "added",
  "additions": 150,
  "deletions": 0,
  "diff": "@@ -0,0 +1,150 @@\n+import jwt from 'jsonwebtoken';\n+import { User } from '../types';\n+\n+export function generateToken(user: User): string {\n+  return jwt.sign(\n+    { userId: user.id, email: user.email },\n+    process.env.JWT_SECRET as string\n+  );\n+}\n+\n+export function verifyToken(token: string): User | null {\n+  try {\n+    return jwt.verify(token, process.env.JWT_SECRET as string) as User;\n+  } catch (err) {\n+    return null;\n+  }\n+}\n...",
  "comments": [
    {
      "id": 123,
      "line": 5,
      "author": "bob",
      "text": "Should we add expiration time to the token?",
      "created": "2024-01-15T11:00:00Z",
      "inline": true,
      "replies": [
        {
          "id": 124,
          "author": "alice",
          "text": "Good point, will add in next commit",
          "created": "2024-01-15T11:30:00Z"
        }
      ]
    },
    {
      "id": 125,
      "line": 12,
      "author": "charlie",
      "text": "Error handling could be more specific",
      "created": "2024-01-15T10:45:00Z",
      "inline": true,
      "replies": []
    }
  ]
}
```

---

### `bb review comment <pr-number> [<file-path> [<line> [<end-line>]]] [--repo REPO] "message"`

**Purpose**: Create a general or inline comment on a PR

**Examples**:
```bash
# General comment
bb review comment 450 --repo test_repo "LGTM!"

# Single line comment
bb review comment 450 src/auth/jwt.ts 42 --repo test_repo "Add error handling here"

# Line range comment
bb review comment 450 src/auth/jwt.ts 42 45 --repo test_repo "This block needs refactoring"
```

**Note**: If end-line ≤ start-line, treated as single line comment.

---

### `bb review reply <pr-number> <comment-id> [--repo REPO] "message"`

**Purpose**: Reply to an existing comment

**Example**:
```bash
bb review reply 450 123456 --repo test_repo "Fixed in commit abc123"
```

---

### `bb review approve <pr-number> [--repo REPO]`

**Purpose**: Approve a PR

**Example**:
```bash
bb review approve 450 --repo test_repo
```

---

### `bb review request-change <pr-number> [--repo REPO]`

**Purpose**: Request changes on a PR

**Example**:
```bash
bb review request-change 450 --repo test_repo
```

**Note**: To add a comment explaining requested changes, use `bb review comment` separately.

---

## Token Comparison

### Before (separate commands):
```bash
bb pr view 450              # 2000 tokens
bb pr checks 450            # 500 tokens
bb pr comments 450          # 1200 tokens
bb pr diff 450 jwt.ts       # 1500 tokens
Total: 5200 tokens
```

### After (combined + unified diff):
```bash
bb review view 450          # 1200 tokens (PR + files + build + reviewers + comments)
bb review view 450 jwt.ts   # 800 tokens (unified diff + inline comments)
Total: 2000 tokens (62% savings)
```

**Unified diff format** is token-efficient and natively understood by LLMs trained on Git diffs.

---

## Implementation Status

✅ **Fully Implemented** - All commands available in current version:

**Authentication**:
- `bb auth` - Interactive login with secure password input (default action)
- `bb auth status` - Check current authentication

**Discovery**:
- `bb list repos` - List repositories in workspace
- `bb review list --repo <repo>` - List PRs with stats

**Review**:
- `bb review view <pr> --repo <repo>` - Complete PR context (metadata, files, build, reviewers, comments)
- `bb review view <pr> <file> --repo <repo>` - File diff with unified diff + inline comments

**Commenting**:
- `bb review comment <pr> --repo <repo> "msg"` - General comment
- `bb review comment <pr> <file> <line> --repo <repo> "msg"` - Inline comment (single line)
- `bb review comment <pr> <file> <start> <end> --repo <repo> "msg"` - Inline comment (line range)
- `bb review reply <pr> <comment-id> --repo <repo> "msg"` - Reply to comment

**Actions**:
- `bb review approve <pr> --repo <repo>` - Approve PR
- `bb review request-change <pr> --repo <repo>` - Request changes

**Removed** (as of v0.2.0):
- `bb auth login` - Use `bb auth` instead
- `bb list prs` - Use `bb review list` instead
- `bb comment *` - Moved to `bb review comment`
- `bb review view --approve` - Use `bb review approve` instead
- `bb review view --request-changes` - Use `bb review request-change` instead
- `bb review view --files` - View returns everything; parse JSON yourself
- `bb review view --comments` - View returns everything; parse JSON yourself
- `bb pr *` - All PR operations under `bb review` namespace
- `bb repo *` - Replaced by `bb list repos`

---

## Example Workflow

```bash
# 1. List open PRs
bb review list --repo test_repo

# 2. Get complete PR context (metadata, files, build status, reviewers, all comments)
bb review view 450 --repo test_repo

# 3. Review specific files (unified diff + inline comments)
bb review view 450 src/auth/jwt.ts --repo test_repo
bb review view 450 src/auth/middleware.ts --repo test_repo

# 4. Add comments
bb review comment 450 --repo test_repo "Looks good overall"
bb review comment 450 src/auth/jwt.ts 42 --repo test_repo "Consider adding type annotation"

# 5. Reply to existing comments
bb review reply 450 123456 --repo test_repo "Fixed in commit abc123"

# 6. Approve or request changes
bb review approve 450 --repo test_repo
# OR
bb review request-change 450 --repo test_repo
```

**Agent workflow**: Use `bb review view` to get full context, parse JSON to find issues, use `bb review comment` to post feedback, then `bb review approve` or `bb review request-change`.

Simple data fetching. Agent does the analysis.
