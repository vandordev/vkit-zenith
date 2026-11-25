import assert from "node:assert/strict";
import test from "node:test";
import { countBun, countGo } from "./check-infra.mjs";

test("Go counter requires every named passing case and rejects skips", () => {
	const passing = ["TestA", "TestB"].map((Test) => JSON.stringify({ Action: "pass", Test })).join("\n");
	assert.equal(countGo(passing, ["TestA", "TestB"]), 2);
	assert.throws(() => countGo(JSON.stringify({ Action: "pass", Test: "TestA" }), ["TestA", "TestB"]));
	assert.throws(() => countGo(`${passing}\n${JSON.stringify({ Action: "skip", Test: "TestB" })}`, ["TestA", "TestB"]));
	assert.throws(() => countGo(passing, []));
});

test("JS counter requires the two named document cases and zero skips", () => {
	const passing = "✓ rejects forbidden YAML features and unknown keys\n✓ accepts infrastructure sections and rejects unknown infrastructure keys\n2 pass\n0 fail";
	assert.equal(countBun(passing), 2);
	assert.throws(() => countBun(passing.replace("2 pass", "1 pass")));
	assert.throws(() => countBun(`${passing}\n1 skip`));
	assert.throws(() => countBun(passing.replace("accepts infrastructure sections and rejects unknown infrastructure keys", "unrelated test")));
});
