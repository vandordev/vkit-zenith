import { spawnSync } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
export const required = {
	"packages/application/customers": ["TestRolePermissionsMatrix"],
	"apps/web-api/internal/securitytest": [
		"TestWebAPIListMyOrganizations", "TestWebAPIRemovedMembership", "TestWebAPIInactiveUser",
		"TestWebAPIRestrictedSession", "TestWebAPIForeignOrganization", "TestWebAPIMalformedOrganizationID",
		"TestWebAPIPermissionDenials", "TestWebAPISuspendedWrite", "TestWebAPIWriteOriginCSRF",
		"TestWebAPICORSNoStore", "TestWebAPIMembersScopedAndNoSecrets",
		"TestWebAPISecretsOnlyOnCreate", "TestWebAPIAPIKeyValidation", "TestWebAPILogoOwnerOnly", "TestWebAPIAdminManagesDevelopers",
	],
};
const modules = {
	"packages/application/customers": { directory: "packages/application", package: "./customers", integration: false },
	"apps/web-api/internal/securitytest": { directory: "apps/web-api", package: "./internal/securitytest", integration: true },
};

function env(home) {
	return {
		PATH: process.env.PATH ?? "/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin",
		HOME: home, TMPDIR: home, LANG: "C.UTF-8", NO_COLOR: "1", GOTOOLCHAIN: "local",
		GOWORK: join(root, "go.work"), GOENV: "off",
		GOMODCACHE: process.env.GOMODCACHE ?? join(tmpdir(), "opencode/web-api-security-go-modcache"),
		GOCACHE: process.env.GOCACHE ?? join(tmpdir(), "opencode/web-api-security-go-cache"),
	};
}

function run(command, args, cwd, environment, timeout = 180000) {
	const result = spawnSync(command, args, { cwd, env: environment, encoding: "utf8", timeout, maxBuffer: 8 * 1024 * 1024 });
	if (result.error || result.status !== 0) throw new Error(`${command} security gate failed (raw output withheld)`);
	return `${result.stdout}\n${result.stderr}`;
}

function countGo(output, expected) {
	const passed = new Set();
	for (const line of output.split("\n").filter(Boolean)) {
		let event;
		try { event = JSON.parse(line); } catch { throw new Error("invalid Go security event") }
		if (event.Action === "fail" || event.Action === "skip") throw new Error("required security case failed or skipped");
		if (event.Action === "pass" && event.Test && !event.Test.includes("/")) passed.add(event.Test);
	}
	if (passed.size !== expected.length || expected.some((name) => !passed.has(name))) throw new Error("required security cases missing or zero");
	return passed.size;
}

function supervisedGo(command, args, cwd, seconds, home) {
	const recovery = mkdtempSync("/tmp/opencode/auth-web-api-security-");
	const result = spawnSync("python3", [join(root, "scripts/auth-e2e/supervise.py"), recovery, String(seconds * 1000), command, ...args], {
		cwd, env: env(home), encoding: "utf8", timeout: (seconds + 60) * 1000, maxBuffer: 1024 * 1024,
	});
	console.log(`Security supervisor evidence: ${recovery}`);
	if (result.error || result.status !== 0 || !existsSync(join(recovery, "result.json"))) {
		const diagnostic = result.error?.code ?? result.stderr?.trim() ?? "supervisor receipt missing";
		throw new Error(`security supervisor failed (${diagnostic}); cleanup unknown; retain ${recovery}`);
	}
	const receipt = JSON.parse(readFileSync(join(recovery, "result.json"), "utf8"));
	if (receipt.cleanup !== "confirmed") throw new Error(`security cleanup unknown; retain ${recovery}`);
	const output = readFileSync(join(recovery, "stdout"), "utf8");
	if (receipt.status !== 0 || receipt.timedOut || receipt.interrupted) {
		const failed = new Set();
		for (const line of output.split("\n")) {
			try { const event = JSON.parse(line); if (event.Action === "fail" && event.Test) failed.add(event.Test); } catch {}
		}
		throw new Error(`security test failed; cleanup confirmed; cases: ${[...failed].join(", ") || "package or setup failure"}; retain ${recovery}`);
	}
	return `${output}\n${readFileSync(join(recovery, "stderr"), "utf8")}`;
}

export function runGate() {
	const home = mkdtempSync("/tmp/opencode/auth-web-api-command-");
	let succeeded = false;
	try {
		for (const [owner, names] of Object.entries(required)) {
			const target = modules[owner];
			const cwd = join(root, target.directory);
			if (!existsSync(join(cwd, "go.mod"))) throw new Error(`${owner}: module unavailable`);
			const integration = target.integration;
			const flags = integration ? ["-tags=integration"] : [];
			const selector = `^(${names.join("|")})$`;
			const listArgs = ["test", "-mod=readonly", ...flags, target.package, "-list", selector];
			const listed = integration
				? supervisedGo("go", listArgs, cwd, 1200, home)
				: run("go", listArgs, cwd, env(home));
			const selected = listed.split("\n").filter((line) => /^Test\w+$/.test(line));
			if (selected.length !== names.length || names.some((name) => !selected.includes(name))) throw new Error(`${owner}: required selection mismatch`);
			console.log(`${owner}: selected ${selected.join(", ")}`);
			const args = ["test", "-mod=readonly", ...flags, target.package, "-run", selector, "-count=1", `-timeout=${integration ? 1200 : 120}s`, "-json"];
			const output = integration
				? supervisedGo("go", args, cwd, 1200, home)
				: run("go", args, cwd, env(home), 123000);
			console.log(`${owner}: PASS ${countGo(output, names)} required tests; zero skips`);
		}
		succeeded = true;
	} finally {
		if (succeeded) rmSync(home, { recursive: true, force: true });
		else console.error(`Security check failed; private HOME retained: ${home}`);
	}
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
	try { runGate(); } catch (error) { console.error(error instanceof Error ? error.message : "Web API security gate failed"); process.exitCode = 1; }
}
