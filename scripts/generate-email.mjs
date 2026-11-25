import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

// Offline TSX compilation/export only. A clean cwd and constructed environment
// prevent Bun's checkout-root dotenv discovery; no shell/lifecycle script runs.
const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const home = mkdtempSync("/tmp/opencode/email-export-");
try {
	const result = spawnSync("bun", ["--no-env-file", "run", join(root, "packages/ui-email/scripts/export.ts")], {
		cwd: home,
		env: { PATH: process.env.PATH ?? "/usr/local/bin:/usr/bin:/bin", HOME: home, LANG: "C.UTF-8" },
		timeout: 60000,
		encoding: "utf8",
		maxBuffer: 1024 * 1024,
	});
	if (result.error || result.status !== 0) throw new Error("offline email export failed; no runtime send attempted");
	console.log("Three email families exported offline; no runtime secrets or provider calls");
} finally {
	rmSync(home, { recursive: true, force: true });
}
