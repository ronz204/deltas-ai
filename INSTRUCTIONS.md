# Working with Delta

A guide for using Delta in your own projects. For what Delta is and why, read the [README](README.md).

The short version: you describe a feature, the agent drafts a **contract** for it, you approve the contract, the agent builds the feature against it, and Delta keeps checking that the code still keeps its promises.

## Contents

- [What you need](#what-you-need)
- [Install the CLI](#install-the-cli)
- [Set up a project](#set-up-a-project)
- [Your first feature](#your-first-feature)
- [The five skills](#the-five-skills)
- [Slice states](#slice-states)
- [Reading and writing a contract](#reading-and-writing-a-contract)
- [Linking a claim to a test](#linking-a-claim-to-a-test)
- [What you will see while you work](#what-you-will-see-while-you-work)
- [Commands you may run yourself](#commands-you-may-run-yourself)
- [Configuration](#configuration)
- [How drift is decided](#how-drift-is-decided)
- [Docs for people](#docs-for-people)
- [Rules](#rules)
- [Keeping Delta up to date](#keeping-delta-up-to-date)
- [Troubleshooting](#troubleshooting)
- [Limits worth knowing](#limits-worth-knowing)

---

## What you need

- A git repository for your project.
- [Claude Code](https://claude.com/claude-code). Delta's skills and hooks are for it.
- Go, only to build the `delta` binary.

## Install the CLI

From a clone of this repository:

```
go build -o dist/delta ./deltas
```

Put `dist/delta` on your `PATH` and check it:

```
delta version
```

It prints `delta (deltas-ai) dev`. The binary has to be called `delta`: the Claude Code hooks and the permissions call it by that name. If another program called `delta` is first on your `PATH` (the `git-delta` pager, for example), `delta status` warns you.

## Set up a project

Inside your project:

```
delta init
```

It installs what Delta needs and never replaces a file you already have:

| Where | What |
|---|---|
| `CLAUDE.md` | A short router for the agent. Replace the comment at the top with two or three lines on what your project is |
| `.claude/skills/` | The five skills |
| `.claude/rules/` | The conventions, which load by themselves when they apply |
| `.claude/settings.json` | Delta's hooks and permissions, merged with whatever you already have |
| `deltas/specs/` | Where the contracts will live |
| `deltas/docs/` | Four templates for docs for people |
| `deltas/budget.conf`, `deltas/testing.conf` | The two config files, with their defaults |
| `deltas/deltas.manifest` | A record of what Delta installed, so it can update it later |

Commit all of it. Anyone who clones the project then has the same skills, rules and hooks, and does not need to run `delta init` again (they do need `delta` on their `PATH`).

Run `delta status` to see the state of the project. A new project says `no specs yet`.

## Your first feature

You can talk plainly or call the skills by name. This is the flow, with both:

1. **Ask for the feature.** *"Add refunds to billing"*, or `/drafter billing`.
2. **Answer the interview.** Drafter reads the real code, then asks what the code cannot tell it: why the feature exists, what it must never do, what it leaves out, and how each claim can be verified. It asks in short batches; take the time to answer well, because the invariants are what will be checked later.
3. **Review the draft.** Drafter saves the contract as a `draft` and summarizes its claims, saying which have no verification yet. Read `deltas/specs/billing.spec.md` yourself. If something is wrong, say so and it rewrites it.
4. **Approve it.** Say yes when asked. Delta runs `delta approve billing`, and Claude Code asks you to confirm that command: approving is always your decision, never the agent's.
5. **Build it.** `/builder billing`, or just keep going. Builder works out what is not verified yet, shows you a plan, and waits. After your OK it writes each test first, then the code, and links every test to its claim.
6. **See it verified.** When every test claim holds, the slice becomes `verified`. `delta status` shows it.
7. **Audit it later.** `/auditor billing` checks that the contract still holds, and reads the code for the claims that have no test.

At any point, `/comper` tells you where each slice stands and which skill comes next.

## The five skills

| Skill | Call it when | What you give it | What it leaves you |
|---|---|---|---|
| `/comper [slice]` | You do not know the next step | Nothing, or a slice name | Where the slice stands and the skill to call |
| `/drafter <slice>` | A feature is new, changes, or its contract looks thin | Your answers about intent, limits and invariants | A `draft` contract, for you to review and approve |
| `/builder <slice>` | A contract is approved, or went stale | An OK on its plan | Tests and code, each test linked to its claim |
| `/auditor <slice>` | You want to know if it still holds | Nothing | A table of verdicts per claim, saying whether each came from a command or from reading. It fixes nothing |
| `/scriber <doc>` | The project needs explaining to people | Your reasons, which the code does not show | One doc written under `deltas/docs/` |

Each skill works on one slice (or one doc) per run. Drafter and Builder never approve; Auditor never edits. When the contract itself is wrong, any of them hands it back to Drafter instead of bending the code.

Plain requests work too: *"is billing still right?"* goes to Auditor, *"change how refunds round"* goes to Drafter first, because a change of behavior starts with the contract. A request that is not about behavior (a typo, a rename, a question about the code) is not routed at all.

## Slice states

| State | Meaning | Who moves it |
|---|---|---|
| `draft` | The contract is written and waits for review | Drafter writes it; **you** approve it |
| `approved` | The contract is accepted | Builder, with `delta built`, once the code is done |
| `built` | The code and tests exist | `delta check`, when every test claim holds |
| `verified` | Every test claim held in the last check | A change to the code its tests cover makes it `stale` |
| `stale` | The code changed after verification: the contract may be out of date | Builder rebuilds it, or Drafter revises the contract |

Two rules to remember:

- Rewriting an `approved` or `built` contract sends it back to `draft`, and it needs a new approval. Delta says so when it happens.
- A claim with no test never blocks `verified`. It stays on your screen as debt.

## Reading and writing a contract

A contract is `deltas/specs/<slice>.spec.md`: a short frontmatter, then free sections, then claims.

```markdown
---
format: 1
slice: billing
status: approved
---

# billing: spec

## Intent
Charge customers and refund them, with amounts that always add up.

## Scope
Owns charges and refunds. Does not own invoices.

## Invariants
- INV-01 [test: go test ./billing]
  When a refund is issued, the system shall never refund more than was charged.
- INV-02 [none]
  The system shall not charge the same order twice.

## Deferred / Open questions
- INV-02 has no test yet. Decide the idempotency key before writing it.

## Acceptance criteria
- ACC-01 [manual: refund an order from the admin screen]
  Given a paid order, when it is refunded, then the customer sees the refund within a minute.
```

**Frontmatter.** `format` is `1`, `slice` is the name of the file in lowercase words joined by hyphens, and `status` is the state. Do not set `status` by hand: the commands move it.

**Claims.** Each claim is a header, `- INV-01 [tag]` for an invariant or `- ACC-01 [tag]` for an acceptance criterion, and under it, indented, one sentence.

- An invariant is one sentence that something observable can confirm. It may open with `When`, `While`, `Where` or `If … then`: *"When a refund is issued, the system shall…"*.
- An acceptance criterion reads `Given <context>, when <action>, then <outcome>`.
- IDs are stable. Never renumber a claim, and never reuse the ID of one you removed.
- Write the spec in English, whatever language you talk in. Delta warns when a spec is not in English.

**Tags** say how a claim is verified:

| Tag | Meaning |
|---|---|
| `test: <command>` | A command that exits 0 when the claim holds |
| `manual: <step>` | Only a person can check it. The step is optional |
| `none` | Nothing verifies it yet. Valid, and listed as debt |

**Do not edit specs through the agent's file tools.** A hook blocks it on purpose: specs are written with `delta spec write`, which validates before it saves, and tests are linked with `delta link`. You, as a person, can edit the files by hand; run `delta spec lint` afterwards to check them.

**What is a slice?** A bounded feature or capability, sized the way a bounded context would be. Not a source file, and never an architecture layer (no "billing domain" spec plus "billing infrastructure" spec). A pattern you apply across many features, like how repositories are written, is a rule, not a slice. A use case that only orchestrates one feature belongs in that feature's contract.

## Linking a claim to a test

A claim with `test: <command>` is checked by running the command, and a test has to say which claim it proves.

1. **The command** exits 0 when the claim holds, and names the file, package or folder it runs. Prefer `go test ./billing` over `go test -run TestRefund`: a filter that matches nothing passes silently.
2. **The test cites the ID**, in its name or in a comment: `INV-01`, `INV_01` and `INV01` all count, in any case.

```go
// INV-01: a refund never exceeds the charge
func TestRefundNeverExceedsTheCharge(t *testing.T) { ... }
```

Builder does both for you, with `delta link billing INV-01 "go test ./billing"`. If you link by hand, quote the command so it is one argument.

`delta check` then gives each claim a verdict:

| Verdict | Means |
|---|---|
| `Holds` | The command exited 0 and a test under its paths cites the ID. It does not prove the test is a good one |
| `Violated` | The command failed. The code or the test is wrong |
| `Unverified` | There is no evidence either way: the command did not run, no test cites the ID, or the claim is `manual` or `none`. It never passes for `Holds` |

A citation only counts under the paths the command names. A command that names no path (`make test`) accepts a citation anywhere.

## What you will see while you work

Delta speaks through the hooks. These are the messages you will meet, and what they mean.

| You see | Why | What to do |
|---|---|---|
| An edit is refused: *"X is part of the knowledge base, which is not edited directly"* | The agent tried to edit `CLAUDE.md`, `.claude/rules|skills|agents|docs`, or a file directly under `deltas/` | Nothing: it should go through `/drafter` or a `delta` command. Edit by hand if you want |
| *"billing was verified and is now stale: …"* after an edit | Code under the paths of its test commands changed | Check the contract still holds, then `/builder billing`, or `/drafter billing` if the contract is out of date |
| *"delta lint found problems in the test file you just wrote"* | A test cites an ID no spec defines, or the file has a skipped test | Fix the test file |
| The agent tries to finish and is sent back: *"delta check has not run on the current changes for: billing"* | Code under billing's paths changed since the last check | Let it run `delta check billing`. It asks once, never in a loop |
| *"billing is approved but not built, and the code changed since"* | Code was written for an approved slice without finishing it | Let `/builder billing` finish, ending with `delta built` |
| At the start of a session, a list of slices | Delta tells the session where each one stands | Nothing |

A broken or missing `delta` never blocks the agent: a hook only blocks when it says so.

## Commands you may run yourself

You rarely need them, because the skills run them. These are the useful ones:

| Command | Use it to |
|---|---|
| `delta status` | See every slice, its state, its claims by tag, and warnings about the setup. Also how many docs are written |
| `delta check [slice]` | Run the test claims. Exit `1` if one is violated |
| `delta lint` | Check every spec and cross it with your tests. Exit `1` on errors; warnings do not fail |
| `delta spec lint [slice]` | Check only the structure of a spec |
| `delta stale` | Mark verified slices stale when their code changed, and list them |
| `delta approve <slice>` | Approve a clean draft. **Only you run this** |
| `delta budget` | See how much context the project loads on every turn, against its ceilings |
| `delta update [--check]` | Bring Delta's installed files up to date |

Exit codes are `0` (fine), `1` (problems found), `2` (an error), so they work in scripts and CI. Findings go to stderr and results to stdout.

## Configuration

Two optional files in `deltas/`, both plain `key=value` lines with `#` comments. A key you leave out takes its default. A line Delta cannot use is ignored and reported by `delta status`, with its file and line.

### `deltas/testing.conf`

```
match=*_test.go,test_*.py,*_test.py,*.test.*,*.spec.*
skip=.git,node_modules,vendor
```

- `match`: patterns on a file's name that make it a test file. Setting it **replaces** the defaults, so list every pattern you use. Markdown is never a test.
- `skip`: folders never searched for tests. It also replaces the default. The `deltas/` folder is never searched.

For a Ruby project: `match=*_spec.rb`.

### `deltas/budget.conf`

The ceilings for the context that loads on every turn, so that `CLAUDE.md`, rules and skill descriptions do not quietly grow.

```
always_tokens=3500
claude_lines=80
rule_always_lines=50
skill_desc_chars=1024
agent_desc_chars=1024
```

`delta budget` weighs the project against them and exits `1` if one is passed. A value of `0` means no ceiling. When something is over, split what is used rarely into something that loads on demand, and leave a pointer; do not trim it away.

## How drift is decided

When `delta check` verifies a slice, Delta remembers a fingerprint of the files the slice depends on. If any of them changes later, the slice becomes `stale`.

A slice depends on **the code its test commands name**:

| Test command | The slice depends on |
|---|---|
| `go test ./billing` | `billing/` |
| `go test ./billing/...` | `billing/` and below |
| `go test ./billing ./shared/money` | both folders |
| `go test ./...`, or `make test` | every code file |

The slice's own spec always counts. Markdown and what Delta installs (`.claude/`, `deltas/`) never do. If any test command of a slice names no path, or the slice has no test claims, it depends on every code file.

So name the folder in your command when you can: it keeps one feature's changes from marking every other slice stale. The cost is a blind spot: a change in a package that `billing` imports, but that its command does not name, does not mark it stale.

The record of checks lives in your user cache folder, not in the repository. On a fresh clone nothing was verified there, so nothing is stale until a check runs.

## Docs for people

`/scriber overview` (then `structure`, `expertise` and `approach`) writes four short docs in `deltas/docs/` for whoever joins the project. They hold what the code does not show: the vision, why the structure is what it is, the words that trip up newcomers, the pillars that cannot break. Scriber reads your code, asks you for the reasons, shows you the draft and writes only after you approve.

`delta status` shows `docs: 1 of 4 written; next: /scriber structure`. A doc counts as written when no `<!-- -->` comment is left in it.

These docs are for people, and the agent rarely reads them. Anything the agent must obey belongs in a rule.

## Rules

`delta init` installs these into `.claude/rules/`:

| Rule | Loads | What it asks of the agent |
|---|---|---|
| `claims-fidelity` | Always | The contract is the scope. Tests cite their claim. If the code would have to contradict a claim, stop and ask |
| `done-needs-proof` | Always | Nothing is "done" until the checks ran in this session. Report which claims were touched and what stayed unverified |
| `project-language` | Always | Talk in your language, write everything the project keeps in English |
| `commit-messages` | Always | `type(scope): subject`, 40 to 50 characters, imperative, lowercase, no final period, a body that says why |
| `code-change-scope` | When you touch source files | Prefer the cheapest answer before writing new code |
| `code-reliability` | When you touch source files | Never cut validation, error handling or security to make a diff shorter |
| `when-write-delta` | When you touch specs | When something deserves its own spec, and how big it should be |
| `kb-deltas-routing` | When you touch the knowledge base | What goes through which skill |
| `context-budgets` | When you touch the knowledge base | What to do when an artifact outgrows its ceiling |

Add your own rules as files in `.claude/rules/`. `delta update` never touches a file it did not install, and a rule with a `paths:` list in its frontmatter loads only when a matching file is read, so it costs nothing the rest of the time. No hook checks commit messages: the rule asks, and the agent follows it.

## Keeping Delta up to date

After you rebuild the `delta` binary:

```
delta update --check    # list what would change; exit 1 if anything would
delta update
```

`update` installs what is missing, replaces the skills and rules you have not touched, removes the ones this version no longer ships, and **keeps every file you changed**, telling you which. The config files, `CLAUDE.md` and the doc templates are yours after `delta init`: `update` never rewrites them.

`delta status` warns when the installed files are out of date.

## Troubleshooting

| Message | What it means |
|---|---|
| `delta is not on the PATH, so the hooks of Claude Code cannot run it` | Put `dist/delta` on your `PATH` and restart Claude Code |
| `the delta on the PATH (…) is not this program` | Another program called `delta` comes first. Reorder your `PATH` |
| `the installed harness is out of date` | Run `delta update` |
| `LINT-NO-CITE`: *a test tag, but no test file under "…" cites its ID* | No test under the command's paths mentions the claim ID. Add `INV-01` to the test, or fix the path in the command |
| `LINT-UNKNOWN-ID` (warning) | A test cites an ID no spec defines. A typo, or a claim that was removed |
| `LINT-SKIP` (warning) | A test file that cites claims has a skipped or focused test. A skipped test passes without running |
| `Unverified: the command exited 0 but no test it runs cites the claim ID` | The test passes but nothing ties it to the claim. Cite the ID |
| `Unverified: could not run the command` | The command is not a program on your machine. Commands run without a shell, from the project root: no pipes, redirections or `$VARIABLES`. Wrap such a check in a script |
| `SPEC-…` codes from `delta spec lint` | The spec's shape is wrong. The message names the line and what to fix |
| *"the spec is not in English"* | Translate it and write it again |
| *"a new spec is written as a draft"* | Specs cannot be written with a status other than `draft`. Only `delta approve`, `built` and `check` move it |
| A slice is `stale` right after you edited code | Expected: that is drift. Check the contract, then rebuild or revise |
| A check takes forever | A command is limited to ten minutes, then it is `Unverified` |

## Limits worth knowing

- **`Holds` is not proof of quality.** It means the command exited 0 and a test cites the ID. A weak test makes a weak claim.
- **Drift is a heuristic.** It follows the paths your commands name. See [How drift is decided](#how-drift-is-decided).
- **Claims with no command are checked by reading, and a reading is a report.** `/auditor` can say a `none` claim holds, but the slice's state does not change because of it. If it holds, write the test.
- **One slice at a time.** Skills do not sweep every spec unasked.
- **The agent follows the rules; nothing enforces them except the hooks above.** Commit messages and code conventions are asked, not checked.
