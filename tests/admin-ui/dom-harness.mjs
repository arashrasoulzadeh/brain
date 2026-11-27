import assert from "node:assert/strict";
import { createRequire } from "node:module";
import { resolve } from "node:path";
import { afterEach } from "node:test";

const require = createRequire(new URL("../../admin-nextjs/package.json", import.meta.url));
const tools = createRequire(resolve(process.env.BRAIN_TEST_TOOLS ?? "tests/admin-ui", "package.json"));
const { JSDOM } = tools("jsdom");
const dom = new JSDOM("<!doctype html><html><body></body></html>", { url: "http://localhost:3000/", pretendToBeVisual: true });
for (const name of ["window", "document", "navigator", "HTMLElement", "HTMLInputElement", "HTMLSelectElement", "HTMLTextAreaElement", "CustomEvent", "Event", "MouseEvent"]) {
  Object.defineProperty(globalThis, name, { value: name === "window" ? dom.window : dom.window[name], configurable: true, writable: true });
}
globalThis.IS_REACT_ACT_ENVIRONMENT = true;
const React = require("react");
globalThis.React = React;
export const { act, createElement: h } = React;
const { createRoot } = require("react-dom/client");
const { AppRouterContext } = require("next/dist/shared/lib/app-router-context.shared-runtime");
const { PathnameContext, PathParamsContext, SearchParamsContext } = require("next/dist/shared/lib/hooks-client-context.shared-runtime");
export const app = path => require("./app/" + path);
const originalFetch = globalThis.fetch;
let root, host;
export const navigations = [];
const router = {
  push: url => navigations.push(["push", url]), replace: url => navigations.push(["replace", url]),
  prefetch: async () => {}, refresh: () => {}, back: () => {}, forward: () => {},
};

export function stubFetch(handler) {
  globalThis.fetch = window.fetch = async (input, init = {}) => handler(String(input), init);
}
export function json(data, status = 200) { return Response.json(data, { status }); }
export function reply(data) { return json({ Success: true, Payload: Buffer.from(JSON.stringify(data)).toString("base64") }); }
export async function mount(Component, props = {}, { pathname = "/", params = {} } = {}) {
  if (root) await act(async () => root.unmount());
  document.body.innerHTML = "";
  host = document.createElement("div");
  document.body.append(host);
  root = createRoot(host);
  await act(async () => root.render(h(AppRouterContext.Provider, { value: router },
    h(PathnameContext.Provider, { value: pathname },
      h(PathParamsContext.Provider, { value: params },
        h(SearchParamsContext.Provider, { value: new URLSearchParams() }, h(Component, props)))))));
  return host;
}
export function find(selector, within = host) {
  const element = within.querySelector(selector);
  assert.ok(element, "missing " + selector + " in " + within.innerHTML);
  return element;
}
export function button(text, within = host) {
  const element = [...within.querySelectorAll("button")].find(node => node.textContent.trim() === text);
  assert.ok(element, "missing button " + text + " in " + within.innerHTML);
  return element;
}
export async function click(element) {
  assert.equal(element.disabled, false, "button unexpectedly disabled: " + element.textContent);
  await act(async () => element.click());
}
export async function change(element, value) {
  const prototype = Object.getPrototypeOf(element);
  const setter = Object.getOwnPropertyDescriptor(prototype, "value").set;
  await act(async () => {
    setter.call(element, value);
    element.dispatchEvent(new window.Event(element.tagName === "SELECT" ? "change" : "input", { bubbles: true }));
  });
}
export async function submit(element) {
  await act(async () => element.dispatchEvent(new window.Event("submit", { bubbles: true, cancelable: true })));
}
export function text() { return host.textContent; }
export function session(project = "dev") {
  window.sessionStorage.setItem("brain-admin-session", "fixture-token");
  window.sessionStorage.setItem("brain-admin-project", project);
}
afterEach(async () => {
  if (root) await act(async () => root.unmount());
  root = undefined;
  document.body.innerHTML = "";
  window.sessionStorage.clear();
  globalThis.fetch = originalFetch;
  window.fetch = originalFetch;
  navigations.length = 0;
  window.confirm = () => true;
});
