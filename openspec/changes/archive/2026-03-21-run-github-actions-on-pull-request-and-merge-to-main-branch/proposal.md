## Why

To enforce CI validation and guardrails, run GitHub Actions automatically on pull requests and require a successful workflow run before merging into `main`.

## What Changes

- Add a scoped GitHub Actions workflow that triggers on `pull_request` and `push` to `main`.
- Enforce branch protection requiring passing status checks from this workflow before merge.
- Add documentation in repository policy about PR gating and auto-merge conditions.

## Capabilities

### New Capabilities
- `github-actions-pr-gating`: Run CI on PRs and enforce merge-only-if-passing via branch protection setup.

### Modified Capabilities
- `none`: No existing capability requirements change (new capability only).

## Impact

- Affected areas: repository CI configuration (`.github/workflows/`), developer workflow, merge policy.
- No code dependencies in Go runtime; primarily repo-level ops and documentation.
