import { expect, test } from "bun:test";
import { publicBOConfigSchema } from "./public";
test("strict public schema excludes server fields and credentialed URLs", () => {
	expect(publicBOConfigSchema.parse({ boApiBaseUrl: "http://localhost:4104" })).toEqual({ boApiBaseUrl: "http://localhost:4104" });
	expect(publicBOConfigSchema.safeParse({ boApiBaseUrl: "http://localhost:4104", password: "forbidden" }).success).toBe(false);
	expect(publicBOConfigSchema.safeParse({ boApiBaseUrl: "http://user:pass@localhost" }).success).toBe(false);
});
