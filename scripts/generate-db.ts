import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

// Reviewed runtime owners only. Empty future owners must not acquire dummy SQL.
function hasQueries(path: string): boolean {
	if (!existsSync(path)) return false;
	return readdirSync(path, { withFileTypes: true }).some((entry) => {
		const child = join(path, entry.name);
		return entry.isDirectory()
			? hasQueries(child)
			: entry.isFile() && entry.name.endsWith(".sql") && /^-- name: /m.test(readFileSync(child, "utf8"));
	});
}
for (const [owner, source] of [
	["tools", "packages/tools/dbtool/testdata"],
	["application", "packages/application/staff"],
	["web-api", "apps/web-api/internal/handlers/v1"],
	["bo-api", "apps/bo-api/internal/handlers/v1"],
] as const) {
	if (owner !== "tools" && !hasQueries(source)) {
		console.log(`db-gen ${owner}: no installed named queries (not selected)`);
		continue;
	}
	const result = Bun.spawnSync(["go", "run", "./packages/tools/cmd/db-gen", "--root", process.cwd(), "--owner", owner], {
		stdout: "inherit",
		stderr: "inherit",
	});
	if (result.exitCode !== 0) process.exit(result.exitCode);
}
