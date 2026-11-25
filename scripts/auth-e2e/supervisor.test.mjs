import { test } from "node:test";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, existsSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { randomBytes } from "node:crypto";

const supervisor = fileURLToPath(new URL("./supervise.py", import.meta.url));

// The child exits without running finally/defers. Its detached grandchild holds
// a real socket and ignores TERM. Only the supervisor can reap this orphan.
for (const mode of ["timeout", "error", "broken-journal"]) {
	test(`outer ${mode} preserves evidence and removes an acquired detached listener`, () => {
		const recovery = mkdtempSync("/tmp/opencode/auth-supervisor-test-");
		const code = `
import os, socket, signal, time, json
pid = os.fork()
if pid == 0:
 os.setsid()
 signal.signal(signal.SIGTERM, signal.SIG_IGN)
 s = socket.socket(); s.bind(('127.0.0.1', 0)); s.listen()
 with open('acquired.json', 'w') as f:
  json.dump({'pid':os.getpid(), 'port':s.getsockname()[1]}, f); f.flush(); os.fsync(f.fileno())
 while True: time.sleep(.05)
while not os.path.exists('acquired.json'): time.sleep(.01)
${mode === "broken-journal" ? "with open(os.environ['AUTH_OWNERSHIP_MANIFEST'],'a') as f: f.write('{incomplete\\n'); f.flush(); os.fsync(f.fileno())\nos._exit(23)" : mode === "error" ? "os._exit(23)" : "while True: time.sleep(.05)"}
`;
		const result = spawnSync("python3", [supervisor, recovery, "2000", "python3", "-c", code], {
			cwd: recovery,
			env: { PATH: "/usr/bin:/bin", HOME: recovery },
			encoding: "utf8",
			timeout: 15000,
		});
		assert.equal(result.status, 0, "supervisor itself must finish, without printing child output");
		assert.equal(result.stdout, "");
		const report = JSON.parse(readFileSync(join(recovery, "result.json"), "utf8"));
		assert.equal(report.cleanup, mode === "broken-journal" ? "unknown" : "confirmed");
		assert.notEqual(report.status, 0, "cleanup must not turn a failed child into success");
		assert.equal(report.timedOut, mode === "timeout");
		const acquired = JSON.parse(readFileSync(join(recovery, "acquired.json"), "utf8"));
		assert.equal(existsSync(`/proc/${acquired.pid}`), false);
		const rebound = spawnSync("python3", ["-c", `import socket; s=socket.socket(); s.bind(('127.0.0.1',${acquired.port})); s.close()`], {
			env: { PATH: "/usr/bin:/bin", HOME: recovery },
			timeout: 2000,
		});
		assert.equal(rebound.status, 0);
		assert.ok(existsSync(join(recovery, "ownership.jsonl")));
		assert.ok(existsSync(join(recovery, "stdout")));
		assert.ok(existsSync(join(recovery, "stderr")));
	});
}

test("an interrupted PG acquisition stays cleanup-unknown; receipt PIDs confer no signal authority", () => {
	const recovery = mkdtempSync("/tmp/opencode/auth-supervisor-unknown-");
	const session = randomBytes(32).toString("hex");
	const code = `import os,json
with open(os.environ['AUTH_OWNERSHIP_MANIFEST'],'a') as f:
 for e in [{'kind':'postgres-session','id':'${session}'},{'kind':'process','id':'${process.pid}'}]:
  f.write(json.dumps(e)+'\\n')
 f.flush(); os.fsync(f.fileno())
os._exit(23)`;
	const result = spawnSync("python3", [supervisor, recovery, "2000", "python3", "-c", code], {
		cwd: recovery,
		env: { PATH: "/usr/bin:/bin", HOME: recovery },
		encoding: "utf8",
		timeout: 10000,
	});
	assert.equal(result.status, 0);
	const report = JSON.parse(readFileSync(join(recovery, "result.json"), "utf8"));
	assert.equal(report.status, 23);
	assert.equal(report.cleanup, "unknown");
	assert.ok(report.issues.includes("in-flight PostgreSQL acquisition"));
	assert.ok(existsSync(`/proc/${process.pid}`));
	assert.ok(existsSync(join(recovery, "ownership.jsonl")));
});
