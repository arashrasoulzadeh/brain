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
