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

## 4. Configs: `deltas.config` split into `budget.conf` and `testing.conf`

### Done in the harness

- `core/configs/deltas.config` is gone.
- `core/configs/budget.conf` holds the context budget, with the same keys as before.
- `core/configs/testing.conf` holds what a test is, grouped in sections:

```
[files]
match=*_test.go,test_*.py,*_test.py,*.test.*,*.spec.*
skip=.git,node_modules,vendor
```

Dropped for good: `test_files` as a flat key, `command_timeout`, `required_tools`, `base` and `scope.<slice>`. The `run` options (timeout, required tools) are not worth a section: a test suite almost always belongs to one library, so how it runs is the suite's own business.

### Decisions behind it

- Sections, not one key per action: the format is INI-like (`[files]`, then `key=value`), so a new topic is a new section and not a new prefix.
- No `scope`. A claim of type `test` is judged by running its command again, so it needs no list of what the slice depends on. A slice goes stale only for claims verified by reading (`manual`, `none`). Still to check in the code: how `verify.Check` treats a slice with only test claims when the code changes, and when the re-check runs (`delta check` on demand or at the Stop hook).
- The timeout stays as the constant in `scripts/delta/run.go:29`; no required tools.

### What each removed key did in the code

- `test_files`: which file names are tests. Used by `Vault.TestFiles` and `IsTest` (`vaults.go`), by `linkcheck.Cited`, by the lints `LINT-NO-CITE`, `LINT-UNKNOWN-ID`, `LINT-SKIP`, and by the guard hook (`guard.go:122`).
- `command_timeout`: `run.go:37`, `runner.go:31`.
- `required_tools`: `setup.go:72`, reported by `delta init` and the install check.
- `scope.<slice>`: `Scopes` in `conffile.go`, `status.New`, `verify.New` (`run.go:45,47`), `linkcheck.NewScope` and `Scope.Covers` (`scope.go`), and through them `ledger.Drift`, `ledger.Standings`, `gitcli.Tree`/`Files` and the guard (`guard.go:162-172`). This is the largest piece to remove or rework.
- `base`: gone with the preset (see section 2).

### Code that changes (pending the go-ahead)

- `source/domain/conffile/conffile.go`: one parser for `budget.conf` (flat keys) and one for `testing.conf` (sections). Drop `TestFiles` as `test_files`, `Timeout`, `Tools`, `Scopes` and `NoBase`; add `files.match` and `files.skip`. `conffile_test.go` follows.
- `source/domain/budget`: `ParseCeilings` and `IsKey` read `budget.conf` only.
- `source/gateway/bundle/bundle.go:64`: it copies `core/configs/deltas.config` to `deltas/deltas.config`; it has to copy the two new files to `deltas/budget.conf` and `deltas/testing.conf` (the destination is not decided yet).
- `source/gateway/vaults/vaults.go:26`: `skipDirs` becomes the default of `skip`.
- `scripts/delta/run.go:28-47`: reads the conf; no timeout or scopes to pass.
- `source/usecase/setup/setup.go:72`: `conf.Tools` goes.
- Scope removal: the files listed above under `scope.<slice>`, plus `linkcheck/scope_test.go`.
- Tests: `conf_test.go`, `update_test.go:157-159`, `assets_test.go:78-79`, `project_test.go:108-116`, `init_test.go`, `budget_test.go`.
- Projects already installed keep `deltas/deltas.config`; `update` has to leave it alone and create the new files if missing.

## 5. `core/configs/settings.json`

### Done in the harness

- The `timeout: 20` of every hook is gone; Claude Code applies its own default.
- Events are in the order they happen in a session (`SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `Stop`) and each group fits on one line, so the five hooks read as a table. The shape is the one Claude Code defines (event, matcher groups, handlers), so it cannot be flattened further.
- All five events call the same command, `delta hook`; `source/adapter/hook/hook.go:14` tells them apart by the `hook_event_name` in the payload.

### Code that changes (pending the go-ahead)

- `source/gateway/bundle/settings_test.go:34`: asserts `h.Timeout != 20`. It has to stop checking the timeout and its message.
- `source/usecase/setup/settings.go:123` (`covers`) compares the matcher and the commands only, never the timeout, so a project that already has the hooks with `timeout: 20` is left as it is by `init` and `update`, and a project without them gets them without a timeout. Nothing to change there, but the old timeout stays in installed projects.
- If the two `PreToolUse` groups (`Write|Edit` and `Bash`) were ever merged into one `Write|Edit|Bash`, `covers` would not recognize the old two as covering it and installed projects would run `delta hook` twice. They are kept apart on purpose.

### Open: the permissions

`allow` ships things that have nothing to do with delta (`WebSearch`, `Skill(update-config)`, `Bash(curl *)`, `Bash(powershell *)`), and `Bash(powershell *)` and `Bash(curl *)` allow almost any command. What the skills run, `delta status`, `delta check`, `delta ready`, `delta link`, is not in the list. To decide: keep only what delta needs (`Bash(delta *)` except `delta approve`, which the person must confirm) and leave the rest to each project.

### Permissions: done

`allow` is now only what delta needs (`delta status`, `stale`, `ready`, `check`, `lint`, `spec`, `link`, `built`, `budget`, and read-only git). `delta approve`, `init` and `update` are in `ask`, so the person confirms them. `WebSearch`, `curl` and `Skill(update-config)` are gone, and `deny` adds `*.pem` and `*.key`. `git push` and `git pull` stay denied.

Installed projects keep what they have: the merge only adds. The `ask` rules and the precedence (deny, ask, allow) come from Claude Code's documentation and are not tried here.
