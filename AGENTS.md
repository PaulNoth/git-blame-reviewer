# AGENTS.md

## Project

Single-module Go CLI (`package main`, flat file layout at repo root — no `cmd/`
or `internal/` dirs). No external dependencies (no `go.sum`); stdlib only.

- `main.go` — CLI arg parsing / entrypoint
- `git.go` — git blame execution, repo root detection, GitHub/GitLab
  detection from remote origin URL
- `github.go` / `gitlab.go` — API clients for PR/MR approver lookups
- `client.go` — shared HTTP client / caching logic
- `formatter.go` — renders `git blame`-compatible output (human + porcelain)

## Commands

Always use the `Makefile` targets, not raw `go` commands:

- `make build` — builds `./git-blame-reviewer` binary
- `make test` — `go test -v ./...`
- `make test-coverage` — generates `coverage.out` + `coverage.html`
- `make lint` — `golangci-lint run` (install via `make install-tools` if missing)
- `make check` — runs `test` then `lint`; run this before considering work done

To run a single test: `go test -run TestName ./...` (or `-v` for verbose).

## Gotchas

- `integration_test.go` (`TestMainIntegration`) compiles a real binary
  (`test-git-review-blame`) via `go build` as part of the test and removes it
  afterward — expect this test to be slower and require a working Go toolchain.
- The compiled binary `git-blame-reviewer` is committed/present in the repo
  root; don't confuse it with source files, and don't rely on it being fresh
  after code changes — rebuild with `make build`.
- `.golangci.yml` sets `goimports.local-prefixes: git-review-blame`, but the
  actual Go module name (`go.mod`) and binary name are `git-blame-reviewer`.
  This is a pre-existing inconsistency, not a typo to "fix" blindly.
- The tool requires `GITHUB_TOKEN` or `GITLAB_TOKEN` env var at runtime
  (auto-detected from git remote origin); no config file is used.
- Linter enables many strict checks (`gomnd`, `funlen`, `dupl`, `gocyclo`
  min-complexity 15, `lll` line-length 140) — expect lint failures on
  magic numbers, long functions, or lines >140 chars in non-test files
  (test files are exempted from `gomnd`/`funlen`/`goconst`).
