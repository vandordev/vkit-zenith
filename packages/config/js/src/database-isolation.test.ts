import { expect, test } from "bun:test";
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import { loadPublicBOConfig, loadRuntimeConfig } from "./runtime";
import { validateConfigDocument } from "./document";

test("DB grammar is recognized but existing selectors never select DB secrets", () => {
	const root = mkdtempSync("/tmp/opencode/js-config-db-");
	try {
		mkdirSync(join(root, "config"));
		let source = readFileSync(new URL("../../../../config/config.yaml", import.meta.url), "utf8");
		if (!source.includes("db:\n"))
			source +=
				'\ndb:\n  dsn: "${DB_DSN}"\n  maxOpenConnections: 10\n  connectionMaxLifetimeSeconds: 1800\n  connectionMaxIdleTimeSeconds: 300\n  connectTimeoutSeconds: 10\n';
		const filePath = join(root, "config/config.yaml");
		writeFileSync(filePath, source);
		const options = { filePath, environment: {} };
		expect(loadRuntimeConfig("api", options).api.port).toBe(4102);
		expect(loadRuntimeConfig("web", options).web.port).toBe(4100);
		expect(loadRuntimeConfig("bo", options).bo.port).toBe(4103);
		expect(loadPublicBOConfig(options)).toEqual({ boApiBaseUrl: "http://localhost:4104" });
		const raw = Bun.YAML.parse(source);
		expect(() => validateConfigDocument(raw)).not.toThrow();
		for (const replacement of [
			'dsn: "${DB_DSN}"\n  maxIdleConnections: 5',
			'dsn: "${DB_DSN}"\n  unknown: sentinel-secret',
			'dsn: "${DB_DSN}"\n  dsn: sentinel-secret',
			'dsn: "${broken"',
		]) {
			writeFileSync(filePath, source.replace('dsn: "${DB_DSN}"', replacement));
			try {
				loadPublicBOConfig(options);
				throw new Error("invalid DB grammar accepted");
			} catch (error) {
				expect(String(error)).not.toContain("sentinel-secret");
				expect(String(error)).not.toContain("invalid DB grammar accepted");
			}
		}
	} finally {
		rmSync(root, { recursive: true, force: true });
	}
});
