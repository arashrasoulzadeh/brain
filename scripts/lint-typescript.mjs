import { readdir, readFile } from "node:fs/promises";
import process from "node:process";

async function collect(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const path = `${directory}/${entry.name}`;
    if (entry.isDirectory()) files.push(...await collect(path));
    else if (entry.name.endsWith(".ts")) files.push(path);
  }
  return files;
}

const sourceFiles = [...await collect("plugins"), ...await collect("tests")];
const violations = [];
for (const file of sourceFiles) {
  const source = await readFile(file, "utf8");
  if (/\bany\b/.test(source)) violations.push(`${file}: explicit any is forbidden`);
  if (/\bconsole\.(log|warn|error|debug)\s*\(/.test(source)) violations.push(`${file}: console logging is forbidden`);
  const lines = source.split("\n");
  let declarationStarted = false;
  lines.forEach((line, index) => {
    if (!line.trim() || line.trim().startsWith("//") || line.trim().startsWith("/*")) return;
    if (/^(export\s+)?(type|interface|const|let|var|class|function|async\s+function)\b/.test(line.trim())) declarationStarted = true;
    if (declarationStarted && /^import\s/.test(line.trim())) violations.push(`${file}:${index + 1}: imports must be at the top of the file`);
  });
}
if (violations.length) {
  process.stderr.write(`${violations.join("\n")}\n`);
  process.exitCode = 1;
} else process.stdout.write(`TypeScript lint passed for ${sourceFiles.length} files\n`);
