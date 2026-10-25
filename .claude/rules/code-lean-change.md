---
paths:
  - "*/cmd/**"
  - "*/source/**"
---

# Lean-Change Code Conventions

This governs how much code a change adds: it is applied whenever source is written or edited, and it is a ladder of cheaper answers to try before writing new code. It decides size only. How code is expressed is the job of the DDD and language rules, and whether something deserves a spec is the job of the delta-scope rule. Whether it improves results has not been measured.

---

## Before writing

- Read the code the change touches and list everything it must reach: callers, tests, fixtures, config, exports. Being lazy about the solution is only safe once the scope is known, and a change that misses a caller is not smaller, only broken.
- When the slice has a spec, its claims are the scope. What no claim and no request asks for is extra: name it in one line and skip it, because each unrequested feature is surface nobody verified.

## The ladder

Take the first rung that fully solves the task:

1. Does it need to exist? A spec claim or an explicit request says yes. A vague request ("build me X") gets the smallest version that does the core job.
2. Is it already in this codebase (a helper, a component, a pattern)? Use it the way the surrounding code does.
3. Does the standard library or the platform do it? Use it, unless the project has its own.
4. Does an installed dependency do it? Use it. Never add one for a few lines.
5. Can it be one line a reader understands at a glance? Write the line.
6. Otherwise, the minimum code that works.

## While writing

- No abstraction, wrapper, option, config or "for later" code nobody asked for. Deletion beats addition.
- Keep the structure the codebase has: its layers, interfaces and conventions. Least code never means crossing a layer boundary to save a line.
- A bug fix starts by finding every caller of what is touched, then fixes the root cause once, in the shared code.
- Code that is moved or merged keeps its validation and error handling.
- Between options of equal size, take the one that is correct on edge cases. A one-liner that needs decoding is not short.
- Comment only the why the code cannot show, in one line.
- A shortcut with a known limit gets a comment: `shortcut: <the limit>, <when to upgrade>`, so the ceiling is visible to the next reader.

## Never cut

- Validation at trust boundaries, error handling that prevents data loss, security, accessibility, and anything the user asked for. A shorter diff that drops one of these is a regression.
- The check for new non-trivial logic (a branch, a loop, a parser, money, security): one small test, or the one a spec claim's verification tag names. Trivial changes need none.

## Closing the reply

- End with one or two lines: what was skipped or left unchecked, and any risk the user must know. Skipped work that goes unmentioned looks like finished work.

---

## Non-goals

- Not a target of fewest lines or tokens: size is the means, a complete and safe change is the goal.
- Does not override a spec: a claim in the contract is requested work, however much code it needs.
