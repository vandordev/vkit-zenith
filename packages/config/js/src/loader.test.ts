import { expect, test } from "bun:test";
import { loadConfigDocument } from "./loader";
import { configFilePath } from "./paths";
import { loadRuntimeConfig } from "./runtime";
test("parses before selected interpolation", () => {
	expect(loadConfigDocument(configFilePath()).web.port).toBe("${WEB_PORT:-4100}");
	expect(() => loadConfigDocument("/tmp/opencode/absent-config.yaml")).toThrow("cannot read document");
	const password = "x:#\n${SECOND}";
	expect(loadRuntimeConfig("api", { environment: { PUBLIC_API_DOCS_ENABLED: "true", PUBLIC_API_DOCS_USERNAME: "fixture", PUBLIC_API_DOCS_PASSWORD: password } }).api.docs.basicAuth.password).toBe(password);
});
