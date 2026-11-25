import { readFileSync } from "node:fs";
import { expect, test } from "bun:test";
test("ships independent nonroot images with runtime YAML and no env copies", () => {
	for (const name of ["api", "web-api", "web"]) {
		const source = readFileSync(`Dockerfile.${name}`, "utf8");
		expect(source).toContain("USER");
		expect(source).toContain("config");
		expect(source).not.toContain("COPY .env");
		if (name === "api") expect(source).not.toMatch(/prisma|packages\/db/);
	}
});
test("web uses pinned prune and nested config without Go compilation", () => {
	const source = readFileSync("Dockerfile.web", "utf8");
	expect(source).toContain("turbo@2.11.10");
	expect(source).toContain("packages/config/js/src/run.ts");
	expect(source).toContain("bun install --production --frozen-lockfile --ignore-scripts");
	expect(source).not.toContain("go build");
});
test("Go images build only selected module and config with workspace disabled", () => {
	for (const name of ["api", "web-api"]) {
		const source = readFileSync(`Dockerfile.${name}`, "utf8");
		expect(source).toContain("golang:1.27.3-alpine");
		expect(source).toContain("GOWORK=off CGO_ENABLED=0");
		expect(source).toContain(`COPY apps/${name} ./`);
		expect(source).not.toContain("COPY . .");
	}
});
