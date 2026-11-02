# Fidelity: the contract rules

The spec of a slice (`deltas/specs/<slice>.spec.md`) is the truth and the scope of the work.

- Every test you write cites the ID of the claim it proves (`INV-01`) in its name or body, and the claim is linked with `delta link`. A test that cites no claim is an orphan: link it or delete it.
- If the code would have to contradict a claim, do not bend either one. Stop and tell the person which claim and why.
- If the task cannot be decided without inventing an answer, stop and write the doubt down as an open question of the spec (through `quill`). Do not settle it silently inside the code.
- Never edit a spec, its status or the knowledge base by hand. The commands do it: `delta link`, `delta built`, `delta check`.
- A claim marked `none` or `manual` is not verified. Do not describe it as if it were.
