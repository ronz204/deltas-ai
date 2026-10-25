// Builds the sealed working copy for one run: repo + mutation + mode overlay,
// committed as a single commit so the mutation cannot be found in git history.

import { BENCH, ROOT, copyTree, hasFiles, run, tempDir, type Case, type Mode } from "./shared.ts";

export async function prepare(c: Case, mode: Mode): Promise<string> {
  const dir = tempDir("bench-run");
  await copyTree(`${BENCH}/repos/${c.repo}`, dir);

  if (c.patch) {
    const applied = run(["git", "apply", "--whitespace=nowarn", `${BENCH}/cases/${c.id}/${c.patch}`], dir);
    if (applied.code !== 0) throw new Error(`patch for ${c.id} does not apply: ${applied.out.trim()}`);
  }

  // Files the mode adds on top of the repo (optional folder).
  if (await hasFiles(`${BENCH}/modes/${mode.id}/files`)) {
    await copyTree(`${BENCH}/modes/${mode.id}/files`, dir);
  }
  // Skills are always copied from the current harness, so a run tests the skill as it is today.
  for (const skill of mode.skills) {
    await copyTree(`${ROOT}/.claude/skills/${skill}`, `${dir}/.claude/skills/${skill}`);
  }

  const git = (...args: string[]) => {
    const r = run(["git", "-c", "core.autocrlf=false", "-c", "core.safecrlf=false", ...args], dir);
    if (r.code !== 0) throw new Error(`git ${args.join(" ")}: ${r.out.trim()}`);
  };
  git("init", "-q");
  git("add", "-A");
  git("-c", "user.email=bench@local", "-c", "user.name=bench", "commit", "-q", "-m", "snapshot");
  return dir;
}
