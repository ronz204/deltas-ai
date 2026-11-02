# Harness refactor notes

Harness first, code later. Nothing in `source/` or `scripts/` is touched until the go-ahead.

## 1. Rules: done

The rules were renamed to a similar length as `kb-deltas-routing`, `when-write-delta` and `context-budgets`: two or three words, kebab-case, 15 to 17 characters.

| Before | Now |
|---|---|
| `fidelity` | `claims-fidelity` |
| `honesty` | `done-needs-proof` |
| `language` | `project-language` |
| `base/base-scope` | `code-change-scope` |
| `base/base-reliability` | `code-reliability` |

`rules/` is flat: 8 files. `base/` is gone, and so is the idea of a rules preset: every rule is installed, with no `base=on|off` switch and no marker in the files. The two former base rules no longer say "Part of the `base` preset". `code-change-scope.md` now points to `claims-fidelity`.

## 2. Where the code is affected (pending the go-ahead)

**Rule renames** (the installed path is `.claude/rules/<name>.md`):

- `source/gateway/bundle/assets_test.go:35` lists the installed rules by name (`fidelity`, `language`, `honesty`, `base-scope`, `base-reliability`).
- `scripts/delta/init_test.go:24`, `:78`, `:82`, `:85` use `.claude/rules/fidelity.md` as the example rule.
- `scripts/delta/update_test.go:25`, `:52`, `:55` use `.claude/rules/fidelity.md` as "a project's own rule".
- Projects that already ran `delta init` keep the old names. Check how update retires a renamed managed rule in `source/usecase/setup` and `source/domain/manifest`: it should remove the old file and not leave it behind. `update_test.go:70` (`old.md`) covers that path.

**Removing the base preset** (code that exists only for it):

- `source/gateway/bundle/bundle.go:47`: drop `Base: strings.HasPrefix(p, "rules/base/")`.
- `source/gateway/bundle/bundle.go:27` and `:62-63`: the comment "rules from any subfolder of rules/" and the `path.Base(p)` flattening are no longer needed, since rules have no subfolders. Simplify both.
- `source/usecase/setup/setup.go:40`: drop the `Base` field of `Asset`.
- `source/usecase/setup/setup.go:168-180`: `assets()` filters on `Base` and `conf.NoBase`. It can return `i.harness.Assets(...)` directly.
- `source/domain/conffile/conffile.go:22` and `:47`: drop `NoBase` and the `base` key. `conffile_test.go:83-87` tests it and goes too.
- `scripts/delta/init_test.go:110-117`: the test "base=off keeps the base preset out, and update retires it" goes, or shrinks to a check that all 8 rules are installed.
- `agents/core/configs/deltas.config:7-8`: remove the `base=on` line and its comment.
- `README.md:112` ("Language and team presets, opt-in") and `:157` (credit for the "lean-change ladder in the `base` preset"): reword so they do not name a preset.
- `source/gateway/bundle/assets_test.go:46` filters embedded paths with `HasPrefix(path, "rules/")`. Review it.
- `agents/agents.go:6` comment mentions `rules/`. It should still be accurate.

**Not affected:**

- `source/domain/budget/measure.go:19` and `:75`, `source/domain/knowledge/kbpaths.go:6`: they match `.claude/rules/*.md` by pattern, not by name.
- Tests that use `.claude/rules/a.md`, `one.md` or `always.md` as fixtures.
- README.md line 100: it only describes `.claude/rules/<topic>.md`.

## 3. Skills: names and how the developer uses them

### What each skill is for

The flow: a request becomes a contract (`drafter`), a person approves it, the contract becomes code and tests (`builder`), the code is checked against the contract (`auditor`). `comper` decides which one the state asks for. `scriber` writes the docs for people and sits outside that loop.

### Names: done

| Before | Now | What the developer reads in it |
|---|---|---|
| `helm` | `comper` | knows where the project is and which skill comes next (6 characters) |
| `quill` | `drafter` | drafts the contract (the `draft` state) |
| `forge` | `builder` | builds what the approved contract asks |
| `seal` | `auditor` | audits the code against the contract |
| `atlas` | `scriber` | writes the docs for people (6 characters) |

Slash forms: `/comper [slice]`, `/drafter <slice>`, `/builder <slice>`, `/auditor <slice>`, `/scriber <doc>`.

### Usage model: decided

`comper` works both ways. A plain request is routed and carried through, as before. `/comper` with no request only reports the state of the slice and names the skill that comes next; it runs it if the person says so. The developer may speak plainly or call the skills by name, and never types `delta` commands. `CLAUDE.md` shows the flow as a table of state to skill.

### Where the code and the harness are affected

**Harness: done** (folders renamed with `git mv`, text updated): the five `SKILL.md`, `agents/core/templates/CLAUDE.md` (usage paragraph and table, plus a `/scriber` row), and the mentions in `claims-fidelity`, `kb-deltas-routing`, `context-budgets` and the `paths:` of `when-write-delta` (now `.claude/skills/drafter/**`).

**Code: pending the go-ahead.** Where it still says the old names:

- `source/usecase/guard/guard.go:105`, `:195`, `:198`: messages that say `/forge <slice>` and `/quill`.
- `source/usecase/guard/guard.go:303`: the context line "use the helm skill before touching any file; the person speaks plainly and never types delta commands". Reword it for the two entrances of `comper`.
- `source/usecase/guard/guard.go:316-324`: "the quill skill", "the forge skill", "the seal skill".
- `source/adapter/cmd/status.go:64`: `next: /atlas <doc>`.
- Tests with the names: `guard_test.go` (`:47`, `:143`, `:153-154`, `:265-266`, `:397`, `:460`), `stop_test.go` (`:168-169`, `:176`), `init_test.go` (`:25`, `:131`, `:140`), `update_test.go` (`:17`, `:26`, `:35`, `:38`, `:46`), `install_test.go:84`, `assets_test.go:34`, `budget_test.go:14,25`, `measure_test.go` (`:54-55`, `:68`, `:79`, `:82`), `kbpaths_test.go:8`, `project_test.go` (`:77-78`, `:90`).
- Projects already installed: update has to remove the old skill folders (same retire logic as the rule renames).
