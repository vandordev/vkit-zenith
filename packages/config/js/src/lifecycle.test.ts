import { expect, test } from "bun:test";
import { getPublicBOConfig, loadPublicBOConfig } from "./runtime";
import { readFileSync } from "node:fs";
test("process getter reuses one process snapshot while pure loaders are isolated", () => {
	expect(getPublicBOConfig()).toBe(getPublicBOConfig());
	expect(loadPublicBOConfig({ environment: { BO_API_BASE_URL: "https://one.test" } }).boApiBaseUrl).toBe("https://one.test");
	expect(loadPublicBOConfig({ environment: { BO_API_BASE_URL: "https://two.test" } }).boApiBaseUrl).toBe("https://two.test");
});
test("public import closure is schema-only and deployment roots are explicit", () => {
	const source = readFileSync(new URL("./public.ts", import.meta.url), "utf8");
	expect(source).not.toMatch(/node:fs|process\.env|\.\/runtime|\.\/server/);
	const launcher = readFileSync(new URL("./run.ts", import.meta.url), "utf8");
	expect(launcher).toContain("--no-env-file");
	expect(launcher).toContain("ZENITH_APPLICATION_ROOT");
});
