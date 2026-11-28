import assert from "node:assert/strict";
import { test } from "node:test";
import { availableSlots, filterNeurons, legacyWorkerState, summarizeCluster, type Neuron } from "../../admin-nextjs/app/lib/neuron-dashboard.js";

const neuron: Neuron = { ID: "neuron-1", Slots: 4, Running: 1, Draining: false, LastPulse: "2026-11-28T00:00:00Z", Runtimes: [{ Kind: "typescript", Profile: "isolated" }] };

test("Cortex metrics distinguish unavailable telemetry from a connected empty pool", () => {
  assert.equal(summarizeCluster(undefined), undefined);
  assert.equal(summarizeCluster({ configured: false, snapshot: { Neurons: null, Running: 0, Queued: 0 } }), undefined);
  assert.deepEqual(summarizeCluster({ configured: true, snapshot: { Neurons: null, Running: 0, Queued: 0 } }), { connected: 0, slots: 0, available: 0, running: 0, queued: 0 });
  assert.deepEqual(summarizeCluster({ configured: true, snapshot: { Neurons: [neuron, { ...neuron, ID: "draining", Draining: true }], Running: 2, Queued: 3 } }), { connected: 2, slots: 8, available: 3, running: 2, queued: 3 });
  assert.equal(availableSlots({ ...neuron, Running: 5 }), 0);
});

test("Neuron filters use real identity, runtime and capacity", () => {
  const busy = { ...neuron, ID: "busy", Running: 4 };
  const draining = { ...neuron, ID: "drain", Draining: true, Runtimes: null };
  const rows = [neuron, busy, draining];
  assert.deepEqual(filterNeurons(rows, "  TYPESCRIPT ", "all"), [neuron, busy]);
  assert.deepEqual(filterNeurons(rows, "", "available"), [neuron]);
  assert.deepEqual(filterNeurons(rows, "", "busy"), [busy]);
  assert.deepEqual(filterNeurons(rows, "", "draining"), [draining]);
  assert.deepEqual(filterNeurons(rows, "drain", "all"), [draining]);
  assert.deepEqual(filterNeurons(rows, "unknown", "all"), []);
});

test("Legacy lifecycle state never claims a generic Neuron connection", () => {
  const worker = { ProjectID: "a", PluginID: "files", Version: "1", Enabled: true, Ready: true, WorkerReady: true };
  assert.equal(legacyWorkerState(worker), "Worker ready");
  assert.equal(legacyWorkerState({ ...worker, Draining: true }), "Draining");
  assert.equal(legacyWorkerState({ ...worker, Enabled: false }), "Disabled");
  assert.equal(legacyWorkerState({ ...worker, Ready: false }), "Artifact not ready");
  assert.equal(legacyWorkerState({ ...worker, WorkerReady: false }), "Worker not ready");
});
