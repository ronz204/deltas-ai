// Deterministic grading: compares the verdict block the agent ended its answer
// with against the truth for the case. No model is involved.

import { VERDICTS, type Claim, type Verdict } from "./shared.ts";

export type Reported = { spec: string; n: number; verdict: string };

export type GradedClaim = {
  spec: string;
  n: number;
  ok: Verdict[];
  reported: string | null;
  correct: boolean;
};

export type Graded = {
  formatFailure: boolean;
  claims: GradedClaim[];
  clean: boolean;
  // Claims whose only acceptable verdicts are Violated/Gone, and how many were caught.
  driftClaims: number;
  caught: number;
  // Claims that must not be flagged, and how many were flagged anyway.
  stableClaims: number;
  falseAlarms: number;
  unverified: number;
};

const isDrift = (v: string) => v === "Violated" || v === "Gone";

export function extractVerdicts(text: string): Reported[] | null {
  const blocks = [...text.matchAll(/```json\s*([\s\S]*?)```/g)].map((m) => m[1]!);
  for (const block of blocks.reverse()) {
    try {
      const parsed = JSON.parse(block);
      if (!Array.isArray(parsed?.verdicts)) continue;
      return parsed.verdicts
        .filter((e: any) => e && typeof e.spec === "string" && Number.isInteger(e.n) && typeof e.verdict === "string")
        .map((e: any) => ({ spec: e.spec, n: e.n, verdict: e.verdict }));
    } catch {
      continue;
    }
  }
  return null;
}

const baseName = (path: string) => path.replaceAll("\\", "/").split("/").pop() ?? path;

export function grade(resultText: string, truth: Claim[]): Graded {
  const reported = extractVerdicts(resultText);
  const claims: GradedClaim[] = truth.map((t) => {
    // Later entries win if the agent repeats a claim.
    const hit = [...(reported ?? [])].reverse().find((r) => r.n === t.n && baseName(r.spec) === baseName(t.spec));
    const verdict = hit && (VERDICTS as string[]).includes(hit.verdict) ? hit.verdict : null;
    return { spec: t.spec, n: t.n, ok: t.ok, reported: verdict, correct: verdict !== null && (t.ok as string[]).includes(verdict) };
  });

  const drift = claims.filter((c) => c.ok.every(isDrift));
  const stable = claims.filter((c) => !c.ok.some(isDrift));
  return {
    formatFailure: reported === null,
    claims,
    clean: reported !== null && claims.every((c) => c.correct),
    driftClaims: drift.length,
    caught: drift.filter((c) => c.correct).length,
    stableClaims: stable.length,
    falseAlarms: stable.filter((c) => c.reported !== null && isDrift(c.reported)).length,
    unverified: claims.filter((c) => c.reported === "Unverified").length,
  };
}
