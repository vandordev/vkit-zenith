import { expect, test } from "bun:test";
import { loadRuntimeConfig } from "./runtime";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { configFilePath } from "./paths";
test("documentation credentials are enabled-only and service isolated", () => {
	expect(() => loadRuntimeConfig("api", { environment: { PUBLIC_API_DOCS_ENABLED: "true", PUBLIC_API_DOCS_USERNAME: "fixture" } })).toThrow();
});
test("disabled credentials do not resolve required environment templates", () => {
	const filePath = join(mkdtempSync(join(tmpdir(), "zenith-config-")), "config.yaml");
	writeFileSync(filePath, readFileSync(configFilePath(), "utf8").replaceAll("${PUBLIC_API_DOCS_USERNAME:-}", "${UNUSED_USERNAME}").replaceAll("${PUBLIC_API_DOCS_PASSWORD:-}", "${UNUSED_PASSWORD}"));
	expect(loadRuntimeConfig("api", { filePath, environment: {} }).api.docs.basicAuth).toEqual({ enabled: false, username: "", password: "" });
	expect(() => loadRuntimeConfig("api", { filePath, environment: { PUBLIC_API_DOCS_ENABLED: "true" } })).toThrow();
});
