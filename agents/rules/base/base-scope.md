---
paths:
  - "**/*.{ts,tsx,js,jsx,py,go,rs,java,kt,cs,rb,php}"
---

# Base preset: scope

Part of the `base` preset that `delta init` installs; delete this file or set `base=off` in `deltas/deltas.config` to turn it off. The ladder is adapted from ponytail (MIT license) with contract rungs. It decides how much code a change adds, not how it is written.

Take the first rung that fully solves the task:

0. Read the code the change touches and list everything it must reach: callers, tests, fixtures, config, exports. A change that misses a caller is not smaller, it is broken.
1. Which claim asks for it? If none and no explicit request does, name it in one line and skip it.
2. Does it already exist in this slice or another? Reuse it, as the surrounding code does.
3. Does the platform, the standard library or an installed dependency do it? Use it. Never add a dependency for a few lines.
4. Can it be one line a reader understands at a glance? Write that line.
5. Otherwise the minimum that satisfies the claim, with its test linked.
6. If it cannot be decided without inventing, stop and record an open question in the spec (see fidelity).

Less code is the means, not the goal: a claim in the spec is requested work, however much code it needs.
