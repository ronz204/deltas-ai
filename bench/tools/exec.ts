// Launches the agent headless in a working copy and extracts the metrics from
// its stream-json output.

export type Usage = { input: number; cacheCreation: number; cacheRead: number; output: number };

export type Execution = {
  exitCode: number;
  timedOut: boolean;
  stderr: string;
  raw: string;
  isError: boolean;
  resultText: string;
  turns: number;
  durationMs: number;
  usage: Usage;
  subagents: number;
  permissionDenials: number;
  toolCalls: number;
  skillInvoked: boolean;
  costUsd: number;
};

type Json = Record<string, any>;

export function parseStream(raw: string): Omit<Execution, "exitCode" | "timedOut" | "stderr" | "raw"> {
  let result: Json | null = null;
  let toolCalls = 0;
  let skillInvoked = false;
  for (const line of raw.split("\n")) {
    if (!line.trim()) continue;
    let event: Json;
    try {
      event = JSON.parse(line);
    } catch {
      continue;
    }
    if (event.type === "result") result = event;
    if (event.type === "assistant") {
      for (const part of event.message?.content ?? []) {
        if (part.type !== "tool_use") continue;
        toolCalls++;
        if (part.name === "Skill" && part.input?.skill === "sentinel") skillInvoked = true;
      }
    }
  }
  const u = result?.usage ?? {};
  return {
    isError: result === null || result.is_error === true,
    resultText: typeof result?.result === "string" ? result.result : "",
    turns: result?.num_turns ?? 0,
    durationMs: result?.duration_ms ?? 0,
    usage: {
      input: u.input_tokens ?? 0,
      cacheCreation: u.cache_creation_input_tokens ?? 0,
      cacheRead: u.cache_read_input_tokens ?? 0,
      output: u.output_tokens ?? 0,
    },
    subagents: result?.subagent_stats?.spawned ?? 0,
    permissionDenials: Array.isArray(result?.permission_denials) ? result.permission_denials.length : 0,
    toolCalls,
    skillInvoked,
    costUsd: result?.total_cost_usd ?? 0,
  };
}

export async function execute(
  dir: string,
  prompt: string,
  tools: string,
  model: string,
  timeoutMs = 20 * 60 * 1000,
): Promise<Execution> {
  // An API key in the environment would bill the API instead of the subscription.
  const env = { ...process.env };
  delete env.ANTHROPIC_API_KEY;

  // The prompt goes through stdin so no shell or argument quoting can alter it.
  const proc = Bun.spawn(
    [
      "claude", "-p",
      "--model", model,
      "--output-format", "stream-json", "--verbose",
      "--no-session-persistence",
      "--setting-sources", "project",
      "--strict-mcp-config",
      "--permission-mode", "dontAsk",
      "--allowedTools", tools,
    ],
    { cwd: dir, env, stdin: new Blob([prompt]), stdout: "pipe", stderr: "pipe", timeout: timeoutMs },
  );
  const [raw, stderr, exitCode] = await Promise.all([
    new Response(proc.stdout).text(),
    new Response(proc.stderr).text(),
    proc.exited,
  ]);
  const timedOut = proc.signalCode !== null;
  return { exitCode, timedOut, stderr, raw, ...parseStream(raw) };
}
