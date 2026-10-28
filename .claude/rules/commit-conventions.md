# Commit Message Conventions

Governs every git commit message in this repository, by a person or by Claude. Pull request titles are out of scope. A `commit-msg` git hook checks the format below and explains each failure; the rest is judgment.

---

## Format

- Subject: `type(scope): subject`, scope optional but preferred. Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
- The whole subject line is 40 to 50 characters, imperative, lowercase after the colon, no trailing period, in English.
- A body is optional: blank line after the subject, then why the change was made, not what the diff shows.

## When to commit

- Claude commits on its own, without asking, when a whole phase of work is finished and checked: a plan phase, a feature, a set of related docs. Never per screen, state or tweak.
- Half-done work, or work the user was asked to approve and has not, waits. Pushing and pulling stay denied.

## Atomic and cohesive

- One logical change per commit, meaning a phase, revertible without collateral damage. "And" in the subject usually means two commits.
- Group by concern: knowledge-base files, project configuration, source and tests each commit apart. A design file and the docs of the same phase travel together.
- Refactors, renames and behavior changes go in separate commits.
- Stage by path (or hunk), never `git add -A` or `git add .`, and review the staged diff before committing.
- Every commit leaves the repo working, so bisecting lands on a usable state.
- Never commit secrets or real environment files; only `.sample` variants are versioned.
- Type by what changed: `docs` for knowledge base, README and roadmap; `chore` for housekeeping and scaffolding; `build` for packaging and dependencies; `ci` for workflows; `feat`, `fix`, `refactor`, `test` for source.
