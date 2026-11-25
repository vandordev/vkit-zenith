import { expect, test } from "bun:test";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import { loadRuntimeConfig, loadWebToolConfig, loadPublicWebConfig } from "./runtime";
import { validateConfigDocument } from "./document";
import { loadPublicBOConfig } from "./runtime";

test("Web selectors resolve only their exact fields, never Core secrets", () => {
	const root = mkdtempSync("/tmp/opencode/auth-config-js-");
	try {
		const filePath = join(root, "config.yaml");
		const source = readFileSync(new URL("../../../../config/config.yaml", import.meta.url), "utf8");
		writeFileSync(filePath, source, { mode: 0o600 });
		const options = { filePath, environment: {} };
		const raw = validateConfigDocument(Bun.YAML.parse(source));
		expect(raw.web.apiBaseUrl).toBe("${API_BASE_URL:-http://localhost:4101}");
		expect(loadRuntimeConfig("web", options).web.port).toBe(4100);
		expect(Object.keys(loadWebToolConfig(options).web).sort()).toEqual(["host", "port"]);
		expect(loadPublicWebConfig(options)).toEqual({ apiBaseUrl: "http://localhost:4101" });
		expect(Object.keys(loadPublicWebConfig(options))).toEqual(["apiBaseUrl"]);
		for (const runtime of ["api", "web", "bo"] as const) expect(() => loadRuntimeConfig(runtime, options)).not.toThrow();
		expect(Object.keys(loadPublicBOConfig(options))).toEqual(["boApiBaseUrl"]);
		writeFileSync(filePath, source.replace("${API_BASE_URL:-http://localhost:4101}", "${UNAVAILABLE_AUTH_API}"));
		expect(() => loadPublicWebConfig(options)).toThrow("missing environment variable UNAVAILABLE_AUTH_API");
		expect(() => loadWebToolConfig(options)).not.toThrow();
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});

test("unselected identity grammar is still strict and secrets stay out of diagnostics", () => {
	const source = readFileSync(new URL("../../../../config/config.yaml", import.meta.url), "utf8");
	const raw = validateConfigDocument(Bun.YAML.parse(source));
	const malformed: unknown = {
		...raw,
		webApi: {
			...raw.webApi,
			identity: {
				...raw.webApi.identity,
				email: { ...raw.webApi.identity.email, unexpected: "fake-sensitive" },
			},
		},
	};
	expect(() => validateConfigDocument(malformed)).toThrow("unknown key or invalid document shape");
	raw.webApi.identity.email.resendKey = "${broken";
	expect(() => validateConfigDocument(raw)).toThrow("malformed template");
});

test("cached Web public getter and configured launcher work from fake root without private credentials", async () => {
	const root = mkdtempSync("/tmp/opencode/auth-getter-js-");
	try {
		mkdirSync(join(root, "config"), { mode: 0o700 });
		writeFileSync(join(root, "config/config.yaml"), readFileSync(new URL("../../../../config/config.yaml", import.meta.url), "utf8"), {
			mode: 0o600,
		});
		const server = new URL("./server.ts", import.meta.url).pathname;
		const run = new URL("./run.ts", import.meta.url).pathname;
		const code = `import {setApplicationRoot,getPublicWebConfig} from ${JSON.stringify(server)};
import {runConfiguredCommand} from ${JSON.stringify(run)};
setApplicationRoot(${JSON.stringify(root)});
const result=await runConfiguredCommand({runtime:"web",root:${JSON.stringify(root)},environment:{PATH:process.env.PATH},command:[process.execPath,"-e",'console.log(JSON.stringify({port:process.env.PORT}))']});
console.log(JSON.stringify({keys:Object.keys(getPublicWebConfig()),child:JSON.parse(result.stdout),exit:result.exitCode}));`;
		const child = Bun.spawn([process.execPath, "--no-env-file", "-e", code], {
			cwd: root,
			env: { PATH: process.env.PATH },
			stdout: "pipe",
			stderr: "pipe",
			timeout: 10000,
		});
		const [stdout, stderr, exit] = await Promise.all([new Response(child.stdout).text(), new Response(child.stderr).text(), child.exited]);
		expect(exit).toBe(0);
		expect(stderr).toBe("");
		expect(JSON.parse(stdout)).toEqual({ keys: ["apiBaseUrl"], child: { port: "4100" }, exit: 0 });
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
