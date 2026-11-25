import { expect, test } from "bun:test";
test("documents independent APIs and direct runtime-configured client", async () => {
	const readme = await Bun.file("README.md").text();
	for (const term of ["TanStack Start", "Echo", "Huma", "apps/web-api", "apps/api", "getPublicConfig", "task gen"]) expect(readme).toContain(term);
	expect(readme).not.toContain("standalone Elysia");
	const taskfile = await Bun.file("Taskfile.yml").text();
	for (const target of ["dev:api", "dev:web-api", "start:api", "start:web-api", "build:api", "build:web-api"]) expect(taskfile).toContain(`  ${target}:`);
});
