import assert from "node:assert/strict";
import { test } from "node:test";
import { app, mount, find, button, click, change, text, session, stubFetch, json, act } from "./dom-harness.mjs";

const Workers = app("workers/view.tsx").default;
const tick = "2026-01-01T00:00:00Z";
const row = (ID, extra = {}) => ({ ID, Slots: 2, Running: 0, Draining: false, LastPulse: tick,
  Runtimes: [{ Kind: "typescript", Profile: "isolated" }], ...extra });
const legacy = (PluginID, extra = {}) => ({ ProjectID: "dev", PluginID, Version: "1", Enabled: true, Ready: true, WorkerReady: false, ...extra });
const cluster = (Neurons, configured = true) => ({ configured, snapshot: { Neurons, Running: 2, Queued: 3 } });

test("Neuron console shows global capacity, filters all statuses, and excludes foreign legacy workers", async () => {
  session();
  const calls = [];
  stubFetch((url, init) => {
    calls.push({ url, auth: init.headers.authorization, cache: init.cache });
    if (url === "/_system/admin/neurons") return json(cluster([
      row("free"), row("busy", { Running: 2 }),
      row("draining", { Draining: true, Slots: 0, Runtimes: null }),
    ]));
    if (url === "/_system/admin/projects") return json([{ ID: "dev", Active: true }]);
    if (url === "/dev/admin/plugins") return json({ installations: [
      legacy("local", { WorkerReady: true }),
      legacy("retiring", { Draining: true }),
      legacy("foreign", { ProjectID: "other" }),
    ] });
    throw new Error("unexpected request " + url);
  });
  await mount(Workers, {}, { pathname: "/_system/ui/workers" });
  assert.match(text(), /2 occupied \/ 2 total · 0 free/);
  assert.match(text(), /Draining/);
  assert.match(text(), /At capacity/);
  assert.match(text(), /Worker ready/);
  assert.doesNotMatch(text(), /foreign/);
  const search = find('input[placeholder="Identity or runtime"]');
  const filter = find(".neuron-filters select");
  await change(filter, "busy");
  assert.ok(find('progress[aria-label="busy occupied slots"]'));
  assert.equal(document.querySelector('progress[aria-label="free occupied slots"]'), null);
  await change(filter, "draining");
  assert.ok(find('progress[aria-label="draining occupied slots"]'));
  await change(search, "isolated");
  assert.match(text(), /No Neurons match the filters/);
  await change(filter, "available");
  assert.ok(find('progress[aria-label="free occupied slots"]'));
  await change(search, "missing");
  assert.match(text(), /No Neurons match the filters/);
  await change(search, "");
  await change(filter, "all");
  assert.equal(document.querySelectorAll(".neuron-capacity progress").length, 3);
  await click(button("Refresh workers"));
  assert.equal(calls.filter(call => call.url === "/_system/admin/neurons").length, 2);
  assert.ok(calls.every(call => call.auth === "Bearer fixture-token" && call.cache === "no-store"));
});

test("Neuron console distinguishes empty, unconfigured and unavailable status", async () => {
  session();
  let mode = "unconfigured";
  let projectMode = "none";
  stubFetch(url => {
    if (url === "/_system/admin/neurons") {
      if (mode === "denied") return json({}, 403);
      return json(cluster(null, mode !== "unconfigured"));
    }
    if (url === "/_system/admin/projects") {
      if (projectMode === "denied") return json({}, 403);
      return json(projectMode === "none" ? null : [{ ID: "dev", Active: true }]);
    }
    if (url === "/dev/admin/plugins") {
      if (projectMode === "plugins-denied") return json({}, 403);
      return json({ installations: null });
    }
    throw new Error("unexpected request " + url);
  });
  await mount(Workers, {}, { pathname: "/_system/ui/workers" });
  assert.match(text(), /Neuron transport is not connected/);
  assert.match(text(), /Neuron inventory unavailable/);
  assert.match(text(), /Choose project/);
  mode = "empty";
  projectMode = "empty";
  await click(button("Refresh workers"));
  assert.match(text(), /No connected Neurons/);
  assert.match(text(), /No Neurons connected/);
  assert.match(text(), /No installed plugin workers/);
  assert.equal(find(".neuron-metrics article strong").textContent, "0");
  mode = "denied";
  projectMode = "denied";
  await click(button("Refresh workers"));
  assert.match(text(), /Neuron status unavailable/);
  assert.match(text(), /No cached metrics are shown/);
  assert.match(text(), /Projects unavailable/);
  assert.equal(find(".neuron-metrics article strong").textContent, "—");
  projectMode = "plugins-denied";
  await click(button("Refresh workers"));
  assert.match(text(), /Legacy worker status unavailable/);
});

