export type NeuronRuntime = { Kind: string; Profile: string };
export type Neuron = { ID: string; Slots: number; Running: number; Draining: boolean; OperatorDraining?: boolean; LastPulse: string; Runtimes: NeuronRuntime[] | null };
export type ExecutionTotals = { Assigned: number; Succeeded: number; Failed: number; Cancelled: number; Lost: number };
export type ExecutionTrace = { ExecutionID: string; TraceID: string; ParentID: string; NeuronID: string; Runtime: NeuronRuntime; Outcome: string; FinishedAt: string; WaitMillis: number; DurationMillis: number; CPUTimeNanos: number; PeakMemoryBytes: number };
export type ManagedNeuron = { ID: string; Running: boolean };
export type NeuronAction = "drain" | "resume" | "start" | "stop" | "restart";
export type ClusterStatus = { configured: boolean; managed?: ManagedNeuron[]; snapshot: { Neurons: Neuron[] | null; Running: number; Queued: number; Executions?: ExecutionTotals; Recent?: ExecutionTrace[] | null; HistoryPersistent?: boolean; HistoryDegraded?: boolean } };
export type Installation = { ProjectID: string; PluginID: string; Version: string; Enabled: boolean; Ready: boolean; WorkerReady?: boolean; Draining?: boolean };

export function availableSlots(neuron: Neuron): number {
  return neuron.Draining ? 0 : Math.max(0, neuron.Slots - neuron.Running);
}

export function summarizeCluster(status: ClusterStatus | undefined) {
  if (!status?.configured) return undefined;
  const neurons = status.snapshot.Neurons ?? [];
  return {
    connected: neurons.length,
    slots: neurons.reduce((sum, neuron) => sum + neuron.Slots, 0),
    available: neurons.reduce((sum, neuron) => sum + availableSlots(neuron), 0),
    running: status.snapshot.Running,
    queued: status.snapshot.Queued,
  };
}

export function filterNeurons(neurons: readonly Neuron[], search: string, state: string) {
  const query = search.trim().toLowerCase();
  return neurons.filter(neuron =>
    (state === "all" || (state === "draining" && neuron.Draining) || (state === "available" && availableSlots(neuron) > 0) || (state === "busy" && !neuron.Draining && availableSlots(neuron) === 0)) &&
    (!query || [neuron.ID, ...(neuron.Runtimes ?? []).flatMap(runtime => [runtime.Kind, runtime.Profile])].join(" ").toLowerCase().includes(query)));
}

export function legacyWorkerState(worker: Installation): string {
  if (worker.Draining) return "Draining";
  if (!worker.Enabled) return "Disabled";
  if (!worker.Ready) return "Artifact not ready";
  return worker.WorkerReady ? "Worker ready" : "Worker not ready";
}
