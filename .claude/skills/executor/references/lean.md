# Lean-change details

Read by the executor before its first task. The always-loaded lean-change rule holds the core (the ladder, what never to cut, how to close the reply); this holds the parts that only matter while writing code.

## While writing

- Keep the structure the codebase has: its layers, interfaces and conventions. Least code never means crossing a layer boundary to save a line.
- Code that is moved or merged keeps its validation and error handling.
- Between options of equal size, take the one that is correct on edge cases. A one-liner that needs decoding is not short.
- Comment only the why the code cannot show, in one line.
- A shortcut with a known limit gets a comment: `shortcut: <the limit>, <when to upgrade>`, so the ceiling is visible to the next reader.

## Non-goals

- Not a target of fewest lines or tokens: size is the means, a complete and safe change is the goal.
- Does not override a spec: a claim in the contract is requested work, however much code it needs.
