# Harness refactor notes

Harness first, code later. Phase 1 (the code follows the harness) is done: see "Phase 1: what was done" at the end. Phase 2 (`package/domain` and the domain renames) is described there too.

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

**Harness: done** (folders renamed with `git mv`, text updated): the five `SKILL.md`, `agents/core/samples/CLAUDE.md` (usage paragraph and table, plus a `/scriber` row), and the mentions in `claims-fidelity`, `kb-deltas-routing`, `context-budgets` and the `paths:` of `when-write-delta` (now `.claude/skills/drafter/**`).

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
- `core/configs/testing.conf` holds what a test is, as flat keys:

```
match=*_test.go,test_*.py,*_test.py,*.test.*,*.spec.*
skip=.git,node_modules,vendor
```

Dropped for good: `test_files` as a flat key, `command_timeout`, `required_tools`, `base` and `scope.<slice>`. The `run` options (timeout, required tools) are not worth a section: a test suite almost always belongs to one library, so how it runs is the suite's own business.

### Decisions behind it

- Plain `key=value` lines: `match` and `skip`. If the file grows, sections can be added later.
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

## 6. Templates and the git hook

### Done in the harness

- `core/templates/spec.md` is now `core/samples/specify.md`: the folder `templates/` is renamed `samples/` and the file `spec.md` is renamed `specify.md`.
- `core/githooks/commit-msg` is removed, and with it the `githooks/` folder. The comment of `agents.go:6` no longer mentions it.
- New rule `rules/commit-messages.md` (15 characters, 9 lines, always loaded): the same convention the validator applied, as plain guidance. Type list, `type(scope): subject`, 40 to 50 characters, imperative, lowercase, no period, blank line before the body, the body says why.

What goes away is the check: nothing refuses a commit message any more. The rule asks, and the agent follows it.

### Code that changes (pending the go-ahead)

**The template rename:**

- `source/gateway/bundle/bundle.go:17`: `ReadFile("core/templates/spec.md")` becomes `core/samples/specify.md`. `bundle.go:42` and `:66` name `core/templates/CLAUDE.md` and become `core/samples/CLAUDE.md`. The method `Spec()` and the `Template` interface of `author.go:33,44` can keep their names.
- `source/gateway/bundle/assets_test.go` and any test that reads the template by name.

**The git hook, installed by `delta init`:**

- `source/gateway/bundle/bundle.go:19-21`: the `GitHook(name)` method.
- `source/usecase/setup/setup.go:16-18`: the constants `hooksDir`, `hookName`, `hookPath`.
- `setup.go:29-30` and `source/gateway/gitcli/gitcli.go:17-29`: `HooksPath` and `SetHooksPath`, only used for this.
- `setup.go:77-92`: the install check that reports `core.hooksPath is not set` and `.githooks/commit-msg is missing`.
- `setup.go:117-145` (`Run`) and `:216-232` (`pointHooksPath`): they write the hook and set `core.hooksPath`.
- `source/usecase/setup/update.go:25`: the hook as a managed asset.
- `source/adapter/cmd/commit.go`: the command `delta commit-msg <file>`.
- `source/domain/linkcheck/scope.go:7,30`: `.githooks/` in the list of paths that are not code (it goes away with the scope itself).
- Tests: `scripts/delta/hooks_test.go`, `run_test.go:292`, `setup_test.go`, `update_test.go`, `install_test.go`, `gitcli_test.go`.
- Projects already installed keep `.githooks/commit-msg` and `core.hooksPath`. `update` has to retire the hook it installed, but a `core.hooksPath` that the project set itself should not be touched.

**Not part of the git hook, but it enforces the same thing:**

- `source/usecase/guard/guard.go:222-224,230-239` (`before` and `commitDecision`): the `PreToolUse` hook on `Bash` validates the message of every `git commit` the agent runs and blocks it. It uses `source/domain/commits` (`commits.go`, `extract.go`).
- Decision to make: keep it (the agent is still checked) or remove it together with the git hook, which leaves the rule as the only guide and lets the `commits` package, the `Bash` matcher in `settings.json` and their tests go.

`agents/agents.go` has no comments any more, the package one included. Its `//go:embed core docs skills rules` line is unchanged.

## Phase 1: what was done

The code now follows sections 1 to 6, and `go vet`, `gofmt` and `go test ./...` pass.

