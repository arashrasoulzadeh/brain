import Link from "next/link";
import { projectUIURL } from "../lib/admin-api";

export function PluginManageLink({ projectID, pluginID, version }: { projectID: string; pluginID: string; version: string }) {
  return <Link className="button-link secondary" href={projectUIURL(projectID, "plugins", pluginID + "@" + version)} aria-label={"Manage " + pluginID + " in " + projectID}>Manage</Link>;
}