test("Neuron console discards obsolete responses after session change", async () => {
  session();
  let finish;
  stubFetch(url => {
    if (url === "/_system/admin/neurons") return new Promise(resolve => { finish = resolve; });
    if (url === "/_system/admin/projects") return json([]);
    throw new Error("unexpected request " + url);
  });
  await mount(Workers, {}, { pathname: "/_system/ui/workers" });
  assert.match(text(), /Loading execution infrastructure/);
  await act(async () => {
    window.sessionStorage.setItem("brain-admin-session", "replacement");
    finish(json(cluster([row("private-old-session")])));
  });
  assert.doesNotMatch(text(), /private-old-session/);
});

test("Neuron console polls while visible and suspends background requests", async () => {
  session();
  const realTimeout = globalThis.setTimeout;
  const realClear = globalThis.clearTimeout;
  const hidden = Object.getOwnPropertyDescriptor(document, "hidden");
  let next, calls = 0;
  let handle = 0;
  const scheduled = new Map();
  globalThis.setTimeout = (callback, delay, ...args) => {
    if (delay !== 5000) return realTimeout(callback, delay, ...args);
    next = ++handle;
    scheduled.set(next, callback);
    return next;
  };
  globalThis.clearTimeout = id => {
    if (scheduled.has(id)) scheduled.delete(id);
    else realClear(id);
  };
  try {
    Object.defineProperty(document, "hidden", { configurable: true, value: false });
    stubFetch(url => {
      if (url === "/_system/admin/neurons") { calls++; return json(cluster([])); }
      if (url === "/_system/admin/projects") return json([]);
      throw new Error("unexpected request " + url);
    });
    await mount(Workers, {}, { pathname: "/_system/ui/workers" });
    assert.equal(calls, 1);
    const fire = async () => {
      const callback = scheduled.get(next);
      assert.ok(callback);
      scheduled.delete(next);
      await act(async () => callback());
    };
    Object.defineProperty(document, "hidden", { configurable: true, value: true });
    await fire();
    assert.equal(calls, 1);
    Object.defineProperty(document, "hidden", { configurable: true, value: false });
    await fire();
    assert.equal(calls, 2);
    // Replacing the page must cancel the next scheduled poll.
    await mount(() => null);
    assert.equal(scheduled.size, 0);
  } finally {
    globalThis.setTimeout = realTimeout;
    globalThis.clearTimeout = realClear;
    if (hidden) Object.defineProperty(document, "hidden", hidden);
    else delete document.hidden;
  }
});

test("Neuron console shows real outcomes and distinguishes unreported resource usage", async () => {
  session();
  stubFetch(url => {
    if (url === "/_system/admin/neurons") return json({ configured: true, snapshot: { Neurons: [], Running: 0, Queued: 0, Executions: { Assigned: 3, Succeeded: 1, Failed: 1, Cancelled: 0, Lost: 1 }, Recent: [{ ExecutionID: "exec-one", TraceID: "trace-one", ParentID: "", NeuronID: "worker-one", Runtime: { Kind: "typescript", Profile: "isolated" }, Outcome: "succeeded", FinishedAt: tick, WaitMillis: 2, DurationMillis: 14, CPUTimeNanos: 1000000, PeakMemoryBytes: 1048576 }] } });
    if (url === "/_system/admin/projects") return json([]);
    throw new Error("unexpected request " + url);
  });
  await mount(Workers, {}, { pathname: "/_system/ui/workers" });
  assert.match(text(), /exec-one/);
  assert.match(text(), /trace-one/);
  assert.match(text(), /2 \/ 14 ms/);
  assert.match(text(), /1.0 ms \/ 1.0 MiB/);
});
