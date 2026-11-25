import { expect, test } from "bun:test";
import { readFileSync } from "node:fs";
test("runtime images include YAML and never copy real env files", () => {
	for (const name of ["api", "web-api", "web"]) {
		const source = readFileSync(new URL(`../../../../Dockerfile.${name}`, import.meta.url), "utf8");
		expect(source).toContain("/app/config");
		expect(source).not.toMatch(/COPY.*\.env/);
	}
});
