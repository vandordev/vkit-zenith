import { expect, test } from "bun:test";
import { loadRuntimeConfig } from "./runtime";
test("isolates consumer secrets and validates enabled docs credentials", () => {
	expect(loadRuntimeConfig("api", { environment: {} }).api.docs.basicAuth.enabled).toBe(false);
	expect(() => loadRuntimeConfig("api", { environment: { PUBLIC_API_DOCS_ENABLED: "true" } })).toThrow("credentials");
	for (const WEB_PORT of [" ", "0", "65536", "1.5"]) expect(() => loadRuntimeConfig("web", { environment: { WEB_PORT } })).toThrow();
});
test("safe errors omit submitted values", () => {
	const secret = "sensitive-value-not-to-log";
	try { loadRuntimeConfig("api", { environment: { PUBLIC_API_PORT: secret } }); throw new Error("expected failure"); }
	catch (error) { expect(error).toBeInstanceOf(Error); if (error instanceof Error) { expect(error.message).toContain("api.port"); expect(error.message).not.toContain(secret); } }
});
