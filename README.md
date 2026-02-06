# bb

Bitbucket Cloud CLI for pull requests, comments, and pipelines.

## Install

```bash
# Homebrew (macOS & Linux)
brew tap ghoseb/bb
brew install bb

# Go
go install github.com/ghoseb/bb/cmd/bb@latest
```

## Authentication

```bash
# Interactive — prompts for workspace, username, app password
bb auth

# Check status and token scopes
bb auth status

# Environment variables (for CI / automation)
export BB_WORKSPACE=myworkspace
export BB_USERNAME=myuser
export BB_TOKEN=mytoken
```

Create an [App Password](https://bitbucket.org/account/settings/app-passwords/) with these scopes:
`read:user`, `read:workspace`, `read:repository`, `read:pullrequest`, `write:pullrequest`

## Usage

### List

```bash
bb list repos                              # List workspace repositories
bb review list --repo <repo>               # List open PRs with stats
```

### View

```bash
bb review view <pr> --repo <repo>          # PR overview (files, build, reviewers, comments)
bb review view <pr> <file> --repo <repo>   # File diff with inline comments
```

### Comment

```bash
bb review comment <pr> --repo <repo> "message"                        # General
bb review comment <pr> <file> <line> --repo <repo> "message"          # Inline
bb review comment <pr> <file> <start> <end> --repo <repo> "message"   # Line range
bb review reply <pr> <comment-id> --repo <repo> "message"             # Reply

# Manage existing comments
bb review comment <pr> --repo <repo> --edit <id> "new text"
bb review comment <pr> --repo <repo> --delete <id>
bb review comment <pr> --repo <repo> --resolve <id>
bb review comment <pr> --repo <repo> --reopen <id>
```

### Actions

```bash
bb review create <branch> --repo <repo> "title"     # Create PR
bb review approve <pr> --repo <repo>                 # Approve
bb review approve <pr> --repo <repo> --undo          # Remove approval
bb review request-change <pr> --repo <repo>          # Request changes
bb review request-change <pr> --repo <repo> --undo   # Remove request-change
```

## Output

Default output is **markdown** (optimized for LLM consumption). Use `--json` for machine-parseable JSON:

```bash
# Markdown (default)
bb review view 450 --repo test

# JSON
bb review view 450 --repo test --json
bb review list --repo test --json | jq '.prs[].title'
```

## License

MIT

## Credits & Thanks

- [avivsinai/bitbucket-cli](https://github.com/avivsinai/bitbucket-cli/)
