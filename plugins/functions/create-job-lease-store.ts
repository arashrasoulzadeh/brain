import { createJobLeaseState } from "./job-lease-state.js";
import { JobLeaseAcquirer } from "./job-lease-acquirer.js";
import { JobLeaseRenewer } from "./job-lease-renewer.js";
export function createJobLeaseStore() {
  const state = createJobLeaseState();
  const acquirer = new JobLeaseAcquirer(state);
  const renewer = new JobLeaseRenewer(state);
  return { acquire: acquirer.acquire.bind(acquirer), renew: renewer.renew.bind(renewer) };
}
