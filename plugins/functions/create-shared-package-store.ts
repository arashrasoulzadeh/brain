import { createSharedPackageState } from "./shared-package-state.js";
import { SharedPackageWriter } from "./shared-package-writer.js";
import { SharedPackageReader } from "./shared-package-reader.js";
export function createSharedPackageStore() { const state = createSharedPackageState(); const writer = new SharedPackageWriter(state); const reader = new SharedPackageReader(state); return { save: writer.save.bind(writer), load: reader.load.bind(reader) }; }
