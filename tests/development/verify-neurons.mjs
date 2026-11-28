// Runs inside the isolated packaged development host, using its real API.
import assert from "node:assert/strict";

const origin = "http://127.0.0.1:8080";
const deadline = Date.now() + 45000;
let token;
while (Date.now() < deadline) {
  try {
    const response = await fetch(origin + "/_system/admin/session", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ username: "ci-operator", secret: "ci-fixture-only" }),
      signal: AbortSignal.timeout(1000),
    });
    if (response.ok) {
      token = (await response.json()).session;
      break;
    }
  } catch {
    // Brain starts listening only after its sandboxed Neurons register.
  }
  await new Promise(resolve => setTimeout(resolve, 200));
}
assert.ok(token, "development Brain did not become available");
const request = async () => {
  const response = await fetch(origin + "/_system/admin/neurons", {
    headers: { authorization: "Bearer " + token },
    signal: AbortSignal.timeout(3000),
  });
  assert.equal(response.status, 200);
  return response.json();
};
const first = await request();
assert.equal(first.configured, true);
assert.equal(first.snapshot.Running, 0);
assert.equal(first.snapshot.Queued, 0);
assert.deepEqual(first.snapshot.Neurons.map(row => row.ID).sort(),
  ["local-neuron-01", "local-neuron-02", "local-neuron-03", "local-neuron-04"]);
assert.ok(first.snapshot.Neurons.every(row => row.Slots === 1 && row.Running === 0 &&
  row.Runtimes.some(rt => rt.Kind === "standalone-typescript" && rt.Profile === "linux-bubblewrap") &&
  row.Runtimes.some(rt => rt.Kind === "plugin-typescript-v1" && rt.Profile === "linux-bubblewrap") &&
  row.Runtimes.some(rt => rt.Kind === "user-function-v1" && rt.Profile === "linux-bubblewrap")));
let later;
for (let i = 0; i < 30; i++) {
  await new Promise(resolve => setTimeout(resolve, 200));
  later = await request();
  if (later.snapshot.Neurons.length === 4 && later.snapshot.Neurons.every(row =>
    Date.parse(row.LastPulse) > Date.parse(first.snapshot.Neurons.find(old => old.ID === row.ID).LastPulse))) break;
}
assert.equal(later.snapshot.Neurons.length, 4);
assert.ok(later.snapshot.Neurons.every(row =>
  Date.parse(row.LastPulse) > Date.parse(first.snapshot.Neurons.find(old => old.ID === row.ID).LastPulse)),
"all Neurons must refresh their heartbeat over Synapse");
console.log("Four isolated Linux Neurons registered and pulsed");
