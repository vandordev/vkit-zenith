import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const goApps = ["apps/web-api", "apps/bo-api", "apps/api", "apps/worker"];
const tsWorkspaces = ["apps/web", "apps/bo", "apps/landing", "apps/storybook", "packages/ui", "packages/ui-email", "packages/ui-utils", "packages/brand", "packages/config/js"];
const tsConsumers = ["apps/web", "apps/bo"];
const bun = ["bun", "--no-env-file", "run"];

// `git status --porcelain=v1 -z`: entries are "XY path"; a rename/copy is followed by its old path.
export function changedPaths(output) {
	const fields = output.split("\0").filter(Boolean);
	const paths = [];
	for (let i = 0; i < fields.length; i++) {
		paths.push(fields[i].slice(3));
		if (/^[RC]/.test(fields[i])) i++;
	}
	return paths;
}

function goModule(path, exists) {
	for (let dir = dirname(path); dir !== "."; dir = dirname(dir)) if (exists(`${dir}/go.mod`)) return dir;
	return null;
}

// shortcut: checks changed owners plus the fixed consumer lists below; upgrade to a dependency graph if packages multiply.
export function plan(paths, exists) {
	const steps = new Map();
	const add = (cwd, cmd) => steps.set(`${cwd}\0${cmd.join(" ")}`, { cwd, cmd });
	const vet = (cwd) => add(cwd, ["go", "vet", "./..."]);
	const types = (cwd) => add(cwd, [...bun, "check-types"]);
	for (const path of paths) {
		if (/\.(go|sql)$/.test(path) || path.endsWith("go.mod")) {
			const mod = goModule(path, exists);
			if (!mod) continue;
			vet(mod);
			if (mod.startsWith("packages/")) goApps.forEach(vet);
		} else if (/\.(ts|tsx|mts)$/.test(path)) {
			if (path.startsWith("scripts/")) {
				add(".", ["bun", "--no-env-file", "x", "tsc", "--noEmit", "-p", "scripts/tsconfig.json"]);
				continue;
			}
			const owner = tsWorkspaces.find((w) => path.startsWith(`${w}/`));
			if (!owner) continue;
			types(owner);
			if (owner.startsWith("packages/")) tsConsumers.forEach(types);
			if (owner === "apps/landing" || owner === "apps/storybook") add(".", [...bun, "check:ui-primitives"]);
		}
	}
	return [...steps.values()];
}

function run() {
	const status = spawnSync("git", ["status", "--porcelain=v1", "-z", "-uall"], { cwd: root, encoding: "utf8" });
	if (status.status !== 0) throw new Error(`git status failed: ${status.stderr}`);
	const steps = plan(changedPaths(status.stdout), (p) => existsSync(resolve(root, p)));
	if (!steps.length) return console.log("ok: nothing to check");
	for (const { cwd, cmd } of steps) {
		const r = spawnSync(cmd[0], cmd.slice(1), { cwd: resolve(root, cwd), encoding: "utf8" });
		if (r.status !== 0) {
			console.error(`FAIL ${cwd}: ${cmd.join(" ")}\n${r.stdout}${r.stderr}`);
			process.exitCode = 1;
			return;
		}
	}
	console.log(`ok: ${steps.length} checks (${[...new Set(steps.map((s) => s.cwd))].join(", ")})`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) run();
