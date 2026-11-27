"use client";
import { useState } from "react";
import type { AdminResultSpec } from "../lib/admin-types";
import { displayAdminValue, readAdminResultPath, recordRows } from "../lib/admin-result-data";

function TableResult({ value, spec }: { value: unknown; spec?: AdminResultSpec }) {
  const [filter, setFilter] = useState("");
  const [page, setPage] = useState(0);
  const rows = recordRows(value);
  if (!rows) return <pre>{JSON.stringify(value, null, 2)}</pre>;
  const columns = spec?.columns?.length ? spec.columns : [...new Set(rows.flatMap(row => Object.keys(row)))];
  const filtered = rows.filter(row => columns.some(column => displayAdminValue(readAdminResultPath(row, column)).toLowerCase().includes(filter.toLowerCase())));
  const size = spec?.page_size || 25;
  const pages = Math.max(1, Math.ceil(filtered.length / size));
  const current = Math.min(page, pages - 1);
  return <section><label className="table-filter">Filter rows<input value={filter} onChange={event => { setFilter(event.target.value); setPage(0); }} /></label>
    {filtered.length ? <div className="table-wrap"><table><thead><tr>{columns.map(column => <th key={column}>{column}</th>)}</tr></thead>
      <tbody>{filtered.slice(current * size, (current + 1) * size).map((row, index) => <tr key={index}>{columns.map(column => <td key={column}>{displayAdminValue(readAdminResultPath(row, column))}</td>)}</tr>)}</tbody></table></div> : <p className="muted empty-state">No rows</p>}
    <div className="lifecycle-actions"><button type="button" disabled={current === 0} onClick={() => setPage(current - 1)}>Previous</button>
      <span>{current + 1} / {pages} · {filtered.length} rows</span><button type="button" disabled={current + 1 >= pages} onClick={() => setPage(current + 1)}>Next</button></div>
  </section>;
}

function ChartResult({ value, spec }: { value: unknown; spec: AdminResultSpec }) {
  const rows = recordRows(value) ?? [];
  const data = rows.slice(0, 100).flatMap(row => {
    const number = readAdminResultPath(row, spec.value_field);
    return typeof number === "number" && Number.isFinite(number) ? [{ label: displayAdminValue(readAdminResultPath(row, spec.label_field)), number }] : [];
  });
  const max = Math.max(1, ...data.map(item => Math.abs(item.number)));
  return <div aria-label={spec.title ?? "Chart"}>{data.length ? data.map((item, index) => <div key={index} style={{ marginBottom: 8 }}><span>{item.label} · {item.number}</span><div style={{ display: "flex", alignItems: "center", height: 12 }}>
    <div style={{ flex: 1, display: "flex", justifyContent: "flex-end", borderRight: "1px solid currentColor" }}><div style={{ width: item.number < 0 ? Math.abs(item.number) / max * 100 + "%" : "0%", background: "currentColor", height: 8 }} /></div>
    <div style={{ flex: 1 }}><div style={{ width: item.number >= 0 ? item.number / max * 100 + "%" : "0%", background: "currentColor", height: 8 }} /></div>
  </div></div>) : <p>No numeric chart data</p>}</div>;
}

function TabsResult({ value, spec }: { value: unknown; spec: AdminResultSpec }) {
  const [index, setIndex] = useState(0);
  const children = spec.children ?? [];
  const selected = Math.min(index, children.length - 1);
  return <section><div className="lifecycle-actions" role="tablist">{children.map((child, i) => <button key={i} type="button" role="tab" aria-selected={selected === i} onClick={() => setIndex(i)}>{child.title ?? child.type}</button>)}</div>
    {children[selected] && <div role="tabpanel"><AdminResultRenderer key={selected} value={value} spec={children[selected]} /></div>}
  </section>;
}

export function AdminResultRenderer({ value, spec }: { value: unknown; spec?: AdminResultSpec }) {
  const selected = readAdminResultPath(value, spec?.path);
  if (selected === undefined) return <p>No result for {spec?.path ?? "this component"}</p>;
  let content;
  switch (spec?.type) {
    case "table": content = <TableResult value={selected} spec={spec} />; break;
    case "detail":
      content = selected && typeof selected === "object" && !Array.isArray(selected)
        ? <dl>{Object.entries(selected).map(([name, val]) => <div key={name}><dt>{name}</dt><dd>{displayAdminValue(val)}</dd></div>)}</dl>
        : <p>{displayAdminValue(selected)}</p>; break;
    case "metric": content = <strong style={{ fontSize: "2rem" }}>{displayAdminValue(selected)}</strong>; break;
    case "chart": content = <ChartResult value={selected} spec={spec} />; break;
    case "logs": content = <pre>{Array.isArray(selected) ? selected.map(displayAdminValue).join("\n") : displayAdminValue(selected)}</pre>; break;
    case "code": content = <pre>{displayAdminValue(selected)}</pre>; break;
    case "grid": content = <div className="plugin-grid">{spec.children?.map((child, i) => <article className="plugin" key={i}><AdminResultRenderer value={selected} spec={child} /></article>)}</div>; break;
    case "tabs": content = <TabsResult value={selected} spec={spec} />; break;
    case "json": content = <pre>{JSON.stringify(selected, null, 2)}</pre>; break;
    default:
      if (recordRows(selected)) content = <TableResult value={selected} />;
      else if (selected && typeof selected === "object" && "rows" in selected && recordRows(selected.rows)) content = <><pre>{JSON.stringify(Object.fromEntries(Object.entries(selected).filter(([key]) => key !== "rows")), null, 2)}</pre><TableResult value={selected.rows} /></>;
      else content = <pre>{JSON.stringify(selected, null, 2)}</pre>;
  }
  return <section>{spec?.title && <h3>{spec.title}</h3>}{content}</section>;
}
