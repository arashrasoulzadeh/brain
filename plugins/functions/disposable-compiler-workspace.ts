import { existsSync, mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
export class DisposableCompilerWorkspace {
  readonly path: string;
  constructor(prefix = "brain-function-") {
    this.path = mkdtempSync(join(tmpdir(), prefix));
  }
  dispose(): void {
    if (existsSync(this.path)) rmSync(this.path, { recursive: true, force: true });
  }
}
