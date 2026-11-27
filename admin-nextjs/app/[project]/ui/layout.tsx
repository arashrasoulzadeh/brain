import { notFound } from "next/navigation";
import type { ReactNode } from "react";

export default async function ProjectLayout({ children, params }: { children: ReactNode; params: Promise<{ project: string }> }) {
  const { project } = await params;
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(project)) notFound();
  return <>{children}</>;
}
