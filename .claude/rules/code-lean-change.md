# Lean-Change Code Conventions

Applies to every change to source, with or without a spec. It decides size only: how code is expressed belongs to the DDD and language rules, whether something deserves a spec to the delta-scope rule. Whether it improves results has not been measured.

---

## Before writing

- Read what the change touches and list everything it must reach (callers, tests, fixtures, config, exports) before choosing the solution: a change that misses a caller is not smaller, only broken.
- When the slice has a spec, its claims are the scope. What no claim and no request asks for is extra: name it in one line and skip it.

## The ladder: take the first rung that fully solves the task

1. Does it need to exist? A spec claim or an explicit request says yes. A vague request gets the smallest version that does the core job.
2. Is it already in this codebase? Use it the way the surrounding code does.
3. Does the standard library or the platform do it? Use it, unless the project has its own.
4. Does an installed dependency do it? Use it. Never add one for a few lines.
5. Can it be one line a reader understands at a glance? Write the line.
6. Otherwise, the minimum code that works.

## While writing

- No abstraction, wrapper, option, config or "for later" code nobody asked for. Deletion beats addition.
- A bug fix finds every caller of what is touched, then fixes the root cause once, in the shared code.

## Never cut

- Validation at trust boundaries, error handling that prevents data loss, security, accessibility, and anything the user asked for.
- New non-trivial logic (a branch, a loop, a parser, money, security) gets one small test, or the one a spec claim's verification tag names.

## Closing the reply

- End with one or two lines: what was skipped or left unchecked, and any risk the user must know.
