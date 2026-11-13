export function validateTicTacToeRuleResult(value: unknown): { board: readonly ("X" | "O" | null)[]; turn: string; status: "active" | "paused" | "finished" } {
  if (!value || typeof value !== "object" || Object.keys(value).some((key) => !["board", "turn", "status"].includes(key))) throw new Error("malformed rule result");
  const result = value as { board?: unknown; turn?: unknown; status?: unknown };
  if (!Array.isArray(result.board) || result.board.length !== 9 || result.board.some((cell) => cell !== null && cell !== "X" && cell !== "O") || typeof result.turn !== "string" || !["active", "paused", "finished"].includes(String(result.status))) throw new Error("malformed rule result");
  return { board: [...result.board] as ("X" | "O" | null)[], turn: result.turn, status: result.status as "active" | "finished" };
}
