# Implementation Plan: Boundary Commit Blanking (`-b` flag)

**Branch**: `001-blame-boundary-flag` | **Date**: 2026-09-24 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-blame-boundary-flag/spec.md`

## Summary

Add a `-b` flag that, in human-readable output only, blanks the commit-identifier
column for lines attributed to a boundary commit (a commit with no further ancestry
within the blamed history — e.g. a repository's root commit), exactly matching
`git blame -b`. Boundary status is already emitted by `git blame --line-porcelain` as a
standalone `boundary` field per commit block, independent of whether `-b` is passed to
the underlying `git blame` invocation, so no change to how the tool shells out to `git`
is needed — only parsing that field and applying it during human-readable rendering.
`-porcelain` output is unaffected by `-b`, matching real `git blame`'s own behavior
(verified empirically against the local `git` binary during planning; see research.md).

## Technical Context

**Language/Version**: Go 1.25.1 (existing `go.mod`)

**Primary Dependencies**: Go standard library only (`flag`, `os/exec`, `bufio`,
`strings`); no new dependencies required

**Storage**: N/A (in-memory per-invocation; existing per-commit approval cache in
`main.go` is unaffected)

**Testing**: `go test ./...` (existing `*_test.go` files, table-driven style already
used in `git_test.go`, `formatter_test.go`, `main_test.go`)

**Target Platform**: Same as existing binary — any platform with `git` on `PATH`
(developed/CI'd on Linux, used cross-platform)

**Project Type**: Single-binary CLI tool (existing flat repo-root layout, no `src/`)

**Performance Goals**: No new performance requirement; must not add extra `git`
subprocess invocations or extra API calls beyond what already runs per line/commit

**Constraints**: Must not change output for any existing flag when `-b` is omitted
(Constitution Principle I); must not add a runtime dependency (Principle III); must
work identically regardless of GitHub/GitLab repository type (Principle IV) — boundary
detection is purely a `git` blame concept, unrelated to the provider, so this is
naturally satisfied and needs no provider-specific code

**Scale/Scope**: Small, additive change confined to `git.go` (parsing), `formatter.go`
(rendering), and `main.go` (flag wiring); no new files needed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Drop-In `git blame` Compatibility** — PASS. `-b` is an existing `git blame` flag
  being added with identical semantics (blank identifier for boundary commits, in
  human-readable output only, per research.md). No existing flag's behavior changes.
- **II. Test-First Development (NON-NEGOTIABLE)** — PASS (planned). New behavior in
  `git.go` (boundary parsing) and `formatter.go` (blanking) will get corresponding
  `*_test.go` cases before/alongside implementation, per tasks.md.
- **III. Zero External Runtime Dependencies** — PASS. No new module required; parsing
  uses `strings`/`bufio` already imported in `git.go`.
- **IV. Equal Treatment of GitHub and GitLab** — PASS. Boundary detection happens before
  any GitHub/GitLab client call and does not touch `client.go`, `github.go`, or
  `gitlab.go`; behavior is identical for both providers by construction.
- **V. Simplicity and API Efficiency** — PASS. No new API calls, no new cache, no new
  configuration surface beyond the single boolean flag `git blame` itself defines.

No violations to justify; Complexity Tracking section is omitted.

## Project Structure

### Documentation (this feature)

```text
specs/001-blame-boundary-flag/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
# Existing flat single-binary CLI layout (no src/ subdirectory) — this feature
# extends existing files in place, no new files or directories.
.
├── main.go              # Flag wiring: add `-b` bool flag, thread through to
│                         # ExecuteGitBlame (unchanged call) and OutputFormatter
├── main_test.go
├── git.go                # BlameLine gains IsBoundary; parseGitBlameOutput parses
│                         # the porcelain "boundary" field
├── git_test.go
├── formatter.go          # OutputFormatter gains ShowBoundary; formatHuman blanks
│                         # the identifier column for boundary lines when set
├── formatter_test.go
├── client.go / github.go / gitlab.go   # Unchanged — boundary detection is
│                                        # provider-agnostic (Constitution IV)
└── integration_test.go   # Extended with an end-to-end `-b` scenario
```

**Structure Decision**: Single-project flat layout (matches the existing repository —
no `src/`, `tests/`, or multi-package split). This feature is purely additive within
the three existing files listed above; no new source files are introduced.

## Complexity Tracking

*No violations — section intentionally left empty.*
