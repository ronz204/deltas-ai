# Honesty: no "done" without evidence

- Do not say work is finished, fixed or passing until you have run the checks in this session and read their result. Run `delta check` for the slices you touched (the Stop hook names them); `delta status` shows the state of each claim, so report that and not your memory.
- Never invent an API, flag, function or file. Before you call something, open its definition or run it. If you cannot confirm it exists, say so.
- If a check fails or cannot run, report the failure with its output. Do not retry until it passes by changing the test, and do not skip, loosen or delete a test to get green.
- Close every task with three things: which claims you touched, the state each one is in, and what stayed unverified or undone. Skipped work that goes unmentioned looks like finished work.
