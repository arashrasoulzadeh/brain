import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve, relative, sep } from "node:path";
import ts from "typescript";
import { pathToFileURL } from "node:url";

const roots = ["plugins", "clients/javascript/src", "admin-nextjs/app", "scripts"];
const source = /\.(?:ts|tsx|mjs)$/;

export function hasExecutableSource(path, content) {
  if (/\.d\.(?:ts|mts)$/.test(path)) return false;
  if (path.endsWith(".mjs")) return true;
  const compiled = ts.transpileModule(content, {
    fileName: path,
    compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX, removeComments: true, verbatimModuleSyntax: true },
    reportDiagnostics: true,
  });
  const errors = compiled.diagnostics?.filter(item => item.category === ts.DiagnosticCategory.Error);
  if (errors?.length) throw new Error("Cannot classify source: " + path + " " + errors.map(item => ts.flattenDiagnosticMessageText(item.messageText, "\n")).join("; "));
  const tree = ts.createSourceFile(path + ".js", compiled.outputText, ts.ScriptTarget.ESNext, true, ts.ScriptKind.JS);
  // A TypeScript-only module can emit an empty ES module marker. It has no
  // executable source statements. Preserve imports, re-exports, enums, code,
  // directives and every other emitted statement in the coverage scope.
  return tree.statements.some(statement => !(
    ts.isExportDeclaration(statement) && !statement.moduleSpecifier &&
    statement.exportClause && ts.isNamedExports(statement.exportClause) &&
    statement.exportClause.elements.length === 0
  ));
}

function collect(directory) {
  const files = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = resolve(directory, entry.name);
    if (entry.isDirectory()) files.push(...collect(path));
    else if (entry.isFile() && source.test(path)) files.push(path);
  }
  return files;
}

export function inventory(root = process.cwd()) {
  const runtime = [], erased = [];
  for (const directory of roots) {
    for (const path of collect(resolve(root, directory))) {
      const name = relative(root, path).split(sep).join("/");
      (hasExecutableSource(name, readFileSync(path, "utf8")) ? runtime : erased).push(name);
    }
  }
  return { runtime: runtime.sort(), erased: erased.sort() };
}

export function validateReport(report, expected, root = process.cwd()) {
  const actual = new Set(Object.keys(report).filter(path => path !== "total").map(path => relative(root, resolve(root, path)).split(sep).join("/")));
  const missing = expected.runtime.filter(path => !actual.has(path));
  if (missing.length) throw new Error("Executable sources missing from coverage: " + missing.join(", "));
  const unexpected = [...actual].filter(path => !expected.runtime.includes(path));
  if (unexpected.length) throw new Error("Unexpected coverage sources: " + unexpected.join(", "));
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const [mode, path] = process.argv.slice(2);
  if (mode === "inventory") {
    const data = inventory();
    writeFileSync(path, JSON.stringify(data, null, 2) + "\n");
    // Literal minimatch escapes are required for bracketed Next route paths.
    for (const name of data.erased) process.stdout.write(name.replace(/([*?{}[\]()!+@])/g, "[$1]") + "\n");
    process.stderr.write("Runtime source files: " + data.runtime.length + "; compiler-erased files: " + data.erased.length + "\n");
  } else if (mode === "verify") {
    validateReport(JSON.parse(readFileSync(path, "utf8")), inventory());
  } else throw new Error("usage: source-inventory.mjs inventory|verify <file>");
}
