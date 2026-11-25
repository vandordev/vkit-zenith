import { readFileSync, readdirSync, existsSync } from "node:fs";
import { createServer } from "node:net";
import { spawnSync } from "node:child_process";

// Read-only process/container inspection and exclusive loopback rebinds. This
// verifier never signals a PID or removes a resource based on receipt contents.
const path = process.argv[2];
if (!path?.startsWith("/tmp/opencode/")) throw new Error("owned receipt required");
const events = readFileSync(path, "utf8").trim().split("\n").map(JSON.parse);
if (events.some((event) => ["fail", "skip"].includes(event.Action))) throw new Error("fixture failure/skip");
const text = events.map((event) => event.Output ?? "").join("");
const groups = new Set([...text.matchAll(/(?:owned-group|group)=(\d+)/g)].map((match) => Number(match[1])));
const containers = new Set([...text.matchAll(/container=([0-9a-f]{64}) removed=true/g)].map((match) => match[1]));
const portRecords = [...text.matchAll(/(?:port=|listener released=127\.0\.0\.1:)(\d+)/g)].map((match) => Number(match[1]));
const ports = new Set(portRecords);
const roots = [...text.matchAll(/owned browser root removed=(\/tmp\/opencode\/(?:auth-recovery-[^/\s]+\/r\/)?auth-browser-[^\s]+)/g)].map(
	(match) => match[1],
);
if (groups.size !== 8 || containers.size !== 4 || portRecords.length !== 29 || roots.length !== 4)
	throw new Error("incomplete fixture cleanup manifest");
for (const entry of readdirSync("/proc")) {
	if (!/^\d+$/.test(entry)) continue;
	let stat;
	try {
		stat = readFileSync(`/proc/${entry}/stat`, "utf8");
	} catch (error) {
		if (["ENOENT", "ESRCH"].includes(error.code)) continue;
		throw new Error("process inspection failed");
	}
	const fields = stat
		.slice(stat.lastIndexOf(")") + 1)
		.trim()
		.split(/\s+/);
	if (groups.has(Number(fields[2])) && !["Z", "X"].includes(fields[0])) throw new Error("owned process remains");
}
for (const id of containers) {
	const result = spawnSync("docker", ["container", "inspect", "--format", "{{.Id}}", id], {
		encoding: "utf8",
		timeout: 3000,
		env: { PATH: process.env.PATH, HOME: "/tmp/opencode" },
	});
	if (result.status !== 1 || !result.stderr.includes("No such container")) throw new Error("container removal not confirmed");
}
for (const port of ports) {
	await new Promise((resolve, reject) => {
		const server = createServer();
		server.on("error", reject);
		server.listen(port, "127.0.0.1", () => server.close((error) => (error ? reject(error) : resolve())));
	});
}
for (const root of roots) if (existsSync(root)) throw new Error("owned browser root remains");
console.log(
	`Independent cleanup PASS: ${groups.size} groups, ${containers.size} containers, ${ports.size} listeners, ${roots.length} browser roots; no live owned resources`,
);
