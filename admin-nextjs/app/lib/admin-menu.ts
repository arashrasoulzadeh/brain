import { projectUIURL, systemUIURL } from "./admin-api.js";
import { adminAliasTitle, type AdminContribution } from "./admin-registry.js";

export type MenuIcon = "projects" | "plugins" | "audit" | "system" | "functions" | "routes" | "monitoring" | "workers" | "resource" | "extension";
export type MenuItem = { href: string; label: string; icon: MenuIcon; detail?: string };
export type MenuGroup = { label: string; items: MenuItem[] };

const platform: MenuItem[] = [
  { href: "/projects", label: "Projects", icon: "projects" },
  { href: "/plugins", label: "Plugins", icon: "plugins" },
  { href: "/system", label: "Capabilities", icon: "system" },
  { href: "/routes", label: "API routes", icon: "routes" },
  { href: "/monitoring", label: "Monitoring", icon: "monitoring" },
  { href: "/workers", label: "Neurons", icon: "workers" },
  { href: "/audit", label: "Audit log", icon: "audit" },
];

export function adminMenu(allowed: readonly string[], extensions: readonly AdminContribution[], search = "", projectID = "", fallback: readonly MenuItem[] = []): MenuGroup[] {
  const aliases = [...new Set(extensions.filter(item => item.projectID === projectID).flatMap(item => item.page.legacy_route ? [item.page.legacy_route] : []))].sort();
  const project: MenuItem[] = [
    ...(projectID && allowed.includes("functions") ? [{ href: projectUIURL(projectID, "functions"), label: "Functions", icon: "functions" as const }] : []),
    ...(projectID ? aliases : []).map(alias => ({ href: projectUIURL(projectID, alias), label: adminAliasTitle(alias), icon: "resource" as const })),
    ...(projectID ? fallback : []),
  ];
  const scopedPlatform = platform.filter(item => projectID || ["projects", "audit", "workers"].includes(item.icon)).map(item => ({ ...item, href: ["projects", "audit", "workers"].includes(item.icon) ? systemUIURL(item.href.slice(1)) : projectUIURL(projectID, item.href.slice(1)) }));
  const groups: MenuGroup[] = [
    { label: "Workspace", items: scopedPlatform.filter(item => allowed.includes(item.icon) && ["projects", "plugins"].includes(item.icon)) },
    { label: "Plugin workspaces", items: project },
    { label: "Extensions", items: extensions.filter(item => projectID && item.projectID === projectID && !item.page.legacy_route).map(item => ({ href: item.href, label: item.page.title, detail: item.pluginID + "@" + item.version, icon: "extension" as const })) },
    { label: "Platform", items: scopedPlatform.filter(item => allowed.includes(item.icon) && !["projects", "plugins"].includes(item.icon)) },
  ];
  const query = search.trim().toLowerCase();
  return groups.map(group => ({ ...group, items: group.items.filter(item => !query || (item.label + " " + (item.detail ?? "") + " " + group.label).toLowerCase().includes(query)) })).filter(group => group.items.length > 0);
}

export function menuItemActive(pathname: string, href: string): boolean {
  if (pathname === systemUIURL()) return href === systemUIURL("projects");
  return pathname === href || (!href.includes("/ui/extensions/") && pathname.startsWith(href + "/"));
}
