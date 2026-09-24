---

description: "Task list for the `-b` (boundary commit blanking) feature"
---

# Tasks: Boundary Commit Blanking (`-b` flag)

**Input**: Design documents from `/specs/001-blame-boundary-flag/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: Included and REQUIRED — Constitution Principle II (Test-First Development,
NON-NEGOTIABLE) mandates tests written before/alongside implementation for all new
behavior, with `make test` passing before any change is considered done.

**Organization**: This feature has a single user story (US1, P1) per spec.md. Tasks are
grouped as Setup → Foundational (shared struct fields) → User Story 1 → Polish.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Maps the task to US1 (the feature's only user story)
- Include exact file paths in descriptions

## Path Conventions

Single flat-project Go CLI at repository root (no `src/`/`tests/` subdirectories) —
matches plan.md's Project Structure. All paths below are relative to repo root.

---

## Phase 1: Setup

**Purpose**: Confirm a clean baseline before making changes

- [X] T001 Run `make check` from the repository root and confirm it passes, establishing
      the pre-change baseline referenced by SC-003 in spec.md (no source changes in
      this task)

**Checkpoint**: Baseline green; safe to start Foundational work

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the data fields both the parsing and rendering logic in User Story 1
depend on. Adding these fields alone changes no observable behavior (both default to
`false`), so it does not violate Test-First ordering — the tests that exercise them
belong to Phase 3.

- [X] T002 [P] Add `IsBoundary bool` field to the `BlameLine` struct in `git.go`
      (immediately after the existing `Content` field), per data-model.md
- [X] T003 [P] Add `ShowBoundary bool` field to the `OutputFormatter` struct in
      `formatter.go` (alongside `ShowEmail`/`Porcelain`/`NoColors`), per data-model.md

**Checkpoint**: Struct fields exist; User Story 1 implementation can begin

---

## Phase 3: User Story 1 - Boundary lines are visually distinguishable with `-b` (Priority: P1) 🎯 MVP

**Goal**: A `-b` flag that blanks the commit-identifier column for boundary-commit
lines in human-readable output only, exactly matching `git blame -b` (contracts/cli.md).

**Independent Test**: Run the built binary with `-b` against `go.mod` in this
repository (whose lines 2-3 trace to the root commit `394b9cee...`) and confirm the
identifier column is blank on those lines while `-porcelain` output stays byte-identical
with/without `-b` (quickstart.md).

### Tests for User Story 1 ⚠️

> Write these tests FIRST; confirm they FAIL before doing the corresponding
> implementation task.

- [X] T004 [P] [US1] Add test(s) in `git_test.go` asserting `parseGitBlameOutput` sets
      `IsBoundary = true` on every `BlameLine` belonging to a commit block containing a
      standalone `boundary` porcelain line, and `false` for commit blocks without one
      (research.md Decision 1) — construct the porcelain input as an inline string
      fixture, following the existing table-driven style in this file
- [X] T005 [P] [US1] Add test(s) in `formatter_test.go` asserting `formatHuman` renders
      the identifier column as `shortHashLength` space characters (preserving column
      alignment) for a `BlameLineWithApproval` where `IsBoundary` is `true` and
      `OutputFormatter.ShowBoundary` is `true`
- [X] T006 [P] [US1] Add test(s) in `formatter_test.go` asserting `formatHuman` renders
      the real (unblanked) identifier for a boundary line when `ShowBoundary` is
      `false` (default/regression case — SC-003), and for a non-boundary line
      regardless of `ShowBoundary`
- [X] T007 [P] [US1] Add test(s) in `formatter_test.go` asserting `formatPorcelain`
      output is byte-identical whether `ShowBoundary` is `true` or `false`, for input
      containing at least one boundary line (FR-004a / contracts/cli.md)
- [X] T008 [P] [US1] Add test(s) in `main_test.go` covering the `-b` flag: that
      `runGitReviewBlame`'s signature/wiring accepts a boundary-flag parameter and
      passes it through to `NewOutputFormatter`, and that `showHelp()`'s output
      documents `-b` (mirrors the existing `-porcelain`/`-show-email` coverage in this
      file)

### Implementation for User Story 1

- [X] T009 [US1] Implement boundary detection in `parseGitBlameOutput` in `git.go`:
      recognize a line equal to `"boundary"` and set `IsBoundary = true` on the
      in-progress `currentLine` (mirroring the existing `author `/`author-mail `/etc.
      `else if` chain) — makes T004 pass
- [X] T010 [US1] Implement blanking in `formatHuman` in `formatter.go`: when
      `f.ShowBoundary && line.IsBoundary`, render the identifier field as
      `strings.Repeat(" ", shortHashLength)` instead of the truncated `CommitHash`,
      leaving every other field's rendering untouched — makes T005 and T006 pass
- [X] T011 [US1] Confirm `formatPorcelain` in `formatter.go` requires no code change
      (it must not read `ShowBoundary` at all) — run T007 and record the result; only
      touch `formatPorcelain` if T007 fails, and if so keep the fix to "never branch on
      `ShowBoundary`" per research.md Decision 2
- [X] T012 [US1] Wire the flag end-to-end in `main.go`: add `boundary =
      flag.Bool("b", false, "Show blank commit identifier for boundary commits")` next
      to the existing flag declarations, thread it through `runGitReviewBlame`'s
      signature and its `NewOutputFormatter(...)` call, update `NewOutputFormatter` in
      `formatter.go` to accept and set `ShowBoundary`, and add a `-b` line (with a
      one-line description) to the `Options`/`Examples` sections of `showHelp()` —
      makes T008 pass
- [X] T013 [US1] Add an end-to-end scenario in `integration_test.go` that runs the
      built blame pipeline with the boundary flag enabled against a fixture repository
      containing a root-commit line, asserting the identifier is blank in
      human-readable output and that `-porcelain` output is unchanged by the flag
      (mirrors quickstart.md's validation steps)

**Checkpoint**: User Story 1 is fully implemented, tested, and independently verifiable
via quickstart.md

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final verification against the constitution's Quality
Gates

- [X] T014 [P] Document the `-b` flag in `README.md`'s "Command Line Options" list and
      add a `-b` example under "Usage", consistent with how `-L`/`-porcelain`/
      `-show-email` are already documented
- [X] T015 Run through every scenario in `quickstart.md` manually against the built
      binary (`make build`) and confirm actual output matches each "Expected" block
- [X] T016 Run `make check` (tests + `golangci-lint run`) and confirm it passes cleanly,
      satisfying the constitution's Quality Gates before this feature is considered done

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS Phase 3
- **User Story 1 (Phase 3)**: Depends on Phase 2. Within the phase: T004-T008 (tests,
  all parallelizable, different files) before their corresponding implementation tasks
  T009-T012; T013 (integration test) depends on T009, T010, T012 all being complete
- **Polish (Phase 4)**: Depends on Phase 3 being complete

### Within User Story 1

- T004 → T009 (test before implementation, same behavior)
- T005, T006 → T010 (tests before implementation, same behavior)
- T007 → T011 (test before verification/guard)
- T008 → T012 (test before implementation, same behavior)
- T009, T010, T012 → T013 (integration test needs all three pieces wired together)

### Parallel Opportunities

- T002 and T003 (Phase 2) can run in parallel — different files
- T004, T005, T006, T007, T008 (Phase 3 tests) can all be written in parallel —
  different files, no shared dependency
- T014 (Phase 4 docs) can run in parallel with T015/T016 since it touches only
  `README.md`

---

## Parallel Example: User Story 1 test-writing batch

```bash
# Launch all Phase 3 test-writing tasks together (before any implementation task):
Task: "Add boundary-parsing test(s) in git_test.go (T004)"
Task: "Add formatHuman blanking test(s) in formatter_test.go (T005)"
Task: "Add formatHuman regression test(s) in formatter_test.go (T006)"
Task: "Add formatPorcelain unaffected-by-flag test(s) in formatter_test.go (T007)"
Task: "Add -b flag wiring test(s) in main_test.go (T008)"
```

---

## Implementation Strategy

### MVP = the entire feature

This feature has exactly one user story (P1), so there is no smaller MVP slice than
"the `-b` flag works end-to-end." Complete phases in order: Setup → Foundational →
User Story 1 → Polish, then stop — the feature is done.

## Notes

- [P] tasks touch different files and have no unmet dependency on an incomplete task
- Verify each Phase 3 test task fails before starting its paired implementation task
- Commit after each task or logical group
- Constitution Principle IV (equal GitHub/GitLab treatment) requires no action here:
  boundary detection happens entirely before any provider-specific code runs
  (plan.md, Constitution Check)
