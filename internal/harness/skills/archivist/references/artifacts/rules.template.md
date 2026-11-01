<!--
Template for .claude/rules/<name>.md — a convention auto-loaded when Claude Code
reads a file matching the `paths:` globs below (same priority as CLAUDE.md).
Fill every section; delete guidance comments before presenting the draft.
- Quote every glob (YAML requires it for patterns starting with * or {).
- One rule file per concern: a convention with different trigger paths is a new
  rule, not an addition to an unrelated one.
- Omit `paths:` only for a rule meant to apply globally (rare). A rule about how
  to work (the size of a change, a process) applies whatever file is open, so it
  is global and stays under the always-loaded rule ceiling, with detail only some
  uses need moved to a reference that a pointer step loads. Scoping it by
  `paths:` would leave it dead in any layout it didn't guess.
-->
---
paths:
  - "<glob pattern, quoted — YAML requires it for patterns starting with * or {>"
---

# <Name> Conventions

<!--
One short paragraph: what this covers, what runtime/toolchain it targets, and
what's explicitly out of scope for now (library-specific conventions not yet
decided, forward-looking vs. descriptive-of-current-code, etc).
-->

---

## <Topic 1>

<!-- Bullet list. Each bullet states the convention AND the reason ("X, because Y") — a bare "must" invites being silently violated the first time it's inconvenient. -->

## <Topic 2>

---

## Non-goals

<!-- Optional. Include only if this rule is easy to over-apply — e.g. a convention that shouldn't be retrofitted onto an explicitly-excepted case. -->
