import { expect, test } from "bun:test";
import { safeReturnTo } from "./redirect";
test("safe BO return destinations reject external, encoded and login-loop paths", () => {
	for (const value of [
		undefined,
		"https://foreign.invalid",
		"//foreign.invalid",
		"/\\foreign.invalid",
		"/%5cforeign.invalid",
		"/%2fforeign.invalid",
		"javascript:alert(1)",
		"data:text/html,hi",
		"/login",
		"/LOGIN?returnTo=/",
		"/%6cogin",
		"/\nforeign",
		"/%00x",
	])
		expect(safeReturnTo(value)).toBe("/");
	expect(safeReturnTo("/organizations?q=zenith&page=2#members")).toBe("/organizations?q=zenith&page=2#members");
});
