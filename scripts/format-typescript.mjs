import { spawn } from "node:child_process";

const command = process.platform === "win32" ? "npx.cmd" : "npx";
const formatter = spawn(command, ["prettier", "--write", "plugins/**/*.ts", "tests/**/*.ts"], {
  stdio: "inherit",
});

formatter.on("exit", (code) => process.exit(code ?? 1));
