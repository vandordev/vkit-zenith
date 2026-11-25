import { describe, expect, test } from "bun:test";

describe("TanStack Start web configuration", () => {
	test("uses TanStack Start with Bun Nitro output", async () => {
		const source = await Bun.file(new URL("./vite.config.ts", import.meta.url)).text();
		expect(source).toContain("tanstackStart");
		expect(source).not.toContain("tanstackRouter(");
		expect(source).toContain('preset: process.env.VERCEL === "1" ? "vercel" : "bun"');
		expect(source).toContain('envDir: false');
		expect(source).toContain('envPrefix: []');
		expect(source).toContain('nitro.options.output.serverDir');
		expect(source).toContain('copyFile');
	});
});
