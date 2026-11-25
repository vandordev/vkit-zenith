import { test } from "node:test";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../", import.meta.url));

for (const mode of ["outer", "go-hard"]) {
	test(`${mode} timeout after real owned PostgreSQL acquisition removes PG, Ryuk and listener`, { timeout: 150000 }, () => {
		const recovery = mkdtempSync("/tmp/opencode/auth-supervisor-pg-");
		const env = {
			PATH: process.env.PATH ?? "/usr/local/go/bin:/usr/bin:/bin",
			HOME: recovery,
			TMPDIR: recovery,
			LANG: "C.UTF-8",
			GOENV: "off",
			GOTOOLCHAIN: "local",
			GOWORK: join(root, "go.work"),
			GOMODCACHE: process.env.GOMODCACHE ?? "/tmp/opencode/auth-go-modcache",
			GOCACHE: process.env.GOCACHE ?? "/tmp/opencode/auth-go-cache",
		};
		const result = spawnSync(
			"python3",
			[
				join(root, "scripts/auth-e2e/supervise.py"),
				recovery,
				"90000",
				...(mode === "outer" ? ["--timeout-after-ready"] : []),
				"go",
				"test",
				"-mod=readonly",
				"-tags=integration",
				"./authtest",
				"-run=^TestSupervisorPostgresChild$",
				"-count=1",
				mode === "outer" ? "-timeout=100s" : "-timeout=15s",
				"-json",
				"-args",
				"auth-supervisor-probe",
			],
			{
				cwd: join(root, "packages/tools"),
				env,
				encoding: "utf8",
				timeout: 140000,
			},
		);
		assert.equal(result.status, 0, `supervisor failed; preserve ${recovery}`);
		const report = JSON.parse(readFileSync(join(recovery, "result.json"), "utf8"));
		assert.equal(report.timedOut, mode === "outer", `failed to reach acquisition: ${recovery}`);
		assert.ok(existsSync(join(recovery, "ready")), `timeout occurred before acquisition: ${recovery}`);
		assert.notEqual(report.status, 0);
		assert.equal(report.cleanup, "confirmed", `cleanup unknown: ${recovery}`);
		const events = readFileSync(join(recovery, "ownership.jsonl"), "utf8").trim().split("\n").map(JSON.parse);
		assert.ok(
			events.some((e) => e.kind === "postgres-acquired"),
			"real PG must have been acquired before forced timeout",
		);
		assert.ok(events.some((e) => e.kind === "listener"));
		for (const event of events.filter((e) => ["supervised-child", "adopted-child"].includes(e.kind)))
			assert.equal(existsSync(`/proc/${event.pid}`), false, "owned child remains");
		const sessions = new Set(events.filter((e) => e.kind === "postgres-session").map((e) => e.id));
		assert.equal(sessions.size, 1);
		for (const session of sessions) {
			const inspected = spawnSync("docker", ["ps", "-aq", "--filter", `label=org.testcontainers.sessionId=${session}`], {
				env,
				encoding: "utf8",
				timeout: 5000,
			});
			assert.equal(inspected.status, 0);
			assert.equal(inspected.stdout.trim(), "", "owned PostgreSQL or Ryuk remains");
		}
		for (const e of events.filter((e) => e.kind === "listener")) {
			const rebound = spawnSync("python3", ["-c", `import socket; s=socket.socket(); s.bind(('127.0.0.1',${Number(e.id)})); s.close()`], {
				env,
				timeout: 3000,
			});
			assert.equal(rebound.status, 0);
		}
		console.log(
			`Independent forced-timeout cleanup: PG/Ryuk session absent, owned processes absent, listener rebound; failed child preserved at ${recovery}`,
		);
	});
}
