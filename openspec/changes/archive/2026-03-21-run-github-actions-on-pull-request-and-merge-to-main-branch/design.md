## Context

Current repository has Go code for a message bus system and no explicit PR gating policy in code or docs. CI may run manually or via existing workflows, but there is no formal workflow to enforce that pull requests against `main` run tests and are blocked until success. This change adds a dedicated GitHub Actions workflow with branch protection expectations.

## Goals / Non-Goals

**Goals:**
- Add a GitHub Actions workflow that runs on pull requests and main pushes.
- Document required checks and enable enforcement via branch protection.
- Keep integration minimal with existing repo structure and workflow naming conventions.

**Non-Goals:**
- Do not introduce new production code changes in Go for core product behavior.
- Do not implement complex GitHub Apps or external approvals process.

## Decisions

- Use `pull_request` and `push` events in a single workflow file (`.github/workflows/pr-main-ci.yml`) for simplicity.
- Include steps: checkout, setup Go, run `go test ./...`, and optionally lint steps if already available or easy to add.
- Keep workflow names clear (e.g., `PR and main CI`) for branch protection readable check names.
- For branch protection, document steps for repository admin to configure status check requirement as `PR and main CI`.

## Risks / Trade-offs

- [Risk] Adding required status checks may block existing pending PRs until configured.
  → Mitigation: Communicate with maintainers and coordinate protection rule updates.
- [Risk] Workflow failure for flaky tests could block merges.
  → Mitigation: Investigate flaky tests, stabilize test suite before enforcing strictly.
- [Risk] If branch protection is not configured correctly, the workflow may run but not block merges.
  → Mitigation: include explicit instructions and verify policies after release.

## Migration Plan

1. Add `.github/workflows/pr-main-ci.yml` workflow file.
2. Create docs update in `README.md`/`CONTRIBUTING.md` describing gating and branch protection steps.
3. Apply branch protection rule in GitHub settings: require `PR and main CI` check for `main` branch.
4. Monitor first PRs and adjust workflow triggers or checks as needed.

## Open Questions

- Should the workflow also include additional linting or static analysis steps beyond `go test`?
- Does the project require separate check names for PR vs main runs, or one combined check is acceptable?
