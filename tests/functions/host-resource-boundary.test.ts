import { rejectPluginHostResources } from "../../plugins/functions/index.js";

rejectPluginHostResources({ artifact: "build.zip" });
for (const input of [
  { dockerSocket: "/var/run/docker.sock" },
  { mount: "/run/containerd/containerd.sock" },
]) {
  try {
    rejectPluginHostResources(input);
    throw new Error("host resource crossed plugin boundary");
  } catch (error) {
    if (!(error instanceof Error) || error.message !== "host resources must remain host-owned")
      throw error;
  }
}
