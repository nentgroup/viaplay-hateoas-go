# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project overview

`viaplay-hateoas-go` (module `github.com/nentgroup/viaplay-hateoas-go/v2`) is a Go
library implementing the [Viaplay HATEOAS standard](http://github.com/nentgroup/api-guidelines)
for building hypermedia-driven RESTful APIs (HAL-style JSON and XML). It is a
single flat-package library at the repo root — there is no `cmd/` or `internal/`
layout.

Key source files (repo root):
- `hal.go` — core `Resource`, `Embedded`, `Links`, `Curies`, `Mapper` interface, and
  the three flavors (`FlavorViaplay`, `FlavorHAL`, `FlavorDefault`).
- `xml.go` — flavor-agnostic HAL+XML marshaling (`xml.Marshaler` implementation).
- `generics.go` — generic helpers (`Payload[T]`, `WithMapper`, `NewResourceCollection`).
- `*_test.go` — table-driven tests and benchmarks colocated with the code they cover.
- `examples/` — runnable example programs (`basic`, `embedded`, `xml`, `generics`, `api-response`).
- `specs/viaplay-hateoas.md` — the canonical spec this library implements.
- `MIGRATION.md` — v1 → v2 breaking-change guide; update it if you change public APIs in a breaking way.

## Knowledge graph (graphify-rs)

A prebuilt knowledge graph of this repo lives in **`graphify-rs-out/`**. Consult it
before large refactors or when you need a map of symbol relationships:
- `graphify-rs-out/GRAPH_REPORT.md` — human-readable summary: god nodes, communities, surprising connections.
- `graphify-rs-out/graph.json` / `graph.graphml` / `graph.cypher` — machine-readable graph data.
- `graphify-rs-out/html/`, `graphify-rs-out/wiki/`, `graphify-rs-out/obsidian/` — browsable/exportable views.

Rebuild after significant code changes:
```bash
task graphify:build     # graphify-rs build --no-llm --output graphify-rs-out
task graphify:serve     # graphify-rs serve --graph graphify-rs-out/graph.json (MCP server)
```

## Setup

```bash
task setup          # installs dev tools (lefthook, commitlint, golangci-lint, Task) + deps
task deps:install    # go mod download && go mod tidy
```

## Build, test, lint

```bash
go build ./...
task test            # go test -v ./... -covermode=count -coverpkg=github.com/nentgroup/viaplay-hateoas-go/v2
task lint            # golangci-lint run --allow-parallel-runners --fix && actionlint
```

Run a single test: `go test -run TestName -v ./...`

Go version: **1.27** (see `go.mod`). golangci-lint config: `.golangci.yml`
(cyclop max-complexity 15, dupl threshold 250, `ireturn`/`errname`/`gosec` enabled, etc.).

## Conventions

- **Commits**: Conventional Commits, enforced by commitlint (`.commitlint.yaml`) via
  lefthook's `commit-msg` hook (e.g. `feat(scope): add feature`, `fix(scope): fix bug`).
- **Git hooks**: managed by `lefthook.yml` — pre-commit runs golangci-lint + actionlint,
  pre-push runs the full test suite, commit-msg runs commitlint.
- **Flavors**: when creating/modifying HAL resource behavior, remember payload flavor
  (`FlavorViaplay` default, `FlavorHAL`, `FlavorDefault`) affects JSON shape but XML output
  is always flavor-agnostic (flattened).
- **Public API changes**: this is a published library (`go get .../v2`) — treat
  exported identifiers as a public contract. Document breaking changes in `MIGRATION.md`
  and keep `README.md` examples in sync with actual behavior.
- Keep new code covered by tests colocated as `*_test.go` next to the file under test,
  following existing table-driven test style.

## Before finishing a task

1. `go build ./...` and `task test` must pass.
2. `task lint` should be clean (or pre-existing issues left untouched).
3. If you changed public API surface, update `README.md` and/or `MIGRATION.md`.
4. Consider rebuilding the graphify-rs graph (`task graphify:build`) after structural changes.
