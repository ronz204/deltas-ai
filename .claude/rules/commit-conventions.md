# Commit Message Conventions

This governs every git commit message written in this repository, by a person or by Claude. It applies globally, so it carries no `paths:` scope. Pull request titles are out of scope.

---

## Format

- The subject follows Conventional Commits: `type(scope): subject`. A scope is optional but preferred, because it makes `git log --oneline` scannable by area.
- Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`. A fixed vocabulary keeps history greppable and lets tooling derive changelogs.
- The whole subject line, type and scope included, is between 40 and 50 characters. Under 40 usually means the message says too little to be useful; over 50 gets truncated in common git UIs.
- The subject is in the imperative mood ("add", not "added" or "adds"), lowercase after the colon, with no trailing period. The imperative reads as what applying the commit does.
- A body is optional. When used, separate it from the subject with a blank line and explain why the change was made, not what the diff already shows.

## Language

- Commit messages are written in English, whatever language the team speaks day to day, because history is read by tooling and future contributors who may not share that language.

## When to commit

- Claude commits on its own when a whole phase of work finishes and has been checked (a phase of a plan, a feature, a set of related docs), without asking first. A phase is the unit, not a single screen, state, or tweak: a history of one-line commits per screen buries the real changes and piles up faster than anyone reviews it.
- A design file and the docs that record the same phase go in one commit, because they describe one change and are reverted together.
- Work that is half done, or that the user was asked to approve and has not, is not committed until it closes.
- This covers committing only: pushing and pulling stay denied by the permission policy.

## Atomic and cohesive commits

- One logical change per commit, where a logical change is a phase of work and not its smallest step, because a commit has to be revertible without collateral damage but also worth reading. If the subject needs "and" to describe unrelated changes, it is probably two commits.
- Group files by concern, not by when they were touched: knowledge-base files, project configuration, source, and tests each go in their own commits, because a mixed commit cannot be reverted without losing unrelated work. The one exception is a design file and the docs recording the same phase, which travel together.
- Stage explicitly by path (or hunk when one file carries two concerns); never `git add -A` or `git add .`, because they sweep unrelated or sensitive files into the commit. Review the staged diff before committing.
- Every commit leaves the repository in a working state: no half-written file that only makes sense with the next commit, because bisecting lands on arbitrary commits.
- Never commit secrets or real environment files; only the `.sample` variants are versioned.
- Keep refactors, renames, and behavior changes in separate commits, because a diff that does both hides the behavior change.
- Choose the type by what the change is: `docs` for knowledge-base, README, and roadmap content; `chore` for repo housekeeping and scaffolding; `build` for packaging and dependency configuration; `ci` for workflow changes; `feat`, `fix`, `refactor`, `test` for source changes.

---

## Non-goals

- Doesn't dictate branch names or pull request titles.