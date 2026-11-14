export const TIC_TAC_TOE_RULES_SOURCE = "package main\n// Brain Tic-tac-toe rules\n";
export function publishTicTacToeRules(compile: (source: string) => string): string { const artifact = compile(TIC_TAC_TOE_RULES_SOURCE); if (!artifact) throw new Error("rules compilation failed"); return artifact; }
