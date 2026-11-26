import { deepEqual, equal, rejects, throws } from "node:assert/strict";
import {
  createProjectMap,
  runCappedRetries,
  runWithTimeout,
} from "../../plugins/sdk-typescript/src/index.js";
import {
  createApprovalStore,
  createFunctionDraftStore,
  createFunctionPublicationStore,
  createInvocationCommitStore,
  createJobLeaseStore,
  createScheduleStore,
  createSharedPackageStore,
  runCappedRetries as functionRetries,
  runWithTimeout as functionTimeout,
} from "../../plugins/functions/index.js";

equal(functionRetries, runCappedRetries);
equal(functionTimeout, runWithTimeout);
let attempts = 0;
equal(await runCappedRetries(() => {
  if (++attempts < 2) throw new Error("retry");
  return "success";
}, 2), "success");
await rejects(runCappedRetries(() => "never", 0), /invalid retry cap/);
await rejects(runWithTimeout(Promise.resolve("never"), 0), /invalid invocation timeout/);

const map = createProjectMap<string>();
map.set("dev", "same", "dev value");
map.set("prod", "same", "prod value");
equal(map.get("dev", "same"), "dev value");
equal(map.get("prod", "same"), "prod value");
throws(() => map.get("other/project", "same"), /invalid project id/);

const drafts = createFunctionDraftStore();
const devDraft = drafts.save("dev", "a", "source", 1);
const prodDraft = drafts.save("prod", "a", "other", 2);
equal(drafts.get("dev", "a")?.source, "source");
equal(drafts.get("prod", "a")?.source, "other");
const publications = createFunctionPublicationStore();
equal(publications.publish(devDraft, "sha256:abc", "go1.23", 3).version, 1);
equal(publications.publish(prodDraft, "sha256:def", "go1.23", 3).version, 1);
equal(publications.publish({ ...devDraft, functionId: "a:b" }, "sha256:def", "go1.23", 4).version, 1);
equal(publications.publish(devDraft, "sha256:def", "go1.23", 5).version, 2);
equal(publications.rollback("dev", "a", 1).version, 1);
equal(publications.active("prod", "a")?.source, "other");

const shared = createSharedPackageStore();
shared.save("dev", "a", "package helper");
throws(() => shared.load("prod", "a"), /shared package not found/);
const leases = createJobLeaseStore();
equal(leases.acquire("dev", "job", "worker-a", 1, 10), true);
equal(leases.acquire("dev", "job", "worker-b", 2, 10), false);
equal(leases.acquire("prod", "job", "worker-b", 2, 10), true);
equal(leases.renew("dev", "job", "worker-b", 3, 10), false);
equal(leases.renew("dev", "job", "worker-a", 11, 10), false);

const approvals = createApprovalStore();
approvals.request("run", "requester");
equal(approvals.approve("run", "requester"), false);
equal(approvals.approve("run", "reviewer"), true);
equal(approvals.isApproved("run"), true);

const commits = createInvocationCommitStore();
commits.commit("id", ["first"], undefined);
commits.commit("id", ["duplicate"], "other");
deepEqual(commits.effects(), ["first"]);
commits.setLifecycleRevision(2);
throws(() => commits.commit("second", [], "result", true, 1), /stale invocation lifecycle revision/);

const schedules = createScheduleStore();
schedules.add("dev", "a", "* * * * *", 1);
schedules.add("prod", "a", "* * * * *", 1);
equal(schedules.claim("dev", "a", "worker", 1, 10), true);
equal(schedules.claim("prod", "a", "worker", 1, 10), true);
equal(schedules.enqueueLatestMissed("dev", "a", 2)?.projectId, "dev");
equal(schedules.enqueueLatestMissed("dev", "a", 2), undefined);
equal(schedules.enqueueLatestMissed("prod", "a", 2)?.projectId, "prod");
