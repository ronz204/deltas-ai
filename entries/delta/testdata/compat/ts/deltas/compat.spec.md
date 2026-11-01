---
format: 1
slice: compat
status: approved
---

## Invariants
- INV-01 [test: node --test pass/pass.test.ts]
  When a test passes and cites the claim, the claim holds.
- INV-02 [test: node --test fail/fail.test.ts]
  When a test fails, the claim is violated.
- INV-03 [test: node --test uncited/uncited.test.ts]
  When the command exits 0 but no test cites the claim, the claim stays unverified.
- INV-04 [test: node --test skip/skip.test.ts]
  When a test is skipped, the command still exits 0.
- INV-05 [none]
  While nothing verifies a claim, it stays unverified.

## Acceptance
- ACC-01 [manual: read the table printed by delta check]
  Given the five claims above, when delta check runs, then each gets the verdict its evidence supports.
