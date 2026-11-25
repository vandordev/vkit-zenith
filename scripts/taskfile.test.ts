import { expect, test } from "bun:test";
const taskfile = await Bun.file("Taskfile.yml").text();
test("exposes only foundation runtime operations", () => {
	for (const name of ["doctor", "install", "dev", "dev:web", "dev:api", "dev:web-api", "dev:storybook", "start:api", "start:web-api", "gen", "gen:handlers", "gen:openapi", "sync:web", "sync:bo", "sync:api", "sync", "check:quick", "create:handler", "create:usecase", "create:migration", "check:scaffold", "check-types", "check:go"]) expect(taskfile).toContain(`  ${name}:`);
	for (const name of ["migrate", "dev:worker", "dev:scheduler", "dev:realtime", "db:generate", "compose:smoke", "quality"]) expect(taskfile).not.toContain(`  ${name}:`);
});
test("Go dev loads development env explicitly and static checks never run services", () => {
	expect(taskfile).toContain("go run ./apps/api --root '{{.ROOT_DIR}}' --development");
	expect(taskfile).toContain("go run ./apps/web-api --root '{{.ROOT_DIR}}' --development");
	// Only the check:go task family: a positional slice also swallowed unrelated db:* and check:db:* tasks.
	const check = taskfile
		.split(/^(?=  [\w:-]+:$)/m)
		.filter((block) => /^  check:go(?::[\w-]+)?:\n/.test(block))
		.join("");
	expect(check).toContain("go build");
	expect(check).not.toMatch(/go (?:run|test|vet)/);
});
