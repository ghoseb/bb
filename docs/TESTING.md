# Testing Guide

## Automated Tests

Run all unit and smoke tests:

```bash
go test ./...
```

Run only smoke tests:

```bash
go test -v ./test/smoke/...
```

Run with coverage:

```bash
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
```

## Manual Testing Checklist

### Prerequisites

1. Create a test Bitbucket Cloud workspace
2. Create a test repository with some commits
3. Create at least one pull request with:
   - Multiple file changes
   - Some comments (general and inline)
   - At least one reviewer
   - Pipeline runs (if available)

### Authentication Tests

- [ ] `bb auth --workspace X --username Y --token Z`
  - Should succeed with valid credentials
  - Should store credentials in system keyring
  - Should output JSON with username and workspace
  
- [ ] `bb auth` (interactive)
  - Should prompt for workspace, username, token
  - Should hide token input
  - Should store credentials on success
  
- [ ] `bb auth status`
  - Should show authenticated status
  - Should display correct username and workspace
  
- [ ] `bb auth` with invalid credentials
  - Should fail with clear error message
  - Should NOT store invalid credentials

### Repository Tests

- [ ] `bb list repos`
  - Should list all repositories in workspace
  - Should output valid JSON array
  - Should include repository names, slugs, and visibility

### Pull Request List Tests

- [ ] `bb review list --repo REPO`
  - Should list PRs with stats
  - Should include: files, additions, deletions, approvals
  
- [ ] `bb review list --repo REPO --state merged`
  - Should filter by state
  
- [ ] `bb review list --repo REPO --limit 5`
  - Should return at most 5 PRs

### Pull Request View Tests

- [ ] `bb review view PR_ID --repo REPO`
  - Should display complete PR context
  - Should include: metadata, files, build status, reviewers, comments
  - Should only show reviewers who took action (approved/declined)
  
- [ ] `bb review view 99999 --repo REPO` (invalid PR)
  - Should fail with clear error message
  
- [ ] `bb review view PR_ID` (missing --repo flag)
  - Should fail with "required flag" error

### File Diff Tests

- [ ] `bb review view PR_ID path/to/file.go --repo REPO`
  - Should display unified diff format
  - Should include inline comments on the file
  - Should show line numbers for inline comments
  - Should include author_id for mentions
  
- [ ] `bb review view PR_ID nonexistent.go --repo REPO`
  - Should fail with clear error (Bitbucket handles validation)

### Comment Tests

- [ ] `bb review comment PR_ID --repo REPO "Test comment"`
  - Should create a general comment
  - Should return comment ID and timestamp
  - Should be visible in review view output
  
- [ ] `bb review comment PR_ID path/to/file.go 42 --repo REPO "Inline comment"`
  - Should create an inline comment on single line
  - Should attach to correct file and line
  - Should be visible in file diff view
  
- [ ] `bb review comment PR_ID path/to/file.go 42 45 --repo REPO "Range comment"`
  - Should create an inline comment on line range
  - Should attach to correct file and lines
  
- [ ] `bb review comment PR_ID path/to/file.go 42 42 --repo REPO "Edge case"`
  - Should treat as single line when end <= start
  
- [ ] `bb review reply PR_ID COMMENT_ID --repo REPO "Reply"`
  - Should create a threaded reply
  - Should have parent_id set correctly
  - Should appear under parent comment

### Approval Tests

- [ ] `bb review approve PR_ID --repo REPO`
  - Should approve the PR
  - Should update reviewer status
  - Should be visible in review view output
  
- [ ] `bb review request-change PR_ID --repo REPO`
  - Should request changes on the PR
  - Should update reviewer status
  - Should NOT include comment (use separate comment command)

### Error Handling Tests

- [ ] Test with missing workspace (no --workspace, no BB_WORKSPACE env)
  - Should fail with clear error about authentication
  
- [ ] Test with network failure (disconnect network)
  - Should fail with timeout/connection error
  
- [ ] Test with invalid API token
  - Should fail with authentication error
  
- [ ] Test with rate limiting (make many requests quickly)
  - Should retry with exponential backoff
  - Should eventually succeed or fail gracefully

### Environment Variable Tests

- [ ] Set `BB_WORKSPACE=test` and run commands without --workspace
  - Should use environment variable
  
- [ ] Set `BB_USERNAME=user` and `BB_TOKEN=token` for auth login
  - Should use environment variables

### Cross-Platform Tests

Test on each platform:

- [ ] macOS (Intel and Apple Silicon)
  - Build and run all commands
  - Test keyring integration
  
- [ ] Linux
  - Build and run all commands
  - Test Secret Service keyring
  
- [ ] Windows
  - Build and run all commands
  - Test Windows Credential Manager

### Performance Tests

- [ ] Test with large PR (100+ files)
  - `bb pr view` should complete in < 5 seconds
  - `bb pr diff` should handle large diffs
  
- [ ] Test pagination with many repos
  - `bb repo list` should handle > 100 repos
  - Pagination should work correctly
  
- [ ] Test with many comments
  - `bb pr comments` with 100+ comments
  - Should paginate correctly

### JSON Output Validation

For each command, verify:

- [ ] Output is valid JSON
- [ ] Can be parsed with `jq`
- [ ] Contains expected fields
- [ ] Timestamps are in ISO 8601 format
- [ ] No extra fields in minimal mode

Example validation:

```bash
bb review view 123 --repo test | jq '.id, .title, .author'
bb review list --repo test | jq '.prs[].id'
```

### Security Tests

- [ ] Credentials stored in keyring
  - Not in plain text
  - Not in environment after auth
  
- [ ] API token not logged
  - Check debug output doesn't leak token
  
- [ ] File permissions on config files
  - Should be restricted

## Integration with AI Agents

Test the documented use cases:

```bash
# Get complete PR context
PR_DATA=$(bb review view 123 --repo test)
FILES=$(echo $PR_DATA | jq -r '.files[].path')

# For each file, get diff with unified diff format
for file in $FILES; do
  bb review view 123 "$file" --repo test > "$file.json"
done

# Post comments
bb review comment 123 --repo test "Automated review feedback"

# Add inline comments
bb review comment 123 src/main.go 42 --repo test "Fix this line"

# Approve if all checks pass
bb review approve 123 --repo test
```

## Regression Testing

After any changes:

1. Run full test suite: `go test ./...`
2. Run smoke tests: `go test -v ./test/smoke/...`
3. Build for all platforms: `goreleaser build --snapshot --clean`
4. Test at least one manual workflow end-to-end

## Known Limitations

- No offline mode
- API rate limiting may affect rapid testing
- Keyring issues in headless environments (use BB_ALLOW_INSECURE_STORE=1)
- Large diffs may be slow to fetch

## Reporting Issues

When reporting issues, include:

1. bb version (`bb --help` shows version in output)
2. Operating system and version
3. Go version used for build
4. Complete command executed
5. Full error output (redact sensitive info)
6. Expected vs actual behavior
