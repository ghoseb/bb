# Agent-Optimized Output Format

## Design Principles

1. **Token Efficiency**: Remove redundant fields, flatten nested structures
2. **Contextual Grouping**: Related information together (e.g., diff + comments)
3. **Summary First**: High-level stats before details
4. **Scannable**: LLMs can quickly identify what needs review
5. **Action-Oriented**: Format guides review workflow

## Command Outputs

### `bb review list --repo <repo>`

**Purpose**: Quick scan of open PRs

**Format**:
```json
{
  "total": 15,
  "prs": [
    {
      "number": 450,
      "title": "feat: add user authentication",
      "author": "alice",
      "state": "OPEN",
      "branch": "feature/auth → main",
      "created": "2h ago",
      "updated": "30m ago",
      "stats": {
        "files": 12,
        "additions": 450,
        "deletions": 120
      },
      "review": {
        "approved": 2,
        "changes_requested": 0,
        "pending": 1
      },
      "build": "passing"
    }
  ]
}
```

**Optimizations**:
- Relative timestamps ("2h ago" vs full ISO dates)
- Branch as single string (source → target)
- Summary stats inline
- Review status summarized
- Build status as simple string

---

### `bb review view <pr-number> --repo <repo>`

**Purpose**: Understand PR scope before diving into code

**Format**:
```json
{
  "pr": {
    "number": 450,
    "title": "feat: add user authentication",
    "description": "Adds JWT-based auth with refresh tokens...",
    "author": "alice",
    "state": "OPEN",
    "branch": "feature/auth → main",
    "created": "2h ago",
    "updated": "30m ago"
  },
  "review": {
    "approved": ["bob", "charlie"],
    "changes_requested": [],
    "pending": ["diane"]
  },
  "build": {
    "status": "passing",
    "checks": [
      {"name": "tests", "status": "passed"},
      {"name": "lint", "status": "passed"}
    ]
  },
  "stats": {
    "files_changed": 12,
    "additions": 450,
    "deletions": 120,
    "comments": 8
  },
  "files": [
    {
      "path": "src/auth/jwt.ts",
      "status": "added",
      "changes": "+150",
      "complexity": "high"
    },
    {
      "path": "src/auth/middleware.ts",
      "status": "modified",
      "changes": "+45-12",
      "complexity": "medium"
    },
    {
      "path": "tests/auth.test.ts",
      "status": "added",
      "changes": "+200",
      "complexity": "low"
    }
  ],
  "activity": {
    "last_commit": "30m ago",
    "last_comment": "1h ago",
    "last_review": "2h ago"
  }
}
```

**Optimizations**:
- Nested but shallow (max 2 levels)
- File complexity hints (based on size/changes)
- Activity summary for recency
- Review status as simple arrays
- Concise change notation (+45-12)

---

### `bb review view <pr-number> <file-path> --repo <repo>`

**Purpose**: Review specific file changes with context (unified diff format)

**Format**:
```json
{
  "file": "src/auth/jwt.ts",
  "status": "added",
  "stats": {
    "additions": 150,
    "deletions": 0
  },
  "diff": "@@ -0,0 +1,150 @@\n+import jwt from 'jsonwebtoken';\n+\n+export function generateToken(userId: string) {\n+  return jwt.sign({ userId }, process.env.JWT_SECRET);\n+}\n...",
  "comments": [
    {
      "line": 5,
      "author": "bob",
      "text": "Should we add expiration time?",
      "created": "1h ago",
      "resolved": false
    },
    {
      "line": 5,
      "author": "alice",
      "text": "Good point, will add in next commit",
      "created": "45m ago",
      "resolved": false
    }
  ]
}
```

**Optimizations**:
- Diff as single string (standard unified format)
- Comments inline with line numbers
- Resolved status for tracking
- File-level comments at line: 0

---

## Implementation Notes

1. **Relative Timestamps**: Convert ISO dates to human-friendly relative times
2. **File Complexity**: Heuristic based on lines changed (>100 = high, >50 = medium, else low)
3. **Comment Threading**: Nest replies under parent comments
4. **Diff Format**: Keep standard unified diff format (LLMs understand it well)
5. **Summary Statistics**: Always provide counts/totals at top level

## Token Savings

**Before (separate commands)**: ~5200 tokens
- pr view: ~2000 tokens
- pr checks: ~500 tokens  
- pr comments: ~1200 tokens
- pr diff: ~1500 tokens

**After (combined + unified diff)**: ~2000 tokens
- review view (all context): ~1200 tokens
- review view file (unified diff): ~800 tokens

**Savings**: 62% overall for combined view approach

## Agent Workflow Example

```bash
# 1. List open PRs
bb review list --repo my-repo

# 2. Get complete PR context (metadata, files, build, reviewers, comments)
bb review view 450 --repo my-repo

# 3. Review specific files (unified diff + inline comments)
bb review view 450 src/auth/jwt.ts --repo my-repo
bb review view 450 src/auth/middleware.ts --repo my-repo

# 4. Add comments
bb review comment 450 --repo my-repo "LGTM overall"
bb review comment 450 src/auth/jwt.ts 42 --repo my-repo "Add error handling"

# 5. Reply to comments
bb review reply 450 123456 --repo my-repo "Fixed"

# 6. Approve or request changes
bb review approve 450 --repo my-repo
# OR
bb review request-change 450 --repo my-repo
```
