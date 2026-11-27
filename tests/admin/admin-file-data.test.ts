import assert from "node:assert/strict";
import { test } from "node:test";
import { ADMIN_FILE_LIMIT, adminDownloadName, decodeAdminFile, encodeAdminFile } from "../../admin-nextjs/app/lib/admin-file-data.js";

test("selected binary and text files round-trip without replacement or loss", () => {
  for (const bytes of [new Uint8Array(), Uint8Array.from([0, 255, 128, 10, 13, 1]), new TextEncoder().encode("سلام 🌍\n"), new Uint8Array(ADMIN_FILE_LIMIT).fill(255)]) {
    const encoded = encodeAdminFile(bytes);
    assert.deepEqual(decodeAdminFile(encoded), bytes);
    assert.ok(encoded.length < 64 * 1024);
  }
});

test("file decoding is bounded and rejects malformed or unrelated data", () => {
  assert.throws(() => encodeAdminFile(new Uint8Array(ADMIN_FILE_LIMIT + 1)), /40 KiB/);
  const prefix = "data:application/octet-stream;base64,";
  for (const data of ["plain text", "data:text/html;base64,AAAA", prefix + "=", prefix + "AB==", prefix + "!AAA", prefix + "AAAA".repeat(15000), prefix + "AA=A"]) {
    assert.equal(decodeAdminFile(data), undefined);
  }
  assert.equal(adminDownloadName("folder/report.txt"), "report.txt");
  assert.equal(adminDownloadName("folder\\report.txt"), "report.txt");
  assert.equal(adminDownloadName(".."), "download.bin");
  assert.equal(adminDownloadName("x\x00.txt"), "x.txt");
});
