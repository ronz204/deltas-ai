# Working with Delta

Welcome! This is the guide for using Delta in your own projects. For what Delta is and why it exists, read the [README](README.md) first.

The short version: you describe a feature, the agent drafts a **contract** for it, you approve the contract, the agent builds the feature against it, and Delta keeps checking that the code still keeps its promises.

**You will almost never type a `delta` command.** You talk to the agent in plain words, answer its questions, and make two decisions: approve the contract and approve the plan. The agent runs the commands.

| | You | The agent | The hooks |
|---|---|---|---|
| Ask for features and answer questions | ✅ | | |
| Read and approve the contract | ✅ | | |
| Approve the plan | ✅ | | |
| Write the contract, tests and code | | ✅ | |
| Run `delta spec write`, `link`, `built`, `check` | | ✅ | |
| Run `delta approve` | confirm it | runs it after your yes | |
| Report drift, block edits, remind it to check | | | ✅ |

### Contents

1. [Getting started](#getting-started): install, set up a project, your first feature
2. [A full example session](#a-full-example-session): what a real conversation looks like
3. [Day to day](#day-to-day): the skills, the states and the messages you will see
4. [Contracts and verification](#contracts-and-verification): how a contract is written and checked
5. [Configuration and conventions](#configuration-and-conventions): config files, rules and docs
6. [Maintenance](#maintenance): updating, troubleshooting and limits

## Getting started

### Requirements

- A git repository for your project.
- [Claude Code](https://claude.com/claude-code). Delta's skills and hooks are for it.
- Go, only if you build `delta` from source. The ready-made binary does not need it.

### Install the CLI

Run the installer for your system. It downloads the latest release, checks its SHA-256 against the published `checksums.txt`, and puts `delta` on your `PATH`. You do not need Go.

**Linux and macOS**

```sh
curl -fsSL https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.sh | sh
```

It installs to `~/.local/bin`. If that folder is not on your `PATH`, the script prints the line to add to your shell profile.

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.ps1 | iex
```

It installs to `%LOCALAPPDATA%\Programs\delta` and adds that folder to your user `PATH`, with no administrator rights. Restart open terminals afterwards.

Check it with `delta version`, which prints `delta (deltas-ai) 0.1.0` (or whichever version you installed).

Both installers read the same settings from the environment: `DELTA_VERSION` to pick a release other than the latest, and `DELTA_INSTALL_DIR` to choose the folder.

**Downloading by hand.** Each [release](https://github.com/ronz204/deltas-ai/releases) has an archive per system (`delta_<version>_<os>_<arch>`, as `.tar.gz` or `.zip`) and a `checksums.txt`. Unpack it and put `delta` on your `PATH`. The binaries are not code-signed, so Windows SmartScreen or macOS Gatekeeper may warn the first time you run one downloaded with a browser; the installers do not trigger it.

**Build from source.** From a clone of this repository, with Go installed:

```
go build -o dist/delta ./cmd/delta
```

Put `dist/delta` on your `PATH`. A local build prints `delta (deltas-ai) dev`.

The binary has to be called `delta`: the hooks and permissions call it by that name. If another program called `delta` comes first on your `PATH` (the `git-delta` pager, for example), `delta status` warns you.

### Set up a project

Inside your project, run `delta init`. It installs what Delta needs and never replaces a file you already have:

- `CLAUDE.md`: a short router for the agent. Replace the comment at the top with two or three lines on what your project is.
- `.claude/skills/`: the five skills.
- `.claude/rules/`: the conventions, which load by themselves when they apply.
- `.claude/settings.json`: Delta's hooks and permissions, merged with what you already have.
- `deltas/specs/`: where the contracts will live.
- `deltas/docs/`: four templates for docs for people.
- `deltas/budget.conf` and `deltas/testing.conf`: the two config files, with their defaults.
- `deltas/deltas.manifest`: a record of what Delta installed, so it can update it later.

Commit all of it. Anyone who clones the project gets the same skills, rules and hooks without running `delta init` again (they do need `delta` on their `PATH`).

Run `delta status` to see where things stand. A new project says `no specs yet`.

### Your first feature

You can talk plainly or call the skills by name.

1. **Ask for the feature.** *"Add refunds to billing"*, or `/drafter billing`.
2. **Answer the interview.** Drafter reads the real code, then asks what the code cannot tell it: why the feature exists, what it must never do, what it leaves out, and how each claim can be verified. Take the time to answer well; the invariants are what will be checked later.
3. **Review the draft.** Drafter saves the contract as a `draft` and summarizes its claims, saying which have no verification yet. Read `deltas/specs/billing.spec.md` yourself. If something is wrong, say so and it rewrites it.
4. **Approve it.** Say yes when asked. Delta runs `delta approve billing` and Claude Code asks you to confirm the command. Approving is always your decision, never the agent's.
5. **Build it.** `/builder billing`, or just keep going. Builder works out what is not verified yet, shows a plan, and waits. After your OK it writes each test first, then the code, and links every test to its claim.
6. **See it verified.** When every test claim holds, the slice becomes `verified`. `delta status` shows it.
7. **Audit it later.** `/auditor billing` checks the contract still holds and reads the code for claims without a test.

At any point, `/comper` tells you where each slice stands and which skill comes next.

## A full example session

Here is what the steps above look like as a conversation. It is an illustration: the agent's exact words will differ, but the order of the steps and who decides at each one will not.

**1. You ask for the feature**

> 👤 **You:** add refunds to billing

A hook has already told the session that `billing` has no contract, so the agent routes the request to Drafter.

**2. Drafter interviews you**

> 🤖 **Drafter:** I read `billing/` and I have a few questions the code can't answer:
> 1. Are partial refunds allowed?
> 2. Is there a time limit after the charge?
> 3. What happens to the original charge: is it cancelled, or does a credit note get issued?
>
> 👤 **You:** Partial refunds yes. 30 days maximum. A credit note; the original charge is never touched.

**3. You review the draft**

> 🤖 **Drafter:** I saved `deltas/specs/billing.spec.md` as a `draft`. It has three claims:
> - `INV-01`: a refund never exceeds what was charged
> - `INV-02`: a refund after 30 days is rejected
> - `ACC-01`: a partial refund issues a credit note
>
> None of them has a test yet, so all three are listed as debt. Please read the file and tell me if anything is wrong.
>
> 👤 **You:** (reads the file) Make it 45 days, not 30.
>
> 🤖 **Drafter:** Done. `INV-02` now says 45 days.

**4. You approve**

> 🤖 **Drafter:** Do you approve this contract?
>
> 👤 **You:** Yes.
>
> Claude Code then shows the command `delta approve billing` and asks you to confirm it. Confirm. The slice is now `approved`.

**5. Builder proposes a plan and you OK it**

> 👤 **You:** /builder billing
>
> 🤖 **Builder:** Nothing is verified yet. My plan:
> 1. Write a test for each claim, each citing its ID.
> 2. Implement `billing/refund.go`.
> 3. Link every claim to `go test ./billing`.
>
> Shall I go ahead?
>
> 👤 **You:** Go ahead.

**6. Builder builds and verifies**

The agent writes each test first. A hook lints every test file as it is written, so a test that cites no claim is caught at once. Then it writes the code, runs `delta link`, `delta built billing` and `delta check billing`.

> 🤖 **Builder:** All three claims hold. `billing` is `verified`. I touched `INV-01`, `INV-02` and `ACC-01`, and nothing stayed unverified.

**7. Weeks later, someone changes the code**

> 👤 **You:** fix the rounding in `billing/refund.go`
>
> 🤖 **Agent:** (edits the file)
>
> A hook reports: *"billing was verified and is now stale"*. When the agent tries to finish, the `Stop` hook asks it, once, to run `delta check billing`. The agent does and reports each claim as `Holds`, `Violated` or `Unverified`. If the change touched behavior, `/builder billing` or `/drafter billing` takes it from there.

**What you did in all this:** asked, answered, read the contract, said yes twice (the contract and the plan) and confirmed one command. Everything else was the agent's.

## Day to day

### The five skills

- **`/comper [slice]`**: call it when you don't know the next step. It leaves you with where the slice stands and the skill to call.
- **`/drafter <slice>`**: call it when a feature is new, changes, or its contract looks thin. You give it your answers about intent, limits and invariants; it leaves you a `draft` contract to review and approve.
- **`/builder <slice>`**: call it when a contract is approved or went stale. You give it an OK on its plan; it leaves you tests and code, each test linked to its claim.
- **`/auditor <slice>`**: call it to know whether a contract still holds. It gives you verdicts per claim, saying whether each came from a command or from reading. It fixes nothing.
- **`/scriber <doc>`**: call it when the project needs explaining to people. You give it the reasons the code doesn't show; it writes one doc under `deltas/docs/`.

Each skill works on one slice (or one doc) per run. Drafter and Builder never approve; Auditor never edits. When the contract itself is wrong, any of them hands it back to Drafter instead of bending the code.

Plain requests work too: *"is billing still right?"* goes to Auditor, and *"change how refunds round"* goes to Drafter first, because a change of behavior starts with the contract. A request not about behavior (a typo, a rename, a question about the code) is not routed at all.

### Slice states

- **`draft`**: the contract is written and waits for review. Drafter writes it; **you** approve it.
- **`approved`**: the contract is accepted. Builder runs `delta built` once the code is done.
- **`built`**: code and tests exist. `delta check` makes it `verified` when every test claim holds.
- **`verified`**: every test claim held in the last check. If the code its tests cover changes afterwards, it stays `verified` and `delta status` shows it as `verified (2 files changed)`. Nothing is rewritten: `delta check` on the slice confirms it still holds and clears the note, and if a claim now fails, Builder fixes the code or Drafter revises the contract.
- **`stale`**: only in specs written by an older Delta, which marked a slice stale on its own. `delta check` brings it back to `verified` once its test claims hold.

Rewriting an `approved`, `built` or `verified` contract sends it back to `draft` and needs a new approval; Delta says so when it happens. A claim with no test never blocks `verified`: it stays on your screen as debt.

### What you will see while you work

Delta speaks through the hooks. These are the messages you will meet.

- **An edit is refused** with *"X is part of the knowledge base, which is not edited directly"*. The agent tried to edit `CLAUDE.md`, `.claude/rules|skills|agents|docs`, or a file directly under `deltas/`. Nothing to do: it should go through `/drafter` or a `delta` command. You can edit by hand.
- ***"billing was verified and is now stale: …"*** after an edit. Code under the paths of its test commands changed. Check the contract still holds, then run `/builder billing`, or `/drafter billing` if the contract is out of date.
- ***"delta lint found problems in the test file you just wrote"***. A test cites an ID no spec defines, or the file has a skipped test. Fix the test file.
- **The agent tries to finish and is sent back** with *"delta check has not run on the current changes for: billing"*. Code under billing's paths changed since the last check. Let it run `delta check billing`. It asks once, never in a loop.
- ***"billing is approved but not built, and the code changed since"***. Code was written for an approved slice without finishing it. Let `/builder billing` finish, ending with `delta built`.
- **A list of slices at session start**. Delta telling the session where each one stands. Nothing to do.

A broken or missing `delta` never blocks the agent. A hook only blocks when it says so.

### Commands you may run yourself

You rarely need them, because the skills run them. The useful ones:

- `delta status`: every slice, its state, its claims by tag, setup warnings and how many docs are written.
- `delta check [slice]`: run the test claims. Exits `1` if one is violated.
- `delta lint`: check every spec and cross it with your tests. Exits `1` on errors; warnings don't fail.
- `delta spec lint [slice]`: check only the structure of a spec.
- `delta stale`: list the verified slices whose code changed since they were verified. It changes nothing; exits `1` if there is any.
- `delta approve <slice>`: approve a clean draft. **Only you run this.**
- `delta budget`: see how much context the project loads on every turn against its ceilings.
- `delta update [--check]`: bring Delta's installed files up to date.

Exit codes are `0` (fine), `1` (problems found) and `2` (an error), so they work in scripts and CI. Findings go to stderr and results to stdout.

## Contracts and verification

### Reading and writing a contract

A contract is `deltas/specs/<slice>.spec.md`: a short frontmatter, free sections, then claims.

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

**Frontmatter.** `format` is `1`, `slice` is the file name in lowercase words joined by hyphens, and `status` is the state. Don't set `status` by hand: the commands move it.

**Claims.** Each claim is a header, `- INV-01 [tag]` for an invariant or `- ACC-01 [tag]` for an acceptance criterion, with one indented sentence under it.

- An invariant is one sentence that something observable can confirm. It may open with `When`, `While`, `Where` or `If … then`.
- An acceptance criterion reads `Given <context>, when <action>, then <outcome>`.
- IDs are stable. Never renumber a claim and never reuse the ID of a removed one.
- Write the spec in English, whatever language you talk in. Delta warns when it isn't.

**Tags** say how a claim is verified:

- `test: <command>`: a command that exits 0 when the claim holds.
- `manual: <step>`: only a person can check it. The step is optional.
- `none`: nothing verifies it yet. Valid, and listed as debt.

**Don't edit specs through the agent's file tools.** A hook blocks it on purpose: specs are written with `delta spec write`, which validates before saving, and tests are linked with `delta link`. As a person you can edit the files by hand; run `delta spec lint` afterwards.

**What is a slice?** A bounded feature or capability, sized like a bounded context. Not a source file, and never an architecture layer (no "billing domain" spec plus "billing infrastructure" spec). A pattern applied across many features, like how repositories are written, is a rule, not a slice. A use case that only orchestrates one feature belongs in that feature's contract.

### Linking a claim to a test

A claim with `test: <command>` is checked by running the command, and the test has to say which claim it proves.

1. **The command** exits 0 when the claim holds and names the file, package or folder it runs. Prefer `go test ./billing` over `go test -run TestRefund`: a filter that matches nothing passes silently.
2. **The test cites the ID**, in its name or a comment. `INV-01`, `INV_01` and `INV01` all count, in any case.

```go
// INV-01: a refund never exceeds the charge
func TestRefundNeverExceedsTheCharge(t *testing.T) { ... }
```

Builder does both for you with `delta link billing INV-01 "go test ./billing"`. If you link by hand, quote the command so it is one argument.

`delta check` then gives each claim a verdict:

- **`Holds`**: the command exited 0 and a test under its paths cites the ID. It does not prove the test is a good one.
- **`Violated`**: the command failed. The code or the test is wrong.
- **`Unverified`**: no evidence either way. The command didn't run, no test cites the ID, or the claim is `manual` or `none`. It never passes for `Holds`.

A citation only counts under the paths the command names. A command that names no path (`make test`) accepts a citation anywhere.

### The footprint of a slice

Delta has to know which files belong to a slice. It does not guess from how your project is laid out: it watches the build. `delta ready billing` takes a snapshot of the project's code, and `delta built billing` compares it with the code now. Every file that is new, changed or deleted in between is recorded in `deltas/specs/billing.files`, one path per line:

```
src/billing/total.ts
testing/unit/billing.test.ts
```

That works whether tests sit beside the code, in a folder of their own, or in a tree shaped differently from the source. The file is committed with the spec, so a teammate's clone sees the same footprint. The agent cannot edit it; a person can, to take out something that does not belong. A later build adds its files to the footprint and drops the ones it deleted.

Once a slice has a footprint, only those files (and the folders its test commands name, if every command names one) count as its code. Changing `go.mod`, an image or another slice's code does not touch it. A slice built before this existed has no footprint and keeps the rule below.

### How drift is decided

When `delta check` verifies a slice, Delta remembers a fingerprint of the files the slice depends on. If any of them changes later, `delta status` and `delta stale` list the slice with the files that changed, and the session start tells the agent to run `delta check` on it. The spec itself is never rewritten for this: drift is read from the files, not stored in the contract.

A slice with no footprint depends on **the code its test commands name**:

- `go test ./billing` → `billing/`
- `go test ./billing/...` → `billing/` and below
- `go test ./billing ./shared/money` → both folders
- `go test ./...` or `make test` → every code file

The slice's own spec always counts. Markdown and what Delta installs (`.claude/`, `deltas/`) never do. If any test command of a slice names no path, or the slice has no test claims, it depends on every code file.

So name the folder in your command when you can: it keeps one feature's changes from marking every other slice stale. The cost is a blind spot: a change in a package that `billing` imports but its command doesn't name won't mark it stale.

The record of checks lives in your user cache folder, not in the repository. On a fresh clone nothing was verified there, so nothing is stale until a check runs.

## Configuration and conventions

### Config files

Two optional files in `deltas/`, both plain `key=value` lines with `#` comments. A key you leave out takes its default. A line Delta can't use is ignored and reported by `delta status` with its file and line.

**`deltas/testing.conf`**

```
match=*_test.go,test_*.py,*_test.py,*.test.*,*.spec.*
skip=.git,node_modules,vendor
```

- `match`: patterns on a file's name that make it a test file. Setting it **replaces** the defaults, so list every pattern you use. Markdown is never a test.
- `skip`: folders never searched for tests. It also replaces the default. The `deltas/` folder is never searched.

For a Ruby project: `match=*_spec.rb`.

**`deltas/budget.conf`**

Ceilings for the context that loads on every turn, so `CLAUDE.md`, rules and skill descriptions don't quietly grow.

```
always_tokens=3500
claude_lines=80
rule_always_lines=50
skill_desc_chars=1024
agent_desc_chars=1024
```

`delta budget` weighs the project against them and exits `1` if one is passed. A value of `0` means no ceiling. When something is over, split what is rarely used into something that loads on demand and leave a pointer; don't trim it away.

### Rules

`delta init` installs these into `.claude/rules/`.

**Always loaded**
- `claims-fidelity`: the contract is the scope. Tests cite their claim. If the code would have to contradict a claim, stop and ask.
- `done-needs-proof`: nothing is "done" until the checks ran in this session. Report which claims were touched and what stayed unverified.
- `project-language`: talk in your language, write everything the project keeps in English.
- `commit-messages`: `type(scope): subject`, 40 to 50 characters, imperative, lowercase, no final period, a body that says why.

**When you touch source files**
- `code-change-scope`: prefer the cheapest answer before writing new code.
- `code-reliability`: never cut validation, error handling or security to make a diff shorter.

**When you touch specs or the knowledge base**
- `when-write-delta`: when something deserves its own spec, and how big it should be.
- `kb-deltas-routing`: what goes through which skill.
- `context-budgets`: what to do when an artifact outgrows its ceiling.

Add your own rules as files in `.claude/rules/`. `delta update` never touches a file it didn't install, and a rule with a `paths:` list in its frontmatter loads only when a matching file is read, so it costs nothing the rest of the time. No hook checks commit messages: the rule asks and the agent follows.

### Docs for people

`/scriber overview` (then `structure`, `expertise` and `approach`) writes four short docs in `deltas/docs/` for whoever joins the project. They hold what the code doesn't show: the vision, why the structure is what it is, the words that trip up newcomers, the pillars that can't break. Scriber reads your code, asks you for the reasons, shows you the draft and writes only after you approve.

`delta status` shows progress, like `docs: 1 of 4 written; next: /scriber structure`. A doc counts as written when no `<!-- -->` comment is left in it.

These docs are for people, and the agent rarely reads them. Anything the agent must obey belongs in a rule.

## Maintenance

### Keeping Delta up to date

The binary and the files it installs in your project are updated separately.

**The binary.** Run the installer again. It replaces the old `delta` with the latest release (or the one in `DELTA_VERSION`).

**The files in your project.** After a new binary is in place:

```
delta update --check    # list what would change; exit 1 if anything would
delta update
```

`update` installs what is missing, replaces the skills and rules you haven't touched, removes the ones this version no longer ships, and **keeps every file you changed**, telling you which. The config files, `CLAUDE.md` and the doc templates are yours after `delta init`: `update` never rewrites them.

`delta status` warns when the installed files are out of date.

### Troubleshooting

**Setup**
- *`delta is not on the PATH, so the hooks of Claude Code cannot run it`*: put the folder with `delta` on your `PATH` and restart Claude Code. The installers use `~/.local/bin` on Linux and macOS and `%LOCALAPPDATA%\Programs\delta` on Windows; a build from source is in `dist/`.
- *`the delta on the PATH (…) is not this program`*: another program called `delta` comes first. Reorder your `PATH`.
- *`the installed harness is out of date`*: run `delta update`.

**Lint**
- `LINT-NO-CITE` (*a test tag, but no test file under "…" cites its ID*): no test under the command's paths mentions the claim ID. Add `INV-01` to the test, or fix the path in the command.
- `LINT-UNKNOWN-ID` (warning): a test cites an ID no spec defines. A typo, or a removed claim.
- `LINT-SKIP` (warning): a test file that cites claims has a skipped or focused test. A skipped test passes without running.
- `SPEC-…` codes from `delta spec lint`: the spec's shape is wrong. The message names the line and what to fix.

**Verdicts**
- *`Unverified: the command exited 0 but no test it runs cites the claim ID`*: the test passes but nothing ties it to the claim. Cite the ID.
- *`Unverified: could not run the command`*: the command is not a program on your machine. Commands run without a shell, from the project root: no pipes, redirections or `$VARIABLES`. Wrap such a check in a script.
- **A check takes forever**: a command is limited to ten minutes, then it is `Unverified`.

**Writing specs**
- *"the spec is not in English"*: translate it and write it again.
- *"a new spec is written as a draft"*: specs can't be written with another status. Only `delta approve`, `built` and `check` move it.
- **A slice is `stale` right after you edited code**: expected, that is drift. Check the contract, then rebuild or revise.

### Limits worth knowing

- **`Holds` is not proof of quality.** It means the command exited 0 and a test cites the ID. A weak test makes a weak claim.
- **Drift is a heuristic.** It follows the footprint of the build, or the paths your commands name when there is none. A file the slice uses but nobody touched during its build is not in the footprint, so changing it does not show as drift.
- **Footprints record what changed, not what was meant.** An edit made during a build, by anyone, lands in the footprint. Two slices built at the same time in one working tree mix theirs. Build one at a time and check the list `delta built` prints.
- **Claims with no command are checked by reading, and a reading is a report.** `/auditor` can say a `none` claim holds, but the slice's state doesn't change because of it. If it holds, write the test.
- **One slice at a time.** Skills don't sweep every spec unasked.
- **The agent follows the rules; nothing enforces them except the hooks above.** Commit messages and code conventions are asked, not checked.
