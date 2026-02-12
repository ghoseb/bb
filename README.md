# bbc

Bitbucket Cloud CLI for pull requests, comments, and pipelines. Designed for AI agent code review workflows with token-efficient output.

## Install

```bash
# Homebrew (macOS & Linux)
brew tap ghoseb/bbc
brew install bbc

# Go
go install github.com/ghoseb/bb/cmd/bbc@latest
```

## Authentication

```bash
# Interactive — prompts for workspace, username, app password
bbc auth

# Check status and token scopes
bbc auth status

# Environment variables (for CI / automation)
export BB_WORKSPACE=myworkspace
export BB_USERNAME=myuser
export BB_TOKEN=mytoken
```

Create an [App Password](https://bitbucket.org/account/settings/app-passwords/) with these scopes:
`read:user`, `read:workspace`, `read:repository`, `read:pullrequest`, `write:pullrequest`, `read:pipeline`

## Usage

### List

```bash
bbc list repos                              # List workspace repositories
bbc review list --repo <repo>               # List open PRs with stats
```

### View

```bash
bbc review view <pr> --repo <repo>          # PR overview (files, build, reviewers, comments)
bbc review view <pr> <file> --repo <repo>   # File diff with inline comments
```

### Comment

```bash
bbc review comment <pr> --repo <repo> "message"                        # General
bbc review comment <pr> <file> <line> --repo <repo> "message"          # Inline
bbc review comment <pr> <file> <start> <end> --repo <repo> "message"   # Line range
bbc review reply <pr> <comment-id> --repo <repo> "message"             # Reply

# Manage existing comments
bbc review comment <pr> --repo <repo> --edit <id> "new text"
bbc review comment <pr> --repo <repo> --delete <id>
bbc review comment <pr> --repo <repo> --resolve <id>
bbc review comment <pr> --repo <repo> --reopen <id>
```

### Actions

```bash
bbc review create <branch> --repo <repo> "title"     # Create PR
bbc review approve <pr> --repo <repo>                 # Approve
bbc review approve <pr> --repo <repo> --undo          # Remove approval
bbc review request-change <pr> --repo <repo>          # Request changes
bbc review request-change <pr> --repo <repo> --undo   # Remove request-change
```

## Output

Default output is **markdown** — optimized for LLM consumption with ~30-50% fewer tokens than JSON. Use `--json` for machine-parseable JSON:

```bash
# Markdown (default)
bbc review view 450 --repo myrepo

# JSON
bbc review view 450 --repo myrepo --json
bbc review list --repo myrepo --json | jq '.prs[].title'
```

### Markdown Features

- **Inline IDs** for API calls: `**Alice** (id:{uuid}) (comment:123456)`
- **Raw unified diffs** in fenced code blocks — zero escaping overhead
- **Smart rename handling** — pure renames show one line, renames with modifications show only the actual changed hunks
- **Scope checking** — `bbc auth status` reports missing token scopes upfront
- **Structured errors** — failed actions return clean JSON, not raw API error blobs

### Rename-Aware Diffs

Renamed files are handled intelligently instead of showing the entire file as new:

```
# Pure rename — no diff noise
renamed: old/path.py → new/path.py

# Rename with modifications — only actual changes shown
renamed: old/path.py → new/path.py

@@ -1,6 +1,6 @@
-from old.module import Foo
+from new.module import Foo
```

## License

MIT

## Credits & Thanks

[avivsinai/bitbucket-cli](https://github.com/avivsinai/bitbucket-cli/)
