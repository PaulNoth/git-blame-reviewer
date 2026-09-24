# Quickstart: Validating the `-b` flag

## Prerequisites

- Repository built: `make build` (produces `./git-blame-reviewer`)
- Run from within any git repository with a file whose oldest line traces to the
  repository's root commit (this repo's own `go.mod` works, since line 2/3 trace to
  the initial commit `394b9cee`)
- A `GITHUB_TOKEN` or `GITLAB_TOKEN` env var set appropriately for the repo's remote
  (only needed for approver lookup; boundary blanking itself doesn't depend on it)

## Validate: default behavior unaffected (SC-003)

```bash
./git-blame-reviewer go.mod > /tmp/before.txt
git diff --no-index /tmp/before.txt /tmp/before.txt  # sanity: identical to itself
```

Confirm output matches what `make check`'s existing tests already assert for `go.mod`
before this feature — i.e., no `-b` flag means no change from current behavior.

## Validate: boundary line blanked in human output (SC-001, contracts/cli.md)

```bash
./git-blame-reviewer -b go.mod
```

Expected: the line(s) attributed to the initial commit (`394b9cee...`) show a blank
identifier column (spaces where the 8-character hash normally appears), while the line
attributed to a later commit still shows its hash. Compare boundary detection against
real `git`:

```bash
git blame -b go.mod   # spaces on the same lines that git-blame-reviewer blanks
```

## Validate: porcelain output unaffected by `-b` (FR-004a, Decision 2)

```bash
diff <(./git-blame-reviewer -porcelain go.mod) <(./git-blame-reviewer -b -porcelain go.mod)
```

Expected: no diff output (empty) — `-b` must not change `-porcelain` output at all.

## Validate: no-boundary-commits case (SC-002)

Pick a file/line range with no boundary commits in scope (e.g. a recently-added line
far from the repo root):

```bash
diff <(./git-blame-reviewer -L <n>,<n> <file>) <(./git-blame-reviewer -b -L <n>,<n> <file>)
```

Expected: no diff output — `-b` has no effect when there's nothing to blank.

## Automated coverage

Run the full suite, which will include new cases added under this feature:

```bash
make check
```

All new/changed behavior is covered by `git_test.go` (boundary parsing),
`formatter_test.go` (blanking in human output, no-op in porcelain), and
`integration_test.go` (end-to-end `-b` scenario against a real repo fixture).
