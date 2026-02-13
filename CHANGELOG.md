# Changelog

All notable changes to this project will be documented here.

## bb v0.2.0 (2026-02-13)

### Additions

- Agent-Optimized `bb review` Commands: Introduced a new set of `bb review` commands specifically designed for AI agent code review workflows, focusing on extreme token efficiency.
  - `bb review list --repo <repo>`: Lists pull requests with key statistics (files, additions, deletions, approvals).
  - `bb review view <pr> --repo <repo>`: Retrieves complete PR context in a single call, including metadata, files, build status, reviewers, and comments.
  - `bb review view <pr> <file> --repo <repo>`: Displays a specific file's diff in raw unified diff format, including inline comments with author IDs.

- Token Efficiency Enhancements:
  - Raw Unified Diff: Diffs are now returned as raw unified diff strings to significantly reduce token usage compared to JSON representations.
  - API Call Aggregation: Commands now aggregate data from multiple Bitbucket API endpoints into single responses, reducing overhead.
  - Optimized Reviewer Data: Only includes reviewers who have taken action (approved or requested changes), filtering out pending ones.
  - Author IDs for Mentions: All comments now include `author_id` (UUID) to facilitate proper Bitbucket @mentions by AI agents.
  - Accurate Build Status: Integrates `GetPRPipelines()` for precise build status reporting, rather than relying on deployment status.

These additions streamline the process for AI agents interacting with Bitbucket Cloud for code review tasks.
