# CLI Contract: `-b` flag

## Flag

```
-b
```

- Type: boolean, no argument (same as `git blame -b`)
- Default: `false` (disabled — current behavior)
- Composable with all existing flags: `-L <start>,<end>`, `-porcelain`, `-show-email`

## Behavioral contract

### Human-readable output (default, no `-porcelain`)

For each output line whose blamed commit is a boundary commit:

- **`-b` absent (current/default behavior)**: identifier column shows the commit hash
  as today (unchanged).
- **`-b` present**: identifier column is blank — rendered as `shortHashLength` (8)
  space characters — preserving column alignment with non-boundary lines. All other
  fields (author/approver name, date, line number, content) render unchanged.

For lines whose blamed commit is NOT a boundary commit, `-b` has no effect regardless
of presence.

### Porcelain output (`-porcelain`)

`-b` has **no effect** on `-porcelain` output in any respect. `-b -porcelain` and
`-porcelain` alone MUST produce byte-identical output.

## Examples

```text
$ git-blame-reviewer go.mod
c5155e2f (Pavol Pidanič 2025-10-03 11:03:44 +0200 1) module git-blame-reviewer
394b9cee (Pavol Pidanič 2025-09-10 14:01:06 +0200 2)
394b9cee (Pavol Pidanič 2025-09-10 14:01:06 +0200 3) go 1.25.1

$ git-blame-reviewer -b go.mod
c5155e2f (Pavol Pidanič 2025-10-03 11:03:44 +0200 1) module git-blame-reviewer
         (Pavol Pidanič 2025-09-10 14:01:06 +0200 2)
         (Pavol Pidanič 2025-09-10 14:01:06 +0200 3) go 1.25.1

$ git-blame-reviewer -b -porcelain go.mod
# identical to `git-blame-reviewer -porcelain go.mod` — identifier fields intact
```

(Note: the two boundary lines above show the underlying git author because no
PR/MR approver could be resolved for that commit — reviewer-attribution fallback is
existing, unrelated behavior.)

## Exit codes / errors

No new exit codes or error conditions. `-b` is parsed the same way `-porcelain` and
`-show-email` already are (via `flag.Bool`); malformed flag usage produces the same
`flag` package error output the tool already relies on.
