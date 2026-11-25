import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const required = {
	"packages/secrets": ["TestKeyringValidation", "TestSealOpenAndRotation", "TestOpenRejectsTamperingAndInvalidFormats", "TestOpenErrorIsOpaque", "TestRecordAADLengthPrefixesFields"],
	"packages/storage": ["TestNewValidatesSettings", "TestKeyBuilders", "TestPresignOffline"],
	"packages/config/go": ["TestInfrastructureSelectorIsolation", "TestInfrastructureBoundsAndSafeDiagnostics", "TestStorageLoopbackRequiresDevelopmentOrTest", "TestEncryptionKeySequenceValidation"],
};

export function countGo(output, expected) {
	const passed = new Set();
	for (const line of output.split("\n").filter(Boolean)) {
		let event;
		try { event = JSON.parse(line); } catch { throw new Error("invalid Go test event"); }
		if (event.Action === "fail" || event.Action === "skip") throw new Error("Go test failed or skipped");
		if (event.Action === "pass" && event.Test && !event.Test.includes("/")) passed.add(event.Test);
	}
	if (!expected.length || passed.size !== expected.length || expected.some((name) => !passed.has(name)))
		throw new Error("required Go tests missing or zero");
	return passed.size;
}

export function countBun(output) {
	const required = [
		"rejects forbidden YAML features and unknown keys",
		"accepts infrastructure sections and rejects unknown infrastructure keys",
	];
	if (!/(?:^|\n)\s*2 pass\b/.test(output) || !/(?:^|\n)\s*0 fail\b/.test(output) || /(?:^|\n)\s*[1-9]\d* (?:skip|todo)\b/.test(output) || required.some((name) => !output.includes(name)))
		throw new Error("required JS document tests missing, skipped or failing");
	return 2;
}

function run(command, args, cwd, env, timeout = 180000) {
	const result = spawnSync(command, args, { cwd, env, encoding: "utf8", timeout, maxBuffer: 8 * 1024 * 1024 });
	if (result.error || result.status !== 0) throw new Error(`${command} gate failed (output withheld)`);
	return `${result.stdout}\n${result.stderr}`;
}

export function runGate() {
	const home = mkdtempSync(join(tmpdir(), "opencode/infra-command-"));
	const env = {
		PATH: process.env.PATH ?? "/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin",
		HOME: home, TMPDIR: home, LANG: "C.UTF-8", NO_COLOR: "1", GOTOOLCHAIN: "local",
		GOWORK: join(root, "go.work"), GOENV: "off",
		GOMODCACHE: process.env.GOMODCACHE ?? join(tmpdir(), "opencode/infra-go-modcache"),
		GOCACHE: process.env.GOCACHE ?? join(tmpdir(), "opencode/infra-go-cache"),
	};
	try {
		for (const [owner, names] of Object.entries(required)) {
			const cwd = join(root, owner);
			const listed = run("go", ["test", "-mod=readonly", ".", "-list", `^(${names.join("|")})$`], cwd, env);
			const selected = listed.split("\n").filter((line) => /^Test\w+$/.test(line));
			if (selected.length !== names.length || names.some((name) => !selected.includes(name))) throw new Error(`${owner}: required test selection mismatch`);
			const output = run("go", ["test", "-mod=readonly", ".", "-run", `^(${names.join("|")})$`, "-count=1", "-timeout=2m", "-json"], cwd, env, 150000);
			console.log(`${owner}: PASS ${countGo(output, names)} required tests; zero skips`);
		}
		const js = run("bun", ["--no-env-file", "test", join(root, "packages/config/js/src/document.test.ts")], home, env, 30000);
		console.log(`config JS document: PASS ${countBun(js)} required tests; zero skips`);
	} finally {
		rmSync(home, { recursive: true, force: true });
	}
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
	try { runGate(); } catch (error) {
		console.error(error instanceof Error ? error.message : "Infrastructure gate failed");
		process.exitCode = 1;
	}
}
