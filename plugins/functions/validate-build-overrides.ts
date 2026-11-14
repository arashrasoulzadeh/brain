export function validateBuildOverrides(overrides: Readonly<Record<string, string>>): void {
  const denied = new Set(["GOWORK", "GOTOOLCHAIN", "GOENV", "GOMOD", "GOPROXY", "GOFLAGS"]);
  if (
    Object.keys(overrides).some(
      (key) =>
        denied.has(key) ||
        key.toLowerCase().includes("workspace") ||
        key.toLowerCase().includes("toolchain"),
    )
  )
    throw new Error("build override denied");
}
