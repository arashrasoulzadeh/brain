export function validateTicTacToeCommandPayload(payload: unknown): { cell: number } {
  if (
    !payload ||
    typeof payload !== "object" ||
    Object.keys(payload).length !== 1 ||
    typeof (payload as { cell?: unknown }).cell !== "number" ||
    !Number.isInteger((payload as { cell: number }).cell) ||
    (payload as { cell: number }).cell < 0 ||
    (payload as { cell: number }).cell > 8
  )
    throw new Error("command payload mismatch");
  return { cell: (payload as { cell: number }).cell };
}
