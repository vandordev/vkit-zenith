import { expect, test } from "bun:test";
import { FormProblem } from "../../../-components/form-input";
import { apiKeyProblems, createApiKeyBody, endOfDay, toRfc3339Offset } from "./api-key-input";

const now = new Date(2026, 10, 19, 12, 0, 0);

test("expiry is an RFC 3339 offset instant for the end of the chosen local day", () => {
	const text = toRfc3339Offset(endOfDay("2026-12-31")!);
	expect(text).toMatch(/^2026-12-31T23:59:59[+-]\d{2}:\d{2}$/);
	// The offset is faithful: parsing it gives back the same instant.
	expect(new Date(text).getTime()).toBe(endOfDay("2026-12-31")!.getTime());
	expect(createApiKeyBody({ name: " Integrasi ", date: "2026-12-31" }, now)).toEqual({ name: "Integrasi", expiresAt: text });
});

test("no expiry sends no expiresAt; past, impossible and malformed dates are refused", () => {
	expect(createApiKeyBody({ name: "Kunci", date: "" }, now)).toEqual({ name: "Kunci" });
	for (const date of ["2026-11-18", "2026-02-30", "31-12-2026", "2026-13-01"])
		expect(apiKeyProblems({ name: "k", date }, now).date).toBeDefined();
	expect(() => createApiKeyBody({ name: "k", date: "2026-11-01" }, now)).toThrow(FormProblem);
});

test("the name is required, trimmed and at most 120 characters", () => {
	expect(apiKeyProblems({ name: "  ", date: "" }, now).name).toBeDefined();
	expect(apiKeyProblems({ name: "x".repeat(121), date: "" }, now).name).toBeDefined();
	expect(apiKeyProblems({ name: "x".repeat(120), date: "" }, now)).toEqual({});
});
