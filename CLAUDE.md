# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

See [AGENTS.md](./AGENTS.md) for the full project guide (structure, build/test/lint
commands, conventions, and the prebuilt `graphify-rs-out/` knowledge graph). Everything
in AGENTS.md applies here; this file only adds Claude-specific notes.

## Quick reference

- Module: `github.com/nentgroup/viaplay-hateoas-go/v2` — flat-package Go library, no `cmd/`/`internal/`.
- Core files: `hal.go` (resources/links/embedded/flavors), `xml.go` (HAL+XML), `generics.go` (typed helpers).
- Build/test/lint: `go build ./...`, `task test`, `task lint`.
- Knowledge graph: browse `graphify-rs-out/GRAPH_REPORT.md` first when exploring unfamiliar
  parts of the codebase; rebuild with `task graphify:build` after significant changes.
- Commits must follow Conventional Commits (enforced by lefthook + commitlint).
- This is a published library — treat exported identifiers as a public API contract;
  update `MIGRATION.md`/`README.md` for any breaking or user-facing change.
