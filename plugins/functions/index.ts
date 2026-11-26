export {
  functionRequest, type FunctionInvocationResult, type HttpsFetcher,
  validateInvocationResult, parseInvocationResult, executeFixture, invokeDirectHttps,
  createInvocationRequest, loadDeclaredInputs, type FunctionWriteIntent,
  validateWriteIntents, createInvocationEffects, createInvocationCommitStore,
  retryInvocation,
} from "./invocation.js";
export {
  publishSharedHelperArtifact, validateStaleInputVersions, createApprovalStore,
  createFunctionDraftStore, createFunctionPublicationStore, createSharedPackageStore,
} from "./publication.js";

export const ENTRY_PACKAGE_FIXTURE =
  'package main\n\nimport "encoding/json"\n\nfunc main() {\n\t_ = json.NewEncoder(nil)\n}\n';
export const PROJECT_HELPER_FIXTURE = "package helper";
export {
  wasmBuildSettings, APPROVED_SDK_IMPORTS, validateSDKImports, requireSDKImports,
  validateBuildOverrides, validateFunctionLimits, DisposableCompilerWorkspace,
  compilerEnvironment, validateCompilerEnvironment, type CompilerLimits,
  validateCompilerLimits, capDiagnostics, APPROVED_WASI_IMPORTS, validateWasiImports,
  validateLinearMemory, validateStartupDuration, type SourceDiagnostic, diagnoseSource,
} from "./compiler.js";
export {
  validateSourceProject, validateSandboxCapabilities, rejectSandboxEscapeAttempt,
  authorizeDeploymentPlugin, type DeploymentHook, type DeploymentHookSet,
  createDeploymentHookSet, rejectPluginCredentials, rejectPluginHostResources,
  validateDeploymentWorkflow,
} from "./sandbox.js";

export type FunctionSchedule = {
  projectId: string;
  functionId: string;
  expression: string;
  dueAt: number;
};
export type EnqueuedFunction = {
  projectId: string;
  functionId: string;
  functionVersion: number;
  enqueuedAt: number;
};
export {
  createScheduleStore, createRecordChangeTriggerStore, createJobLeaseStore,
} from "./scheduling.js";
export type RecordChangeTrigger = {
  projectId: string;
  collection: string;
  operation: string;
  target: { functionId: string; version: number };
};
export { commitWithGrant, runCappedRetries, runWithTimeout } from "../sdk-typescript/src/index.js";
export {
  runIdempotentWorkflow, redactRunStatus, type WorkflowNode, validateWorkflowGraph,
  type WorkflowCondition, evaluateCondition, registerProjectWorkflowAction,
  type ProjectWorkflowAction, replaceProjectWorkflowStep,
  authorizeProjectWorkflowAction,
} from "./workflow.js";
export type FunctionDraft = {
  projectId: string;
  functionId: string;
  source: string;
  savedAt: number;
};
export type FunctionPublication = FunctionDraft & {
  version: number;
  artifactDigest: string;
  toolchain: string;
  publishedAt: number;
};
export { invoke, publish, call, dispatchRoute } from "./routes.js";
