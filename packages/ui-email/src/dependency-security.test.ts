import { expect, test } from "bun:test";
import merge from "deepmerge";
import { createRequire } from "node:module";

test("patched email tooling merge cannot poison returned object prototypes", () => {
	const poisoned: unknown = JSON.parse('{"__proto__":{"inheritedPrivilege":true},"safe":{"original":1}}');
	if (poisoned === null || typeof poisoned !== "object") throw new Error("invalid regression fixture");
	// Exercise the advisory's unsafe TARGET-copy path, not only the already
	// guarded source-key path. JSON parsing creates an enumerable own __proto__.
	const result = merge(poisoned, { safe: { added: 2 } });
	expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
	expect(Reflect.get(result, "inheritedPrivilege")).toBeUndefined();
	const safe: unknown = Reflect.get(result, "safe");
	expect(safe).toEqual({ original: 1, added: 2 });
	const empty: Record<string, unknown> = Object.create(null);
	const fromSource = merge(empty, poisoned);
	expect(Object.getPrototypeOf(fromSource)).toBe(Object.prototype);
	expect(Reflect.get(fromSource, "inheritedPrivilege")).toBeUndefined();
	expect(merge({ list: [1] }, { list: [2] })).toEqual({ list: [1, 2] });
});

test("security backport covers CommonJS and UMD runtime distribution entrypoints", () => {
	const require = createRequire(import.meta.url);
	for (const entry of ["deepmerge/dist/cjs.js", "deepmerge/dist/umd.js"]) {
		const candidate: unknown = require(entry);
		if (typeof candidate !== "function") throw new Error("merge entrypoint unavailable");
		const target: unknown = JSON.parse('{"__proto__":{"inheritedPrivilege":true}}');
		if (target === null || typeof target !== "object") throw new Error("invalid regression fixture");
		const output: unknown = candidate(target, {});
		if (output === null || typeof output !== "object") throw new Error("merge returned non-object");
		expect(Object.getPrototypeOf(output)).toBe(Object.prototype);
		expect(Reflect.get(output, "inheritedPrivilege")).toBeUndefined();
	}
});
