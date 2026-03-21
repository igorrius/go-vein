## ADDED Requirements

### Requirement: GitHub Actions workflow runs for pull requests and main pushes
The repository SHALL have a GitHub Actions workflow that triggers on pull request events and push events to `main`.

#### Scenario: Workflow triggers on pull request
- **WHEN** a pull request is opened, synchronized, or reopened targeting any branch
- **THEN** the workflow runs and reports status checks on the pull request

#### Scenario: Workflow triggers on main push
- **WHEN** a commit is pushed to `main`
- **THEN** the workflow runs and reports status checks on the push

### Requirement: Branch protection requires passing status checks for main
The repository SHALL configure `main` branch protection to require the new workflow status checks to pass before allowing merge.

#### Scenario: Merge blocked when workflow fails
- **WHEN** a pull request targeting `main` has a failing or pending workflow status check
- **THEN** the PR cannot be merged to `main`

#### Scenario: Merge allowed when workflow passes
- **WHEN** a pull request targeting `main` has all required workflow status checks passing
- **THEN** the PR can be merged to `main`

### Requirement: Merge process updates documentation
The repository SHALL include documentation describing the PR CI gating policy and merge criteria in the repository README or CONTRIBUTING guidelines.

#### Scenario: Documentation exists
- **WHEN** a maintainer checks repository policy
- **THEN** they find a section describing that PRs require passing GitHub Actions checks before `main` merge
