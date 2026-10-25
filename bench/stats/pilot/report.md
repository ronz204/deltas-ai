# Sentinel bench report — pilot

| | |
|---|---|
| Model | `claude-sonnet-5-5` |
| Claude Code | 2.1.300 (Claude Code) |
| Bun | 1.3.14 |
| Harness commit | `6e94472` |
| Started | 2026-10-25T16:10:21.622Z |
| Repetitions per case and mode | 3 |
| Hold-out cases | excluded (tuning batch) |
| Runs | 42 |

Reading guide: counts are shown next to percentages because there are few runs; a gap of under about 15 points is noise. *Catch rate* is drift claims reported correctly out of drift claims; *False alarms* is stable claims flagged `Violated`/`Gone` out of stable claims. Tokens are the sum of input, cache creation, cache read and output; dollar cost is not reported because the runs use a subscription.

## By mode

| Mode | Runs | Clean runs | Catch rate | False alarms | Claim accuracy | Unverified | Format failures | Skill invoked | Median output tok | Median total tok | Median turns | Median s | Median subagents |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| plain | 21 | 21/21 (100%) | 27/27 (100%) | 0/87 (0%) | 114/114 (100%) | 0/114 (0%) | 0/21 (0%) | 0/21 (0%) | 1.3k | 61.7k | 7 | 12 | 0 |
| skill | 21 | 21/21 (100%) | 27/27 (100%) | 0/87 (0%) | 114/114 (100%) | 0/114 (0%) | 0/21 (0%) | 21/21 (100%) | 1.9k | 121.3k | 11 | 18 | 0 |

## By category and mode

| Category / mode | Runs | Clean runs | Catch rate | False alarms | Claim accuracy | Unverified | Format failures | Skill invoked | Median output tok | Median total tok | Median turns | Median s | Median subagents |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| beht / plain | 3 | 3/3 (100%) | 3/3 (100%) | 0/15 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 0/3 (0%) | 1.4k | 61.8k | 8 | 11 | 0 |
| beht / skill | 3 | 3/3 (100%) | 3/3 (100%) | 0/15 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 3/3 (100%) | 2.1k | 121.9k | 13 | 17 | 0 |
| beni / plain | 3 | 3/3 (100%) | n/a | 0/18 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 0/3 (0%) | 1.4k | 61.7k | 8 | 10 | 0 |
| beni / skill | 3 | 3/3 (100%) | n/a | 0/18 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 3/3 (100%) | 1.9k | 121.6k | 12 | 16 | 0 |
| bent / plain | 3 | 3/3 (100%) | 3/3 (100%) | 0/12 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 0/3 (0%) | 1.3k | 82.5k | 7 | 13 | 0 |
| bent / skill | 3 | 3/3 (100%) | 3/3 (100%) | 0/12 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 3/3 (100%) | 1.9k | 119.6k | 10 | 24 | 0 |
| ctrl / plain | 3 | 3/3 (100%) | n/a | 0/15 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 0/3 (0%) | 1.1k | 60.3k | 5 | 11 | 0 |
| ctrl / skill | 3 | 3/3 (100%) | n/a | 0/15 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 3/3 (100%) | 1.4k | 93.4k | 7 | 17 | 0 |
| gone / plain | 3 | 3/3 (100%) | 15/15 (100%) | n/a | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 0/3 (0%) | 1.1k | 103.4k | 6 | 16 | 0 |
| gone / skill | 3 | 3/3 (100%) | 15/15 (100%) | n/a | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 3/3 (100%) | 1.6k | 142.7k | 9 | 18 | 0 |
| pres / plain | 3 | 3/3 (100%) | 3/3 (100%) | 0/12 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 0/3 (0%) | 1.5k | 61.4k | 7 | 13 | 0 |
| pres / skill | 3 | 3/3 (100%) | 3/3 (100%) | 0/12 (0%) | 15/15 (100%) | 0/15 (0%) | 0/3 (0%) | 3/3 (100%) | 1.9k | 120.6k | 10 | 19 | 0 |
| spec / plain | 3 | 3/3 (100%) | 3/3 (100%) | 0/15 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 0/3 (0%) | 1.3k | 61.7k | 8 | 10 | 0 |
| spec / skill | 3 | 3/3 (100%) | 3/3 (100%) | 0/15 (0%) | 18/18 (100%) | 0/18 (0%) | 0/3 (0%) | 3/3 (100%) | 2.1k | 122.2k | 13 | 21 | 0 |

## Clean runs by case

| Case | plain | skill |
|---|---|---|
| 01-ctrl-checkout | 3/3 | 3/3 |
| 03-pres-stockevent | 3/3 | 3/3 |
| 06-beht-rounding | 3/3 | 3/3 |
| 08-bent-ttl | 3/3 | 3/3 |
| 09-gone-stock | 3/3 | 3/3 |
| 10-beni-rename | 3/3 | 3/3 |
| 12-spec-rounding | 3/3 | 3/3 |
