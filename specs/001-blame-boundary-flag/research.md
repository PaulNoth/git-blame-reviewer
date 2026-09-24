# Research: Boundary Commit Blanking (`-b` flag)

## Decision 1: What counts as a "boundary commit" and how the tool learns about it

**Decision**: A boundary commit is exactly what `git blame`'s own porcelain output
tags as one: for a given commit block in `git blame --line-porcelain` (or
`--porcelain`) output, `git` emits a standalone line containing the literal word
`boundary` when that commit has no further ancestry considered within the current
blame walk (typically the repository's root commit, since the tool never passes
`--root` and requests no revision range). This field is emitted unconditionally,
regardless of whether `-b` is passed to the underlying `git blame` invocation.

**Rationale**: Verified empirically against the local `git` binary (see Decision 2)
against this repository's own history — `go.mod`'s line pinned to the initial commit
(`394b9cee...`) produces a `boundary` line in `--line-porcelain` output. This means the
tool does not need to change how it invokes `git blame` (`git.go:ExecuteGitBlame`
already always uses `--line-porcelain`/`--porcelain`); it only needs to parse a field
that's already present in the output it already requests.

**Alternatives considered**:
- *Pass `-b` through to the underlying `git blame` subprocess and blank the hash
  ourselves by detecting an empty/whitespace hash field.* Rejected: `-b` does not alter
  porcelain output at all (see Decision 2), so this would have no effect and silently
  fail to blank anything.
- *Reimplement boundary detection by checking `git log --root` reachability manually.*
  Rejected: unnecessary — `git blame` already computes and reports this per line;
  duplicating it would add complexity and risk disagreeing with `git`'s own notion of
  "boundary" (Constitution Principle V: simplicity; Principle I: exact compatibility).

## Decision 2: Whether `-b` affects `--porcelain`/`--line-porcelain` output

**Decision**: `-b` has **no effect** on porcelain output. It only changes the default
human-readable format, where it blanks the SHA (normally shown as `^abcdef12` for a
boundary commit without `--root`, or a plain hash with `--root`) to spaces of the same
width. Porcelain output already conveys boundary status via the `boundary` field
unconditionally, so `git` does not also blank the hash there — machine consumers need
the real hash regardless.

**Rationale**: Directly tested against the local `git` installation used by this
project:
```
git blame --line-porcelain go.mod      # commit block includes "boundary" line
git blame --line-porcelain -b go.mod   # byte-identical output, "boundary" line still present, hash line unaffected
git blame -b go.mod                    # boundary line's hash column is blanked (spaces) in human format
```
This directly resolves the ambiguity found in the initial spec draft (which assumed
`-b` also blanks `-porcelain` output) — spec.md has been corrected accordingly
(FR-002, FR-004a, acceptance scenario 2, edge cases, SC-001).

**Alternatives considered**:
- *Blank the identifier in our own `-porcelain` output too, since it seems more
  "consistent."* Rejected: Constitution Principle I requires matching `git blame`'s
  actual behavior, not inventing a more "consistent" one; real `git blame -b -porcelain`
  users rely on porcelain always carrying the true hash.

## Decision 3: Where blanking happens in the existing formatter

**Decision**: Blanking is applied only inside `OutputFormatter.formatHuman`, driven by
a new `ShowBoundary bool` field on `OutputFormatter` (set from the new `-b` CLI flag).
`formatPorcelain` is left untouched (per Decision 2).

**Rationale**: `formatter.go` already separates `formatHuman` and `formatPorcelain` as
distinct methods with independent field-selection logic (`getAuthorName`,
`getDateString`), so this follows the existing pattern with minimal surface area —
no new abstraction needed (Constitution Principle V).

**Alternatives considered**:
- *Blank `CommitHash` on the `BlameLine`/`BlameLineWithApproval` struct itself before
  formatting.* Rejected: this would destroy the real hash before the porcelain
  formatter (which must not blank it) ever sees it, and would also make the value
  unavailable for any future feature needing the true hash (e.g., PR/MR lookup already
  keys off `CommitHash` in `main.go:buildBlameLinesWithApprovals` — mutating it in place
  would be a correctness hazard, not just a layering one).

## Decision 4: Column width for the blanked identifier in human output

**Decision**: The blanked field occupies exactly `shortHashLength` (8) characters —
the same width as a normal abbreviated hash already printed by `formatHuman` — so
column alignment with the parenthesis/author block is unaffected.

**Rationale**: Matches real `git blame -b`'s behavior (blank field is the same width as
the abbreviated hash column) and requires no change to the existing width-calculation
logic in `formatHuman` (`maxAuthorWidth`, `maxLineNumWidth`), which is computed
independently of the hash column already.

**Alternatives considered**: None meaningfully different — the hash column is
currently a fixed `shortHashLength` width in `formatHuman`, so blanking it to spaces of
the same length is the only option consistent with existing output alignment.
