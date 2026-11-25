import { resolve, join } from "node:path";
import { readFileSync } from "node:fs";
import { parseEnv } from "node:util";
import { loadRuntimeConfig, loadWebToolConfig, loadBOToolConfig } from "./runtime";
import { configFilePath, setApplicationRoot } from "./paths";
import { webToolEnvironment, boToolEnvironment } from "./tooling";

export type RunConfiguredCommandOptions = {
	command: readonly string[];
	runtime: "web" | "bo";
	root?: string;
	tool?: "vite";
	development?: boolean;
	environment?: Record<string, string | undefined>;
	stdio?: "inherit" | "pipe";
};
export async function runConfiguredCommand({
	command,
	runtime,
	root,
	tool,
	development = false,
	environment = process.env,
	stdio = "pipe",
}: RunConfiguredCommandOptions) {
	if (!command.length) throw new Error("A command is required after --");
	if (root) setApplicationRoot(root);
	const applicationRoot = resolve(configFilePath(), "../..");
	let snapshot = { ...environment };
	if (development) {
		let source: string;
		try {
			source = readFileSync(join(applicationRoot, ".env"), "utf8");
		} catch {
			throw new Error("Cannot read explicitly requested root development env file");
		}
		snapshot = { ...parseEnv(source), ...snapshot };
	}
	const options = { filePath: configFilePath(), environment: snapshot };
	if (tool !== "vite") loadRuntimeConfig(runtime, options);
	const configured = runtime === "bo" ? loadBOToolConfig(options) : loadWebToolConfig(options);
	const listener = "bo" in configured ? configured.bo : configured.web;
	const emitted = "bo" in configured ? boToolEnvironment(configured) : webToolEnvironment(configured);
	const executable = command[0];
	if (executable !== "bun" && executable !== process.execPath) throw new Error("Configured commands require Bun");
	const args = [...command.slice(1)];
	if (tool === "vite") args.push("--host", listener.host, "--port", String(listener.port));
	const child = Bun.spawn({
		cmd: [process.execPath, "--no-env-file", ...args],
		env: { ...snapshot, ...emitted, ZENITH_APPLICATION_ROOT: applicationRoot },
		stdin: "inherit",
		stderr: stdio,
		stdout: stdio,
	});
	if (stdio === "inherit") return { exitCode: await child.exited, stderr: "", stdout: "" };
	const [stdout, stderr, exitCode] = await Promise.all([
		new Response(child.stdout).text(),
		new Response(child.stderr).text(),
		child.exited,
	]);
	return { exitCode, stderr, stdout };
}
if (import.meta.main) {
	const args = process.argv.slice(2);
	const separator = args.indexOf("--");
	if (separator < 0) throw new Error("Expected -- followed by a Bun command");
	const flags = args.slice(0, separator);
	const flag = (name: string) => {
		const index = flags.indexOf(name);
		return index < 0 ? undefined : flags[index + 1];
	};
	const runtime = flag("--runtime");
	if (runtime !== "web" && runtime !== "bo") throw new Error("Expected --runtime web or bo");
	const tool = flag("--tool");
	if (tool !== undefined && tool !== "vite") throw new Error("Unknown tooling adapter");
	process.exitCode = (
		await runConfiguredCommand({
			runtime,
			root: flag("--root"),
			tool,
			development: flags.includes("--development"),
			command: args.slice(separator + 1),
			stdio: "inherit",
		})
	).exitCode;
}
