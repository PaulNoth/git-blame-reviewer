# Feature Specification: Boundary Commit Blanking (`-b` flag)

**Feature Branch**: `001-blame-boundary-flag`

**Created**: 2026-09-24

**Status**: Draft

**Input**: User description: "I want that the project mimics `-b` argument from `git blame`"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hide noisy boundary identifiers in human-readable output (Priority: P1)

A developer runs the tool on a file whose history reaches back to the repository's
initial commit (or another commit `git blame` treats as a boundary, such as the root of
a shallow clone or a revision-limited blame). Today those lines show a full identifier
column the same as any other line, which can be misread as a meaningful reviewer
attribution when in fact no further history/approval exists to inspect. The developer
passes the same `-b` flag they already use with `git blame` and expects the commit
identifier column for those specific lines to be blanked out, exactly as `git blame -b`
does, so boundary lines are visually distinguishable from normally-attributed lines.

**Why this priority**: This is the entire scope of the requested feature — without it,
`-b` has no observable effect and the "drop-in replacement for `git blame`" promise is
broken for this flag.

**Independent Test**: Run the tool with `-b` against a file whose blame reaches a
boundary commit, and confirm the commit identifier field on boundary lines is blank
(rendered as spaces) while non-boundary lines are unaffected. Can be tested standalone
without any other flag.

**Acceptance Scenarios**:

1. **Given** a file whose oldest attributed line traces to the repository's root
   commit, **When** the tool is run with `-b` in human-readable (default) output,
   **Then** the commit identifier shown for that line is blank (space-padded to the
   same column width as non-boundary identifiers), while the reviewer/approver name,
   date, and line content for that line are still shown normally.
2. **Given** the same file, **When** the tool is run with `-b -porcelain`, **Then**
   output is unaffected by `-b`: the commit identifier and all other porcelain fields
   are emitted in full, matching `git blame -b -porcelain`'s actual behavior, where
   boundary status is already conveyed machine-readably and the identifier is never
   blanked in porcelain output.
3. **Given** the same file, **When** the tool is run without `-b`, **Then** boundary
   lines show their full commit identifier exactly as they do today (no regression to
   current default behavior).
4. **Given** a file with no boundary commits in its blame range, **When** the tool is
   run with `-b`, **Then** output is identical to running without `-b`.

### Edge Cases

- What happens when every line in the file traces back to a boundary commit (e.g., the
  entire file was added in the initial commit)? All identifier columns MUST be blank,
  and human-readable column alignment MUST still be computed correctly.
- How does the system handle `-b` combined with `-show-email`? The identifier blanking
  applies only to the commit identifier field; the reviewer name/email field is
  unaffected and continues to follow `-show-email`.
- How does the system handle `-b` combined with `-L <range>`? Blanking applies only to
  boundary lines that fall within the requested range, consistent with how `-L` already
  restricts output.
- What happens when `-b` is combined with `-porcelain`? `-b` has no effect on porcelain
  output at all — matching real `git blame`, which only blanks the identifier in
  human-readable output and leaves porcelain output unchanged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The tool MUST accept a `-b` command-line flag with no argument, matching
  `git blame -b`'s syntax.
- **FR-002**: When `-b` is supplied, the tool MUST blank the commit identifier field, in
  human-readable output only, for any line whose blamed commit is a boundary commit.
  `-porcelain` output MUST be unaffected by `-b`, matching real `git blame`'s behavior.
- **FR-003**: A commit MUST be treated as a boundary commit under the same conditions
  `git blame` itself treats it as one (e.g., the commit has no parent reachable within
  the blame's history, such as a repository's root commit).
- **FR-004**: When `-b` is supplied, all non-identifier fields for a boundary line
  (reviewer/approver name, email, approval or commit date, line number, and line
  content) MUST continue to render exactly as they do without `-b`.

- **FR-004a**: `-b` MUST have no effect whatsoever on `-porcelain` output; passing `-b`
  together with `-porcelain` MUST produce output identical to `-porcelain` alone.
- **FR-005**: In human-readable output, a blanked identifier MUST be rendered as
  whitespace occupying the same column width as a normal identifier, preserving output
  alignment.
- **FR-006**: When `-b` is omitted, output MUST be unchanged from current behavior (no
  identifiers are blanked).
- **FR-007**: `-b` MUST compose with existing flags (`-L`, `-porcelain`, `-show-email`)
  without altering their independent behavior.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running the tool with `-b` on a file with known boundary commits produces
  human-readable output where 100% of boundary-commit lines have a blank identifier
  field, verifiable by comparing against `git blame -b`'s own boundary detection on the
  same file.
- **SC-002**: Running the tool with `-b` on a file with no boundary commits produces
  output byte-for-byte identical to running without `-b`.
- **SC-003**: Existing output for all flags and formats remains unchanged when `-b` is
  not supplied (zero regressions in current default behavior).

## Assumptions

- "Boundary commit" carries the same meaning `git` itself uses for `git blame -b`
  (primarily: a commit with no parent within the blamed history, such as a root commit).
  No custom or tool-specific definition of "boundary" is introduced.
- The reviewer/approver identity shown for a boundary line is unaffected by `-b`; only
  the commit identifier is blanked, matching upstream `git blame` behavior where the
  author name/date are still shown for boundary commits.
