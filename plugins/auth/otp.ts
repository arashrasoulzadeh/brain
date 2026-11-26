import { createDigestToken, consumeDigestToken, numericOTP, randomToken, type DigestToken } from "../sdk-typescript/src/tokens.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export type OTPChallenge = DigestToken;
export type RecoveryToken = OTPChallenge;
type Inbox = { deliver(destination: string, body: string): void };

function deliverToken(
  records: Map<string, OTPChallenge>, secret: string, inbox: Inbox,
  key: string, destination: string, token: string, now: number, ttlSeconds: number, error: string,
): void {
  records.set(key, createDigestToken(secret, token, now, ttlSeconds * 1000, error));
  inbox.deliver(destination, token);
}

function consumeToken(
  records: ReadonlyMap<string, OTPChallenge>, secret: string, key: string,
  token: string, now: number,
): boolean {
  return consumeDigestToken(records.get(key), secret, token, now);
}

class OTPIssuer {
  constructor(
    private readonly challenges: Map<string, OTPChallenge>,
    private readonly secret: string,
    private readonly inbox: Inbox,
    private readonly validDestination: (destination: string) => boolean,
    private readonly error: string,
  ) {}
  issue(projectId: string, destination: string, purpose: string, now: number, ttlSeconds: number): void {
    validateProjectId(projectId);
    if (!this.validDestination(destination) || !purpose || ttlSeconds <= 0) throw new Error(this.error);
    const code = numericOTP();
    deliverToken(this.challenges, this.secret, this.inbox, `${projectId}:${destination}:${purpose}`, destination, code, now, ttlSeconds, this.error);
  }
}

export class EmailOTPIssuer extends OTPIssuer {
  constructor(challenges: Map<string, OTPChallenge>, secret: string, inbox: Inbox) {
    super(challenges, secret, inbox, (destination) => destination.includes("@"), "invalid email OTP");
  }
}

export class SMSOTPIssuer extends OTPIssuer {
  constructor(challenges: Map<string, OTPChallenge>, secret: string, inbox: Inbox) {
    super(challenges, secret, inbox, (destination) => /^\+[1-9][0-9]{7,14}$/.test(destination), "invalid SMS OTP");
  }
}

class OTPVerifier {
  constructor(private readonly challenges: ReadonlyMap<string, OTPChallenge>, private readonly secret: string) {}
  verify(projectId: string, destination: string, purpose: string, code: string, now: number): boolean {
    validateProjectId(projectId);
    return consumeToken(this.challenges, this.secret, `${projectId}:${destination}:${purpose}`, code, now);
  }
}

export class EmailOTPVerifier extends OTPVerifier {}
export class SMSOTPVerifier extends OTPVerifier {}

export class RecoveryTokenIssuer {
  constructor(private readonly tokens: Map<string, RecoveryToken>, private readonly secret: string, private readonly inbox: Inbox) {}
  issue(projectId: string, destination: string, now: number, ttlSeconds: number): void {
    validateProjectId(projectId);
    if (!destination.includes("@") || ttlSeconds <= 0) throw new Error("invalid recovery request");
    deliverToken(this.tokens, this.secret, this.inbox, `${projectId}:${destination}`, destination, randomToken(), now, ttlSeconds, "invalid recovery request");
  }
}

export class RecoveryTokenConsumer {
  constructor(private readonly tokens: ReadonlyMap<string, RecoveryToken>, private readonly secret: string) {}
  consume(projectId: string, destination: string, token: string, now: number): boolean {
    validateProjectId(projectId);
    return consumeToken(this.tokens, this.secret, `${projectId}:${destination}`, token, now);
  }
}

export function normalizePhoneNumber(value: string): string {
  const compact = value.replace(/[\s().-]/g, "");
  const normalized = compact.startsWith("00") ? `+${compact.slice(2)}` : compact;
  if (!/^\+[1-9][0-9]{7,14}$/.test(normalized)) throw new Error("invalid phone number");
  return normalized;
}
