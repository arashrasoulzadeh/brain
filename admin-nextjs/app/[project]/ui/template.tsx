"use client";
import { useParams } from "next/navigation";
import { Fragment, type ReactNode } from "react";

// A project change must discard every page's local form, result and polling state.
export default function ProjectTemplate({ children }: { children: ReactNode }) {
  const { project } = useParams<{ project: string }>();
  return <Fragment key={project}>{children}</Fragment>;
}
