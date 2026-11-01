# deltas-ai

Delta: a deterministic Go CLI (`delta`) plus Claude Code hooks that keep a verifiable contract between what a project says (specs with claim IDs) and what its code does. The model proposes, the code decides.

---

## Where to look

| If you need... | Read |
|---|---|
| What Delta is and why | `docs/direction.md` |
| How the CLI is built: layers, packages, hooks, `delta.conf` | `docs/blueprint.md` |
| What is done and what is next | `docs/waypoints.md` |
| The contract of a capability before changing its behavior | `deltas/`: list the folder and open the spec that matches |
| Which contracts are stale or unverified | `delta status` |
| A convention for the files you are editing | Nothing to do: `.claude/rules/` loads by itself when a matching file is read |

## Commands

- `go test ./...`, `go vet ./...`, `golangci-lint run`: the three checks CI runs.
- `go run ./scripts/delta <command>`: run the CLI from source.
- `.claude/` is written by `delta init` from `source/harness/` and is not versioned: change the harness, then run `go run ./scripts/delta init` again (it keeps files that differ, so delete the one you changed first).

## Non-goals

- Delta calls no model; skills (Quill, Forge, Seal) do the interviewing and writing.
- No third-party library besides cobra, and no shell in claim commands.
