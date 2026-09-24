<!--
Sync Impact Report
- Version change: [TEMPLATE] → 1.0.0 (initial ratification; prior file contained only unfilled placeholders)
- Modified principles: n/a (first concrete version)
- Added sections: Core Principles (5), Quality Gates, Development Workflow, Governance
- Removed sections: none
- Templates checked: .specify/templates/ (plan/spec/tasks) — no agent-specific references to principle
  names found requiring updates; templates read this file at runtime, no edits needed here per scope guard.
- Follow-up TODOs: none — RATIFICATION_DATE inferred from first commit date (2025-09-10)
-->

# git-blame-reviewer Constitution

## Core Principles

### I. Drop-In `git blame` Compatibility
git-blame-reviewer MUST accept the same command-line arguments and positional file
argument as `git blame`, and MUST produce output in the same human-readable and
`-porcelain` formats users already expect. New flags MAY be added, but existing
`git blame` flags MUST NOT change meaning or be removed. Any deviation from `git blame`'s
interface breaks the tool's core value proposition — letting users substitute it for
`git blame` without relearning a command.

### II. Test-First Development (NON-NEGOTIABLE)
All new behavior (parsing, formatting, GitHub/GitLab client logic) MUST be covered by
tests written before or alongside the implementation, following red-green-refactor.
Every source file with logic (`*.go`, excluding `main.go` wiring) MUST have a
corresponding `*_test.go`. `make test` MUST pass before any change is considered done.
Untested logic in a review-attribution tool is worse than no tool: a silent formatting
or parsing bug misattributes code ownership without any visible failure.

### III. Zero External Runtime Dependencies
The built artifact MUST remain a single, statically-linked Go binary with no runtime
dependency beyond the `git` executable and network access to the configured GitHub or
GitLab API. New third-party Go modules MAY be added only when they meaningfully reduce
implementation risk (e.g., a well-vetted API client), never for convenience wrappers
around functionality the standard library already provides. This keeps installation to
`go build`/a single binary copy, matching the project's stated distribution model.

### IV. Equal Treatment of GitHub and GitLab
Every feature that queries pull/merge request approvers, caching, or formatting MUST
work identically across both supported providers, with repository-type detection
handled transparently. A change that adds a capability to only one provider MUST either
add the GitLab/GitHub equivalent in the same change or be explicitly scoped and tracked
as a known gap in the PR description — it MUST NOT be presented as complete otherwise.
Users select this tool based on multi-provider support; silent asymmetry erodes trust.

### V. Simplicity and API Efficiency
Prefer the simplest implementation that satisfies the `git blame`-compatible contract;
do not add configuration surface, abstraction layers, or speculative extensibility
points not required by a current feature (YAGNI). Because GitHub/GitLab API calls are
rate-limited and commit-to-approver mappings are stable once merged, any code path that
resolves approvers MUST use the existing commit-level cache rather than re-querying the
API for previously-resolved commits.

## Quality Gates

- `make check` (tests + `golangci-lint run`) MUST pass locally before a change is
  proposed for merge, and MUST pass in CI before merge.
- `golangci-lint` version pinning in `Makefile` and `.github/workflows/ci.yml` MUST stay
  in sync; bumping one without the other is a defect, not a style choice, since a
  mismatch can silently disable or misconfigure lint rules.
- Coverage reporting (`make test-coverage`) is informational via Codecov; a coverage
  drop is not itself a merge blocker but MUST be explainable in review (e.g., new code
  is genuinely untestable glue, not skipped tests).

## Development Workflow

- Changes are proposed via pull request against `main`; CI (test, lint, build) MUST be
  green before merge.
- Commit messages and PR descriptions MUST state which provider(s) (GitHub, GitLab, or
  both) a change affects, per Principle IV.
- Tool version bumps affecting reproducibility (Go toolchain, golangci-lint) MUST update
  every reference location (`go.mod`, `Makefile`, CI workflow) in the same change.

## Governance

This constitution supersedes ad-hoc practice for this repository. Amendments are made
by editing `.specify/memory/constitution.md` directly (or via `/speckit-constitution`)
and MUST include an updated Sync Impact Report and version bump per the rules below.

**Versioning policy** (semantic versioning for governance):
- MAJOR: Backward-incompatible removal or redefinition of a principle (e.g., dropping
  the `git blame` compatibility guarantee).
- MINOR: A new principle or materially expanded section is added.
- PATCH: Wording clarifications, typo fixes, or non-semantic refinements.

**Compliance review**: Every PR MUST be checked against the Core Principles above
before merge, at minimum by confirming CI is green (Quality Gates) and that provider
parity (Principle IV) and test coverage (Principle II) are addressed in the diff or
explicitly called out as deferred.

**Version**: 1.0.0 | **Ratified**: 2025-09-10 | **Last Amended**: 2026-09-24
