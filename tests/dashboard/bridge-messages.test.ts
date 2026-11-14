import { validateBridgeMessage } from "../../plugins/dashboard/validate-bridge-message.js";

const context = { projectId: "dev", invocationId: "frame-1" } as const;

const valid = validateBridgeMessage(
  {
    protocol: 1,
    projectId: "dev",
    invocationId: "frame-1",
    operation: "record.list",
    payload: { table: "users" },
  },
  context,
  ["record.list"],
);

if (valid.operation !== "record.list" || valid.projectId !== "dev") {
  throw new Error("valid bridge message was not accepted");
}

for (const message of [
  null,
  { protocol: 1, projectId: "dev", invocationId: "frame-1" },
  {
    protocol: 1,
    projectId: "dev",
    invocationId: "frame-1",
    operation: "record.list",
    unexpected: true,
  },
  { protocol: 1, projectId: "prod", invocationId: "frame-1", operation: "record.list" },
  { protocol: 1, projectId: "dev", invocationId: "other", operation: "record.list" },
  { protocol: 1, projectId: "dev", invocationId: "frame-1", operation: "plugin.grant" },
  { protocol: 2, projectId: "dev", invocationId: "frame-1", operation: "record.list" },
]) {
  try {
    validateBridgeMessage(message, context, ["record.list"]);
    throw new Error("invalid bridge message was accepted");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "invalid bridge message") throw error;
  }
}

try {
  validateBridgeMessage(
    {
      protocol: 1,
      projectId: "dev",
      invocationId: "frame-1",
      operation: "record.list",
      payload: { operatorToken: "secret" },
    },
    context,
    ["record.list"],
  );
  throw new Error("operator token crossed the bridge");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid bridge message") throw error;
}

try {
  validateBridgeMessage(
    {
      protocol: 1,
      projectId: "dev",
      invocationId: "frame-1",
      operation: "record.list",
      payload: { destructiveConfirmation: true },
    },
    context,
    ["record.list"],
  );
  throw new Error("sensitive confirmation crossed the bridge");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid bridge message") throw error;
}
