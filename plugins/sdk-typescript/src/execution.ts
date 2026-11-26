export async function runCappedRetries<T>(
  task: () => T | Promise<T>,
  maxAttempts: number,
): Promise<T> {
  if (!Number.isSafeInteger(maxAttempts) || maxAttempts < 1) throw new Error("invalid retry cap");
  let lastError: unknown;
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      return await task();
    } catch (error) {
      lastError = error;
    }
  }
  throw lastError instanceof Error ? lastError : new Error("retry attempts exhausted");
}

export async function runWithTimeout<T>(task: Promise<T>, timeoutMs: number): Promise<T> {
  if (!Number.isSafeInteger(timeoutMs) || timeoutMs < 1)
    throw new Error("invalid invocation timeout");
  let timer: ReturnType<typeof setTimeout>;
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error("function invocation timed out")), timeoutMs);
  });
  try {
    return await Promise.race([task, timeout]);
  } finally {
    clearTimeout(timer!);
  }
}

export function commitWithGrant<T>(
  hasGrant: () => boolean,
  commit: () => T,
  capability: string,
): T {
  if (!capability || !hasGrant()) throw new Error("grant revoked before commit");
  return commit();
}