- **Skills and rules:** the old names are replaced by `comper`, `drafter`, `builder`, `auditor`, `scriber` in the hook messages (`guard.go`), `delta status` and the tests. The prompt context line now says the person may speak plainly or call a skill by name.
- **Base preset:** `Asset.Base`, `NoBase`, the `base` key and the `rules/base/` marker are gone.
- **Configs:** `conffile` reads `deltas/budget.conf` (the budget keys) and `deltas/testing.conf` (`match`, `skip`), and reports problems with the file name and line. `Timeout`, `Tools`, `Scopes` and the legacy `.claude/budget.conf` fallback are gone; the command timeout is the constant in `run.go`. `Vault.Skipping` makes `skip` configurable, with `.git`, `node_modules` and `vendor` as the default (the folder `deltas` is never searched).
- **Git hook and commit validation:** `core/githooks`, `GitHook`, `HooksPath`/`SetHooksPath`, `pointHooksPath`, the install check for `core.hooksPath`, `delta commit-msg`, `source/domain/commits` and the `Bash` commit check of `guard.go` are removed, and with them the `Bash` group of `PreToolUse` in `settings.json`. `setup.New` no longer takes a git.
- **Samples:** `core/samples/CLAUDE.md` and `core/samples/specify.md` are read by `bundle.go`.
- **Scope:** `Scope` has no paths any more: every code file counts, except Markdown and what delta installs (`.claude/`, `deltas/`). `status.New`, `verify.New` and `run.go` no longer take scopes.
- **README:** names, rules, install list and layout follow the new design.

### Not done: stale only for claims verified by reading

Section 4 decided that a slice with only `test` claims needs no "stale" because the check runs its commands again. That is a change of behavior in `lifecycle`, `status.Stale`, `ledger` and `guard.stop`, and it needs a design pass of its own (when the re-check runs, what `verified` means for a slice with reading claims). Today a verified slice still goes stale when any code file changes, and the Stop hook still asks for `delta check` when code changed.

### Phase 2: done

