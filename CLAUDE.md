# deltas-ai

Delta: a deterministic Go CLI (`delta`) plus Claude Code hooks that keep a verifiable contract between what a project says (specs with claim IDs) and what its code does. The model proposes, the code decides.

---

## Where to look

| If you need... | Read |
|---|---|
| What Delta is and why | `docs/direction.md` |
| How the CLI is built: layers, packages, hooks, `deltas/deltas.config` | `docs/blueprint.md` |
| What is done and what is next | `docs/waypoints.md` |

## Commands

- `go test ./...`, `go vet ./...`, `golangci-lint run`: the three checks CI runs.
- `go run ./scripts/delta <command>`: run the CLI from source.

## Non-goals

- This repo does not use Delta on itself yet: no `deltas/`, no `.claude/`, no git hook. The first real use is a separate project (phase 6).
- Delta calls no model; skills (Quill, Forge, Seal) do the interviewing and writing.
- No third-party library besides cobra, and no shell in claim commands.
