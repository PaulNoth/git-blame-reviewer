# Data Model: Boundary Commit Blanking (`-b` flag)

No new entities are introduced. This feature extends two existing structs with a single
boolean field each.

## `BlameLine` (git.go)

| Field | Type | Change | Notes |
|-------|------|--------|-------|
| `CommitHash` | `string` | unchanged | Full commit SHA parsed from the porcelain commit-block header line |
| `Author` | `string` | unchanged | |
| `AuthorEmail` | `string` | unchanged | |
| `Date` | `string` | unchanged | |
| `LineNumber` | `int` | unchanged | |
| `Content` | `string` | unchanged | |
| `IsBoundary` | `bool` | **new** | Set `true` when the commit block's porcelain output contains a standalone `boundary` line (see research.md, Decision 1). Defaults `false`. Set once per commit block; every `BlameLine` sharing that commit carries the same value. |

**Validation rule**: `IsBoundary` MUST be derived solely from the presence of the
`boundary` porcelain field for that commit block — no other heuristic (e.g., missing
parent lookup) is used, keeping it consistent with `git`'s own determination.

## `BlameLineWithApproval` (formatter.go)

Embeds `BlameLine` (via struct embedding); no new fields required. `IsBoundary` is
already available through the embedded `BlameLine`.

## `OutputFormatter` (formatter.go)

| Field | Type | Change | Notes |
|-------|------|--------|-------|
| `ShowEmail` | `bool` | unchanged | |
| `Porcelain` | `bool` | unchanged | |
| `NoColors` | `bool` | unchanged | |
| `ShowBoundary` | `bool` | **new** | Mirrors the `-b` CLI flag. When `true`, `formatHuman` blanks the identifier column for lines where `IsBoundary` is `true`. Has no effect on `formatPorcelain` (research.md, Decision 2). |

## State / Transitions

None — all fields above are set once at parse/formatter-construction time and read
thereafter; there is no mutation or lifecycle to model.
