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
