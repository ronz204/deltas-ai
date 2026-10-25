import { describe, expect, test } from "bun:test";
import { extractVerdicts, grade } from "./grade.ts";
import type { Claim } from "./shared.ts";

const spec = "deltas/pricing.spec.md";
const truth: Claim[] = [
  { spec, n: 1, ok: ["Holds"] },
  { spec, n: 2, ok: ["Violated"] },
  { spec, n: 3, ok: ["Holds", "Unverified"] },
];
const block = (verdicts: unknown) => "Report.\n\n```json\n" + JSON.stringify({ verdicts }) + "\n```\n";

describe("extractVerdicts", () => {
  test("reads the last json block", () => {
    const text = block([{ spec, n: 1, verdict: "Violated" }]) + block([{ spec, n: 1, verdict: "Holds" }]);
    expect(extractVerdicts(text)).toEqual([{ spec, n: 1, verdict: "Holds" }]);
  });

  test("returns null when there is no usable block", () => {
    expect(extractVerdicts("no block here")).toBeNull();
    expect(extractVerdicts("```json\n{ broken\n```")).toBeNull();
    expect(extractVerdicts('```json\n{"other":1}\n```')).toBeNull();
  });
});

describe("grade", () => {
  test("a fully correct answer is a clean run that catches the drift", () => {
    const g = grade(
      block([
        { spec, n: 1, verdict: "Holds" },
        { spec, n: 2, verdict: "Violated" },
        { spec, n: 3, verdict: "Unverified" },
      ]),
      truth,
    );
    expect(g.clean).toBe(true);
    expect(g.caught).toBe(1);
    expect(g.driftClaims).toBe(1);
    expect(g.falseAlarms).toBe(0);
    expect(g.stableClaims).toBe(2);
    expect(g.unverified).toBe(1);
  });

  test("missing the drift and crying wolf are both counted", () => {
    const g = grade(
      block([
        { spec, n: 1, verdict: "Violated" },
        { spec, n: 2, verdict: "Holds" },
        { spec, n: 3, verdict: "Holds" },
      ]),
      truth,
    );
    expect(g.clean).toBe(false);
    expect(g.caught).toBe(0);
    expect(g.falseAlarms).toBe(1);
  });

  test("Unverified on the drifted claim is a miss", () => {
    const g = grade(block([{ spec, n: 2, verdict: "Unverified" }]), truth);
    expect(g.caught).toBe(0);
  });

  test("an omitted claim is incorrect and an unknown verdict is ignored", () => {
    const g = grade(block([{ spec, n: 1, verdict: "Holds" }, { spec, n: 2, verdict: "Maybe" }]), truth);
    expect(g.claims.find((c) => c.n === 2)!.reported).toBeNull();
    expect(g.claims.find((c) => c.n === 3)!.correct).toBe(false);
    expect(g.clean).toBe(false);
  });

  test("a missing block is a format failure", () => {
    const g = grade("I checked everything and it looks fine.", truth);
    expect(g.formatFailure).toBe(true);
    expect(g.clean).toBe(false);
  });

  test("the spec path may be written differently", () => {
    const g = grade(block([{ spec: "./deltas\\pricing.spec.md", n: 1, verdict: "Holds" }]), truth);
    expect(g.claims.find((c) => c.n === 1)!.correct).toBe(true);
  });

  test("Gone counts as drift for a retired slice", () => {
    const gone: Claim[] = [{ spec, n: 1, ok: ["Gone"] }];
    expect(grade(block([{ spec, n: 1, verdict: "Gone" }]), gone).caught).toBe(1);
    expect(grade(block([{ spec, n: 1, verdict: "Violated" }]), gone).caught).toBe(0);
  });
});
