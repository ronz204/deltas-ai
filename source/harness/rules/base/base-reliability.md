---
paths:
  - "**/*.{ts,tsx,js,jsx,py,go,rs,java,kt,cs,rb,php}"
---

# Base preset: reliability

Part of the `base` preset; delete this file or set `base=off` in `deltas/deltas.config` to turn it off.

Brevity never cuts these. A shorter diff that drops one is a regression:

- Validation at trust boundaries (user input, network, files, other services).
- Error handling that prevents data loss. Never swallow an error with an empty catch or a bare ignore.
- Security and accessibility.
- Code that is moved or merged keeps its validation and error handling.

New non-trivial logic (a branch, a loop, a parser, money, security) gets a test linked to its claim with `delta link`. Trivial changes need none.
