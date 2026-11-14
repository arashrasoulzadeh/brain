import { hash as argon2Hash } from "@node-rs/argon2";
export function passwordHash(password: string): Promise<string> {
  return argon2Hash(password, { memoryCost: 19456, timeCost: 2, parallelism: 1, outputLen: 32 });
}
