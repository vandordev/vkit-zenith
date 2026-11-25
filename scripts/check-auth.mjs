import { spawnSync } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

// Pure counters are exported for zero-case/skip/failure regression tests.
export function goResult(output, selected, required = {}) {
	const passed = new Set();
	const passedSubtests = new Set();
	const seen = new Set();
	let subcases = 0;
	for (const line of output.split("\n").filter(Boolean)) {
		let event;
		try {
			event = JSON.parse(line);
		} catch {
			throw new Error("invalid Go event; raw output withheld");
		}
		if (["fail", "skip"].includes(event.Action)) throw new Error("required Go case failed or skipped");
		if (event.Action === "pass" && event.Test) {
			if (selected.includes(event.Test) || required.subtests?.includes(event.Test)) {
				if (seen.has(event.Test)) throw new Error("duplicate Go case");
				seen.add(event.Test);
			}
			if (event.Test.includes("/")) {
				subcases++;
				passedSubtests.add(event.Test);
			}
			else passed.add(event.Test);
		}
	}
	if (!selected.length || passed.size !== selected.length || selected.some((name) => !passed.has(name)))
		throw new Error("missing or zero Go cases");
	if (required.tests?.some((name) => !selected.includes(name) || !passed.has(name)))
		throw new Error("missing required Go test");
	if (required.subtests?.some((name) => !passedSubtests.has(name)))
		throw new Error("missing required Go subtest");
	return { tests: passed.size, subcases };
}

export function bunResult(output) {
	const passes = output.match(/(?:^|\n)\s*(\d+) pass\b/);
	if (!passes || Number(passes[1]) === 0 || !/(?:^|\n)\s*0 fail\b/.test(output) || /(?:^|\n)\s*[1-9]\d* (?:skip|todo)\b/.test(output))
		throw new Error("missing, skipped or failing Bun cases");
	return Number(passes[1]);
}

function environment(home) {
	return {
		PATH: process.env.PATH ?? "/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin",
		HOME: home,
		TMPDIR: home,
		LANG: "C.UTF-8",
		NO_COLOR: "1",
		GOTOOLCHAIN: "local",
		GOWORK: join(root, "go.work"),
		GOENV: "off",
		GOMODCACHE: process.env.GOMODCACHE ?? join(tmpdir(), "opencode/auth-go-modcache"),
		GOCACHE: process.env.GOCACHE ?? join(tmpdir(), "opencode/auth-go-cache"),
	};
}

