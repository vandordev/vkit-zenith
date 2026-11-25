import { test } from "node:test";
import assert from "node:assert/strict";
import { goResult, bunResult, runGate } from "./check-auth.mjs";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { join } from "node:path";

test("Go counter requires every selected actual case and rejects skips/failures", () => {
	const output = [
		{ Action: "pass", Test: "TestAuthOne/sub" },
		{ Action: "pass", Test: "TestAuthOne" },
		{ Action: "pass", Package: "fixture" },
	]
		.map(JSON.stringify)
		.join("\n");
	assert.deepEqual(goResult(output, ["TestAuthOne"]), { tests: 1, subcases: 1 });
	for (const [data, selected] of [
		["", []],
		[output, ["TestAuthTwo"]],
		[output, ["TestAuthOne", "TestAuthTwo"]],
		[output + '\n{"Action":"skip","Test":"TestAuthTwo"}', ["TestAuthOne"]],
		[output + '\n{"Action":"fail"}', ["TestAuthOne"]],
	])
		assert.throws(() => goResult(data, selected));
});
test("Go counter rejects duplicated required subtests", () => {
	const output = [
		{ Action: "pass", Test: "TestAuthLife/startup-failure" },
		{ Action: "pass", Test: "TestAuthLife/startup-failure" },
		{ Action: "pass", Test: "TestAuthLife" },
	].map(JSON.stringify).join("\n");
	assert.throws(() => goResult(output, ["TestAuthLife"], { subtests: ["TestAuthLife/startup-failure"] }));
});
test("Bun counter rejects zero, missing, failed and skipped selections", () => {
	assert.equal(bunResult("\n 9 pass\n 0 fail\n"), 9);
	for (const output of ["", "0 pass\n 0 fail", "1 pass\n 1 fail", "1 pass\n 0 fail\n 1 skip", "1 pass\n 0 fail\n 1 todo"])
		assert.throws(() => bunResult(output));
});
test("malformed Go output never exposes its secret-bearing parse excerpt", () => {
	assert.throws(() => goResult("fake-sensitive-action-url", ["TestAuthOne"]), { message: "invalid Go event; raw output withheld" });
});
test("Go counter rejects missing required tests and subtests", () => {
 const output = JSON.stringify({ Action: "pass", Test: "TestAuthOne" });
 assert.throws(() => goResult(output, ["TestAuthOne"], { tests: ["TestAuthTwo"] }));
 assert.throws(() => goResult(output, ["TestAuthOne"], { subtests: ["TestAuthOne/sub"] }));
});
test("unknown commands fail without launching a gate", () => {
	assert.throws(() => runGate("absent"));
});

test("independent cleanup rejects an empty manifest before any resource inspection", () => {
	const root = mkdtempSync("/tmp/opencode/auth-receipt-negative-");
	try {
		const receipt = join(root, "empty.jsonl");
		writeFileSync(receipt, '{"Action":"pass"}\n', { mode: 0o600 });
		const result = spawnSync(process.execPath, [fileURLToPath(new URL("./auth-e2e/verify-receipt.mjs", import.meta.url)), receipt], {
			cwd: root,
			env: { PATH: "/usr/bin:/bin", HOME: root },
			encoding: "utf8",
			timeout: 5000,
		});
		assert.equal(result.status, 1);
		assert.match(result.stderr, /incomplete fixture cleanup manifest/);
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
