import type { AdminContribution } from "./admin-registry.js";

export type MenuIcon = "projects" | "plugins" | "audit" | "system" | "functions" | "routes" | "monitoring" | "workers" | "resource" | "extension";
export type MenuItem = { href: string; label: string; icon: MenuIcon; detail?: string };
export type MenuGroup = { label: string; items: MenuItem[] };

const platform: MenuItem[] = [
  { href: "/projects", label: "Projects", icon: "projects" },
  { href: "/plugins", label: "Plugins", icon: "plugins" },
  { href: "/system", label: "Capabilities", icon: "system" },
  { href: "/routes", label: "API routes", icon: "routes" },
  { href: "/monitoring", label: "Monitoring", icon: "monitoring" },
  { href: "/workers", label: "Workers", icon: "workers" },
  { href: "/audit", label: "Audit log", icon: "audit" },
];

export function adminMenu(allowed: readonly string[], extensions: readonly AdminContribution[], search = ""): MenuGroup[] {
  const aliases = [...new Set(extensions.flatMap(item => item.page.legacy_route ? [item.page.legacy_route] : []))].sort();
  const project: MenuItem[] = [
    ...(allowed.includes("functions") ? [{ href: "/functions", label: "Functions", icon: "functions" as const }] : []),
    ...aliases.map(alias => ({ href: "/" + alias, label: alias[0].toUpperCase() + alias.slice(1).replaceAll("-", " "), icon: "resource" as const })),
  ];
  const groups: MenuGroup[] = [
    { label: "Workspace", items: platform.filter(item => allowed.includes(item.href.slice(1)) && ["projects", "plugins"].includes(item.icon)) },
    { label: "Project tools", items: project },
    { label: "Extensions", items: extensions.filter(item => !item.page.legacy_route).map(item => ({ href: item.href, label: item.page.title, detail: item.pluginID + "@" + item.version, icon: "extension" as const })) },
    { label: "Platform", items: platform.filter(item => allowed.includes(item.href.slice(1)) && !["projects", "plugins"].includes(item.icon)) },
  ];
  const query = search.trim().toLowerCase();
  return groups.map(group => ({ ...group, items: group.items.filter(item => !query || (item.label + " " + (item.detail ?? "") + " " + group.label).toLowerCase().includes(query)) })).filter(group => group.items.length > 0);
}

export function menuItemActive(pathname: string, href: string): boolean {
  if (pathname === "/") return href === "/projects";
  return pathname === href || (!href.startsWith("/extensions/") && pathname.startsWith(href + "/"));
}