export function runGate(name) {
	const home = mkdtempSync("/tmp/opencode/auth-command-");
	const env = environment(home);
	let succeeded = false;
	let invocation = 0;
	const invoke = (command, args, cwd, timeout = 180000) => {
		if (command === "go" && args.includes("-json")) {
			// Durable outside HOME, before child acquisition. Even a nonzero/timeout
			// preserves stdout/stderr and fsynced ownership, without printing either.
			const recovery = mkdtempSync("/tmp/opencode/auth-recovery-");
			const result = spawnSync("python3", [join(root, "scripts/auth-e2e/supervise.py"), recovery, String(timeout), command, ...args], {
				cwd,
				env,
				encoding: "utf8",
				timeout: timeout + 60000,
				maxBuffer: 1024 * 1024,
			});
			console.log(`Auth recovery evidence: ${recovery}`);
			if (result.error || result.status !== 0 || !existsSync(join(recovery, "result.json")))
				throw new Error(`Auth supervisor failed; cleanup unknown; retain ${home} and ${recovery}`);
			const report = JSON.parse(readFileSync(join(recovery, "result.json"), "utf8"));
			if (report.cleanup !== "confirmed") throw new Error(`Auth cleanup unknown; retain ${home} and ${recovery}`);
			if (report.status !== 0 || report.timedOut || report.interrupted)
				throw new Error(`go focused check failed; owned cleanup confirmed; raw output withheld in ${recovery}`);
			return `${readFileSync(join(recovery, "stdout"), "utf8")}\n${readFileSync(join(recovery, "stderr"), "utf8")}`;
		}
		const result = spawnSync(command, args, { cwd, env, encoding: "utf8", timeout, maxBuffer: 16 * 1024 * 1024 });
		const receipt = join(home, `command-${++invocation}`);
		writeFileSync(`${receipt}.stdout`, result.stdout ?? "", { mode: 0o600 });
		writeFileSync(`${receipt}.stderr`, result.stderr ?? "", { mode: 0o600 });
		if (result.error || result.status !== 0) throw new Error(`${command} focused check failed (raw output withheld)`);
		return `${result.stdout}\n${result.stderr}`;
	};
	const go = (owner, pkg, selector, integration = false, seconds = 120, required) => {
		const cwd = join(root, owner);
		if (!existsSync(join(cwd, "go.mod"))) throw new Error(`${name}: unavailable owner`);
		const flags = integration ? ["-tags=integration"] : [];
		const listed = invoke("go", ["test", "-mod=readonly", ...flags, pkg, "-list", selector], cwd);
		const selected = listed.split("\n").filter((line) => /^Test\w+$/.test(line));
		if (!selected.length) throw new Error(`${name}: unavailable (zero selected cases)`);
		const output = invoke(
			"go",
			["test", "-mod=readonly", ...flags, pkg, "-run", selector, "-count=1", `-timeout=${seconds}s`, "-json"],
			cwd,
			(seconds + 30) * 1000,
		);
		const count = goResult(output, selected, required);
		console.log(`${owner}/${pkg}: PASS ${count.tests} tests / ${count.subcases} subcases; zero skips`);
		return output;
	};
	const bun = (files, expected) => {
		const paths = files.map((file) => join(root, file));
		if (paths.some((path) => !existsSync(path))) throw new Error(`${name}: unavailable source`);
		const count = bunResult(invoke("bun", ["--no-env-file", "test", ...paths], home));
		if (expected !== undefined && count !== expected) throw new Error(`${name}: unexpected Bun case count`);
		console.log(`Bun focused cases: PASS ${count} tests; zero skips`);
	};
	const types = (config) =>
		invoke(process.execPath, [join(root, "node_modules/typescript/bin/tsc"), "--noEmit", "-p", join(root, config)], home);
	try {
		switch (name) {
			case "config":
				go("packages/config/go", ".", "^(TestAuthConfig|TestDatabase)");
				bun(
					["auth-isolation", "database-isolation", "runtime", "public", "tooling", "run"].map(
						(name) => `packages/config/js/src/${name}.test.ts`,
					),
				);
				types("packages/config/js/tsconfig.json");
				break;
			case "contracts":
				go("packages/application", "./identity", "^TestAuthPolicy|^TestAuthDelivery|^TestAuthNormalize");
				bun(["scripts/check-architecture.test.ts"], 7);
				const authCounterTests = invoke(process.execPath, ["--test", "--test-reporter=tap", join(root, "scripts/check-auth.test.mjs")], home, 30000);
				if (!/^# tests 7$/m.test(authCounterTests) || !/^# pass 7$/m.test(authCounterTests) || !/^# skipped 0$/m.test(authCounterTests))
					throw new Error("Auth counter regression tests missing, skipped or failing");
				console.log("Auth counter regressions: PASS 7 tests; zero skips");
				invoke("bun", ["--no-env-file", "run", join(root, "scripts/check-architecture.ts"), root, "--auth"], home);
				console.log(
					"Actual Auth-owned architecture and all-runtime test-import/identity-secret-log guards: PASS (not the unrelated BO presentation baseline gate)",
				);
				console.log("G-LICENSE conditional: reviewed candidates only; actual consumer delta review still required");
				break;
			case "unit":
				go(
					"packages/application",
					"./identity",
					"^TestAuthPolicy|^TestAuthDelivery|^TestAuthNormalize|^TestAuthIdentityPassword|^TestAuthAnonymousCSRF|^TestAuthSessionCSRF|^TestAuthActionRedaction|^TestAuthKDFBounded",
				);
				go("apps/web-api", "./internal/identityhttp", "^TestAuth(Delivery|Source)");
				// Email is a required part of this gate, not silently omitted while absent.
				go("packages/email", "./...", "^TestAuthEmail");
				go("packages/tools", "./authtest", "^TestAuthEmail(CaptureBoundary|DeliveryPostCommit)$", true, 240);
				invoke(process.execPath, [join(root, "scripts/generate-email.mjs")], home, 90000);
				bun(["packages/ui-email/src/export.test.ts", "packages/ui-email/src/dependency-security.test.ts"]);
				types("packages/ui-email/tsconfig.json");
				break;
			case "schema":
				go("packages/tools", "./authtest", "^TestAuthSchema|^TestAuthAccount", true, 1200);
				go("packages/tools", "./dbtool", "^TestSchema(Manifest|StaffAuth|Identity)$", true, 1200);
				go(
					"apps/bo-api",
					"./internal/securitytest",
					"^TestAuthSchemaBO|^TestBOCustomer(CreationUnverifiedWithoutCredentials|RevocationDisableAndEmail|RevocationAuditFailureRollsBackEverything)$",
					true,
					1200,
				);
				break;
			case "accounts":
				go("packages/tools", "./authtest", "^TestAuthAccount", true, 1200);
				go("apps/web-api", "./internal/securitytest", "^TestAuthAccount", true, 1200);
				go("apps/bo-api", "./internal/securitytest", "^TestBOCustomerRevocationCoreResetCompatibility", true, 1200);
				break;
			case "sessions":
				go("packages/tools", "./authtest", "^TestAuthFullSessionForegroundActivityHasNoAbsoluteCap", true, 1200);
				go("packages/tools", "./authtest", "^TestAuthAccountVerifyRotatesRestrictedSession", true, 1200);
				go("apps/web-api", "./internal/handlers", "^TestAuthCookie", true);
				go("apps/web-api", "./internal/securitytest", "^TestAuthAccountHTTPRegistrationPrivacyAndCapture", true, 1200);
				go("apps/bo-api", "./internal/securitytest", "^TestBOCustomerRevocationCoreResetCompatibility", true, 1200);
				break;
			case "adapters":
				bun([
					"apps/web/src/lib/api/client.test.ts",
					"apps/web/src/lib/session/token.test.ts",
					"apps/web/src/lib/session/single-flight.test.ts",
					"apps/web/src/lib/session/flows.test.ts",
				]);
				types("apps/web/tsconfig.json");
				break;
			default:
				throw new Error("unknown Auth gate");
		}
		succeeded = true;
	} finally {
		if (succeeded) rmSync(home, { recursive: true, force: true });
		else console.error(`Auth check failed; private HOME retained: ${home}`);
	}
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
	try {
		runGate(process.argv[2]);
	} catch (error) {
		console.error(error instanceof Error ? error.message : "Auth gate failed");
		process.exitCode = 1;
	}
}