`source/domain` is now `package/domain`, with folders and files of 6 to 8 characters and no folder with an icon in Material Icon Theme (checked against the extension's own list, version 5.40.0): `assertion` is `claimset`, `linkcheck` is `linking`, `lifecycle` is `states`, `knowledge` is `kbpaths`, `language` is `english`, `docs` is `guides` and `drift` is `diffing`. `budget`, `conffile`, `findings`, `judgment`, `manifest` and `specfile` keep their names (`conffile` was planned as `testconf`, but it reads `budget.conf` as well as `testing.conf`). `commits` went in phase 1. `docs.go`, `drift.go` and `language.go` follow their folder (`guides.go`, `diffing.go`, `english.go`).

The other layers (`source/usecase`, `source/gateway`, `source/adapter`) stay where they are.

`package` does have an icon in that theme (`folder-packages`), and so do `source` (`folder-src`) and `scripts`. Left as they are.

## Phase 3: the domain reviewed

`package/domain` has no comments (package docs included), and the code is simpler where it was repeating itself:

- `budget`: the budget keys live in one place (`Ceilings.fields`), `Measure` sorts with `slices.SortFunc` and adds the total as it goes, and `weigh` builds its items through one helper.
- `english`: lowercases the text once.
- `kbpaths`: the folder test is one `slices.ContainsFunc`.
- `judgment`: the two manual cases are one.
- `specfile`: `SetStatus` and `SetTag` share `replaceLine`, and `frontmatterEnd` is the one place that finds the closing `---`.
- `conffile`: the readers are `readBudget`, `readTesting` and `readMatch`, and a problem takes a format.
- `linking`: the test-file matcher moved to `matcher.go`; `citing.go` keeps the citation search (`inScope` is now `underAny`).

Behavior is the same, and the existing tests pass unchanged. Without comments, the reasons behind some choices (why a Spanish word list, why hashes ignore CRLF, why `Cited` only counts files under the command's paths) live in the tests and in the git history of this branch, not in the code.

## Phase 4: the gateway reviewed

`source/gateway` has no comments and its files are 6 to 8 characters (the folders `bundle`, `gitcli`, `ledger`, `runner` and `vaults` already were, with no icon).

- `bundle`: the executable-script rule (`mode`, `.sh`) is gone, since no asset is a script; every asset is `0644`.
- `gitcli`: `tree.go` is `hashing.go`, and `Tree` reads HEAD through its own `head` helper.
- `ledger`: `read` returns a map and never an error (a missing or damaged record is an empty one), so `Record`, `Drift` and `Standings` lose their ignored errors.
- `runner`: `split.go` is `command.go`, and the whitespace case and the end of the command share one `flush`.
- `vaults`: the 260-line file is four: `vaults.go` (the type and its construction), `catalog.go` (the specs: `Path`, `List`, `Load`, `Create`, `Save`), `project.go` (project files: `ReadFile`, `WriteFile`, `RemoveFile`, `Glob`, `writeAtomic`) and `suites.go` (the test files).

Still open from section 4: `Setup.Asset.Mode` and the `perm` of `Files.WriteFile` are always `0644` now, so they can go when the use-case layer is reviewed. `ledger.Drift`, `Standings` and `gitcli.Tree/Files` stay until the decision on "stale only for claims verified by reading" is made.

## Phase 5: the gateway moved

`source/gateway` is now `package/gateway`, with every file name at 6 characters: `gitcli/hashing.go` is `hashes.go`, `runner/command.go` is `tokens.go`, `vaults/catalog.go` is `slices.go`, `vaults/project.go` is `inside.go`, and the tests follow (`inside_test.go`, `claude_test.go`). Only imports and paths changed.

The file and line references in sections 2 to 6 above are the historical record of what was pending when they were written, and use the old `source/gateway` and `source/domain` paths.

Open: `package/gateway/bundle` imports `source/usecase/setup` and `package/gateway/ledger` imports `source/usecase/verify`, so the gateway depends on a layer that still lives in `source/`. It compiles, but it is the reason to move `usecase` into `package/` next, or to give the gateway its own types.

## Phase 6: the gateway folders renamed

The five gateway folders are 8 characters long and have no icon in Material Icon Theme: `bundle` is `embedded`, `gitcli` is `gitshell`, `ledger` is `recorder`, `runner` is `executor` and `vaults` is `diskbase`. The package names follow, so `gitcli.New` is `gitshell.New`, and the main file and its test take the package name (`embedded.go`, `gitshell.go`, `recorder.go`, `executor.go`, `diskbase.go`). The other files keep their 6 characters (`hashes`, `tokens`, `locate`, `inside`, `slices`, `suites`, `assets_test`, `claude_test`, `inside_test`).

The older sections of this file use the names the packages had when they were written.

## Phase 7: the use cases moved and reviewed

`source/usecase` is now `package/usecase`, with folders of 8 characters and no icon in Material Icon Theme: `author` is `drafting`, `guard` is `sentinel`, `measure` is `weighing`, `setup` is `deployer`, `status` is `tracking` and `verify` is `verifier`. The package names follow, so `guard.New` is `sentinel.New`. The gateway no longer depends on a layer outside `package/`.

- No comments, and the big files are split into files of 6 to 8 characters: `tracking` (`gating`, `drifts`, `guides`), `verifier` (`checks`), `sentinel` (`events`, `session`, `finish`, `edited`) and `deployer` (`hooks`, `permits`, `records`). Imports are in two blocks, the standard library and the rest.
- The file mode is gone: `Asset.Mode`, the `perm` of `Files.WriteFile` and the gateway's `WriteFile` took only `0644` since no asset is a script, so every file is written `0644`.
- `deployer` hashes each asset once in its plan, `drafting.hasError` is one `slices.ContainsFunc`, and the next doc to write in `tracking` is the first open one of the usual order followed by the others sorted.

Still open: `source/adapter` (the command line and the hook) keeps its comments and its names, and "stale only for claims verified by reading" is not decided.

## Stale: decided and done

The question left open in Phase 1 ("stale only for claims verified by reading") was wrong: `verified` never depended on `manual` or `none` claims, and what `stale` tells the person is that the contract may be out of date, not that a test may fail (the Stop hook already asks for a check when code changes). Dropping it for slices with tests would have lost that signal.

Decided instead: a slice depends on the code its test commands name. `linking.NewScope(spec, claims)` takes the paths of every `test:` command (`go test ./billing` is `billing/`, `./billing/...` is the package and below, `./...` or `.` is the whole project), and the slice depends on every code file as soon as one command names no path or it has no test claim. The spec of the slice always counts; Markdown and what delta installs never do. There is no configuration for it.

It reaches `Scope.Covers`, so it narrows at once the part of the tree the ledger fingerprints, what `delta stale` and the `PostToolUse` hook report as drift, and when the `Stop` hook asks for a check. Known blind spot: a slice whose commands name `./billing` does not go stale when `shared/money`, which `billing` imports, changes.

Tests: `TestScopeFollowsThePathsOfTheTestCommands`, `TestAChangeOutsideWhatTheCommandsNameDoesNotMakeASliceStale`, and the two `Stop` tests of `scripts/delta/stop_test.go`.

## Phase 8: the adapters moved

`source/adapter` is now `package/adapter`, and `source/` is gone: `cmd` is `terminal` (the command line) and `hook` is `listener` (the Claude Code hook events), both 8 characters with no icon in Material Icon Theme. `terminal/program.go` is `terminal.go`, `stale.go` is `drifts.go` and `hook.go` is `listener.go`. No comments; `foundIn` is one `slices.ContainsFunc`, and `init` and `update` print their steps through `printSteps`.

The Go code is now `scripts/delta` (the wiring) and `package/{domain,usecase,gateway,adapter}`.
