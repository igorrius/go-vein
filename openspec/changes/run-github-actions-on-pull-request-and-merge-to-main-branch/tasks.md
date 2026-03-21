## 1. Workflow Setup

- [x] 1.1 Add `.github/workflows/pr-main-ci.yml` with `pull_request` and `push` triggers
- [x] 1.2 Include steps to checkout code, setup Go toolchain, and run `go test ./...`
- [x] 1.3 Ensure workflow name / status check name is stable for branch protection (`PR and main CI`)

## 2. Documentation

- [x] 2.1 Add or update README/CONTRIBUTING section describing the PR check requirement with branch protection details
- [x] 2.2 Document expected behavior for PR gating and merge rules

## 3. Branch Protection & Validation

- [x] 3.1 Configure `main` branch protection requiring `PR and main CI` status check
- [x] 3.2 Validate PRs are blocked until check passes and allowed after success
- [x] 3.3 Communicate and coordinate with maintainers about enforcement
