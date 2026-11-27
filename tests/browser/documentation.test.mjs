import assert from "node:assert/strict";
import { resolve } from "node:path";
import { test } from "node:test";
import { chromium, artifacts, start, visible, waitFor } from "./browser-runtime.mjs";

test("documentation homepage, navigation, search and responsive dark/light layouts", { timeout: 150000 }, async () => {
  const base = "http://127.0.0.1:3080";
  const server = start("python", ["-m", "http.server", "3080", "--bind", "127.0.0.1"], ".generated-site", "documentation");
  let browser;
  try {
    await server.ready(base);
    browser = await chromium.launch({ headless: true });
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 }, deviceScaleFactor: 1 });
    await page.goto(base + "/index.html", { waitUntil: "networkidle" });
    await visible(page.locator(".md-content__inner").getByRole("heading", { name: "Build apps, websites and games on a composable backend." }));
    await visible(page.locator(".brain-hero"));
    assert.ok(await page.locator(".grid.cards > ul > li").count() >= 4, "Markdown feature cards must render as cards, not raw text");
    await page.screenshot({ path: resolve(artifacts, "documentation-desktop-light.png"), fullPage: true });
    await page.getByTitle("Switch to dark mode", { exact: true }).click();
    await waitFor(() => page.locator("body").getAttribute("data-md-color-scheme").then(value => value === "slate"), "dark palette did not activate");
    await page.screenshot({ path: resolve(artifacts, "documentation-desktop-dark.png"), fullPage: true });
    await page.getByRole("link", { name: "Start developing", exact: true }).click();
    await page.waitForURL("**/developers.html");
    await visible(page.locator("#developer-guide"));
    await page.goto(base + "/index.html", { waitUntil: "networkidle" });
    // Follow the generated documentation links instead of assuming MkDocs URL layout.
    await page.getByRole("link", { name: "Explore plugins", exact: true }).click();
    await visible(page.locator(".md-content__inner").getByText("Generated from the host", { exact: false }));
    const reference = page.getByText(/^Manifest reference coverage: (\d+)\/\d+\.$/);
    await visible(reference);
    const counts = (await reference.innerText()).match(/(\d+)\/(\d+)/);
    assert.ok(counts && Number(counts[1]) > 0 && counts[1] === counts[2]);
    await page.locator(".md-content__inner").getByRole("link", { name: "arena", exact: true }).click();
    await page.getByRole("link", { name: "Developer and administrator guide", exact: true }).click();
    await visible(page.getByRole("heading", { name: "Developer", exact: true }));
    await visible(page.getByRole("heading", { name: "Administrator", exact: true }));
    const search = page.locator(".md-search__input");
    await search.fill("arena");
    await waitFor(async () => await page.locator(".md-search-result__link").count() > 0, "search index did not return Arena results");
    const mobile = await browser.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 1, isMobile: true, hasTouch: true });
    await mobile.goto(base, { waitUntil: "networkidle" });
    await visible(mobile.locator(".brain-hero"));
    const overflow = await mobile.evaluate(() => document.documentElement.scrollWidth - innerWidth);
    assert.ok(overflow <= 2, "mobile page horizontally overflows: " + overflow);
    await mobile.screenshot({ path: resolve(artifacts, "documentation-mobile.png"), fullPage: true });
  } finally { await browser?.close(); server.close(); }
});
