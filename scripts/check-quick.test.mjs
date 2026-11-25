import assert from "node:assert/strict";
import test from "node:test";
import { changedPaths, plan } from "./check-quick.mjs";

const goMods = new Set(["apps/bo-api", "apps/web-api", "apps/api", "apps/worker", "packages/db"]);
const exists = (p) => goMods.has(p.replace(/\/go\.mod$/, "")) && p.endsWith("/go.mod");
const labels = (paths) => plan(paths, exists).map((s) => `${s.cwd}: ${s.cmd.join(" ")}`);

test("changedPaths parses -z porcelain, skipping the old name of a rename", () => {
	const out = " M apps/bo-api/main.go\0R  packages/db/b.go\0a.go\0?? scripts/x.ts\0";
	assert.deepEqual(changedPaths(out), ["apps/bo-api/main.go", "packages/db/b.go", "scripts/x.ts"]);
});

test("two files in one Go module give a single vet step", () => {
	assert.deepEqual(labels(["apps/bo-api/internal/a.go", "apps/bo-api/internal/b/c.go"]), ["apps/bo-api: go vet ./..."]);
});

test("a shared Go package also vets the consumer apps, without duplicates", () => {
	const got = labels(["packages/db/x.go", "apps/api/y.go"]);
	assert.equal(got.length, 5);
	assert.equal(new Set(got).size, 5);
});

test("a deleted file is still mapped by its directory", () => {
	assert.deepEqual(labels(["apps/bo-api/internal/gone.go"]), ["apps/bo-api: go vet ./..."]);
});

test("a Web or BO change checks only that workspace", () => {
	assert.deepEqual(labels(["apps/bo/src/x.tsx"]), ["apps/bo: bun --no-env-file run check-types"]);
});

test("a shared UI package checks its consumers", () => {
	assert.deepEqual(labels(["packages/ui/src/y.tsx"]).map((s) => s.split(":")[0]), ["packages/ui", "apps/web", "apps/bo"]);
});

test("landing adds the ui-primitives step", () => {
	assert.deepEqual(labels(["apps/landing/src/z.tsx"]), ["apps/landing: bun --no-env-file run check-types", ".: bun --no-env-file run check:ui-primitives"]);
});

test("files outside any owner are ignored", () => {
	assert.deepEqual(plan(["README.md", "docs/a.md"], exists), []);
});
