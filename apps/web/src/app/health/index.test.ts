import { expect, test } from "bun:test";
test("web health has no legacy API or external service dependency", async () => {
	const source = await Bun.file(new URL("./index.ts", import.meta.url)).text();
	expect(source).toContain('Response.json({ status: "ok" })');
	expect(source).not.toMatch(/elysia|@repo\/application|@repo\/db/);
});
